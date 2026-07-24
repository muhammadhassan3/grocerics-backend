package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/query"
	"grocerics-backend/internal/repository"
	"grocerics-backend/internal/service"
	"grocerics-backend/internal/util"

	"github.com/gin-gonic/gin"
)

// ConsumerDeps bundles what the consumer read/write endpoints need.
type ConsumerDeps struct {
	JWT       *auth.JWTService
	Auth      *middleware.AuthDeps
	Users     *repository.UserRepository
	Cities    *repository.CityRepository
	Catalog   *service.CatalogService
	Cart      *service.CartService
	Loc       *service.LocationResolver
	Analytics *repository.AnalyticsRepository
}

// RegisterConsumerRoutes wires the mobile-app read/write endpoints. All are
// authed; every price read is served from stored data (no live API calls).
func RegisterConsumerRoutes(r *gin.Engine, d ConsumerDeps) {
	g := r.Group("/v1")
	g.Use(middleware.AuthMiddleware(d.Auth))
	g.Use(middleware.ClientOnly())
	g.Use(middleware.ActivityTracker(d.Analytics))

	g.GET("/cities", listCities(d))
	g.GET("/home", getHome(d))
	g.GET("/categories/:id/products", getCategoryProducts(d))
	g.GET("/subcategories/:subcategory_id/products", getSubcategoryProducts(d))
	g.GET("/brands/:brand_id/products", getBrandProducts(d))
	g.GET("/search/variants", searchVariants(d))
	g.GET("/deals", getDeals(d))
	g.GET("/products/:id", getProduct(d))

	g.GET("/cart", getCart(d))
	g.POST("/cart/items", addCartItem(d))
	g.PATCH("/cart/items/:id", updateCartItem(d))
	g.DELETE("/cart/items/:id", removeCartItem(d))
	g.DELETE("/cart", clearCart(d))

	g.GET("/wishlist", getWishlist(d))
	g.POST("/wishlist", addWishlist(d))
	g.DELETE("/wishlist/:variantId", removeWishlist(d))
}

func ok(c *gin.Context, data any) { c.JSON(200, dto.Response{Status: "success", Data: data}) }

// resolveCity returns (cityID, pincode) for the current user, or writes an
// error and returns ok=false.
func resolveCity(c *gin.Context, d ConsumerDeps) (string, string, bool) {
	u := auth.MustUser(c)
	cityID, pincode, err := d.Loc.Resolve(u.ID)
	if err != nil {
		c.Error(errs.Internal("LOCATION_RESOLVE_FAILED", err))
		return "", "", false
	}
	if cityID == "" {
		c.Error(errs.BadRequest("NO_CITY", "no serviceable city configured"))
		return "", "", false
	}
	return cityID, pincode, true
}

// @Summary List serviceable cities
// @Description Cities the app serves — used by the location picker.
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]dto.CityDTO}
// @Failure 401 {object} dto.Response
// @Router /v1/cities [get]
func listCities(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		cities, err := d.Cities.ListEnabled()
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]dto.CityDTO, 0, len(cities))
		for _, ci := range cities {
			out = append(out, dto.CityDTO{ID: ci.ID, Name: ci.Name, Slug: ci.Slug})
		}
		ok(c, out)
	}
}

// @Summary Home screen
// @Description Banners, top stores (platforms), trending categories and trending items for the user's city.
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=dto.HomeResponse}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Router /v1/home [get]
func getHome(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		cityID, _, good := resolveCity(c, d)
		if !good {
			return
		}
		home, err := d.Catalog.Home(auth.MustUser(c).ID, cityID)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, home)
	}
}

// @Summary Variants in a category (PLP)
// @Description Paginated grid of variant cards for a category in the user's city. Variant-first — one card per pack. Optional ?platforms= filters which reference prices are shown.
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param platforms query string false "comma-separated platform codes; omitted = all enabled"
// @Param page query int false "page number (default 1)"
// @Param page_size query int false "page size, max 100 (default 20)"
// @Success 200 {object} dto.Response{data=dto.VariantSearchListDTO}
// @Failure 401 {object} dto.Response
// @Router /v1/categories/{id}/products [get]
func getCategoryProducts(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		cityID, _, good := resolveCity(c, d)
		if !good {
			return
		}
		items, meta, err := d.Catalog.ProductsByCategory(auth.MustUser(c).ID, c.Param("id"), cityID, util.SplitCSV(c.Query("platforms")), query.PageFromContext(c))
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, dto.VariantSearchListDTO{Items: items, Meta: meta})
	}
}

// @Summary Variants in a subcategory (PLP)
// @Description Paginated grid of variant cards for a subcategory in the user's city. Variant-first. Optional ?platforms= filters which reference prices are shown.
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Param subcategory_id path string true "Subcategory ID"
// @Param platforms query string false "comma-separated platform codes; omitted = all enabled"
// @Param page query int false "page number (default 1)"
// @Param page_size query int false "page size, max 100 (default 20)"
// @Success 200 {object} dto.Response{data=dto.VariantSearchListDTO}
// @Failure 401 {object} dto.Response
// @Router /v1/subcategories/{subcategory_id}/products [get]
func getSubcategoryProducts(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		cityID, _, good := resolveCity(c, d)
		if !good {
			return
		}
		items, meta, err := d.Catalog.ProductsBySubcategory(auth.MustUser(c).ID, c.Param("subcategory_id"), cityID, util.SplitCSV(c.Query("platforms")), query.PageFromContext(c))
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, dto.VariantSearchListDTO{Items: items, Meta: meta})
	}
}

// @Summary Variants of a brand (PLP)
// @Description Paginated grid of variant cards for all products of a brand in the user's city. Variant-first — one card per pack. Optional ?platforms= filters which reference prices are shown.
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Param brand_id path string true "Brand ID"
// @Param platforms query string false "comma-separated platform codes; omitted = all enabled"
// @Param page query int false "page number (default 1)"
// @Param page_size query int false "page size, max 100 (default 20)"
// @Success 200 {object} dto.Response{data=dto.VariantSearchListDTO}
// @Failure 401 {object} dto.Response
// @Router /v1/brands/{brand_id}/products [get]
func getBrandProducts(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		cityID, _, good := resolveCity(c, d)
		if !good {
			return
		}
		items, meta, err := d.Catalog.ProductsByBrand(auth.MustUser(c).ID, c.Param("brand_id"), cityID, util.SplitCSV(c.Query("platforms")), query.PageFromContext(c))
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, dto.VariantSearchListDTO{Items: items, Meta: meta})
	}
}

// @Summary Variant search (reference prices)
// @Description Variant-level search: matches products by name/brand, flattens to one row per variant, attaches stored REFERENCE prices for the user's city filtered to the selected platforms. Zero QuickCommerce calls. City defaults to the user's current city; pass ?city_id= to override.
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Param q query string true "search term (min 2 chars)"
// @Param platforms query string false "comma-separated platform codes; omitted = all enabled"
// @Param city_id query string false "override city (defaults to the user's current city)"
// @Param page query int false "page number (default 1)"
// @Param page_size query int false "page size, max 100 (default 20)"
// @Success 200 {object} dto.Response{data=dto.VariantSearchListDTO}
// @Failure 400 {object} dto.Response
// @Router /v1/search/variants [get]
func searchVariants(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		term := c.Query("q")
		if len([]rune(term)) < 2 {
			c.Error(errs.BadRequest("VALIDATION", "search query must be at least 2 characters"))
			return
		}
		cityID := c.Query("city_id")
		if cityID == "" {
			var good bool
			if cityID, _, good = resolveCity(c, d); !good {
				return
			}
		}
		codes := util.SplitCSV(c.Query("platforms"))
		items, meta, err := d.Catalog.SearchVariants(auth.MustUser(c).ID, term, cityID, codes, query.PageFromContext(c))
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, dto.VariantSearchListDTO{Items: items, Meta: meta})
	}
}

// @Summary Top Deals
// @Description Variants with a discounted platform price (mrp > price) in the user's city, as variant cards. Optional ?platforms= filters which reference prices are shown.
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Param platforms query string false "comma-separated platform codes; omitted = all enabled"
// @Success 200 {object} dto.Response{data=[]dto.VariantSearchItemDTO}
// @Router /v1/deals [get]
func getDeals(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		cityID, _, good := resolveCity(c, d)
		if !good {
			return
		}
		items, err := d.Catalog.Deals(auth.MustUser(c).ID, cityID, util.SplitCSV(c.Query("platforms")))
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, items)
	}
}

// @Summary Product detail (PDP)
// @Description Variants with per-platform prices, average price, unit price, and similar products.
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} dto.Response{data=dto.ProductDetailDTO}
// @Failure 404 {object} dto.Response
// @Router /v1/products/{id} [get]
func getProduct(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		cityID, _, good := resolveCity(c, d)
		if !good {
			return
		}
		detail, err := d.Catalog.ProductDetail(auth.MustUser(c).ID, c.Param("id"), cityID)
		if err != nil {
			c.Error(err)
			return
		}
		if detail == nil {
			c.Error(errs.NotFound("PRODUCT_NOT_FOUND", "product not found"))
			return
		}
		ok(c, detail)
	}
}

// @Summary Get cart with per-platform breakdown
// @Description The cart plus, for every enabled platform, the available/not-available split and totals.
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=dto.CartResponse}
// @Router /v1/cart [get]
func getCart(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		cityID, pincode, good := resolveCity(c, d)
		if !good {
			return
		}
		resp, err := d.Cart.GetCart(auth.MustUser(c).ID, cityID, pincode)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, resp)
	}
}

type addCartItemRequest struct {
	VariantID string `json:"variant_id" binding:"required"`
	Quantity  int    `json:"quantity"`
}

// @Summary Add an item to the cart
// @Description Add a variant (upserts quantity if already present). Returns the updated cart.
// @Tags consumer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body addCartItemRequest true "variant + quantity"
// @Success 200 {object} dto.Response{data=dto.CartResponse}
// @Failure 400 {object} dto.Response
// @Router /v1/cart/items [post]
func addCartItem(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req addCartItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		cityID, pincode, good := resolveCity(c, d)
		if !good {
			return
		}
		uid := auth.MustUser(c).ID
		if err := d.Cart.AddItem(uid, req.VariantID, req.Quantity); err != nil {
			c.Error(err)
			return
		}
		resp, err := d.Cart.GetCart(uid, cityID, pincode)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, resp)
	}
}

type quantityRequest struct {
	Quantity int `json:"quantity" binding:"required"`
}

// @Summary Update a cart item's quantity
// @Tags consumer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cart item ID"
// @Param request body quantityRequest true "new quantity"
// @Success 200 {object} dto.Response{data=dto.CartResponse}
// @Failure 400 {object} dto.Response
// @Router /v1/cart/items/{id} [patch]
func updateCartItem(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req quantityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Cart.UpdateItem(c.Param("id"), req.Quantity); err != nil {
			c.Error(err)
			return
		}
		cityID, pincode, good := resolveCity(c, d)
		if !good {
			return
		}
		resp, err := d.Cart.GetCart(auth.MustUser(c).ID, cityID, pincode)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, resp)
	}
}

// @Summary Remove a cart item
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cart item ID"
// @Success 200 {object} dto.Response{data=dto.CartResponse}
// @Router /v1/cart/items/{id} [delete]
func removeCartItem(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := d.Cart.RemoveItem(c.Param("id")); err != nil {
			c.Error(err)
			return
		}
		cityID, pincode, good := resolveCity(c, d)
		if !good {
			return
		}
		resp, err := d.Cart.GetCart(auth.MustUser(c).ID, cityID, pincode)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, resp)
	}
}

// @Summary Clear the cart
// @Description Remove every item from the user's cart. Returns the emptied cart (same shape as GET /v1/cart).
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=dto.CartResponse}
// @Router /v1/cart [delete]
func clearCart(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		cityID, pincode, good := resolveCity(c, d)
		if !good {
			return
		}
		uid := auth.MustUser(c).ID
		if err := d.Cart.ClearCart(uid); err != nil {
			c.Error(err)
			return
		}
		resp, err := d.Cart.GetCart(uid, cityID, pincode)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, resp)
	}
}

// @Summary Get wishlist with per-platform breakdown
// @Description Same per-platform available/not-available shape as the cart.
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=dto.CartResponse}
// @Router /v1/wishlist [get]
func getWishlist(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		cityID, pincode, good := resolveCity(c, d)
		if !good {
			return
		}
		resp, err := d.Cart.GetWishlist(auth.MustUser(c).ID, cityID, pincode)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, resp)
	}
}

type wishlistRequest struct {
	VariantID string `json:"variant_id" binding:"required"`
}

// @Summary Add a variant to the wishlist
// @Tags consumer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body wishlistRequest true "variant"
// @Success 200 {object} dto.Response{data=dto.CartResponse}
// @Failure 400 {object} dto.Response
// @Router /v1/wishlist [post]
func addWishlist(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req wishlistRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		uid := auth.MustUser(c).ID
		if err := d.Cart.AddWishlist(uid, req.VariantID); err != nil {
			c.Error(err)
			return
		}
		cityID, pincode, good := resolveCity(c, d)
		if !good {
			return
		}
		resp, err := d.Cart.GetWishlist(uid, cityID, pincode)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, resp)
	}
}

// @Summary Remove a variant from the wishlist
// @Tags consumer
// @Produce json
// @Security BearerAuth
// @Param variantId path string true "Variant ID"
// @Success 200 {object} dto.Response{data=dto.CartResponse}
// @Router /v1/wishlist/{variantId} [delete]
func removeWishlist(d ConsumerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := auth.MustUser(c).ID
		if err := d.Cart.RemoveWishlist(uid, c.Param("variantId")); err != nil {
			c.Error(err)
			return
		}
		cityID, pincode, good := resolveCity(c, d)
		if !good {
			return
		}
		resp, err := d.Cart.GetWishlist(uid, cityID, pincode)
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, resp)
	}
}
