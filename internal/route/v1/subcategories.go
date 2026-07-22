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

type SubcategoryDeps struct {
	JWT           *auth.JWTService
	Auth          *middleware.AuthDeps
	Users         *repository.UserRepository
	Subcategories *repository.SubcategoryRepository
	Categories    *repository.CategoryRepository
}

// Routes live under /v1/subcategories (not /v1/categories/...) to avoid colliding
// with the GET /v1/categories/:category_id wildcard.
func RegisterSubcategoryRoutes(r *gin.Engine, d SubcategoryDeps) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(d.Auth))
	group.GET("/subcategories", listSubcategories(d))
	group.GET("/subcategories/:subcategory_id", getSubcategoryByID(d))

	admin := group.Group("")
	admin.Use(middleware.AdminOnly())
	admin.POST("/subcategories", createSubcategory(d))
	admin.PATCH("/subcategories", updateSubcategory(d))
	admin.PATCH("/subcategories/reorder", reorderSubcategories(d))
	admin.DELETE("/subcategories", deleteSubcategory(d))
}

// @Summary Reorder subcategories
// @Description Sets display_order from the given order (drag-to-reorder). Send the ids in the desired order.
// @Tags subcategories
// @Accept json
// @Produce json
// @Param request body ReorderRequest true "Ordered subcategory IDs"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/subcategories/reorder [patch]
func reorderSubcategories(d SubcategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReorderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Subcategories.Reorder(req.IDs); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Subcategories reordered"})
	}
}

func toSubcategoryDTO(s domain.Subcategory, categoryName string) dto.SubCategory {
	return dto.SubCategory{
		SubCategoryID:    s.ID,
		SubCategoryName:  s.Name,
		ImageURL:         util.Deref(s.ImageURL),
		Status:           string(s.Status),
		IsTopSubCategory: s.IsTopSubcategory,
		DisplayOrder:     s.DisplayOrder,
		CategoryID:       s.CategoryID,
		CategoryName:     categoryName,
		CreatedAt:        s.CreatedAt.Format(time.RFC3339),
	}
}

func (d SubcategoryDeps) categoryName(categoryID string) string {
	names, err := d.Categories.NamesByIDs([]string{categoryID})
	if err != nil {
		return ""
	}
	return names[categoryID]
}

// @Summary Get subcategories
// @Tags subcategories
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param category_id query string false "Filter by parent category"
// @Param search query string false "Filter by name"
// @Success 200 {object} dto.Response{data=dto.SubCategories}
// @Security BearerAuth
// @Param has_products query bool false "if true, only subcategories with at least one active variant (use for the consumer browse to avoid empty tiles)"
// @Router /v1/subcategories [get]
func listSubcategories(d SubcategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := query.PageFromContext(c)
		categoryID := c.Query("category_id")
		items, total, err := d.Subcategories.ListAdmin(p, categoryID, c.Query("search"), c.Query("has_products") == "true")
		if err != nil {
			c.Error(err)
			return
		}
		catIDs := make([]string, 0, len(items))
		for _, it := range items {
			catIDs = append(catIDs, it.CategoryID)
		}
		catNames, err := d.Categories.NamesByIDs(catIDs)
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]dto.SubCategory, len(items))
		for i, it := range items {
			out[i] = toSubcategoryDTO(it, catNames[it.CategoryID])
		}
		resp := dto.SubCategories{Meta: query.BuildMeta(total, p), SubCategories: out}
		if categoryID != "" {
			if cat, _ := d.Categories.FindByID(categoryID); cat != nil {
				resp.Category = dto.CategoryData{CategoryID: cat.ID, CategoryName: cat.Name}
			}
		}
		ok(c, resp)
	}
}

// @Summary Get subcategory by ID
// @Tags subcategories
// @Produce json
// @Param subcategory_id path string true "Subcategory ID"
// @Success 200 {object} dto.Response{data=dto.SubCategory}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/subcategories/{subcategory_id} [get]
func getSubcategoryByID(d SubcategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		s, err := d.Subcategories.FindByID(c.Param("subcategory_id"))
		if err != nil {
			c.Error(err)
			return
		}
		if s == nil {
			c.Error(errs.NotFound("SUBCATEGORY_NOT_FOUND", "subcategory not found"))
			return
		}
		ok(c, toSubcategoryDTO(*s, d.categoryName(s.CategoryID)))
	}
}

type CreateSubcategoryRequest struct {
	SubCategoryName  string `json:"sub_category_name" binding:"required"`
	CategoryID       string `json:"category_id" binding:"required"`
	ImageURL         string `json:"image_url" binding:"required"`
	Status           string `json:"status" binding:"required,oneof=active disabled"`
	IsTopSubCategory bool   `json:"is_top_sub_category"`
}

// @Summary Create a subcategory
// @Tags subcategories
// @Accept json
// @Produce json
// @Param subcategory body CreateSubcategoryRequest true "Create Subcategory Request"
// @Success 201 {object} dto.Response{data=dto.SubCategory}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/subcategories [post]
func createSubcategory(d SubcategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateSubcategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		slug := util.Slugify(req.SubCategoryName)
		created, err := d.Subcategories.Create(&domain.Subcategory{
			CategoryID:       req.CategoryID,
			Name:             req.SubCategoryName,
			Slug:             util.PtrIfSet(slug),
			ImageURL:         util.PtrIfSet(req.ImageURL),
			Status:           domain.Status(req.Status),
			IsTopSubcategory: req.IsTopSubCategory,
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(201, dto.Response{Status: "success", Data: toSubcategoryDTO(*created, d.categoryName(created.CategoryID)), Message: "Subcategory created successfully"})
	}
}

type UpdateSubcategoryRequest struct {
	SubCategoryID    string `json:"subcategory_id" binding:"required"`
	SubCategoryName  string `json:"sub_category_name"`
	CategoryID       string `json:"category_id"`
	ImageURL         string `json:"image_url"`
	Status           string `json:"status" binding:"omitempty,oneof=active disabled"`
	IsTopSubCategory *bool  `json:"is_top_sub_category"`
}

// @Summary Update a subcategory
// @Tags subcategories
// @Accept json
// @Produce json
// @Param subcategory body UpdateSubcategoryRequest true "Update Subcategory Request"
// @Success 200 {object} dto.Response{data=dto.SubCategory}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/subcategories [patch]
func updateSubcategory(d SubcategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateSubcategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		fields := map[string]any{}
		if req.SubCategoryName != "" {
			fields["name"] = req.SubCategoryName
			fields["slug"] = util.Slugify(req.SubCategoryName)
		}
		if req.CategoryID != "" {
			fields["category_id"] = req.CategoryID
		}
		if req.ImageURL != "" {
			fields["image_url"] = req.ImageURL
		}
		if req.Status != "" {
			fields["status"] = req.Status
		}
		if req.IsTopSubCategory != nil {
			fields["is_top_subcategory"] = *req.IsTopSubCategory
		}
		updated, err := d.Subcategories.Update(req.SubCategoryID, fields)
		if err != nil {
			c.Error(err)
			return
		}
		if updated == nil {
			c.Error(errs.NotFound("SUBCATEGORY_NOT_FOUND", "subcategory not found"))
			return
		}
		ok(c, toSubcategoryDTO(*updated, d.categoryName(updated.CategoryID)))
	}
}

type DeleteSubcategoryRequest struct {
	SubCategoryID string `json:"subcategory_id" binding:"required"`
}

// @Summary Delete a subcategory
// @Tags subcategories
// @Accept json
// @Produce json
// @Param DeleteSubcategoryRequest body DeleteSubcategoryRequest true "Delete Subcategory Request"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/subcategories [delete]
func deleteSubcategory(d SubcategoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeleteSubcategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Subcategories.SoftDelete(req.SubCategoryID, auth.MustUser(c).ID); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Subcategory deleted successfully"})
	}
}
