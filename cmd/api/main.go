package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "ecommers/docs"
	"ecommers/internal/config"
	"ecommers/internal/database"
	"ecommers/internal/handler"
	"ecommers/internal/middleware"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title                       Minimal Ecommerce API
// @version                     1.0
// @description                 Go + Echo + GORM + PostgreSQL + JWT.
// @BasePath                    /api/v1
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer " followed by the JWT from /auth/login.
func main() {
	_ = godotenv.Load() // optional: .env for local runs; Docker passes env directly

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db); err != nil {
		log.Fatalf("seed: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Recover())
	if os.Getenv("ACCESS_LOG") != "false" { // disable during heavy k6 runs
		e.Use(echomw.Logger())
	}

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	auth := &handler.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret, JWTTTL: cfg.JWTTTL}
	products := &handler.ProductHandler{DB: db}
	checkout := &handler.CheckoutHandler{DB: db}

	api := e.Group("/api/v1")
	api.POST("/auth/login", auth.Login)
	api.GET("/products", products.List)
	api.GET("/products/:id", products.Get)
	api.POST("/checkout", checkout.Checkout, middleware.JWT(cfg.JWTSecret))

	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
