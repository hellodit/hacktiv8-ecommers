package handler

import (
	"errors"
	"net/http"

	"ecommers/internal/dto"
	"ecommers/internal/middleware"
	"ecommers/internal/model"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

const maxQuantity = 100

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	errNotImplemented    = errors.New("reserveStock not implemented yet")
)

type CheckoutHandler struct {
	DB *gorm.DB
}

// Checkout godoc
// @Summary      Checkout
// @Description  Buy a product. Decrements stock and stores an order with a product snapshot.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CheckoutRequest  true  "Product & quantity"
// @Success      201   {object}  model.Order
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Failure      404   {object}  dto.ErrorResponse
// @Failure      409   {object}  dto.ErrorResponse
// @Router       /checkout [post]
func (h *CheckoutHandler) Checkout(c echo.Context) error {
	var req dto.CheckoutRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.ProductID == 0 || req.Quantity < 1 || req.Quantity > maxQuantity {
		return echo.NewHTTPError(http.StatusBadRequest, "product_id is required and quantity must be 1-100")
	}

	userID := c.Get(middleware.UserIDKey).(uint)
	var order model.Order

	err := h.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		product, err := reserveStock(tx, req.ProductID, req.Quantity)
		if err != nil {
			return err
		}

		order = model.Order{
			UserID: userID,
			ProductSnapshot: model.ProductSnapshot{
				ProductID:   product.ID,
				Name:        product.Name,
				Description: product.Description,
				ImageURL:    product.ImageURL,
				Price:       product.Price,
				Quantity:    req.Quantity,
				Total:       product.Price * int64(req.Quantity),
			},
		}
		return tx.Create(&order).Error
	})

	switch {
	case errors.Is(err, ErrProductNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInsufficientStock):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, errNotImplemented):
		return echo.NewHTTPError(http.StatusNotImplemented, err.Error())
	case err != nil:
		return err
	}
	return c.JSON(http.StatusCreated, order)
}

// reserveStock runs INSIDE the checkout transaction. It must:
//  1. load the product with a row lock (so concurrent checkouts on the same
//     product wait for each other instead of both reading the same stock),
//  2. return ErrProductNotFound if it doesn't exist,
//  3. return ErrInsufficientStock if product.Stock < qty,
//  4. decrement the stock and return the (pre-decrement-loaded) product.
//
// Hints: tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, id)
//
//	tx.Model(&p).Update("stock", gorm.Expr("stock - ?", qty))
func reserveStock(tx *gorm.DB, productID uint, qty int) (*model.Product, error) {
	// TODO(you): implement — see the comment above.
	return nil, errNotImplemented
}
