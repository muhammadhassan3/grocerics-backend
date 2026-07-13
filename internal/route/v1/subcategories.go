package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterSubcategoryRoutes(jwt *auth.JWTService, user *repository.UserRepository, r *gin.RouterGroup) {
	group := r.Group("/v1/categories")
	group.Use(middleware.AuthMiddleware(jwt, user))
	group.GET("/:category_id/subcategories", getSubcategories())
	adminGroup := group.Group("")
	adminGroup.Use(middleware.RequireRole(domain.RoleAdmin))
	adminGroup.POST("/subcategories", CreateSubcategory())
	adminGroup.PATCH("/subcategories", UpdateSubcategory())
	adminGroup.DELETE("/subcategories", DeleteSubcategory())
}

// @Swagger:route GET /v1/categories/{category_id}/subcategories subcategories getSubcategories
// @Summary Get subcategories
// @Description Fetches a paginated list of subcategories.
// @Tags subcategories
// @Accept json
// @Produce json
// @Param category_id path string true "Unique identifier for the category"
// @Param page query int true "Page number"
// @Param limit query int true "Number of items per page"
// @Success 200 {object} dto.Response{data=dto.SubCategories}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories/{category_id}/subcategories [get]
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

// @Swagger:route POST /v1/categories/subcategories subcategories createSubcategory
// @Summary Create a new subcategory
// @Description Creates a new subcategory. This endpoint is intended for internal use and should be secured appropriately.
// @Tags subcategories
// @Accept multipart/form-data
// @Produce json
// @Param sub_category_name formData string true "Display name of the subcategory"
// @Param category_id formData string true "Identifier of the parent category"
// @Param image formData file true "Image file for the subcategory"
// @Param status formData string true "Status of the subcategory" enums(active,disabled)
// @Param is_top_sub_category formData bool true "Whether the subcategory is flagged as a top/featured subcategory"
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

// @Swagger:route PATCH /v1/categories/subcategories subcategories updateSubcategory
// @Summary Update an existing subcategory (id supplied via form data)
// @Description Updates an existing subcategory. This endpoint is intended for internal use and should be secured appropriately.
// @Tags subcategories
// @Accept multipart/form-data
// @Produce json
// @Param sub_category_id formData string true "Unique identifier for the subcategory"
// @Param sub_category_name formData string false "Display name of the subcategory"
// @Param category_id formData string false "Identifier of the parent category"
// @Param image formData file false "Image file for the subcategory"
// @Param status formData string false "Status of the subcategory" enums(active,disabled)
// @Param is_top_sub_category formData bool false "Whether the subcategory is flagged as a top/featured subcategory"
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
