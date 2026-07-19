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

type BrandDeps struct {
	JWT    *auth.JWTService
	Auth   *middleware.AuthDeps
	Users  *repository.UserRepository
	Brands *repository.BrandRepository
}

func RegisterBrandsRoutes(r *gin.Engine, d BrandDeps) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(d.Auth))
	group.GET("/brands", listBrands(d))
	group.GET("/brands/:brand_id", getBrandByID(d))

	admin := group.Group("")
	admin.Use(middleware.AdminOnly())
	admin.POST("/brands", createBrand(d))
	admin.PATCH("/brands", updateBrand(d))
	admin.PATCH("/brands/reorder", reorderBrands(d))
	admin.DELETE("/brands", deleteBrand(d))
}

// @Summary Reorder brands
// @Description Sets display_order from the given order (drag-to-reorder). Send the ids in the desired order.
// @Tags brands
// @Accept json
// @Produce json
// @Param request body ReorderRequest true "Ordered brand IDs"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/brands/reorder [patch]
func reorderBrands(d BrandDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReorderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Brands.Reorder(req.IDs); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Brands reordered"})
	}
}

func toBrandDTO(b domain.Brand, productCount int) dto.BrandItem {
	return dto.BrandItem{
		BrandID:      b.ID,
		BrandName:    b.Name,
		ImageURL:     util.Deref(b.ImageURL),
		Status:       string(b.Status),
		IsTopBrand:   b.IsTopBrand,
		DisplayOrder: b.DisplayOrder,
		ProductCount: productCount,
		CreatedAt:    b.CreatedAt.Format(time.RFC3339),
	}
}

// @Summary Get brands
// @Tags brands
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param search query string false "Filter by name"
// @Success 200 {object} dto.Response{data=dto.BrandList}
// @Security BearerAuth
// @Router /v1/brands [get]
func listBrands(d BrandDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := query.PageFromContext(c)
		items, total, err := d.Brands.ListAdmin(p, c.Query("search"))
		if err != nil {
			c.Error(err)
			return
		}
		ids := make([]string, len(items))
		for i, it := range items {
			ids[i] = it.ID
		}
		counts, err := d.Brands.CountProducts(ids)
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]dto.BrandItem, len(items))
		for i, it := range items {
			out[i] = toBrandDTO(it, counts[it.ID])
		}
		ok(c, dto.BrandList{Meta: query.BuildMeta(total, p), Brands: out})
	}
}

// @Summary Get brand by ID
// @Tags brands
// @Produce json
// @Param brand_id path string true "Brand ID"
// @Success 200 {object} dto.Response{data=dto.BrandItem}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/brands/{brand_id} [get]
func getBrandByID(d BrandDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		b, err := d.Brands.FindByID(c.Param("brand_id"))
		if err != nil {
			c.Error(err)
			return
		}
		if b == nil {
			c.Error(errs.NotFound("BRAND_NOT_FOUND", "brand not found"))
			return
		}
		counts, err := d.Brands.CountProducts([]string{b.ID})
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, toBrandDTO(*b, counts[b.ID]))
	}
}

type CreateBrandRequest struct {
	BrandName  string `json:"brand_name" binding:"required"`
	ImageURL   string `json:"image_url" binding:"required"`
	Status     string `json:"status" binding:"required,oneof=active disabled"`
	IsTopBrand bool   `json:"is_top_brand"`
}

// @Summary Create a brand
// @Tags brands
// @Accept json
// @Produce json
// @Param brand body CreateBrandRequest true "Create Brand Request"
// @Success 201 {object} dto.Response{data=dto.BrandItem}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/brands [post]
func createBrand(d BrandDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateBrandRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		slug := util.Slugify(req.BrandName)
		created, err := d.Brands.Create(&domain.Brand{
			Name:       req.BrandName,
			Slug:       util.PtrIfSet(slug),
			ImageURL:   util.PtrIfSet(req.ImageURL),
			Status:     domain.Status(req.Status),
			IsTopBrand: req.IsTopBrand,
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(201, dto.Response{Status: "success", Data: toBrandDTO(*created, 0), Message: "Brand created successfully"})
	}
}

type UpdateBrandRequest struct {
	BrandID    string `json:"brand_id" binding:"required"`
	BrandName  string `json:"brand_name"`
	ImageURL   string `json:"image_url"`
	Status     string `json:"status" binding:"omitempty,oneof=active disabled"`
	IsTopBrand *bool  `json:"is_top_brand"`
}

// @Summary Update a brand
// @Tags brands
// @Accept json
// @Produce json
// @Param brand body UpdateBrandRequest true "Update Brand Request"
// @Success 200 {object} dto.Response{data=dto.BrandItem}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/brands [patch]
func updateBrand(d BrandDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateBrandRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		fields := map[string]any{}
		if req.BrandName != "" {
			fields["name"] = req.BrandName
			fields["slug"] = util.Slugify(req.BrandName)
		}
		if req.ImageURL != "" {
			fields["image_url"] = req.ImageURL
		}
		if req.Status != "" {
			fields["status"] = req.Status
		}
		if req.IsTopBrand != nil {
			fields["is_top_brand"] = *req.IsTopBrand
		}
		updated, err := d.Brands.Update(req.BrandID, fields)
		if err != nil {
			c.Error(err)
			return
		}
		if updated == nil {
			c.Error(errs.NotFound("BRAND_NOT_FOUND", "brand not found"))
			return
		}
		ok(c, toBrandDTO(*updated, 0))
	}
}

type DeleteBrandRequest struct {
	BrandID string `json:"brand_id" binding:"required"`
}

// @Summary Delete a brand
// @Tags brands
// @Accept json
// @Produce json
// @Param DeleteBrandRequest body DeleteBrandRequest true "Delete Brand Request"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/brands [delete]
func deleteBrand(d BrandDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeleteBrandRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Brands.SoftDelete(req.BrandID, auth.MustUser(c).ID); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Brand deleted successfully"})
	}
}
