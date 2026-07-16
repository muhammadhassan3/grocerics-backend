package quickcommerce

type Location struct {
	Lat     float64
	Lon     float64
	Pincode string
}

type Product struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Brand      string   `json:"brand"`
	Available  bool     `json:"available"`
	PricePaise int64    `json:"price_paise"`
	MRPPaise   int64    `json:"mrp_paise"`
	Quantity   string   `json:"quantity"`
	Multipack  int      `json:"multipack"`
	Rating     float64  `json:"rating"`
	Inventory  *int     `json:"inventory"`
	StockLabel string   `json:"stock_label"`
	Images     []string `json:"images"`
	DeepLink   string   `json:"deeplink"`
}

type SearchResult struct {
	Query            string    `json:"query"`
	Platform         string    `json:"platform"`
	TotalResults     int       `json:"total_results"`
	Products         []Product `json:"products"`
	CreditsRemaining int       `json:"credits_remaining"`
}

type ItemDetail struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Quantity   string `json:"quantity"`
	Available  bool   `json:"available"`
	PricePaise int64  `json:"price_paise"`
	MRPPaise   int64  `json:"mrp_paise"`
	Inventory  *int   `json:"inventory"`
	StockLabel string `json:"stock_label"`
	DeepLink   string `json:"deeplink"`
}

type ETAResult struct {
	Platform    string `json:"platform"`
	ETAMinutes  int    `json:"eta_minutes"`
	ETAText     string `json:"eta_text"`
	Serviceable bool   `json:"serviceable"`
}

type GroupSearchResult struct {
	Query            string               `json:"query"`
	ByPlatform       map[string][]Product `json:"by_platform"`
	CreditsRemaining int                  `json:"credits_remaining"`
}

type GroupETAResult struct {
	ByPlatform       map[string]ETAResult `json:"by_platform"`
	CreditsRemaining int                  `json:"credits_remaining"`
}

type Credits struct {
	Remaining int `json:"credits_remaining"`
}
