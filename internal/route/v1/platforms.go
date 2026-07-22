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

type PlatformDeps struct {
	JWT       *auth.JWTService
	Auth      *middleware.AuthDeps
	Users     *repository.UserRepository
	Platforms *repository.PlatformRepository
}

func RegisterPlatformRoutes(r *gin.Engine, d PlatformDeps) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(d.Auth))

	admin := group.Group("")
	admin.Use(middleware.AdminOnly())
	admin.GET("/platforms/all", listPlatformsAdmin(d))
	admin.GET("/platforms/:platform_id", getPlatformByID(d))
	admin.POST("/platforms", createPlatform(d))
	admin.PATCH("/platforms", updatePlatform(d))
	admin.PATCH("/platforms/reorder", reorderPlatforms(d))
	admin.DELETE("/platforms", deletePlatform(d))
}

// @Summary Reorder platforms
// @Description Sets display_order from the given order (drag-to-reorder). Send the ids in the desired order.
// @Tags platforms
// @Accept json
// @Produce json
// @Param request body ReorderRequest true "Ordered platform IDs"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/platforms/reorder [patch]
func reorderPlatforms(d PlatformDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReorderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Platforms.Reorder(req.IDs); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Platforms reordered"})
	}
}

func toPlatformDTO(p domain.Platform) dto.PlatformItem {
	qc := util.Deref(p.QCName)
	return dto.PlatformItem{
		PlatformID:      p.ID,
		Code:            p.Code,
		DisplayName:     p.DisplayName,
		QCName:          qc,
		Searchable:      qc != "",
		LogoURL:         util.Deref(p.LogoURL),
		DeliveryETAText: util.Deref(p.DeliveryETAText),
		Status:          dto.StatusLabel(p.Enabled),
		DisplayOrder:    p.DisplayOrder,
		CreatedAt:       p.CreatedAt.Format(time.RFC3339),
	}
}

// @Summary List all platforms (admin)
// @Description Every platform including disabled ones. searchable=true means the platform has a QuickCommerce mapping (qc_name) and can be searched/linked.
// @Tags platforms
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param search query string false "Filter by name or code"
// @Success 200 {object} dto.Response{data=dto.Platforms}
// @Security BearerAuth
// @Router /v1/platforms/all [get]
func listPlatformsAdmin(d PlatformDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := query.PageFromContext(c)
		items, total, err := d.Platforms.ListAdmin(p, c.Query("search"))
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]dto.PlatformItem, len(items))
		for i, it := range items {
			out[i] = toPlatformDTO(it)
		}
		ok(c, dto.Platforms{Meta: query.BuildMeta(total, p), Platforms: out})
	}
}

// @Summary Get a platform by ID
// @Tags platforms
// @Produce json
// @Param platform_id path string true "Platform ID"
// @Success 200 {object} dto.Response{data=dto.PlatformItem}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/platforms/{platform_id} [get]
func getPlatformByID(d PlatformDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, err := d.Platforms.FindByID(c.Param("platform_id"))
		if err != nil {
			c.Error(err)
			return
		}
		if p == nil {
			c.Error(errs.NotFound("PLATFORM_NOT_FOUND", "platform not found"))
			return
		}
		ok(c, toPlatformDTO(*p))
	}
}

type CreatePlatformRequest struct {
	Code            string `json:"code" binding:"required"`
	DisplayName     string `json:"display_name" binding:"required"`
	QCName          string `json:"qc_name"`
	LogoURL         string `json:"logo_url"`
	DeliveryETAText string `json:"delivery_eta_text"`
	Enabled         *bool  `json:"enabled"`
}

// @Summary Create a platform
// @Description qc_name is the platform's name on QuickCommerce (e.g. "BlinkIt", "Zepto", "Swiggy"). Leave it blank and the platform exists but cannot be searched or linked.
// @Tags platforms
// @Accept json
// @Produce json
// @Param platform body CreatePlatformRequest true "Create Platform Request"
// @Success 201 {object} dto.Response{data=dto.PlatformItem}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/platforms [post]
func createPlatform(d PlatformDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreatePlatformRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		enabled := true
		if req.Enabled != nil {
			enabled = *req.Enabled
		}
		created, err := d.Platforms.Create(&domain.Platform{
			Code:            req.Code,
			DisplayName:     req.DisplayName,
			QCName:          util.PtrIfSet(req.QCName),
			LogoURL:         util.PtrIfSet(req.LogoURL),
			DeliveryETAText: util.PtrIfSet(req.DeliveryETAText),
			Enabled:         enabled,
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(201, dto.Response{Status: "success", Data: toPlatformDTO(*created), Message: "Platform created successfully"})
	}
}

type UpdatePlatformRequest struct {
	PlatformID      string `json:"platform_id" binding:"required"`
	Code            string `json:"code"`
	DisplayName     string `json:"display_name"`
	QCName          string `json:"qc_name"`
	LogoURL         string `json:"logo_url"`
	DeliveryETAText string `json:"delivery_eta_text"`
	Enabled         *bool  `json:"enabled"`
}

// @Summary Update a platform
// @Description Set qc_name to make a platform searchable on QuickCommerce; clear it (send "-") to make it unsearchable.
// @Tags platforms
// @Accept json
// @Produce json
// @Param platform body UpdatePlatformRequest true "Update Platform Request"
// @Success 200 {object} dto.Response{data=dto.PlatformItem}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/platforms [patch]
func updatePlatform(d PlatformDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdatePlatformRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		fields := map[string]any{}
		if req.Code != "" {
			fields["code"] = req.Code
		}
		if req.DisplayName != "" {
			fields["display_name"] = req.DisplayName
		}
		switch req.QCName {
		case "":
			// not supplied — leave as-is
		case "-":
			fields["qc_name"] = nil // explicit clear: platform becomes unsearchable
		default:
			fields["qc_name"] = req.QCName
		}
		if req.LogoURL != "" {
			fields["logo_url"] = req.LogoURL
		}
		if req.DeliveryETAText != "" {
			fields["delivery_eta_text"] = req.DeliveryETAText
		}
		if req.Enabled != nil {
			fields["enabled"] = *req.Enabled
		}
		updated, err := d.Platforms.Update(req.PlatformID, fields)
		if err != nil {
			c.Error(err)
			return
		}
		if updated == nil {
			c.Error(errs.NotFound("PLATFORM_NOT_FOUND", "platform not found"))
			return
		}
		ok(c, toPlatformDTO(*updated))
	}
}

type DeletePlatformRequest struct {
	PlatformID string `json:"platform_id" binding:"required"`
}

// @Summary Delete a platform
// @Tags platforms
// @Accept json
// @Produce json
// @Param DeletePlatformRequest body DeletePlatformRequest true "Delete Platform Request"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/platforms [delete]
func deletePlatform(d PlatformDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeletePlatformRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Platforms.SoftDelete(req.PlatformID, auth.MustUser(c).ID); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Platform deleted successfully"})
	}
}
