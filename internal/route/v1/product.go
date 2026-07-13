package v1

import (
	"grocerics-backend/internal/dto"

	"github.com/gin-gonic/gin"
)

func RegisterProductRoutes(r *gin.Engine) {
	group := r.Group("/v1/products")
	group.GET("/:product_id", getInventoryItem())
}

// @Swagger:route GET /v1/products/{product_id} products getInventoryItem
// @Summary Get inventory item
// @Description Fetches an inventory item by its unique identifier.
// @Tags products
// @Accept json
// @Produce json
// @Param product_id path string true "Unique identifier for the product"
// @Success 200 {object} dto.Response{data=dto.ProductItem}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/products/{product_id} [get]
func getInventoryItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.ProductItem{},
			Message: "Inventory item fetched successfully",
			Status:  "success",
		})
	}
}

// @Swagger:route GET /v1/products/search products searchProducts
// @Summary Search products
// @Description Searches for products based on a query string.
// @Tags products
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param platforms query []string false "Filter by delivery platforms"
// @Param page query int true "Page number"
// @Param limit query int true "Number of items per page"
// @Success 200 {object} dto.Response{data=dto.SearchResultMobile}
func searchProducts() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.SearchResultMobile{},
			Message: "Products fetched successfully",
			Status:  "success",
		})
	}
}
