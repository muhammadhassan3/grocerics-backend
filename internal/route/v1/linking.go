package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/repository"
	"grocerics-backend/internal/service"
	"grocerics-backend/internal/util"

	"github.com/gin-gonic/gin"
)

type LinkingDeps struct {
	JWT            *auth.JWTService
	Users          *repository.UserRepository
	Platforms      *repository.PlatformRepository
	Links          *repository.ProductPlatformLinkRepository
	PlatformPrices *repository.PlatformPriceRepository
	Linking        *service.LinkingService
	Pricing        *service.PricingService
}

func RegisterLinkingRoutes(r *gin.Engine, d LinkingDeps) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(d.JWT, d.Users))
	group.Use(middleware.RequireRole(domain.RoleAdmin))

	group.GET("/platforms", listPlatforms(d))
	group.GET("/inventory-management/link/search", searchLinkCandidates(d))
	group.POST("/inventory-management/variants/:variant_id/links", confirmLink(d))
	group.GET("/inventory-management/variants/:variant_id/prices", variantPrices(d))
	group.POST("/inventory-management/variants/manual-price", setManualPrice(d))
	group.POST("/inventory-management/refresh", refreshPrices(d))
}

type PlatformOption struct {
	PlatformID  string `json:"platform_id"`
	Code        string `json:"code"`
	DisplayName string `json:"display_name"`
	QCName      string `json:"qc_name,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	Searchable  bool   `json:"searchable"`
}

type LinkSearchResponse struct {
	CreditsRemaining int                            `json:"credits_remaining"`
	Results          map[string][]service.Candidate `json:"results"`
}

// VariantPriceRow is one platform's price for a variant in a city.
// Money is integer paise, everywhere, always (₹38.00 => 3800). Divide by 100 to display.
type VariantPriceRow struct {
	PlatformID   string `json:"platform_id"`
	PlatformCode string `json:"platform_code"`
	PlatformName string `json:"platform_name"`
	// Selling price in paise. 3800 = ₹38.00
	PricePaise int64 `json:"price_paise"`
	// MRP in paise. 0 when the platform reports no MRP.
	MRPPaise    int64  `json:"mrp_paise"`
	Available   bool   `json:"available"`
	Inventory   *int   `json:"inventory,omitempty"`
	Source      string `json:"source"`
	PlatformSKU string `json:"platform_sku,omitempty"`
	DeepLink    string `json:"deep_link,omitempty"`
	LastUpdated string `json:"last_updated_at,omitempty"`
}

// @Summary List enabled platforms
// @Description Enabled delivery platforms — powers the linking picker's platform checkboxes. Only those with searchable=true can be queried on QuickCommerce.
// @Tags linking
// @Produce json
// @Success 200 {object} dto.Response{data=[]v1.PlatformOption}
// @Security BearerAuth
// @Router /v1/platforms [get]
func listPlatforms(d LinkingDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		plats, err := d.Platforms.ListEnabled()
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]PlatformOption, len(plats))
		for i, p := range plats {
			qc := util.Deref(p.QCName)
			out[i] = PlatformOption{
				PlatformID:  p.ID,
				Code:        p.Code,
				DisplayName: p.DisplayName,
				QCName:      qc,
				LogoURL:     util.Deref(p.LogoURL),
				Searchable:  qc != "",
			}
		}
		ok(c, out)
	}
}

// @Summary Search QuickCommerce candidates for linking
// @Description Discovery step: groupsearch across platforms so the admin can confirm the right item per platform. Costs 1 credit per platform. Never called on a consumer request.
// @Tags linking
// @Produce json
// @Param q query string true "search term"
// @Param city query string true "city ID (supplies the QC location anchor)"
// @Param platforms query string false "comma-separated platform codes to search, e.g. blinkit,zepto. Omit to search every enabled QC-mapped platform. Costs 1 credit per platform."
// @Success 200 {object} dto.Response{data=v1.LinkSearchResponse}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/link/search [get]
func searchLinkCandidates(d LinkingDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		cityID := c.Query("city")
		// platforms is optional: empty = every enabled QC-mapped platform.
		codes := util.SplitCSV(c.Query("platforms"))
		if len(q) < 2 || cityID == "" {
			c.Error(errs.BadRequest("VALIDATION", "q (min 2 chars) and city are required"))
			return
		}
		results, credits, err := d.Linking.SearchCandidates(q, cityID, codes)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, LinkSearchResponse{CreditsRemaining: credits, Results: results})
	}
}

type ConfirmLinkRequest struct {
	PlatformCode string `json:"platform_code" binding:"required"`
	QCItemID     string `json:"qc_item_id" binding:"required"`
	CityID       string `json:"city_id" binding:"required"`
	DeepLink     string `json:"deep_link"`
}

// @Summary Confirm a platform link for a variant
// @Description Pins a (variant, platform) to the chosen QC item id and seeds its price/stock via GetItem. The link is city-independent; the price is stored per city.
// @Tags linking
// @Accept json
// @Produce json
// @Param variant_id path string true "Variant ID"
// @Param request body ConfirmLinkRequest true "Confirm link request"
// @Success 200 {object} dto.Response{data=string}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/variants/{variant_id}/links [post]
func confirmLink(d LinkingDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ConfirmLinkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Linking.ConfirmLink(c.Param("variant_id"), req.PlatformCode, req.CityID, req.QCItemID, req.DeepLink); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Link confirmed and price seeded"})
	}
}

type ManualPriceRequest struct {
	VariantID    string `json:"variant_id" binding:"required"`
	PlatformCode string `json:"platform_code" binding:"required"`
	CityID       string `json:"city_id" binding:"required"`
	MRPPaise     int64  `json:"mrp_paise" binding:"required"`
	PricePaise   int64  `json:"price_paise" binding:"required"`
	PlatformSKU  string `json:"platform_sku"`
	DeepLink     string `json:"deep_link"`
}

// @Summary Set a manual price (fallback when not on QuickCommerce)
// @Description Admin fallback: pin the link and write the price by hand. price = MRP, discounted_price = shown price.
// @Tags linking
// @Accept json
// @Produce json
// @Param request body ManualPriceRequest true "Manual price request"
// @Success 200 {object} dto.Response{data=string}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/variants/manual-price [post]
func setManualPrice(d LinkingDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ManualPriceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		pl, err := d.Platforms.FindByCode(req.PlatformCode)
		if err != nil {
			c.Error(err)
			return
		}
		if pl == nil {
			c.Error(errs.NotFound("PLATFORM_NOT_FOUND", "platform not found"))
			return
		}
		if err := d.Pricing.SetManualPrice(req.VariantID, pl.ID, req.CityID, req.MRPPaise, req.PricePaise, req.PlatformSKU, req.DeepLink); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Manual price set"})
	}
}

type RefreshRequest struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	CityID    string `json:"city_id" binding:"required"`
}

// @Summary Refresh live prices/stock from QuickCommerce
// @Description Re-pulls live price+stock via GetItem for a variant or a whole product's variants in one city. Costs 1 credit per linked platform. One platform failing never fails the refresh.
// @Tags linking
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh request (one of product_id / variant_id)"
// @Success 200 {object} dto.Response{data=service.RefreshResult}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/refresh [post]
func refreshPrices(d LinkingDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if req.VariantID == "" && req.ProductID == "" {
			c.Error(errs.BadRequest("VALIDATION", "one of variant_id or product_id is required"))
			return
		}
		var (
			res service.RefreshResult
			err error
		)
		if req.VariantID != "" {
			res, err = d.Linking.RefreshVariant(req.VariantID, req.CityID)
		} else {
			res, err = d.Linking.RefreshProduct(req.ProductID, req.CityID)
		}
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, res)
	}
}

// @Summary Per-platform prices for a variant
// @Description Joins links + platform_prices for a variant in a city — powers the wizard's Variations grid.
// @Tags linking
// @Produce json
// @Param variant_id path string true "Variant ID"
// @Param city query string true "City ID"
// @Success 200 {object} dto.Response{data=[]v1.VariantPriceRow}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/variants/{variant_id}/prices [get]
func variantPrices(d LinkingDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		variantID := c.Param("variant_id")
		cityID := c.Query("city")
		if cityID == "" {
			c.Error(errs.BadRequest("VALIDATION", "city is required"))
			return
		}
		links, err := d.Links.ListByVariant(variantID)
		if err != nil {
			c.Error(err)
			return
		}
		prices, err := d.PlatformPrices.ListByVariantCity(variantID, cityID)
		if err != nil {
			c.Error(err)
			return
		}
		priceByPlatform := make(map[string]domain.PlatformPrice, len(prices))
		platformIDs := make([]string, 0, len(links))
		for _, p := range prices {
			priceByPlatform[p.PlatformID] = p
		}
		for _, l := range links {
			platformIDs = append(platformIDs, l.PlatformID)
		}
		plats, err := d.Platforms.FindByIDs(platformIDs)
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]VariantPriceRow, 0, len(links))
		for _, l := range links {
			pl := plats[l.PlatformID]
			row := VariantPriceRow{
				PlatformID:   l.PlatformID,
				PlatformCode: pl.Code,
				PlatformName: pl.DisplayName,
				PlatformSKU:  util.Deref(l.PlatformSKU),
				DeepLink:     util.Deref(l.DeepLink),
			}
			if p, okp := priceByPlatform[l.PlatformID]; okp {
				row.PricePaise = p.PricePaise
				if p.MRPPaise != nil {
					row.MRPPaise = *p.MRPPaise
				}
				row.Available = p.Available
				row.Inventory = p.Inventory
				row.Source = string(p.Source)
				row.LastUpdated = p.LastUpdatedAt.Format("2006-01-02T15:04:05Z07:00")
			}
			out = append(out, row)
		}
		ok(c, out)
	}
}
