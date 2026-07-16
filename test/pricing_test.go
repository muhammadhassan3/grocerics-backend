package test

import (
	"testing"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/service"
)

func price(id string, paise int64, available bool) domain.PlatformPrice {
	return domain.PlatformPrice{PlatformID: id, PricePaise: paise, Available: available}
}

func TestComputeSummary(t *testing.T) {
	// avg = round((15000+20000+10000)/3) = 15000; min = blinkit @ 10000.
	s := service.ComputeSummary("v1", "c1", []domain.PlatformPrice{
		price("zepto", 15000, true),
		price("swiggy", 20000, true),
		price("blinkit", 10000, true),
		price("jiomart", 5000, false), // cheapest but out of stock — must be ignored
	})
	if s.AvailablePlatformCount != 3 {
		t.Fatalf("count = %d, want 3", s.AvailablePlatformCount)
	}
	if s.AvgPricePaise == nil || *s.AvgPricePaise != 15000 {
		t.Fatalf("avg = %v, want 15000", s.AvgPricePaise)
	}
	if s.MinPricePaise == nil || *s.MinPricePaise != 10000 {
		t.Fatalf("min = %v, want 10000", s.MinPricePaise)
	}
	if s.MinPlatformID == nil || *s.MinPlatformID != "blinkit" {
		t.Fatalf("minPlatform = %v, want blinkit", s.MinPlatformID)
	}
}

func TestComputeSummaryNoneAvailable(t *testing.T) {
	s := service.ComputeSummary("v1", "c1", []domain.PlatformPrice{price("zepto", 15000, false)})
	if s.AvailablePlatformCount != 0 || s.AvgPricePaise != nil || s.MinPricePaise != nil {
		t.Fatalf("expected empty summary, got %+v", s)
	}
}
