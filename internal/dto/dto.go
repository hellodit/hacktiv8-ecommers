package dto

import "ecommers/internal/model"

type ErrorResponse struct {
	Message string `json:"message" example:"invalid credentials"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"user1@example.com"`
	Password string `json:"password" example:"password123"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at" example:"1767225600"`
}

type PaginationMeta struct {
	Page  int   `json:"page" example:"1"`
	Limit int   `json:"limit" example:"20"`
	Total int64 `json:"total" example:"100"`
}

type ProductListResponse struct {
	Data []model.Product `json:"data"`
	Meta PaginationMeta  `json:"meta"`
}

type CheckoutRequest struct {
	ProductID uint `json:"product_id" example:"1"`
	Quantity  int  `json:"quantity" example:"2"`
}
