package v1

import (
	"grocerics-backend/internal/dto"

	"github.com/gin-gonic/gin"
)

func RegiterCartRoutes(r *gin.Engine) {
	group := r.Group("/v1/cart")
	group.GET("/", getCart())
}

// @Swagger:route GET /v1/cart cart getCart
// @Summary Get cart
// @Description Fetches a cart by its unique identifier.
// @Tags cart
// @Accept json
// @Produce json
// @Param cart_id path string true "Unique identifier for the cart"
// @Success 200 {object} dto.Response{data=dto.CartMobile}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/cart [get]
func getCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.CartMobile{},
			Message: "Cart fetched successfully",
			Status:  "success",
		})
	}
}

type AddToCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
}

// @Swagger:route POST /v1/cart cart upsertToCart
// @Summary Add or update item in cart
// @Description Adds an item to the user's cart or updates its quantity.
// @Tags cart
// @Accept json
// @Produce json
// @Param request body AddToCartRequest true "Request payload containing product ID and quantity"
// @Success 200 {object} dto.Response{data=dto.CartMobile}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/cart [post]
func upsertToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddToCartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, dto.Response{
				Data:    nil,
				Message: err.Error(),
				Status:  "error",
			})
			return
		}

		c.JSON(200, dto.Response{
			Data:    dto.CartMobile{},
			Message: "Item added to cart successfully",
			Status:  "success",
		})
	}
}

type RemoveFromCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
}

// @Swagger:route DELETE /v1/cart cart removeFromCart
// @Summary Remove item from cart
// @Description Removes an item from the user's cart.
// @Tags cart
// @Accept json
// @Produce json
// @Param request body RemoveFromCartRequest true "Request payload containing product ID"
// @Success 200 {object} dto.Response{data=dto.CartMobile}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/cart [delete]
func removeFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RemoveFromCartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, dto.Response{
				Data:    nil,
				Message: err.Error(),
				Status:  "error",
			})
			return

		}
		c.JSON(200, dto.Response{
			Data:    dto.CartMobile{},
			Message: "Item removed from cart successfully",
			Status:  "success",
		})
	}
}
