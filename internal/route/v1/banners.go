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

type BannerDeps struct {
	JWT     *auth.JWTService
	Auth    *middleware.AuthDeps
	Users   *repository.UserRepository
	Banners *repository.BannerRepository
}

func RegisterBannerRoutes(r *gin.Engine, d BannerDeps) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(d.Auth))
	group.GET("/banners", listBanners(d))
	group.GET("/banners/:banner_id", getBannerByID(d))

	admin := group.Group("")
	admin.Use(middleware.AdminOnly())
	admin.POST("/banners", createBanner(d))
	admin.PATCH("/banners", updateBanner(d))
	admin.DELETE("/banners", deleteBanner(d))
}

func toBannerDTO(b domain.Banner) dto.BannerItem {
	return dto.BannerItem{
		BannerID:  b.ID,
		Title:     b.Title,
		StartDate: util.FmtDate(b.StartDate),
		EndDate:   util.FmtDate(b.EndDate),
		ImageURL:  b.ImageURL,
		Status:    dto.StatusLabel(b.IsActive),
		IsLive:    b.IsLive(time.Now()),
		CreatedAt: b.CreatedAt.Format(time.RFC3339),
	}
}

// @Summary Get banners
// @Description Paginated list of all banners (admin sees inactive/expired too).
// @Tags banners
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} dto.Response{data=dto.Banners}
// @Security BearerAuth
// @Router /v1/banners [get]
func listBanners(d BannerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := query.PageFromContext(c)
		items, total, err := d.Banners.ListAdmin(p)
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]dto.BannerItem, len(items))
		for i, it := range items {
			out[i] = toBannerDTO(it)
		}
		ok(c, dto.Banners{Meta: query.BuildMeta(total, p), Banners: out})
	}
}

// @Summary Get banner by ID
// @Tags banners
// @Produce json
// @Param banner_id path string true "Banner ID"
// @Success 200 {object} dto.Response{data=dto.BannerItem}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/banners/{banner_id} [get]
func getBannerByID(d BannerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		b, err := d.Banners.FindByID(c.Param("banner_id"))
		if err != nil {
			c.Error(err)
			return
		}
		if b == nil {
			c.Error(errs.NotFound("BANNER_NOT_FOUND", "banner not found"))
			return
		}
		ok(c, toBannerDTO(*b))
	}
}

type CreateBannerRequest struct {
	Title     string `json:"title" binding:"required"`
	ImageURL  string `json:"image_url" binding:"required"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	IsActive  *bool  `json:"is_active"`
}

// @Summary Create a banner
// @Tags banners
// @Accept json
// @Produce json
// @Param banner body CreateBannerRequest true "Create Banner Request"
// @Success 201 {object} dto.Response{data=dto.BannerItem}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/banners [post]
func createBanner(d BannerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateBannerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		start, ok1 := util.ParseDate(req.StartDate)
		end, ok2 := util.ParseDate(req.EndDate)
		if !ok1 || !ok2 {
			c.Error(errs.BadRequest("VALIDATION", "start_date/end_date must be RFC3339 or YYYY-MM-DD"))
			return
		}
		active := true
		if req.IsActive != nil {
			active = *req.IsActive
		}
		created, err := d.Banners.Create(&domain.Banner{
			Title:      req.Title,
			ImageURL:   req.ImageURL,
			TargetType: domain.BannerTargetNone,
			StartDate:  start,
			EndDate:    end,
			IsActive:   active,
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(201, dto.Response{Status: "success", Data: toBannerDTO(*created), Message: "Banner created successfully"})
	}
}

type UpdateBannerRequest struct {
	BannerID  string `json:"banner_id" binding:"required"`
	Title     string `json:"title"`
	ImageURL  string `json:"image_url"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	IsActive  *bool  `json:"is_active"`
}

// @Summary Update a banner
// @Tags banners
// @Accept json
// @Produce json
// @Param banner body UpdateBannerRequest true "Update Banner Request"
// @Success 200 {object} dto.Response{data=dto.BannerItem}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/banners [patch]
func updateBanner(d BannerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateBannerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		fields := map[string]any{}
		if req.Title != "" {
			fields["title"] = req.Title
		}
		if req.ImageURL != "" {
			fields["image_url"] = req.ImageURL
		}
		if req.StartDate != "" {
			t, valid := util.ParseDate(req.StartDate)
			if !valid {
				c.Error(errs.BadRequest("VALIDATION", "start_date must be RFC3339 or YYYY-MM-DD"))
				return
			}
			fields["start_date"] = t
		}
		if req.EndDate != "" {
			t, valid := util.ParseDate(req.EndDate)
			if !valid {
				c.Error(errs.BadRequest("VALIDATION", "end_date must be RFC3339 or YYYY-MM-DD"))
				return
			}
			fields["end_date"] = t
		}
		if req.IsActive != nil {
			fields["is_active"] = *req.IsActive
		}
		updated, err := d.Banners.Update(req.BannerID, fields)
		if err != nil {
			c.Error(err)
			return
		}
		if updated == nil {
			c.Error(errs.NotFound("BANNER_NOT_FOUND", "banner not found"))
			return
		}
		ok(c, toBannerDTO(*updated))
	}
}

type DeleteBannerRequest struct {
	BannerID string `json:"banner_id" binding:"required"`
}

// @Summary Delete a banner
// @Tags banners
// @Accept json
// @Produce json
// @Param DeleteBannerRequest body DeleteBannerRequest true "Delete Banner Request"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/banners [delete]
func deleteBanner(d BannerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeleteBannerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Banners.SoftDelete(req.BannerID, auth.MustUser(c).ID); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Banner deleted successfully"})
	}
}
