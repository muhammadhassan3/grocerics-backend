package v1

import (
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

type CityDeps struct {
	JWT    *auth.JWTService
	Auth   *middleware.AuthDeps
	Users  *repository.UserRepository
	Cities *repository.CityRepository
}

func toCityDTO(c domain.City) dto.CityItem {
	return dto.CityItem{
		CityID:         c.ID,
		Name:           c.Name,
		Slug:           c.Slug,
		State:          util.Deref(c.State),
		Lat:            c.Lat,
		Lng:            c.Lng,
		DefaultPincode: util.Deref(c.DefaultPincode),
		Enabled:        c.Enabled,
	}
}

func RegisterCityRoutes(r *gin.Engine, d CityDeps) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(d.Auth))

	admin := group.Group("")
	admin.Use(middleware.AdminOnly())
	admin.GET("/cities/all", listCitiesAdmin(d))
	admin.GET("/cities/:id", getCityByID(d))
	admin.POST("/cities", createCity(d))
	admin.PATCH("/cities", updateCity(d))
	admin.DELETE("/cities", deleteCity(d))
}

// @Summary List all cities (admin)
// @Description Paginated list of serviceable cities, including disabled.
// @Tags cities
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param search query string false "Filter by name"
// @Success 200 {object} dto.Response{data=dto.Cities}
// @Security BearerAuth
// @Router /v1/cities/all [get]
func listCitiesAdmin(d CityDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := query.PageFromContext(c)
		items, total, err := d.Cities.ListAdmin(p, c.Query("search"))
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]dto.CityItem, len(items))
		for i, it := range items {
			out[i] = toCityDTO(it)
		}
		ok(c, dto.Cities{Meta: query.BuildMeta(total, p), Cities: out})
	}
}

// @Summary Get a city by ID (admin)
// @Tags cities
// @Produce json
// @Param id path string true "City ID"
// @Success 200 {object} dto.Response{data=dto.CityItem}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/cities/{id} [get]
func getCityByID(d CityDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		city, err := d.Cities.FindByID(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}
		if city == nil {
			c.Error(errs.NotFound("CITY_NOT_FOUND", "city not found"))
			return
		}
		ok(c, toCityDTO(*city))
	}
}

type CreateCityRequest struct {
	Name           string   `json:"name" binding:"required"`
	State          string   `json:"state"`
	Lat            *float64 `json:"lat" binding:"required"`
	Lng            *float64 `json:"lng" binding:"required"`
	DefaultPincode string   `json:"default_pincode" binding:"required"`
	Enabled        *bool    `json:"enabled"`
}

// @Summary Create a serviceable city
// @Tags cities
// @Accept json
// @Produce json
// @Param city body CreateCityRequest true "Create City Request"
// @Success 201 {object} dto.Response{data=dto.CityItem}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/cities [post]
func createCity(d CityDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateCityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		enabled := true
		if req.Enabled != nil {
			enabled = *req.Enabled
		}
		created, err := d.Cities.Create(&domain.City{
			Name:           req.Name,
			Slug:           util.Slugify(req.Name),
			State:          util.PtrIfSet(req.State),
			Lat:            req.Lat,
			Lng:            req.Lng,
			DefaultPincode: util.PtrIfSet(req.DefaultPincode),
			Enabled:        enabled,
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(201, dto.Response{Status: "success", Data: toCityDTO(*created), Message: "City created successfully"})
	}
}

type UpdateCityRequest struct {
	CityID         string   `json:"city_id" binding:"required"`
	Name           string   `json:"name"`
	State          string   `json:"state"`
	Lat            *float64 `json:"lat"`
	Lng            *float64 `json:"lng"`
	DefaultPincode string   `json:"default_pincode"`
	Enabled        *bool    `json:"enabled"`
}

// @Summary Update a city
// @Tags cities
// @Accept json
// @Produce json
// @Param city body UpdateCityRequest true "Update City Request"
// @Success 200 {object} dto.Response{data=dto.CityItem}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/cities [patch]
func updateCity(d CityDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateCityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		fields := map[string]any{}
		if req.Name != "" {
			fields["name"] = req.Name
			fields["slug"] = util.Slugify(req.Name)
		}
		if req.State != "" {
			fields["state"] = req.State
		}
		if req.Lat != nil {
			fields["lat"] = *req.Lat
		}
		if req.Lng != nil {
			fields["lng"] = *req.Lng
		}
		if req.DefaultPincode != "" {
			fields["default_pincode"] = req.DefaultPincode
		}
		if req.Enabled != nil {
			fields["enabled"] = *req.Enabled
		}
		updated, err := d.Cities.Update(req.CityID, fields)
		if err != nil {
			c.Error(err)
			return
		}
		if updated == nil {
			c.Error(errs.NotFound("CITY_NOT_FOUND", "city not found"))
			return
		}
		ok(c, toCityDTO(*updated))
	}
}

type DeleteCityRequest struct {
	CityID string `json:"city_id" binding:"required"`
}

// @Summary Delete a city
// @Tags cities
// @Accept json
// @Produce json
// @Param DeleteCityRequest body DeleteCityRequest true "Delete City Request"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/cities [delete]
func deleteCity(d CityDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeleteCityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Cities.SoftDelete(req.CityID, auth.MustUser(c).ID); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "City deleted successfully"})
	}
}
