package v1

import (
	"grocerics-backend/internal/dto"

	"github.com/gin-gonic/gin"
)

func RegisterTopDealsRoutes(r *gin.Engine) {
	group := r.Group("/v1/top-deals")
	group.GET("/", getTopDeals())
}

// @Swagger:route GET /v1/top-deals top-deals getTopDeals
// @Summary Get top deals
// @Description Fetches a list of top deals.
// @Tags top-deals
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=dto.TopDealsMobileResponse}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/top-deals [get]
func getTopDeals() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, dto.Response{
			Data:    dto.TopDealsMobileResponse{},
			Message: "Top deals fetched successfully",
			Status:  "success",
		})
	}
}
