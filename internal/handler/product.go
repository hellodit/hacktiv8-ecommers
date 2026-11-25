package handler

import (
	"errors"
	"net/http"
	"strconv"

	"ecommers/internal/dto"
	"ecommers/internal/model"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ProductHandler struct {
	DB *gorm.DB
}

// List godoc
// @Summary      List products
// @Description  Paginated list of products, ordered by ID.
// @Tags         products
// @Produce      json
// @Param        page   query     int  false  "Page number"     default(1)
// @Param        limit  query     int  false  "Items per page (max 100)"  default(20)
// @Success      200    {object}  dto.ProductListResponse
// @Router       /products [get]
func (h *ProductHandler) List(c echo.Context) error {
	page := queryInt(c, "page", 1, 1, 1_000_000)
	limit := queryInt(c, "limit", 20, 1, 100)
	db := h.DB.WithContext(c.Request().Context())

	var total int64
	if err := db.Model(&model.Product{}).Count(&total).Error; err != nil {
		return err
	}

	products := []model.Product{}
	if err := db.Order("id").Limit(limit).Offset((page - 1) * limit).Find(&products).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ProductListResponse{
		Data: products,
		Meta: dto.PaginationMeta{Page: page, Limit: limit, Total: total},
	})
}

// Get godoc
// @Summary      Get product detail
// @Tags         products
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  model.Product
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /products/{id} [get]
func (h *ProductHandler) Get(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid product id")
	}

	var product model.Product
	err = h.DB.WithContext(c.Request().Context()).First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "product not found")
	}
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, product)
}

// queryInt parses an int query param, falling back to def when missing or
// invalid, and clamping the result to [min, max].
func queryInt(c echo.Context, key string, def, min, max int) int {
	v, err := strconv.Atoi(c.QueryParam(key))
	if err != nil {
		return def
	}
	return clamp(v, min, max)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
