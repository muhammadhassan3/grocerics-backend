package service

import (
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/repository"

	"gorm.io/gorm"
)

type CartService struct {
	cart     *repository.CartRepository
	item     *repository.CartItemRepository
	variant  *repository.ProductVariantRepository
	product  *repository.ProductRepository
	summary  *repository.VariantPriceSummaryRepository
	price    *repository.PlatformPriceRepository
	platform *repository.PlatformRepository
	eta      *repository.PlatformDeliveryETARepository
	wishlist *repository.WishlistRepository
	link     *repository.ProductPlatformLinkRepository
}

func NewCartService(db *gorm.DB) *CartService {
	return &CartService{
		cart:     repository.NewCartRepository(db),
		item:     repository.NewCartItemRepository(db),
		variant:  repository.NewProductVariantRepository(db),
		product:  repository.NewProductRepository(db),
		summary:  repository.NewVariantPriceSummaryRepository(db),
		price:    repository.NewPlatformPriceRepository(db),
		platform: repository.NewPlatformRepository(db),
		eta:      repository.NewPlatformDeliveryETARepository(db),
		wishlist: repository.NewWishlistRepository(db),
		link:     repository.NewProductPlatformLinkRepository(db),
	}
}

type breakdownLine struct {
	ID        string
	VariantID string
	Quantity  int
}

func (s *CartService) GetCart(userID, cityID, pincode string) (*dto.CartResponse, error) {
	cart, err := s.cart.FindOrCreateByUser(userID)
	if err != nil {
		return nil, err
	}
	items, err := s.item.ListByCart(cart.ID)
	if err != nil {
		return nil, err
	}
	lines := make([]breakdownLine, 0, len(items))
	for _, it := range items {
		lines = append(lines, breakdownLine{ID: it.ID, VariantID: it.VariantID, Quantity: it.Quantity})
	}
	resp, err := s.buildResponse(lines, cityID, pincode)
	if err != nil {
		return nil, err
	}
	resp.CartID = cart.ID
	return resp, nil
}

func (s *CartService) GetWishlist(userID, cityID, pincode string) (*dto.CartResponse, error) {
	rows, err := s.wishlist.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	lines := make([]breakdownLine, 0, len(rows))
	for _, w := range rows {
		lines = append(lines, breakdownLine{ID: w.ID, VariantID: w.VariantID, Quantity: 1})
	}
	return s.buildResponse(lines, cityID, pincode)
}

func (s *CartService) AddItem(userID, variantID string, quantity int) error {
	if quantity < 1 {
		quantity = 1
	}
	v, err := s.variant.FindByID(variantID)
	if err != nil {
		return err
	}
	if v == nil {
		return errs.NotFound("VARIANT_NOT_FOUND", "product variant not found")
	}
	cart, err := s.cart.FindOrCreateByUser(userID)
	if err != nil {
		return err
	}
	_, err = s.item.Upsert(cart.ID, variantID, quantity)
	return err
}

func (s *CartService) UpdateItem(itemID string, quantity int) error {
	if quantity < 1 {
		return errs.BadRequest("VALIDATION", "quantity must be at least 1")
	}
	return s.item.UpdateQuantity(itemID, quantity)
}

func (s *CartService) RemoveItem(itemID string) error { return s.item.Delete(itemID) }

func (s *CartService) AddWishlist(userID, variantID string) error {
	v, err := s.variant.FindByID(variantID)
	if err != nil {
		return err
	}
	if v == nil {
		return errs.NotFound("VARIANT_NOT_FOUND", "product variant not found")
	}
	return s.wishlist.Add(userID, variantID)
}

func (s *CartService) RemoveWishlist(userID, variantID string) error {
	return s.wishlist.Delete(userID, variantID)
}

func (s *CartService) buildResponse(lines []breakdownLine, cityID, pincode string) (*dto.CartResponse, error) {
	resp := &dto.CartResponse{Items: []dto.CartLineDTO{}, Platforms: []dto.PlatformBreakdownDTO{}}
	if len(lines) == 0 {
		return resp, nil
	}

	variantIDs := make([]string, 0, len(lines))
	for _, l := range lines {
		variantIDs = append(variantIDs, l.VariantID)
	}
	variants, err := s.variant.FindByIDs(variantIDs)
	if err != nil {
		return nil, err
	}
	productIDs := make([]string, 0, len(variants))
	for _, v := range variants {
		productIDs = append(productIDs, v.ProductID)
	}
	products, err := s.product.FindByIDs(productIDs)
	if err != nil {
		return nil, err
	}
	summaries, err := s.summary.GetMany(variantIDs, cityID)
	if err != nil {
		return nil, err
	}
	imageByVariant, err := s.link.PrimaryImagesByVariants(variantIDs)
	if err != nil {
		return nil, err
	}
	prices, err := s.price.ListByVariantsCity(variantIDs, cityID)
	if err != nil {
		return nil, err
	}
	platforms, err := s.platform.ListEnabled()
	if err != nil {
		return nil, err
	}

	priceIdx := make(map[string]map[string]domain.PlatformPrice)
	for _, p := range prices {
		if priceIdx[p.VariantID] == nil {
			priceIdx[p.VariantID] = make(map[string]domain.PlatformPrice)
		}
		priceIdx[p.VariantID][p.PlatformID] = p
	}

	etaByPlatform := map[string]int{}
	if pincode != "" {
		if etas, eErr := s.eta.ListByPincode(pincode); eErr == nil {
			for _, e := range etas {
				if e.ETAMinutes != nil {
					etaByPlatform[e.PlatformID] = *e.ETAMinutes
				}
			}
		}
	}

	for _, l := range lines {
		v := variants[l.VariantID]
		line := dto.CartLineDTO{ItemID: l.ID, VariantID: l.VariantID, PackLabel: packLabel(v), Quantity: l.Quantity, ImageURL: imageByVariant[l.VariantID]}
		if p, ok := products[v.ProductID]; ok {
			line.ProductName = p.Name
		}
		if sum, ok := summaries[l.VariantID]; ok {
			line.AveragePrice = dto.MoneyPtr(sum.AvgPricePaise)
		}
		resp.Items = append(resp.Items, line)
	}

	cheapestIdx, cheapestTotal := -1, int64(-1)
	for _, plat := range platforms {
		b := dto.PlatformBreakdownDTO{
			PlatformCode: plat.Code, PlatformName: plat.DisplayName, LogoURL: strPtr(plat.LogoURL),
			AvailableItemIDs: []string{}, UnavailableItemIDs: []string{},
		}
		if m, ok := etaByPlatform[plat.ID]; ok {
			eta := m
			b.DeliveryETAMinutes = &eta
		}
		var itemTotal, availTotal int64
		for _, l := range lines {
			pp, ok := priceIdx[l.VariantID][plat.ID]
			if !ok {
				b.UnavailableItemIDs = append(b.UnavailableItemIDs, l.ID)
				continue
			}
			itemTotal += pp.PricePaise * int64(l.Quantity)
			if pp.Available {
				b.AvailableItemIDs = append(b.AvailableItemIDs, l.ID)
				availTotal += pp.PricePaise * int64(l.Quantity)
			} else {
				b.UnavailableItemIDs = append(b.UnavailableItemIDs, l.ID)
			}
		}
		b.ItemTotal = dto.Money(itemTotal)
		b.AvailableTotal = dto.Money(availTotal)
		resp.Platforms = append(resp.Platforms, b)
		// is_cheapest only among platforms carrying the WHOLE basket (no
		// unavailable/missing lines). If none do, nobody is flagged.
		fullCoverage := len(b.UnavailableItemIDs) == 0 && len(b.AvailableItemIDs) > 0
		if fullCoverage && (cheapestTotal < 0 || availTotal < cheapestTotal) {
			cheapestTotal = availTotal
			cheapestIdx = len(resp.Platforms) - 1
		}
	}
	if cheapestIdx >= 0 {
		resp.Platforms[cheapestIdx].IsCheapest = true
	}
	return resp, nil
}
