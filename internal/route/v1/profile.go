package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/repository"
	"grocerics-backend/internal/service"
	"grocerics-backend/internal/util"

	"github.com/gin-gonic/gin"
)

type ProfileDeps struct {
	JWT     *auth.JWTService
	Auth    *middleware.AuthDeps
	Users   *repository.UserRepository
	Profile *service.ProfileService
}

func RegisterProfileRoutes(r *gin.Engine, d ProfileDeps) {
	g := r.Group("/v1/me")
	g.Use(middleware.AuthMiddleware(d.Auth))
	g.Use(middleware.ClientOnly())

	g.GET("", getMe(d))
	g.PATCH("", updateMe(d))

	g.GET("/addresses", listAddresses(d))
	g.POST("/addresses", createAddress(d))
	g.PATCH("/addresses/:id", updateAddress(d))
	g.DELETE("/addresses/:id", deleteAddress(d))

	g.GET("/notification-preferences", getNotificationPreferences(d))
	g.PUT("/notification-preferences", updateNotificationPreferences(d))

	g.POST("/fcm-token", registerFcmToken(d))
	g.DELETE("/fcm-token", removeFcmToken(d))
}

// @Summary Get my profile
// @Tags profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=dto.MeDTO}
// @Router /v1/me [get]
func getMe(d ProfileDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		me, err := d.Profile.GetMe(auth.MustUser(c).ID)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, me)
	}
}

type updateMeRequest struct {
	Name string `json:"name" binding:"required"`
}

// @Summary Update my profile
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body updateMeRequest true "profile fields"
// @Success 200 {object} dto.Response{data=dto.MeDTO}
// @Failure 400 {object} dto.Response
// @Router /v1/me [patch]
func updateMe(d ProfileDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req updateMeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		me, err := d.Profile.UpdateMe(auth.MustUser(c).ID, req.Name)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, me)
	}
}

// @Summary List my delivery addresses
// @Tags profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]dto.AddressDTO}
// @Router /v1/me/addresses [get]
func listAddresses(d ProfileDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		addrs, err := d.Profile.ListAddresses(auth.MustUser(c).ID)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, addrs)
	}
}

type addressRequest struct {
	Label     *string  `json:"label"`
	Line1     string   `json:"line1" binding:"required"`
	Line2     *string  `json:"line2"`
	Pincode   string   `json:"pincode" binding:"required"`
	Lat       *float64 `json:"lat"`
	Lng       *float64 `json:"lng"`
	IsDefault bool     `json:"is_default"`
}

func (r addressRequest) toInput() service.AddressInput {
	return service.AddressInput{
		Label: r.Label, Line1: r.Line1, Line2: r.Line2,
		Pincode: r.Pincode, Lat: r.Lat, Lng: r.Lng, IsDefault: r.IsDefault,
	}
}

// @Summary Add a delivery address
// @Description Saves an address; the pincode resolves to a serving city. Setting is_default updates the user's current city.
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body addressRequest true "address"
// @Success 200 {object} dto.Response{data=dto.AddressDTO}
// @Failure 400 {object} dto.Response
// @Router /v1/me/addresses [post]
func createAddress(d ProfileDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req addressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		a, err := d.Profile.CreateAddress(auth.MustUser(c).ID, req.toInput())
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, a)
	}
}

// @Summary Update a delivery address
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Address ID"
// @Param request body addressRequest true "address"
// @Success 200 {object} dto.Response{data=dto.AddressDTO}
// @Failure 400 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /v1/me/addresses/{id} [patch]
func updateAddress(d ProfileDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req addressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		a, err := d.Profile.UpdateAddress(auth.MustUser(c).ID, c.Param("id"), req.toInput())
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, a)
	}
}

// @Summary Delete a delivery address
// @Tags profile
// @Produce json
// @Security BearerAuth
// @Param id path string true "Address ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /v1/me/addresses/{id} [delete]
func deleteAddress(d ProfileDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := d.Profile.DeleteAddress(auth.MustUser(c).ID, c.Param("id")); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "address deleted"})
	}
}

// @Summary Get my notification preferences
// @Tags profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=dto.NotificationPreferencesDTO}
// @Router /v1/me/notification-preferences [get]
func getNotificationPreferences(d ProfileDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		prefs, err := d.Profile.GetNotificationPreferences(auth.MustUser(c).ID)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, prefs)
	}
}

// @Summary Update my notification preferences
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.NotificationPreferencesDTO true "preferences"
// @Success 200 {object} dto.Response{data=dto.NotificationPreferencesDTO}
// @Failure 400 {object} dto.Response
// @Router /v1/me/notification-preferences [put]
func updateNotificationPreferences(d ProfileDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.NotificationPreferencesDTO
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		prefs, err := d.Profile.UpdateNotificationPreferences(auth.MustUser(c).ID, req)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, prefs)
	}
}

type fcmTokenRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform"`
}

// @Summary Register a device push token
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body fcmTokenRequest true "device token"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /v1/me/fcm-token [post]
func registerFcmToken(d ProfileDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req fcmTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Profile.RegisterFcmToken(auth.MustUser(c).ID, req.Token, req.Platform); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "token registered"})
	}
}

type fcmDeleteRequest struct {
	Token string `json:"token" binding:"required"`
}

// @Summary Remove a device push token
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body fcmDeleteRequest true "device token"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /v1/me/fcm-token [delete]
func removeFcmToken(d ProfileDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req fcmDeleteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Profile.RemoveFcmToken(req.Token); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "token removed"})
	}
}
