package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/query"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type DashboardDeps struct {
	JWT            *auth.JWTService
	Users          *repository.UserRepository
	Analytics      *repository.AnalyticsRepository
	Products       *repository.ProductRepository
	Variants       *repository.ProductVariantRepository
	Categories     *repository.CategoryRepository
	PlatformPrices *repository.PlatformPriceRepository
	Platforms      *repository.PlatformRepository
	Cities         *repository.CityRepository
}

func RegisterDashboardRoutes(r *gin.Engine, d DashboardDeps) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(d.JWT, d.Users))
	group.Use(middleware.RequireRole(domain.RoleAdmin))
	group.GET("/dashboard", getDashboard(d))
	group.GET("/dashboard/stats", getDashboardStats(d))
	group.GET("/dashboard/live-price-comparison", getLivePriceComparison(d))
	group.GET("/dashboard/top-searched-products", getTopSearchedProducts(d))

	group.GET("/dashboard/mobile", getDashboardMobile()) // out of scope for Slice D (mobile stub)
}

func buildStats(d DashboardDeps) (dto.DashboardStats, error) {
	total, err := d.Users.Count()
	if err != nil {
		return dto.DashboardStats{}, err
	}
	sTotal, sThis, sLast, err := d.Analytics.SearchStats()
	if err != nil {
		return dto.DashboardStats{}, err
	}
	avg, err := d.Analytics.AverageBasketItems()
	if err != nil {
		return dto.DashboardStats{}, err
	}
	uDiff, _ := d.Analytics.NewUserMonthlyDiff()

	return dto.DashboardStats{
		TotalUsers:        dto.StatsItem{Value: int(total), DiffLastMonth: uDiff},
		AverageBasketSize: dto.StatsItem{Value: avg, DiffLastMonth: 0},
		TotalSearches:     dto.StatsItem{Value: sTotal, DiffLastMonth: sThis - sLast},
	}, nil
}

// @Summary Get dashboard data
// @Description Fetches the data needed to populate the admin dashboard, including headline stats, daily active users, and monthly active users.
// @Tags dashboard
// @Produce json
// @Success 200 {object} dto.Response{data=dto.DashboardResponse}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/dashboard [get]
func getDashboard(d DashboardDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := buildStats(d)
		if err != nil {
			c.Error(err)
			return
		}
		dau, err := d.Analytics.DAU()
		if err != nil {
			c.Error(err)
			return
		}
		mau, err := d.Analytics.MAU()
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, dto.DashboardResponse{
			Stats: stats,
			DailyActiveUsers: dto.DailyActiveUsers{
				Monday: dau[1], Tuesday: dau[2], Wednesday: dau[3], Thursday: dau[4],
				Friday: dau[5], Saturday: dau[6], Sunday: dau[7],
			},
			MonthlyActiveUsers: dto.MonthlyActiveUsers{
				January: mau[1], February: mau[2], March: mau[3], April: mau[4],
				May: mau[5], June: mau[6], July: mau[7], August: mau[8],
				September: mau[9], October: mau[10], November: mau[11], December: mau[12],
			},
		})
	}
}

// @Summary Get dashboard stats
// @Description Fetches the headline stat cards for the admin dashboard: total users, average basket size, and total searches.
// @Tags dashboard
// @Produce json
// @Param interval query string false "Accepted but currently inert — the response always carries a month-over-month diff" enums(daily,weekly,monthly) default(daily)
// @Success 200 {object} dto.Response{data=dto.DashboardStats}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/dashboard/stats [get]
func getDashboardStats(d DashboardDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := buildStats(d)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, stats)
	}
}

// @Summary Get live price comparison
// @Description Per-product prices across every active platform for the default (first enabled) city.
// @Tags dashboard
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param q query string false "Filter by product name"
// @Success 200 {object} dto.Response{data=dto.LivePrice}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/dashboard/live-price-comparison [get]
func getLivePriceComparison(d DashboardDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := query.PageFromContext(c)
		cities, err := d.Cities.ListEnabled()
		if err != nil {
			c.Error(err)
			return
		}
		if len(cities) == 0 {
			ok(c, dto.LivePrice{Meta: query.BuildMeta(0, p), Products: []dto.ProductPrice{}})
			return
		}
		cityID := cities[0].ID

		prods, total, err := d.Products.ListAdmin(p, c.Query("q"))
		if err != nil {
			c.Error(err)
			return
		}
		pids := make([]string, len(prods))
		catIDs := make([]string, 0, len(prods))
		for i, pr := range prods {
			pids[i] = pr.ID
			catIDs = append(catIDs, pr.CategoryID)
		}
		defaults, err := d.Variants.DefaultsForProducts(pids)
		if err != nil {
			c.Error(err)
			return
		}
		vIDs := make([]string, 0, len(defaults))
		for _, v := range defaults {
			vIDs = append(vIDs, v.ID)
		}
		prices, err := d.PlatformPrices.ListByVariantsCity(vIDs, cityID)
		if err != nil {
			c.Error(err)
			return
		}
		plats, err := d.Platforms.ListEnabled()
		if err != nil {
			c.Error(err)
			return
		}
		priceByVariantPlatform := make(map[string]map[string]domain.PlatformPrice, len(vIDs))
		for _, pp := range prices {
			m := priceByVariantPlatform[pp.VariantID]
			if m == nil {
				m = make(map[string]domain.PlatformPrice)
				priceByVariantPlatform[pp.VariantID] = m
			}
			m[pp.PlatformID] = pp
		}
		catNames, err := d.Categories.NamesByIDs(catIDs)
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]dto.ProductPrice, 0, len(prods))
		for _, pr := range prods {
			v := defaults[pr.ID]
			items := make([]dto.PlatformPriceItem, 0, len(plats))
			for _, pl := range plats {
				it := dto.PlatformPriceItem{PlatformID: pl.ID, PlatformCode: pl.Code, PlatformName: pl.DisplayName}
				if pp, okp := priceByVariantPlatform[v.ID][pl.ID]; okp {
					it.PricePaise = pp.PricePaise
					if pp.MRPPaise != nil {
						it.MRPPaise = *pp.MRPPaise
					}
					it.Available = pp.Available
				}
				items = append(items, it)
			}
			out = append(out, dto.ProductPrice{
				ProductID:       pr.ID,
				ProductName:     pr.Name,
				ProductCategory: catNames[pr.CategoryID],
				PlatformPrices:  items,
			})
		}
		ok(c, dto.LivePrice{Meta: query.BuildMeta(total, p), Products: out})
	}
}

// @Summary Get top searched products
// @Description Products ranked by how often they were the top search result.
// @Tags dashboard
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} dto.Response{data=dto.TopSearchProduct}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/dashboard/top-searched-products [get]
func getTopSearchedProducts(d DashboardDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := query.PageFromContext(c)
		rows, total, err := d.Analytics.TopSearchedProducts(p)
		if err != nil {
			c.Error(err)
			return
		}
		ids := make([]string, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ProductID)
		}
		prods, err := d.Products.FindByIDs(ids)
		if err != nil {
			c.Error(err)
			return
		}
		catIDs := make([]string, 0, len(prods))
		for _, pr := range prods {
			catIDs = append(catIDs, pr.CategoryID)
		}
		catNames, err := d.Categories.NamesByIDs(catIDs)
		if err != nil {
			c.Error(err)
			return
		}
		items := make([]dto.TopSearchProductItem, 0, len(rows))
		for _, r := range rows {
			pr := prods[r.ProductID]
			items = append(items, dto.TopSearchProductItem{
				ProductID:       r.ProductID,
				ProductName:     pr.Name,
				ProductCategory: catNames[pr.CategoryID],
				SearchCount:     r.SearchCount,
			})
		}
		ok(c, dto.TopSearchProduct{Products: items, Meta: query.BuildMeta(total, p)})
	}
}

// @Summary Get mobile dashboard data
// @Description STUB — out of scope for the admin dashboard slice; overlaps /v1/home.
// @Tags dashboard
// @Produce json
// @Success 200 {object} dto.Response{data=dto.DashboardMobile}
// @Security BearerAuth
// @Router /v1/dashboard/mobile [get]
func getDashboardMobile() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.DashboardMobile{},
			Message: "Dashboard data fetched successfully",
			Status:  "success",
		})
	}
}
