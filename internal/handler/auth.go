package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ecommers/internal/dto"
	"ecommers/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB        *gorm.DB
	JWTSecret []byte
	JWTTTL    time.Duration
}

// dummyHash is compared against when the email doesn't exist, so a login for
// an unknown email takes as long as a wrong password (no user enumeration
// via response timing).
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

// Login godoc
// @Summary      Login
// @Description  Authenticate with email & password and receive a JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Credentials"
// @Success      200   {object}  dto.LoginResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}

	var user model.User
	err := h.DB.WithContext(c.Request().Context()).Where("email = ?", req.Email).First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash := user.Password
	if hash == "" {
		hash = string(dummyHash)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil || user.ID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	expiresAt := time.Now().Add(h.JWTTTL)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(user.ID), 10),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	})
	signed, err := token.SignedString(h.JWTSecret)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.LoginResponse{Token: signed, ExpiresAt: expiresAt.Unix()})
}
