package v1

import (
	"grocerics-backend/internal/dto"

	"github.com/gin-gonic/gin"
)

func RegiterWishlistRoutes(r *gin.Engine) {
	group := r.Group("/v1/wishlist")
	group.GET("/", getWishlist())
}

// @Swagger:route GET /v1/wishlist wishlist getWishlist
// @Summary Get wishlist
// @Description Fetches a wishlist by its unique identifier.
// @Tags wishlist
// @Accept json
// @Produce json
// @Param wishlist_id path string true "Unique identifier for the wishlist"
// @Success 200 {object} dto.Response{data=dto.WishlistMobile}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/wishlist [get]
func getWishlist() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.WishlistMobile{},
			Message: "Wishlist fetched successfully",
			Status:  "success",
		})
	}
}

type AddToWishlistRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
}

// @Swagger:route POST /v1/wishlist wishlist upsertToWishlist
// @Summary Add or update item in wishlist
// @Description Adds an item to the user's wishlist or updates its quantity.
// @Tags wishlist
// @Accept json
// @Produce json
// @Param request body AddToWishlistRequest true "Request payload containing product ID and quantity"
// @Success 200 {object} dto.Response{data=dto.WishlistMobile}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/wishlist [post]
func upsertToWishlist() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddToWishlistRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, dto.Response{
				Data:    nil,
				Message: err.Error(),
				Status:  "error",
			})
			return
		}

		c.JSON(200, dto.Response{
			Data:    dto.WishlistMobile{},
			Message: "Item added to wishlist successfully",
			Status:  "success",
		})
	}
}

type RemoveFromWishlistRequest struct {
	ProductID string `json:"product_id" binding:"required"`
}

// @Swagger:route DELETE /v1/wishlist wishlist removeFromWishlist
// @Summary Remove item from cart
// @Description Removes an item from the user's cart.
// @Tags cart
// @Accept json
// @Produce json
// @Param request body RemoveFromCartRequest true "Request payload containing product ID"
// @Success 200 {object} dto.Response{data=dto.WishlistMobile}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/wishlist [delete]
func removeFromWishlist() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RemoveFromWishlistRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, dto.Response{
				Data:    nil,
				Message: err.Error(),
				Status:  "error",
			})
			return

		}
		c.JSON(200, dto.Response{
			Data:    dto.WishlistMobile{},
			Message: "Item removed from wishlist successfully",
			Status:  "success",
		})
	}
}
