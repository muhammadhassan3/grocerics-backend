package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterBrandsRoutes(jwt *auth.JWTService, users *repository.UserRepository, r *gin.RouterGroup) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(jwt, users))
	group.GET("/brands", getBrands())
	adminGroup := group.Group("")
	adminGroup.Use(middleware.RequireRole(domain.RoleAdmin))
	adminGroup.POST("/brands", CreateNewBrand())
	adminGroup.PATCH("/brands", UpdateBrand())
	adminGroup.DELETE("/brands", DeleteBrand())
}

// @Swagger:route GET /v1/brands brands getBrands
// @Summary Get brands
// @Description Fetches a paginated list of brands.
// @Tags brands
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param limit query int true "Number of items per page"
// @Success 200 {object} dto.Response{data=dto.BrandList}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/brands [get]
func getBrands() gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = c.Param("page")
		_ = c.Param("limit")
		c.JSON(200, dto.Response{
			Message: "Brands fetched successfully",
			Status:  "success",
			Data:    dto.BrandList{},
		})
	}
}

// @Swagger:route POST /v1/brands brands createBrand
// @Summary Create a new brand
// @Description Creates a new brand. This endpoint is intended for internal use and should be secured appropriately.
// @Tags brands
// @Accept multipart/form-data
// @Produce json
// @Param brand_name formData string true "Display name of the brand"
// @Param image formData file true "Image file for the brand"
// @Param status formData string true "Status of the brand" enums(active,disabled)
// @Param is_top_brand formData bool true "Whether the brand is flagged as a top/featured brand"
// @Success 201 {object} dto.Response{data=dto.BrandItem}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/brands [post]
func CreateNewBrand() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(201, dto.Response{
			Message: "Brand created successfully",
			Status:  "success",
			Data:    dto.BrandItem{},
		})
	}
}

// @Swagger:route PATCH /v1/brands brands updateBrand
// @Summary Update an existing brand
// @Description Updates an existing brand. This endpoint is intended for internal use and should be secured appropriately.
// @Tags brands
// @Accept multipart/form-data
// @Produce json
// @Param brand_id formData string true "Unique identifier for the brand"
// @Param brand_name formData string false "Display name of the brand"
// @Param image formData file false "Image file for the brand"
// @Param status formData string false "Status of the brand" enums(active,disabled)
// @Param is_top_brand formData bool false "Whether the brand is flagged as a top/featured brand"
// @Success 200 {object} dto.Response{data=dto.BrandItem}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/brands [patch]
func UpdateBrand() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Message: "Brand updated successfully",
			Status:  "success",
			Data:    dto.BrandItem{},
		})
	}
}

type DeleteBrandRequest struct {
	// Unique identifier for the brand
	BrandID string `json:"brand_id" binding:"required"`
}

// @Swagger:route DELETE /v1/brands brands deleteBrand
// @Summary Delete a brand
// @Description Deletes a brand. This endpoint is intended for internal use and should be secured appropriately.
// @Tags brands
// @Accept json
// @Produce json
// @Param DeleteBrandRequest body DeleteBrandRequest true "Unique identifier for the brand"
// @Success 200 {object} dto.Response{data=string}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/brands [delete]
func DeleteBrand() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeleteBrandRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, dto.Response{
				Message: "Invalid request payload",
				Status:  "error",
				Data:    err.Error(),
			})
			return
		}

		c.JSON(200, dto.Response{
			Message: "Brand deleted successfully",
			Status:  "success",
			Data:    nil,
		})
	}
}
