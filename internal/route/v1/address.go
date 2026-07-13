package v1

import (
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/util"

	"github.com/gin-gonic/gin"
)

func RegisterAddressRoutes(r *gin.Engine) {
	group := r.Group("/v1/address")
	group.POST("", CreateAddress())
}

// @Swagger:route POST /v1/address address CreateAddress
// @Summary Create a new address
// @Description Creates a new address for the authenticated user.
// @Tags address
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param address body dto.AddAddressRequest true "Address details"
// @Success 201 {object} dto.Response{data=dto.AddressResponse}
// @Failure 400 {object} dto.Response{data=string}
// @Failure 401 {object} dto.Response{data=string}
// @Failure 403 {object} dto.Response{data=string}
// @Router /v1/address [post]
func CreateAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.AddAddressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
		}

		c.JSON(201, dto.Response{
			Data:    nil,
			Message: "Address created successfully",
			Status:  "success",
		})
	}
}
