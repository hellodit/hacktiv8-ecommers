package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Order struct {
	ID              uint            `gorm:"primaryKey" json:"id"`
	UserID          uint            `gorm:"not null;index" json:"user_id"`
	User            User            `gorm:"constraint:OnDelete:RESTRICT" json:"-"`
	ProductSnapshot ProductSnapshot `gorm:"type:jsonb;not null" json:"product_snapshot"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ProductSnapshot freezes the product as it was at checkout time, so later
// edits to the product (price, name, ...) never rewrite order history.
type ProductSnapshot struct {
	ProductID   uint   `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Price       int64  `json:"price"`
	Quantity    int    `json:"quantity"`
	Total       int64  `json:"total"`
}

// Value / Scan let GORM store the snapshot as a JSONB column.
func (s ProductSnapshot) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *ProductSnapshot) Scan(src any) error {
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("product_snapshot: unsupported type")
	}
	return json.Unmarshal(b, s)
}
