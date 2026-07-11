package quickcommerce

import "context"

// mockClient returns deterministic data so the whole backend runs and is
// testable before a real API key exists. Prices vary per platform so compare
// flows produce a genuine cheapest-platform.
type mockClient struct{}

func NewMock() Client { return &mockClient{} }

var mockPlatforms = []string{"blinkit", "zepto", "instamart", "flipkart", "jiomart", "amazon_now"}

var mockPlatformOffset = map[string]int64{
	"blinkit": 0, "zepto": 500, "instamart": 1000,
	"flipkart": -500, "jiomart": 1500, "amazon_now": 2000,
}

func mockProduct(platform, id, name, brand, qty string, basePaise int64) Product {
	price := basePaise + mockPlatformOffset[platform]
	return Product{
		ID: id, Name: name, Brand: brand, Available: true,
		PricePaise: price, MRPPaise: price + 5000, Quantity: qty,
		Rating: 4.2, Inventory: 25,
		Images:   []string{"https://picsum.photos/seed/" + id + "/200"},
		DeepLink: "https://" + platform + ".example/item/" + id,
	}
}

func mockCatalog(platform string) []Product {
	return []Product{
		mockProduct(platform, "1001", "Sunfeast Whole Grain Bread 500gm", "Sunfeast", "500 g", 30000),
		mockProduct(platform, "1002", "Amul Taaza Toned Milk 1L", "Amul", "1 L", 6600),
	}
}

func (m *mockClient) Search(_ context.Context, query string, _ Location, platform string) (*SearchResult, error) {
	ps := mockCatalog(platform)
	return &SearchResult{Query: query, Platform: platform, TotalResults: len(ps), Products: ps, CreditsRemaining: 100}, nil
}

func (m *mockClient) GetItem(_ context.Context, itemID, platform string, _ Location) (*ItemDetail, error) {
	price := int64(30000) + mockPlatformOffset[platform]
	return &ItemDetail{ID: itemID, PricePaise: price, MRPPaise: price + 5000, Available: true, Stock: 25}, nil
}

func (m *mockClient) ETA(_ context.Context, platform string, _ Location) (*ETAResult, error) {
	return &ETAResult{Platform: platform, ETAMinutes: 10, Serviceable: true}, nil
}

func (m *mockClient) GroupSearch(_ context.Context, query string, _ Location, platforms []string) (*GroupSearchResult, error) {
	by := make(map[string][]Product, len(platforms))
	for _, p := range platforms {
		by[p] = []Product{mockProduct(p, "1001", "Sunfeast Whole Grain Bread 500gm", "Sunfeast", "500 g", 30000)}
	}
	return &GroupSearchResult{Query: query, ByPlatform: by, CreditsRemaining: 100}, nil
}

func (m *mockClient) GroupETA(_ context.Context, _ Location, platforms []string) (*GroupETAResult, error) {
	by := make(map[string]ETAResult, len(platforms))
	for i, p := range platforms {
		by[p] = ETAResult{Platform: p, ETAMinutes: 10 + i*2, Serviceable: true}
	}
	return &GroupETAResult{ByPlatform: by, CreditsRemaining: 100}, nil
}

func (m *mockClient) Credits(_ context.Context) (*Credits, error) {
	return &Credits{Remaining: 100}, nil
}

func (m *mockClient) ListPlatforms(_ context.Context) ([]string, error) {
	return mockPlatforms, nil
}
