package v1

import (
	"grocerics-backend/internal/dto"

	"github.com/gin-gonic/gin"
)

func RegisterSettingsRoutes(r *gin.Engine) {
	group := r.Group("/v1/settings")
	group.GET("/about-us", getAboutUs())
	group.GET("/terms-and-conditions", getTermsAndConditions())
}

// @Swagger:route GET /v1/settings/about-us settings getAboutUs
// @Summary Get About Us
// @Description Fetches the About Us information.
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=dto.AboutUsResponse}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/settings/about-us [get]
func getAboutUs() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.AboutUsResponse{},
			Message: "About Us fetched successfully",
			Status:  "success",
		})
	}
}

// @Swagger:route GET /v1/settings/terms-and-conditions settings getTermsAndConditions
// @Summary Get Terms and Conditions
// @Description Fetches the Terms and Conditions information.
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=dto.TermsAndConditionsResponse}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/settings/terms-and-conditions [get]
func getTermsAndConditions() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.TermsAndConditionsResponse{},
			Message: "Terms and Conditions fetched successfully",
			Status:  "success",
		})
	}
}
