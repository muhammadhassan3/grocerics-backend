package service

import (
	"strconv"
	"strings"
	"time"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/query"
	"grocerics-backend/internal/repository"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

type CatalogService struct {
	banners   *repository.BannerRepository
	platforms *repository.PlatformRepository
	category  *repository.CategoryRepository
	product   *repository.ProductRepository
	brand     *repository.BrandRepository
	variant   *repository.ProductVariantRepository
	image     *repository.ProductImageRepository
	price     *repository.PlatformPriceRepository
	link      *repository.ProductPlatformLinkRepository
	wishlist  *repository.WishlistRepository
}

func NewCatalogService(db *gorm.DB) *CatalogService {
	return &CatalogService{
		banners:   repository.NewBannerRepository(db),
		platforms: repository.NewPlatformRepository(db),
		category:  repository.NewCategoryRepository(db),
		product:   repository.NewProductRepository(db),
		brand:     repository.NewBrandRepository(db),
		variant:   repository.NewProductVariantRepository(db),
		image:     repository.NewProductImageRepository(db),
		price:     repository.NewPlatformPriceRepository(db),
		link:      repository.NewProductPlatformLinkRepository(db),
		wishlist:  repository.NewWishlistRepository(db),
	}
}

// wishlistSet returns the set of variant IDs in the user's wishlist (empty when
// userID is blank). Used to stamp in_wishlist on variant cards while browsing.
func (s *CatalogService) wishlistSet(userID string) (map[string]bool, error) {
	if userID == "" {
		return map[string]bool{}, nil
	}
	rows, err := s.wishlist.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(rows))
	for _, w := range rows {
		set[w.VariantID] = true
	}
	return set, nil
}

const homePreviewCap = 6

func capList[T any](xs []T, n int) []T {
	if len(xs) > n {
		return xs[:n]
	}
	return xs
}

func (s *CatalogService) Home(userID, cityID string) (*dto.HomeResponse, error) {
	wl, err := s.wishlistSet(userID)
	if err != nil {
		return nil, err
	}
	banners, err := s.banners.ListActive()
	if err != nil {
		return nil, err
	}
	plats, err := s.platforms.ListEnabled()
	if err != nil {
		return nil, err
	}
	cats, err := s.category.ListVisibleWithProducts(true)
	if err != nil {
		return nil, err
	}
	banners = capList(banners, homePreviewCap)
	plats = capList(plats, homePreviewCap)
	cats = capList(cats, homePreviewCap)
	top, err := s.product.ListTop(10)
	if err != nil {
		return nil, err
	}
	topVariants, err := s.defaultVariantsFor(top)
	if err != nil {
		return nil, err
	}
	cards, err := s.variantCards(topVariants, cityID, nil, wl)
	if err != nil {
		return nil, err
	}

	resp := &dto.HomeResponse{
		Banners:       []dto.BannerCardDTO{},
		TopStores:     []dto.PlatformDTO{},
		Categories:    []dto.CategoryCardDTO{},
		TrendingItems: cards,
	}
	for _, b := range banners {
		card := dto.BannerCardDTO{ImageURL: b.ImageURL, TargetType: string(b.TargetType)}
		if b.TargetID != nil {
			card.TargetID = *b.TargetID
		}
		if b.TargetURL != nil {
			card.TargetURL = *b.TargetURL
		}
		resp.Banners = append(resp.Banners, card)
	}
	for _, p := range plats {
		resp.TopStores = append(resp.TopStores, platformDTO(p))
	}
	for _, c := range cats {
		resp.Categories = append(resp.Categories, dto.CategoryCardDTO{
			ID: c.ID, Name: c.Name, Slug: c.Slug, ImageURL: strPtr(c.ImageURL),
		})
	}
	return resp, nil
}

func (s *CatalogService) ProductDetail(userID, productID, cityID string) (*dto.ProductDetailDTO, error) {
	wl, err := s.wishlistSet(userID)
	if err != nil {
		return nil, err
	}
	product, err := s.product.FindByID(productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	platMap, err := s.platformMap()
	if err != nil {
		return nil, err
	}

	out := &dto.ProductDetailDTO{
		ProductID:   product.ID,
		Name:        product.Name,
		Description: strPtr(product.Description),
		CategoryID:  product.CategoryID,
	}
	if product.BrandID != nil {
		if brands, bErr := s.brand.FindByIDs([]string{*product.BrandID}); bErr == nil {
			if b, ok := brands[*product.BrandID]; ok {
				out.BrandName = b.Name
			}
		}
	}
	imgs, err := s.image.ListByProduct(product.ID)
	if err != nil {
		return nil, err
	}
	for _, im := range imgs {
		out.Images = append(out.Images, im.ImageURL)
	}
	if len(out.Images) == 0 && product.ImageURL != nil {
		out.Images = append(out.Images, *product.ImageURL)
	}

	variants, err := s.variant.ListByProduct(product.ID)
	if err != nil {
		return nil, err
	}
	variantIDs := make([]string, 0, len(variants))
	for _, v := range variants {
		variantIDs = append(variantIDs, v.ID)
	}
	imageByVariant, err := s.link.PrimaryImagesByVariants(variantIDs)
	if err != nil {
		return nil, err
	}
	for _, v := range variants {
		vd, vErr := s.variantDetail(v, cityID, platMap, wl)
		if vErr != nil {
			return nil, vErr
		}
		vd.ImageURL = imageByVariant[v.ID]
		out.Variants = append(out.Variants, vd)
	}

	similar, err := s.product.ListSimilar(product.CategoryID, product.ID, 6)
	if err != nil {
		return nil, err
	}
	similarVariants, err := s.defaultVariantsFor(similar)
	if err != nil {
		return nil, err
	}
	if out.Similar, err = s.variantCards(similarVariants, cityID, nil, wl); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *CatalogService) defaultVariantsFor(products []domain.Product) ([]domain.ProductVariant, error) {
	ids := make([]string, 0, len(products))
	for _, p := range products {
		ids = append(ids, p.ID)
	}
	defaults, err := s.variant.DefaultsForProducts(ids)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProductVariant, 0, len(products))
	for _, p := range products {
		if v, ok := defaults[p.ID]; ok {
			out = append(out, v)
		}
	}
	return out, nil
}

func (s *CatalogService) variantDetail(v domain.ProductVariant, cityID string, platMap map[string]domain.Platform, wl map[string]bool) (dto.VariantDetailDTO, error) {
	vd := dto.VariantDetailDTO{VariantID: v.ID, PackLabel: packLabel(v), InWishlist: wl[v.ID]}
	prices, err := s.price.ListByVariantCity(v.ID, cityID)
	if err != nil {
		return vd, err
	}
	links, err := s.link.ListByVariant(v.ID)
	if err != nil {
		return vd, err
	}
	deepByPlatform := make(map[string]string, len(links))
	for _, l := range links {
		if l.DeepLink != nil {
			deepByPlatform[l.PlatformID] = *l.DeepLink
		}
	}

	var available []int64
	for _, p := range prices {
		plat := platMap[p.PlatformID]
		chip := dto.PlatformPriceChipDTO{
			PlatformCode: plat.Code, PlatformName: plat.DisplayName, LogoURL: strPtr(plat.LogoURL),
			Price: dto.Money(p.PricePaise), MRP: dto.MoneyPtr(p.MRPPaise),
			Available: p.Available, DeepLink: deepByPlatform[p.PlatformID],
		}
		vd.PlatformPrices = append(vd.PlatformPrices, chip)
		if p.Available {
			available = append(available, p.PricePaise)
		}
	}
	if avg, ok := util.AveragePaise(available); ok {
		m := dto.Money(avg)
		vd.AveragePrice = &m
		if per, label, uok := util.UnitPricePaise(avg, v.VolumeValue, v.VolumeUnit); uok {
			vd.UnitPrice = util.FormatPaise(per) + "/" + label
		}
	}
	return vd, nil
}

func (s *CatalogService) ProductsByCategory(userID, categoryID, cityID string, platformCodes []string, page query.Page) ([]dto.VariantSearchItemDTO, query.Meta, error) {
	wl, err := s.wishlistSet(userID)
	if err != nil {
		return nil, query.Meta{}, err
	}
	products, total, err := s.product.ListByCategory(categoryID, page)
	if err != nil {
		return nil, query.Meta{}, err
	}
	return s.variantCardsForProducts(products, total, cityID, platformCodes, page, wl)
}

func (s *CatalogService) ProductsBySubcategory(userID, subcategoryID, cityID string, platformCodes []string, page query.Page) ([]dto.VariantSearchItemDTO, query.Meta, error) {
	wl, err := s.wishlistSet(userID)
	if err != nil {
		return nil, query.Meta{}, err
	}
	products, total, err := s.product.ListBySubcategory(subcategoryID, page)
	if err != nil {
		return nil, query.Meta{}, err
	}
	return s.variantCardsForProducts(products, total, cityID, platformCodes, page, wl)
}

func (s *CatalogService) ProductsByBrand(userID, brandID, cityID string, platformCodes []string, page query.Page) ([]dto.VariantSearchItemDTO, query.Meta, error) {
	wl, err := s.wishlistSet(userID)
	if err != nil {
		return nil, query.Meta{}, err
	}
	products, total, err := s.product.ListByBrand(brandID, page)
	if err != nil {
		return nil, query.Meta{}, err
	}
	return s.variantCardsForProducts(products, total, cityID, platformCodes, page, wl)
}

func (s *CatalogService) variantCardsForProducts(products []domain.Product, total int64, cityID string, platformCodes []string, page query.Page, wl map[string]bool) ([]dto.VariantSearchItemDTO, query.Meta, error) {
	productIDs := make([]string, 0, len(products))
	for _, p := range products {
		productIDs = append(productIDs, p.ID)
	}
	variants, err := s.variant.ListByProducts(productIDs)
	if err != nil {
		return nil, query.Meta{}, err
	}
	items, err := s.variantCards(variants, cityID, platformCodes, wl)
	return items, query.BuildMeta(total, page), err
}

func (s *CatalogService) SearchVariants(userID, term, cityID string, platformCodes []string, page query.Page) ([]dto.VariantSearchItemDTO, query.Meta, error) {
	wl, err := s.wishlistSet(userID)
	if err != nil {
		return nil, query.Meta{}, err
	}
	products, total, err := s.product.SearchByNameOrBrand(term, page)
	if err != nil {
		return nil, query.Meta{}, err
	}
	if len(products) == 0 {
		return []dto.VariantSearchItemDTO{}, query.BuildMeta(0, page), nil
	}

	productIDs := make([]string, 0, len(products))
	for _, p := range products {
		productIDs = append(productIDs, p.ID)
	}
	variants, err := s.variant.ListByProducts(productIDs)
	if err != nil {
		return nil, query.Meta{}, err
	}
	items, err := s.variantCards(variants, cityID, platformCodes, wl)
	if err != nil {
		return nil, query.Meta{}, err
	}
	return items, query.BuildMeta(total, page), nil
}

func (s *CatalogService) variantCards(variants []domain.ProductVariant, cityID string, platformCodes []string, wl map[string]bool) ([]dto.VariantSearchItemDTO, error) {
	items := make([]dto.VariantSearchItemDTO, 0, len(variants))
	if len(variants) == 0 {
		return items, nil
	}
	plats, err := s.platforms.ListEnabled()
	if err != nil {
		return nil, err
	}
	platByID := make(map[string]domain.Platform, len(plats))
	for _, p := range plats {
		platByID[p.ID] = p
	}
	want := make(map[string]bool, len(platformCodes))
	for _, code := range platformCodes {
		want[code] = true
	}

	productIDs := make([]string, 0, len(variants))
	variantIDs := make([]string, 0, len(variants))
	seenProduct := make(map[string]bool)
	for _, v := range variants {
		variantIDs = append(variantIDs, v.ID)
		if !seenProduct[v.ProductID] {
			seenProduct[v.ProductID] = true
			productIDs = append(productIDs, v.ProductID)
		}
	}
	prods, err := s.product.FindByIDs(productIDs)
	if err != nil {
		return nil, err
	}
	brandIDs := make([]string, 0, len(prods))
	for _, p := range prods {
		if p.BrandID != nil {
			brandIDs = append(brandIDs, *p.BrandID)
		}
	}
	brands, err := s.brand.FindByIDs(brandIDs)
	if err != nil {
		return nil, err
	}
	prices, err := s.price.ListByVariantsCity(variantIDs, cityID)
	if err != nil {
		return nil, err
	}
	priceByVariant := make(map[string][]domain.PlatformPrice, len(variantIDs))
	for _, pr := range prices {
		priceByVariant[pr.VariantID] = append(priceByVariant[pr.VariantID], pr)
	}
	imageByVariant, err := s.link.PrimaryImagesByVariants(variantIDs)
	if err != nil {
		return nil, err
	}

	for _, v := range variants {
		prod := prods[v.ProductID]
		row := dto.VariantSearchItemDTO{
			VariantID: v.ID, ProductID: v.ProductID, ProductName: prod.Name,
			ImageURL: imageByVariant[v.ID], PackLabel: packLabel(v),
			ReferencePrices: []dto.ReferencePriceDTO{},
			InWishlist:      wl[v.ID],
		}
		if prod.BrandID != nil {
			if b, ok := brands[*prod.BrandID]; ok {
				row.BrandName = b.Name
			}
		}
		var minPaise *int64
		for _, pr := range priceByVariant[v.ID] {
			pl, ok := platByID[pr.PlatformID]
			if !ok {
				continue
			}
			if len(want) > 0 && !want[pl.Code] {
				continue
			}
			mrp := int64(0)
			if pr.MRPPaise != nil {
				mrp = *pr.MRPPaise
			}
			row.ReferencePrices = append(row.ReferencePrices, dto.ReferencePriceDTO{
				PlatformCode: pl.Code, PlatformName: pl.DisplayName,
				PricePaise: pr.PricePaise, MRPPaise: mrp, Available: pr.Available,
				LastUpdatedAt: pr.LastUpdatedAt.Format(time.RFC3339),
			})
			if pr.Available && (minPaise == nil || pr.PricePaise < *minPaise) {
				m := pr.PricePaise
				minPaise = &m
			}
		}
		row.ReferenceFromPaise = minPaise
		if minPaise != nil {
			if per, label, ok := util.UnitPricePaise(*minPaise, v.VolumeValue, v.VolumeUnit); ok {
				row.UnitPrice = util.FormatPaise(per) + "/" + label
			}
		}
		items = append(items, row)
	}
	return items, nil
}

func (s *CatalogService) Deals(userID, cityID string, platformCodes []string) ([]dto.VariantSearchItemDTO, error) {
	wl, err := s.wishlistSet(userID)
	if err != nil {
		return nil, err
	}
	ids, err := s.product.ListDealVariantIDs(cityID, 30)
	if err != nil {
		return nil, err
	}
	variants, err := s.variant.ListByIDsOrdered(ids)
	if err != nil {
		return nil, err
	}
	return s.variantCards(variants, cityID, platformCodes, wl)
}

func (s *CatalogService) platformMap() (map[string]domain.Platform, error) {
	plats, err := s.platforms.ListEnabled()
	if err != nil {
		return nil, err
	}
	m := make(map[string]domain.Platform, len(plats))
	for _, p := range plats {
		m[p.ID] = p
	}
	return m, nil
}

func (s *CatalogService) StoreVariants(userID, storeCode, cityID string, platformCodes []string, page query.Page) ([]dto.VariantSearchItemDTO, query.Meta, error) {
	pl, err := s.platforms.FindByCode(storeCode)
	if err != nil {
		return nil, query.Meta{}, err
	}
	if pl == nil || !pl.Enabled {
		return nil, query.Meta{}, errs.NotFound("STORE_NOT_FOUND", "store not found")
	}
	wl, err := s.wishlistSet(userID)
	if err != nil {
		return nil, query.Meta{}, err
	}
	ids, total, err := s.product.ListVariantIDsByPlatformCity(pl.ID, cityID, page)
	if err != nil {
		return nil, query.Meta{}, err
	}
	variants, err := s.variant.ListByIDsOrdered(ids)
	if err != nil {
		return nil, query.Meta{}, err
	}
	items, err := s.variantCards(variants, cityID, platformCodes, wl)
	if err != nil {
		return nil, query.Meta{}, err
	}
	return items, query.BuildMeta(total, page), nil
}

func (s *CatalogService) TrendingItems(userID, cityID string, platformCodes []string, page query.Page) ([]dto.VariantSearchItemDTO, query.Meta, error) {
	wl, err := s.wishlistSet(userID)
	if err != nil {
		return nil, query.Meta{}, err
	}
	products, total, err := s.product.ListTopPaged(page)
	if err != nil {
		return nil, query.Meta{}, err
	}
	variants, err := s.defaultVariantsFor(products)
	if err != nil {
		return nil, query.Meta{}, err
	}
	items, err := s.variantCards(variants, cityID, platformCodes, wl)
	if err != nil {
		return nil, query.Meta{}, err
	}
	return items, query.BuildMeta(total, page), nil
}

func (s *CatalogService) TopCategories(page query.Page) ([]dto.CategoryCardDTO, query.Meta, error) {
	cats, total, err := s.category.ListVisibleWithProductsPaged(true, page)
	if err != nil {
		return nil, query.Meta{}, err
	}
	items := make([]dto.CategoryCardDTO, 0, len(cats))
	for _, c := range cats {
		items = append(items, dto.CategoryCardDTO{ID: c.ID, Name: c.Name, Slug: c.Slug, ImageURL: strPtr(c.ImageURL)})
	}
	return items, query.BuildMeta(total, page), nil
}

func (s *CatalogService) Stores(page query.Page) ([]dto.PlatformDTO, query.Meta, error) {
	plats, total, err := s.platforms.ListEnabledPaged(page)
	if err != nil {
		return nil, query.Meta{}, err
	}
	items := make([]dto.PlatformDTO, 0, len(plats))
	for _, p := range plats {
		items = append(items, platformDTO(p))
	}
	return items, query.BuildMeta(total, page), nil
}

func platformDTO(p domain.Platform) dto.PlatformDTO {
	return dto.PlatformDTO{Code: p.Code, Name: p.DisplayName, LogoURL: strPtr(p.LogoURL), DeliveryETAText: strPtr(p.DeliveryETAText)}
}

func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func packLabel(v domain.ProductVariant) string {
	val := strconv.FormatFloat(v.VolumeValue, 'f', -1, 64)
	unit := strings.TrimSpace(string(v.VolumeUnit))
	return val + " " + unit
}
