package service

import (
	"strconv"
	"strings"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
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
	summary   *repository.VariantPriceSummaryRepository
	link      *repository.ProductPlatformLinkRepository
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
		summary:   repository.NewVariantPriceSummaryRepository(db),
		link:      repository.NewProductPlatformLinkRepository(db),
	}
}

func (s *CatalogService) Home(cityID string) (*dto.HomeResponse, error) {
	banners, err := s.banners.ListActive()
	if err != nil {
		return nil, err
	}
	plats, err := s.platforms.ListEnabled()
	if err != nil {
		return nil, err
	}
	cats, err := s.category.ListVisible(true)
	if err != nil {
		return nil, err
	}
	top, err := s.product.ListTop(10)
	if err != nil {
		return nil, err
	}
	cards, err := s.productCards(top, cityID)
	if err != nil {
		return nil, err
	}

	resp := &dto.HomeResponse{TrendingItems: cards}
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

func (s *CatalogService) ProductDetail(productID, cityID string) (*dto.ProductDetailDTO, error) {
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
	for _, v := range variants {
		vd, vErr := s.variantDetail(v, cityID, platMap)
		if vErr != nil {
			return nil, vErr
		}
		out.Variants = append(out.Variants, vd)
	}

	similar, err := s.product.ListSimilar(product.CategoryID, product.ID, 6)
	if err != nil {
		return nil, err
	}
	if out.Similar, err = s.productCards(similar, cityID); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *CatalogService) variantDetail(v domain.ProductVariant, cityID string, platMap map[string]domain.Platform) (dto.VariantDetailDTO, error) {
	vd := dto.VariantDetailDTO{VariantID: v.ID, PackLabel: packLabel(v)}
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

func (s *CatalogService) ProductsByCategory(categoryID, cityID string, page query.Page) ([]dto.ProductCardDTO, query.Meta, error) {
	products, total, err := s.product.ListByCategory(categoryID, page)
	if err != nil {
		return nil, query.Meta{}, err
	}
	cards, err := s.productCards(products, cityID)
	return cards, query.BuildMeta(total, page), err
}

func (s *CatalogService) Search(term, cityID string, page query.Page) ([]dto.ProductCardDTO, query.Meta, error) {
	products, total, err := s.product.Search(term, page)
	if err != nil {
		return nil, query.Meta{}, err
	}
	cards, err := s.productCards(products, cityID)
	return cards, query.BuildMeta(total, page), err
}

func (s *CatalogService) Deals(cityID string) ([]dto.ProductCardDTO, error) {
	products, err := s.product.ListDeals(cityID, 30)
	if err != nil {
		return nil, err
	}
	return s.productCards(products, cityID)
}

func (s *CatalogService) productCards(products []domain.Product, cityID string) ([]dto.ProductCardDTO, error) {
	if len(products) == 0 {
		return []dto.ProductCardDTO{}, nil
	}
	productIDs := make([]string, 0, len(products))
	brandIDs := make([]string, 0, len(products))
	for _, p := range products {
		productIDs = append(productIDs, p.ID)
		if p.BrandID != nil {
			brandIDs = append(brandIDs, *p.BrandID)
		}
	}
	brands, err := s.brand.FindByIDs(brandIDs)
	if err != nil {
		return nil, err
	}
	defaults, err := s.variant.DefaultsForProducts(productIDs)
	if err != nil {
		return nil, err
	}
	variantIDs := make([]string, 0, len(defaults))
	for _, v := range defaults {
		variantIDs = append(variantIDs, v.ID)
	}
	summaries, err := s.summary.GetMany(variantIDs, cityID)
	if err != nil {
		return nil, err
	}

	cards := make([]dto.ProductCardDTO, 0, len(products))
	for _, p := range products {
		card := dto.ProductCardDTO{ProductID: p.ID, Name: p.Name, ImageURL: strPtr(p.ImageURL)}
		if p.BrandID != nil {
			if b, ok := brands[*p.BrandID]; ok {
				card.BrandName = b.Name
			}
		}
		if dv, ok := defaults[p.ID]; ok {
			card.DefaultVariantID = dv.ID
			if sum, ok := summaries[dv.ID]; ok && sum.MinPricePaise != nil {
				card.StartingPrice = dto.MoneyPtr(sum.MinPricePaise)
			}
		}
		cards = append(cards, card)
	}
	return cards, nil
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
