package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterBrandsRoutes(jwt *auth.JWTService, users *repository.UserRepository, r *gin.Engine) {
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

type CreateBrandRequest struct {
	BrandName  string `json:"brand_name" binding:"required"`
	ImageURL   string `json:"image_url" binding:"required"`
	Status     string `json:"status" binding:"required,oneof=active disabled"`
	IsTopBrand bool   `json:"is_top_brand" binding:"required"`
}

// @Swagger:route POST /v1/brands brands createBrand
// @Summary Create a new brand
// @Description Creates a new brand. This endpoint is intended for internal use and should be secured appropriately.
// @Tags brands
// @Accept application/json
// @Produce json
// @Param brand body CreateBrandRequest true "Create Brand Request"
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

type UpdateBrandRequest struct {
	BrandID    string `json:"brand_id" binding:"required"`
	BrandName  string `json:"brand_name"`
	ImageURL   string `json:"image_url"`
	Status     string `json:"status" binding:"omitempty,oneof=active disabled"`
	IsTopBrand *bool  `json:"is_top_brand"`
}

// @Swagger:route PATCH /v1/brands brands updateBrand
// @Summary Update an existing brand
// @Description Updates an existing brand. This endpoint is intended for internal use and should be secured appropriately.
// @Tags brands
// @Accept application/json
// @Produce json
// @Param brand body UpdateBrandRequest true "Update Brand Request"
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
