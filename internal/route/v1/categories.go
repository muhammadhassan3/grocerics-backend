package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterCategoryRoutes(jwt *auth.JWTService, user *repository.UserRepository, r *gin.Engine) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(jwt, user))
	group.GET("/categories", getCategories())
	group.GET("/categories/:category_id", getCategoryByID())

	adminGroup := group.Group("")
	adminGroup.Use(middleware.RequireRole("admin"))
	adminGroup.POST("/categories", CreateCategory())
	adminGroup.PATCH("/categories", UpdateCategory())
	adminGroup.DELETE("/categories", DeleteCategory())
}

// @Swagger:route GET /v1/categories categories getCategories
// @Summary Get categories
// @Description Fetches a paginated list of categories.
// @Tags categories
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param limit query int true "Number of items per page"
// @Success 200 {object} dto.Response{data=dto.Categories}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories [get]
func getCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = c.Param("page")
		_ = c.Param("limit")
		c.JSON(200, dto.Response{
			Data:    dto.Categories{},
			Message: "Categories fetched successfully",
			Status:  "success",
		})
	}
}

// @Swagger:route GET /v1/categories/{category_id} categories getCategoryByID
// @Summary Get category by ID
// @Description Fetches a category by its unique identifier.
// @Tags categories
// @Accept json
// @Produce json
// @Param category_id path string true "Unique identifier for the category"
// @Success 200 {object} dto.Response{data=dto.Category}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories/{category_id} [get]
func getCategoryByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.Category{},
			Message: "Category fetched successfully",
			Status:  "success",
		})
	}
}

type CreateCategoryRequest struct {
	CategoryName  string `json:"category_name" binding:"required"`
	ImageURL      string `json:"image_url" binding:"required"`
	Status        string `json:"status" binding:"required,oneof=active disabled"`
	IsTopCategory bool   `json:"is_top_category" binding:"required"`
}

// @Swagger:route POST /v1/categories categories createCategory
// @Summary Create a new category
// @Description Creates a new category. This endpoint is intended for internal use and should be secured appropriately.
// @Tags categories
// @Accept application/json
// @Produce json
// @Param category body CreateCategoryRequest true "Create Category Request"
// @Success 201 {object} dto.Response{data=dto.Category}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories [post]
func CreateCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.Category{},
			Message: "Category created successfully",
			Status:  "success",
		})
	}
}

type UpdateCategoryRequest struct {
	CategoryID    string `json:"category_id" binding:"required"`
	CategoryName  string `json:"category_name"`
	ImageURL      string `json:"image_url"`
	Status        string `json:"status" binding:"omitempty,oneof=active disabled"`
	IsTopCategory *bool  `json:"is_top_category"`
}

// @Swagger:route PATCH /v1/categories categories updateCategory
// @Summary Update an existing category
// @Description Updates an existing category. This endpoint is intended for internal use and should be secured appropriately.
// @Tags categories
// @Accept application/json
// @Produce json
// @Param category body UpdateCategoryRequest true "Update Category Request"
// @Success 200 {object} dto.Response{data=dto.Category}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories [patch]
func UpdateCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.Category{},
			Message: "Category updated successfully",
			Status:  "success",
		})
	}
}

type DeleteCategoryRequest struct {
	CategoryID string `json:"category_id" binding:"required"`
}

// @Swagger:route DELETE /v1/categories categories deleteCategory
// @Summary Delete an existing category
// @Description Deletes an existing category. This endpoint is intended for internal use and should be secured appropriately.
// @Tags categories
// @Accept json
// @Produce json
// @Param DeleteCategoryRequest body DeleteCategoryRequest true "Unique identifier for the category"
// @Success 200 {object} dto.Response{data=string}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories [delete]
func DeleteCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    nil,
			Message: "Category deleted successfully",
			Status:  "success",
		})
	}
}
