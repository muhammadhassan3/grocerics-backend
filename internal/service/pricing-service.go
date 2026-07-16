package service

import (
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/repository"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

type PricingService struct {
	platformPrices *repository.PlatformPriceRepository
	summaries      *repository.VariantPriceSummaryRepository
	links          *repository.ProductPlatformLinkRepository
}

func NewPricingService(db *gorm.DB) *PricingService {
	return &PricingService{
		platformPrices: repository.NewPlatformPriceRepository(db),
		summaries:      repository.NewVariantPriceSummaryRepository(db),
		links:          repository.NewProductPlatformLinkRepository(db),
	}
}

func ComputeSummary(variantID, cityID string, prices []domain.PlatformPrice) domain.VariantPriceSummary {
	s := domain.VariantPriceSummary{VariantID: variantID, CityID: cityID}
	amounts := make([]int64, 0, len(prices))
	for _, p := range prices {
		if !p.Available {
			continue
		}
		amounts = append(amounts, p.PricePaise)
		if s.MinPricePaise == nil || p.PricePaise < *s.MinPricePaise {
			min := p.PricePaise
			pid := p.PlatformID
			s.MinPricePaise = &min
			s.MinPlatformID = &pid
		}
	}
	s.AvailablePlatformCount = len(amounts)
	if avg, ok := util.AveragePaise(amounts); ok {
		s.AvgPricePaise = &avg
	}
	return s
}

func (s *PricingService) RecomputeVariantSummary(variantID, cityID string) error {
	prices, err := s.platformPrices.ListByVariantCity(variantID, cityID)
	if err != nil {
		return err
	}
	summary := ComputeSummary(variantID, cityID, prices)
	return s.summaries.Upsert(&summary)
}

func (s *PricingService) SetManualPrice(variantID, platformID, cityID string, mrpPaise, pricePaise int64, platformSKU, deepLink string) error {
	if _, err := s.links.Upsert(&domain.ProductPlatformLink{
		VariantID:   variantID,
		PlatformID:  platformID,
		PlatformSKU: util.PtrIfSet(platformSKU),
		DeepLink:    util.PtrIfSet(deepLink),
	}); err != nil {
		return err
	}
	mrp := mrpPaise
	if err := s.platformPrices.Upsert(&domain.PlatformPrice{
		VariantID:  variantID,
		PlatformID: platformID,
		CityID:     cityID,
		PricePaise: pricePaise,
		MRPPaise:   &mrp,
		Available:  true,
		Source:     domain.PriceSourceManual,
	}); err != nil {
		return err
	}
	return s.RecomputeVariantSummary(variantID, cityID)
}
