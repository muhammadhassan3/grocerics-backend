package test

import (
	"context"
	"testing"

	"grocerics-backend/internal/integration/quickcommerce"
)

func TestNewSelectsMockWhenNoKey(t *testing.T) {

	c := quickcommerce.New(quickcommerce.Config{})
	if _, err := c.Credits(context.Background()); err != nil {
		t.Fatalf("mock Credits failed: %v", err)
	}
}

func TestMockSearchReturnsProducts(t *testing.T) {
	c := quickcommerce.NewMock()
	res, err := c.Search(context.Background(), "bread", quickcommerce.Location{Lat: 28.6, Lon: 77.2}, "blinkit")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Products) == 0 {
		t.Fatal("expected products")
	}
	if p := res.Products[0]; p.ID == "" || p.PricePaise <= 0 {
		t.Errorf("bad product: %+v", p)
	}
}

func TestMockPricesDifferPerPlatform(t *testing.T) {
	c := quickcommerce.NewMock()
	ctx := context.Background()
	loc := quickcommerce.Location{Lat: 28.6, Lon: 77.2}
	blink, _ := c.GetItem(ctx, "1001", "blinkit", loc)
	zepto, _ := c.GetItem(ctx, "1001", "zepto", loc)
	if blink.PricePaise == zepto.PricePaise {
		t.Errorf("expected per-platform difference, got equal %d", blink.PricePaise)
	}
	if blink.PricePaise >= zepto.PricePaise {
		t.Errorf("blinkit should be cheaper than zepto in the mock: %d vs %d", blink.PricePaise, zepto.PricePaise)
	}
}

func TestMockGroupSearchCoversPlatforms(t *testing.T) {
	c := quickcommerce.NewMock()
	res, err := c.GroupSearch(context.Background(), "bread", quickcommerce.Location{}, []string{"blinkit", "zepto"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.ByPlatform) != 2 {
		t.Errorf("expected 2 platforms, got %d", len(res.ByPlatform))
	}
}
