package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterBannerRoutes(jwt *auth.JWTService, users *repository.UserRepository, r *gin.Engine) {
	group := r.Group("/v1/banners")
	group.Use(middleware.AuthMiddleware(jwt, users))
	group.GET("", listBanners())
	adminGroup := group.Group("")
	adminGroup.Use(middleware.RequireRole("admin"))
	adminGroup.POST("", CreateBanner())
	adminGroup.PATCH("", UpdateBanner())
	adminGroup.DELETE("", DeleteBanner())
}

// @Swagger:route GET /v1/banners banners listBanners
// @Summary Get banners
// @Description Fetches a paginated list of banners.
// @Tags banners
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param limit query int true "Number of items per page"
// @Success 200 {object} dto.Response{data=dto.Banners}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/banners [get]
func listBanners() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.Banners{},
			Message: "Banners fetched successfully",
			Status:  "success",
		})
	}
}

// @Swagger:route POST /v1/banners banners createBanner
// @Summary Create a new banner
// @Description Creates a new banner. This endpoint is intended for internal use and should be secured appropriately.
// @Tags banners
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file for the banner"
// @Param start_date formData string true "Start date of the banner in YYYY-MM-DD format"
// @Param end_date formData string true "End date of the banner in YYYY-MM-DD format"
// @Param is_active formData bool false "Whether the banner is currently enabled" default(true)
// @Success 201 {object} dto.Response{data=dto.BannerItem}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/banners [post]
func CreateBanner() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.BannerItem{},
			Message: "Banner created successfully",
			Status:  "success",
		})
	}
}

// @Swagger:route PATCH /v1/banners banners updateBanner
// @Summary Update an existing banner (id supplied via form data)
// @Description Updates an existing banner. This endpoint is intended for internal use and should be secured appropriately.
// @Tags banners
// @Accept multipart/form-data
// @Produce json
// @Param banner_id formData string true "Unique identifier for the banner"
// @Param image formData file false "Image file for the banner"
// @Param start_date formData string false "Start date of the banner in YYYY-MM-DD format"
// @Param end_date formData string false "End date of the banner in YYYY-MM-DD format"
// @Param is_active formData bool false "Whether the banner is currently enabled"
// @Success 200 {object} dto.Response{data=dto.BannerItem}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/banners [patch]
func UpdateBanner() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.BannerItem{},
			Message: "Banner updated successfully",
			Status:  "success",
		})
	}
}

type DeleteBannerRequest struct {
	BannerID string `json:"banner_id" binding:"required"`
}

// @Swagger:route DELETE /v1/banners banners deleteBanner
// @Summary Delete an existing banner
// @Description Deletes an existing banner. This endpoint is intended for internal use and should be secured appropriately.
// @Tags banners
// @Accept json
// @Produce json
// @Param DeleteBannerRequest body DeleteBannerRequest true "Delete Banner Request"
// @Success 200 {object} dto.Response{data=string}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/banners [delete]
func DeleteBanner() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Message: "Banner deleted successfully",
			Status:  "success",
		})
	}
}
