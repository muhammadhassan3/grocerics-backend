package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterSubcategoryRoutes(jwt *auth.JWTService, user *repository.UserRepository, r *gin.Engine) {
	group := r.Group("/v1/categories")
	group.Use(middleware.AuthMiddleware(jwt, user))
	group.GET("/subcategories", getSubcategories())
	group.GET("/:id/subcategories/:subcategory_id", getSubcategoryByID())
	adminGroup := group.Group("")
	adminGroup.Use(middleware.RequireRole(domain.RoleAdmin))
	adminGroup.POST("/subcategories", CreateSubcategory())
	adminGroup.PATCH("/subcategories", UpdateSubcategory())
	adminGroup.DELETE("/subcategories", DeleteSubcategory())
}

// @Swagger:route GET /v1/categories/subcategories subcategories getSubcategories
// @Summary Get subcategories
// @Description Fetches a paginated list of subcategories.
// @Tags subcategories
// @Accept json
// @Produce json
// @Param search query string false "Search term to filter subcategories by name"
// @Param category_id query string false "Unique identifier for the category whose subcategories are to be fetched"
// @Param page query int true "Page number"
// @Param limit query int true "Number of items per page"
// @Success 200 {object} dto.Response{data=dto.SubCategories}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories/subcategories [get]
func getSubcategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = c.Param("page")
		_ = c.Param("limit")
		c.JSON(200, dto.Response{
			Data:    dto.SubCategories{},
			Message: "Subcategories fetched successfully",
			Status:  "success",
		})
	}
}

// @Swagger:route GET /v1/categories/{id}/subcategories/{subcategory_id} subcategories getSubcategoryByID
// @Summary Get subcategory by ID
// @Description Fetches a subcategory by its unique identifier.
// @Tags subcategories
// @Accept json
// @Produce json
// @Param id path string true "Unique identifier for the category"
// @Param subcategory_id path string true "Unique identifier for the subcategory"
// @Success 200 {object} dto.Response{data=dto.SubCategory}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories/{id}/subcategories/{subcategory_id} [get]
func getSubcategoryByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.SubCategory{},
			Message: "Subcategory fetched successfully",
			Status:  "success",
		})
	}
}

type CreateSubcategoryRequest struct {
	SubCategoryName  string `json:"sub_category_name" binding:"required"`
	CategoryID       string `json:"category_id" binding:"required"`
	ImageURL         string `json:"image_url" binding:"required"`
	Status           string `json:"status" binding:"required,oneof=active disabled"`
	IsTopSubCategory bool   `json:"is_top_sub_category" binding:"required"`
}

// @Swagger:route POST /v1/categories/subcategories subcategories createSubcategory
// @Summary Create a new subcategory
// @Description Creates a new subcategory. This endpoint is intended for internal use and should be secured appropriately.
// @Tags subcategories
// @Accept application/json
// @Produce json
// @Param subcategory body CreateSubcategoryRequest true "Create Subcategory Request"
// @Success 201 {object} dto.Response{data=dto.SubCategory}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories/subcategories [post]
func CreateSubcategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.SubCategory{},
			Message: "Subcategory created successfully",
			Status:  "success",
		})
	}
}

type UpdateSubcategoryRequest struct {
	SubCategoryID    string `json:"sub_category_id" binding:"required"`
	SubCategoryName  string `json:"sub_category_name"`
	CategoryID       string `json:"category_id"`
	ImageURL         string `json:"image_url"`
	Status           string `json:"status" binding:"omitempty,oneof=active disabled"`
	IsTopSubCategory *bool  `json:"is_top_sub_category"`
}

// @Swagger:route PATCH /v1/categories/subcategories subcategories updateSubcategory
// @Summary Update an existing subcategory (id supplied via form data)
// @Description Updates an existing subcategory. This endpoint is intended for internal use and should be secured appropriately.
// @Tags subcategories
// @Accept application/json
// @Produce json
// @Param subcategory body UpdateSubcategoryRequest true "Update Subcategory Request"
// @Success 200 {object} dto.Response{data=dto.SubCategory}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories/subcategories [patch]
func UpdateSubcategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.SubCategory{},
			Message: "Subcategory updated successfully",
			Status:  "success",
		})
	}
}

type DeleteSubcategoryRequest struct {
	SubCategoryID string `json:"sub_category_id" binding:"required"`
}

// @Swagger:route DELETE /v1/categories/subcategories subcategories deleteSubcategory
// @Summary Delete a subcategory
// @Description Deletes a subcategory. This endpoint is intended for internal use and should be secured appropriately.
// @Tags subcategories
// @Accept json
// @Produce json
// @Param DeleteSubcategoryRequest body DeleteSubcategoryRequest true "Delete Subcategory Request"
// @Success 200 {object} dto.Response{data=string}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories/subcategories [delete]
func DeleteSubcategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    nil,
			Message: "Subcategory deleted successfully",
			Status:  "success",
		})
	}
}
