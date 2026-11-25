package database

import (
	"fmt"
	"log"

	"ecommers/internal/model"

	"github.com/brianvoe/gofakeit/v7"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	SeedUserCount    = 30
	SeedProductCount = 100
	SeedPassword     = "password123"
)

// Seed is idempotent: each table is only seeded when it is empty.
func Seed(db *gorm.DB) error {
	if err := seedUsers(db); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	if err := seedProducts(db); err != nil {
		return fmt.Errorf("seed products: %w", err)
	}
	return nil
}

func seedUsers(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(SeedPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := make([]model.User, 0, SeedUserCount)
	for i := 1; i <= SeedUserCount; i++ {
		users = append(users, model.User{
			Name:     gofakeit.Name(),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: string(hash),
		})
	}
	if err := db.CreateInBatches(users, 100).Error; err != nil {
		return err
	}
	log.Printf("seeded %d users (password: %s)", SeedUserCount, SeedPassword)
	return nil
}

func seedProducts(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.Product{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	products := make([]model.Product, 0, SeedProductCount)
	for i := 1; i <= SeedProductCount; i++ {
		products = append(products, model.Product{
			Name:        gofakeit.ProductName(),
			Description: gofakeit.ProductDescription(),
			ImageURL:    fmt.Sprintf("https://picsum.photos/seed/product-%d/600/400", i),
			Price:       int64(gofakeit.IntRange(10, 5000)) * 1000, // Rp10.000 – Rp5.000.000
			Stock:       1_000_000,                                 // large so load tests never run out
		})
	}
	if err := db.CreateInBatches(products, 100).Error; err != nil {
		return err
	}
	log.Printf("seeded %d products", SeedProductCount)
	return nil
}
