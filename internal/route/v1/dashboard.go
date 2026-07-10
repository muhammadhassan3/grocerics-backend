package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterDashboardRoutes(jwt *auth.JWTService, users *repository.UserRepository, r *gin.Engine) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(jwt, users))
	group.Use(middleware.RequireRole(domain.RoleAdmin))
	group.GET("/dashboard", getDashboard())
	group.GET("/dashboard/stats", getDashboardStats())
	group.GET("/dashboard/live-price-comparison", getLivePriceComparison())
	group.GET("/dashboard/top-searched-products", getTopSearchedProducts())
}

// @Swagger:route GET /v1/dashboard dashboard getDashboard
// @Summary Get dashboard data
// @Description Fetches the data needed to populate the admin dashboard, including headline stats, daily active users, and monthly active users.
// @Tags dashboard
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=dto.DashboardResponse}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/dashboard [get]
func getDashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.DashboardResponse{},
			Message: "Dashboard data fetched successfully",
			Status:  "success",
		})
	}
}

// @Swagger:route GET /v1/dashboard/live-price-comparison dashboard getLivePriceComparison
// @Summary Get live price comparison
// @Description Fetches the live price comparison data for products across different delivery platforms.
// @Tags dashboard
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param limit query int true "Number of items per page"
// @Success 200 {object} dto.Response{data=dto.LivePrice}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/dashboard/live-price-comparison [get]
func getLivePriceComparison() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.LivePrice{},
			Message: "Live price comparison data fetched successfully",
			Status:  "success",
		})
	}
}

// @Swagger:route GET /v1/dashboard/top-searched-products dashboard getTopSearchedProducts
// @Summary Get top searched products
// @Description Fetches the top searched products data for the admin dashboard.
// @Tags dashboard
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param limit query int true "Number of items per page"
// @Success 200 {object} dto.Response{data=dto.TopSearchProduct}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/dashboard/top-searched-products [get]
func getTopSearchedProducts() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.TopSearchProduct{},
			Message: "Top searched products data fetched successfully",
			Status:  "success",
		})
	}
}

// @Swagger:route GET /v1/dashboard/stats dashboard getDashboardStats
// @Summary Get dashboard stats
// @Description Fetches the headline stat cards for the admin dashboard: total users, average basket size, and total searches.
// @Tags dashboard
// @Accept json
// @Produce json
// @Param interval query string true "Interval for stats aggregation (e.g., daily, weekly, monthly)" enums(daily,weekly,monthly) default(daily)
// @Success 200 {object} dto.Response{data=dto.DashboardStats}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/dashboard/stats [get]
func getDashboardStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.DashboardStats{},
			Message: "Dashboard stats fetched successfully",
			Status:  "success",
		})
	}
}
