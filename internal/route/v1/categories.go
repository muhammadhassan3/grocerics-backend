package v1

import (
	"time"

	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/query"
	"grocerics-backend/internal/repository"
	"grocerics-backend/internal/util"

	"github.com/gin-gonic/gin"
)

type CategoryDeps struct {
	JWT        *auth.JWTService
	Auth       *middleware.AuthDeps
	Users      *repository.UserRepository
	Categories *repository.CategoryRepository
}

func RegisterCategoryRoutes(r *gin.Engine, d CategoryDeps) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(d.Auth))
	group.GET("/categories", listCategories(d))
	group.GET("/categories/:id", getCategoryByID(d))

	admin := group.Group("")
	admin.Use(middleware.AdminOnly())
	admin.POST("/categories", createCategory(d))
	admin.PATCH("/categories", updateCategory(d))
	admin.PATCH("/categories/reorder", reorderCategories(d))
	admin.DELETE("/categories", deleteCategory(d))
}

type ReorderRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// @Summary Reorder categories
// @Description Sets display_order from the given order (drag-to-reorder). Send every id in the desired order.
// @Tags categories
// @Accept json
// @Produce json
// @Param request body ReorderRequest true "Ordered category IDs"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories/reorder [patch]
func reorderCategories(d CategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReorderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Categories.Reorder(req.IDs); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Categories reordered"})
	}
}

func toCategoryDTO(c domain.Category, subCount int) dto.Category {
	return dto.Category{
		CategoryID:       c.ID,
		CategoryName:     c.Name,
		ImageURL:         util.Deref(c.ImageURL),
		SubCategoryCount: subCount,
		Status:           string(c.Status),
		IsTopCategory:    c.IsTopCategory,
		DisplayOrder:     c.DisplayOrder,
		CreatedAt:        c.CreatedAt.Format(time.RFC3339),
	}
}

// @Summary Get categories
// @Description Paginated list of categories (admin sees all, including disabled).
// @Tags categories
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param search query string false "Filter by name"
// @Success 200 {object} dto.Response{data=dto.Categories}
// @Security BearerAuth
// @Router /v1/categories [get]
func listCategories(d CategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := query.PageFromContext(c)
		items, total, err := d.Categories.ListAdmin(p, c.Query("search"))
		if err != nil {
			c.Error(err)
			return
		}
		ids := make([]string, len(items))
		for i, it := range items {
			ids[i] = it.ID
		}
		counts, err := d.Categories.CountSubcategories(ids)
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]dto.Category, len(items))
		for i, it := range items {
			out[i] = toCategoryDTO(it, counts[it.ID])
		}
		ok(c, dto.Categories{Meta: query.BuildMeta(total, p), Categories: out})
	}
}

// @Summary Get category by ID
// @Tags categories
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} dto.Response{data=dto.Category}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories/{id} [get]
func getCategoryByID(d CategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		cat, err := d.Categories.FindByID(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}
		if cat == nil {
			c.Error(errs.NotFound("CATEGORY_NOT_FOUND", "category not found"))
			return
		}
		counts, err := d.Categories.CountSubcategories([]string{cat.ID})
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, toCategoryDTO(*cat, counts[cat.ID]))
	}
}

type CreateCategoryRequest struct {
	CategoryName  string `json:"category_name" binding:"required"`
	ImageURL      string `json:"image_url" binding:"required"`
	Status        string `json:"status" binding:"required,oneof=active disabled"`
	IsTopCategory bool   `json:"is_top_category"`
}

// @Summary Create a category
// @Tags categories
// @Accept json
// @Produce json
// @Param category body CreateCategoryRequest true "Create Category Request"
// @Success 201 {object} dto.Response{data=dto.Category}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories [post]
func createCategory(d CategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateCategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		created, err := d.Categories.Create(&domain.Category{
			Name:          req.CategoryName,
			Slug:          util.Slugify(req.CategoryName),
			ImageURL:      util.PtrIfSet(req.ImageURL),
			Status:        domain.Status(req.Status),
			IsTopCategory: req.IsTopCategory,
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(201, dto.Response{Status: "success", Data: toCategoryDTO(*created, 0), Message: "Category created successfully"})
	}
}

type UpdateCategoryRequest struct {
	CategoryID    string `json:"category_id" binding:"required"`
	CategoryName  string `json:"category_name"`
	ImageURL      string `json:"image_url"`
	Status        string `json:"status" binding:"omitempty,oneof=active disabled"`
	IsTopCategory *bool  `json:"is_top_category"`
}

// @Summary Update a category
// @Tags categories
// @Accept json
// @Produce json
// @Param category body UpdateCategoryRequest true "Update Category Request"
// @Success 200 {object} dto.Response{data=dto.Category}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories [patch]
func updateCategory(d CategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateCategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		fields := map[string]any{}
		if req.CategoryName != "" {
			fields["name"] = req.CategoryName
			fields["slug"] = util.Slugify(req.CategoryName)
		}
		if req.ImageURL != "" {
			fields["image_url"] = req.ImageURL
		}
		if req.Status != "" {
			fields["status"] = req.Status
		}
		if req.IsTopCategory != nil {
			fields["is_top_category"] = *req.IsTopCategory
		}
		updated, err := d.Categories.Update(req.CategoryID, fields)
		if err != nil {
			c.Error(err)
			return
		}
		if updated == nil {
			c.Error(errs.NotFound("CATEGORY_NOT_FOUND", "category not found"))
			return
		}
		ok(c, toCategoryDTO(*updated, 0))
	}
}

type DeleteCategoryRequest struct {
	CategoryID string `json:"category_id" binding:"required"`
}

// @Summary Delete a category
// @Tags categories
// @Accept json
// @Produce json
// @Param DeleteCategoryRequest body DeleteCategoryRequest true "Delete Category Request"
// @Success 200 {object} dto.Response{data=string}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/categories [delete]
func deleteCategory(d CategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeleteCategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Categories.SoftDelete(req.CategoryID, auth.MustUser(c).ID); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Category deleted successfully"})
	}
}
