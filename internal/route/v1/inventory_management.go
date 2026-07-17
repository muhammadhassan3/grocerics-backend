package v1

import (
	"time"

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

type InventoryDeps struct {
	JWT           *auth.JWTService
	Users         *repository.UserRepository
	Products      *repository.ProductRepository
	Variants      *repository.ProductVariantRepository
	Categories    *repository.CategoryRepository
	Subcategories *repository.SubcategoryRepository
	Brands        *repository.BrandRepository
}

func RegisterInventoryManagementRoutes(r *gin.Engine, d InventoryDeps) {
	group := r.Group("/v1")
	group.Use(middleware.AuthMiddleware(d.JWT, d.Users))
	group.Use(middleware.RequireRole(domain.RoleAdmin))

	group.GET("/inventory-management", listInventory(d))
	group.GET("/inventory-management/stats", inventoryStats(d))
	group.GET("/inventory-management/:product_id", getInventoryItem(d))
	group.POST("/inventory-management", createItem(d))
	group.PATCH("/inventory-management", updateItem(d))
	group.PATCH("/inventory-management/reorder", reorderProducts(d))
	group.DELETE("/inventory-management", deleteItem(d))

	group.GET("/inventory-management/:product_id/variants", listVariants(d))
	group.GET("/inventory-management/:product_id/variants/:variant_id", getVariant(d))
	group.POST("/inventory-management/variants", createVariant(d))
	group.PATCH("/inventory-management/variants", updateVariant(d))
	group.PATCH("/inventory-management/variants/reorder", reorderVariants(d))
	group.DELETE("/inventory-management/variants", deleteVariant(d))
}

// @Summary Reorder products
// @Description Sets display_order from the given order (drag-to-reorder). Send product ids in the desired order.
// @Tags inventory-management
// @Accept json
// @Produce json
// @Param request body ReorderRequest true "Ordered product IDs"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/reorder [patch]
func reorderProducts(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReorderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Products.Reorder(req.IDs); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Products reordered"})
	}
}

// @Summary Reorder a product's variants
// @Description Sets display_order from the given order (drag-to-reorder). Send variant ids in the desired order.
// @Tags inventory-management
// @Accept json
// @Produce json
// @Param request body ReorderRequest true "Ordered variant IDs"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/variants/reorder [patch]
func reorderVariants(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReorderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Variants.Reorder(req.IDs); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Variants reordered"})
	}
}

func toVariantItemDTO(v domain.ProductVariant) dto.ProductVariantItem {
	return dto.ProductVariantItem{
		ProductID:        v.ProductID,
		ProductVariantID: v.ID,
		VariantCustomID:  util.Deref(v.CustomVariantID),
		ProductVolume:    dto.ProductVariantUnit{Value: v.VolumeValue, Unit: string(v.VolumeUnit)},
	}
}

// @Summary Inventory list
// @Tags inventory-management
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param search query string false "Filter by product name"
// @Success 200 {object} dto.Response{data=dto.InventoryManagements}
// @Security BearerAuth
// @Router /v1/inventory-management [get]
func listInventory(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := query.PageFromContext(c)
		items, total, err := d.Products.ListAdmin(p, c.Query("search"))
		if err != nil {
			c.Error(err)
			return
		}
		ids := make([]string, len(items))
		catIDs := make([]string, 0, len(items))
		subIDs := make([]string, 0, len(items))
		for i, it := range items {
			ids[i] = it.ID
			catIDs = append(catIDs, it.CategoryID)
			if it.SubcategoryID != nil {
				subIDs = append(subIDs, *it.SubcategoryID)
			}
		}
		catNames, err := d.Categories.NamesByIDs(catIDs)
		if err != nil {
			c.Error(err)
			return
		}
		subNames, err := d.Subcategories.NamesByIDs(subIDs)
		if err != nil {
			c.Error(err)
			return
		}
		varCounts, err := d.Products.CountVariants(ids)
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]dto.InventoryManagementsItem, len(items))
		for i, it := range items {
			subID := util.Deref(it.SubcategoryID)
			out[i] = dto.InventoryManagementsItem{
				ProductID:          it.ID,
				ProductName:        it.Name,
				ImageURL:           util.Deref(it.ImageURL),
				ProductCategory:    catNames[it.CategoryID],
				ProductSubCategory: subNames[subID],
				SubcategoryID:      subID,
				TotalVariants:      varCounts[it.ID],
				Status:             string(it.Status),
				TopItem:            it.IsTopItem,
				StockCount:         0, // QC-owned; filled by refresh (Slice C)
			}
		}
		ok(c, dto.InventoryManagements{Meta: query.BuildMeta(total, p), Products: out})
	}
}

// @Summary Inventory headline stats
// @Tags inventory-management
// @Produce json
// @Success 200 {object} dto.Response{data=dto.InventoryManagementsStats}
// @Security BearerAuth
// @Router /v1/inventory-management/stats [get]
func inventoryStats(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		s, err := d.Products.Stats()
		if err != nil {
			c.Error(err)
			return
		}
		ok(c, dto.InventoryManagementsStats{
			TotalCategories: dto.StatsItem{Value: int(s.Categories)},
			TotalProducts:   dto.StatsItem{Value: int(s.Products)},
			TotalBrands:     dto.StatsItem{Value: int(s.Brands)},
			Platforms:       dto.StatsItem{Value: int(s.Platforms)},
		})
	}
}

// @Summary Get inventory item
// @Tags inventory-management
// @Produce json
// @Param product_id path string true "Product ID"
// @Success 200 {object} dto.Response{data=dto.ProductItem}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/{product_id} [get]
func getInventoryItem(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		prod, err := d.Products.FindByID(c.Param("product_id"))
		if err != nil {
			c.Error(err)
			return
		}
		if prod == nil {
			c.Error(errs.NotFound("PRODUCT_NOT_FOUND", "product not found"))
			return
		}
		catNames, _ := d.Categories.NamesByIDs([]string{prod.CategoryID})
		var brandName string
		if prod.BrandID != nil {
			if bm, _ := d.Brands.FindByIDs([]string{*prod.BrandID}); bm != nil {
				brandName = bm[*prod.BrandID].Name
			}
		}
		subID := util.Deref(prod.SubcategoryID)
		var subName string
		if subID != "" {
			if sn, _ := d.Subcategories.NamesByIDs([]string{subID}); sn != nil {
				subName = sn[subID]
			}
		}
		ok(c, dto.ProductItem{
			ProductID:          prod.ID,
			ImageURL:           util.Deref(prod.ImageURL),
			ProductName:        prod.Name,
			ProductDescription: util.Deref(prod.Description),
			ProductCategory:    dto.ProductCategory{ProductCategoryID: prod.CategoryID, ProductCategoryName: catNames[prod.CategoryID]},
			ProductSubCategory: dto.ProductSubCategory{ProductSubCategoryID: subID, ProductSubCategoryName: subName},
			ProductBrand:       dto.Brand{ProductBrandID: util.Deref(prod.BrandID), ProductBrandName: brandName},
			CategoryID:         prod.CategoryID,
			SubcategoryID:      subID,
			BrandID:            util.Deref(prod.BrandID),
			IsTopItem:          prod.IsTopItem,
			Status:             string(prod.Status),
			CreatedAt:          prod.CreatedAt.Format(time.RFC3339),
		})
	}
}

type CreateNewItemRequest struct {
	ImageURL           string `json:"image_url" binding:"required"`
	ProductName        string `json:"product_name" binding:"required"`
	ProductDescription string `json:"product_description"`
	BrandID            string `json:"brand_id" binding:"required"`
	CategoryID         string `json:"category_id" binding:"required"`
	SubcategoryID      string `json:"subcategory_id"`
	IsTopItem          bool   `json:"is_top_item"`
	Status             string `json:"status" binding:"required,oneof=active disabled"`
}

// @Summary Create an inventory item (product)
// @Tags inventory-management
// @Accept json
// @Produce json
// @Param item body CreateNewItemRequest true "Create Item Request"
// @Success 201 {object} dto.Response{data=dto.ProductItem}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management [post]
func createItem(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateNewItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		created, err := d.Products.Create(&domain.Product{
			CategoryID:    req.CategoryID,
			SubcategoryID: util.PtrIfSet(req.SubcategoryID),
			BrandID:       util.PtrIfSet(req.BrandID),
			Name:          req.ProductName,
			Description:   util.PtrIfSet(req.ProductDescription),
			ImageURL:      util.PtrIfSet(req.ImageURL),
			IsTopItem:     req.IsTopItem,
			Status:        domain.Status(req.Status),
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(201, dto.Response{Status: "success", Data: gin.H{"product_id": created.ID}, Message: "Product created successfully"})
	}
}

type UpdateItemRequest struct {
	ProductID          string `json:"product_id" binding:"required"`
	ImageURL           string `json:"image_url"`
	ProductName        string `json:"product_name"`
	ProductDescription string `json:"product_description"`
	BrandID            string `json:"brand_id"`
	CategoryID         string `json:"category_id"`
	SubcategoryID      string `json:"subcategory_id"`
	IsTopItem          *bool  `json:"is_top_item"`
	Status             string `json:"status" binding:"omitempty,oneof=active disabled"`
}

// @Summary Update an inventory item
// @Tags inventory-management
// @Accept json
// @Produce json
// @Param item body UpdateItemRequest true "Update Item Request"
// @Success 200 {object} dto.Response{data=string}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management [patch]
func updateItem(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		fields := map[string]any{}
		if req.ImageURL != "" {
			fields["image_url"] = req.ImageURL
		}
		if req.ProductName != "" {
			fields["name"] = req.ProductName
		}
		if req.ProductDescription != "" {
			fields["description"] = req.ProductDescription
		}
		if req.BrandID != "" {
			fields["brand_id"] = req.BrandID
		}
		if req.CategoryID != "" {
			fields["category_id"] = req.CategoryID
		}
		if req.SubcategoryID != "" {
			fields["subcategory_id"] = req.SubcategoryID
		}
		if req.IsTopItem != nil {
			fields["is_top_item"] = *req.IsTopItem
		}
		if req.Status != "" {
			fields["status"] = req.Status
		}
		updated, err := d.Products.Update(req.ProductID, fields)
		if err != nil {
			c.Error(err)
			return
		}
		if updated == nil {
			c.Error(errs.NotFound("PRODUCT_NOT_FOUND", "product not found"))
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Product updated successfully"})
	}
}

type DeleteItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
}

// @Summary Delete an inventory item
// @Tags inventory-management
// @Accept json
// @Produce json
// @Param DeleteItemRequest body DeleteItemRequest true "Delete Item Request"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management [delete]
func deleteItem(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeleteItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Products.SoftDelete(req.ProductID, auth.MustUser(c).ID); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Product deleted successfully"})
	}
}

// @Summary List a product's variants
// @Tags inventory-management
// @Produce json
// @Param product_id path string true "Product ID"
// @Success 200 {object} dto.Response{data=dto.ProductVariantItems}
// @Security BearerAuth
// @Router /v1/inventory-management/{product_id}/variants [get]
func listVariants(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		vs, err := d.Variants.ListByProduct(c.Param("product_id"))
		if err != nil {
			c.Error(err)
			return
		}
		out := make([]dto.ProductVariantItem, len(vs))
		for i, v := range vs {
			out[i] = toVariantItemDTO(v)
		}
		ok(c, dto.ProductVariantItems{Variants: out})
	}
}

// @Summary Get a single variant
// @Tags inventory-management
// @Produce json
// @Param product_id path string true "Product ID"
// @Param variant_id path string true "Variant ID"
// @Success 200 {object} dto.Response{data=dto.ProductVariantItem}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/{product_id}/variants/{variant_id} [get]
func getVariant(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, err := d.Variants.FindByID(c.Param("variant_id"))
		if err != nil {
			c.Error(err)
			return
		}
		if v == nil {
			c.Error(errs.NotFound("VARIANT_NOT_FOUND", "variant not found"))
			return
		}
		ok(c, toVariantItemDTO(*v))
	}
}

type CreateVariantRequest struct {
	ProductID              string                 `json:"product_id" binding:"required"`
	Volume                 dto.ProductVariantUnit `json:"volume" binding:"required"`
	CustomProductVariantID string                 `json:"custom_product_variant_id"`
}

// @Summary Create a variant (pack size)
// @Description Creates a pack-size variant. Per-platform linking + price is set separately via the linking endpoints.
// @Tags inventory-management
// @Accept json
// @Produce json
// @Param CreateVariantRequest body CreateVariantRequest true "Create Variant Request"
// @Success 201 {object} dto.Response{data=dto.ProductVariantItem}
// @Failure 400 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/variants [post]
func createVariant(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateVariantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		created, err := d.Variants.Create(&domain.ProductVariant{
			ProductID:       req.ProductID,
			VolumeValue:     req.Volume.Value,
			VolumeUnit:      domain.VolumeUnit(req.Volume.Unit),
			CustomVariantID: util.PtrIfSet(req.CustomProductVariantID),
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(201, dto.Response{Status: "success", Data: toVariantItemDTO(*created), Message: "Variant created successfully"})
	}
}

type UpdateVariantRequest struct {
	ProductVariantID       string                 `json:"product_variant_id" binding:"required"`
	Volume                 dto.ProductVariantUnit `json:"volume"`
	CustomProductVariantID string                 `json:"custom_product_variant_id"`
}

// @Summary Update a variant
// @Tags inventory-management
// @Accept json
// @Produce json
// @Param UpdateVariantRequest body UpdateVariantRequest true "Update Variant Request"
// @Success 200 {object} dto.Response{data=dto.ProductVariantItem}
// @Failure 404 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/variants [patch]
func updateVariant(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateVariantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		fields := map[string]any{}
		if req.Volume.Value != 0 {
			fields["volume_value"] = req.Volume.Value
		}
		if req.Volume.Unit != "" {
			fields["volume_unit"] = req.Volume.Unit
		}
		if req.CustomProductVariantID != "" {
			fields["custom_variant_id"] = req.CustomProductVariantID
		}
		updated, err := d.Variants.Update(req.ProductVariantID, fields)
		if err != nil {
			c.Error(err)
			return
		}
		if updated == nil {
			c.Error(errs.NotFound("VARIANT_NOT_FOUND", "variant not found"))
			return
		}
		ok(c, toVariantItemDTO(*updated))
	}
}

type DeleteVariantRequest struct {
	ProductVariantID string `json:"product_variant_id" binding:"required"`
}

// @Summary Delete a variant
// @Tags inventory-management
// @Accept json
// @Produce json
// @Param DeleteVariantRequest body DeleteVariantRequest true "Delete Variant Request"
// @Success 200 {object} dto.Response{data=string}
// @Security BearerAuth
// @Router /v1/inventory-management/variants [delete]
func deleteVariant(d InventoryDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeleteVariantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := d.Variants.SoftDelete(req.ProductVariantID, auth.MustUser(c).ID); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Variant deleted successfully"})
	}
}
