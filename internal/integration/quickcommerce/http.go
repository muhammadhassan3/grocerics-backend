package quickcommerce

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"grocerics-backend/internal/util"
)

const defaultBaseURL = "https://api.quickcommerceapi.com/v1"

type httpClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func NewHTTPClient(cfg Config) Client {
	base := cfg.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	return &httpClient{
		apiKey:  cfg.APIKey,
		baseURL: strings.TrimRight(base, "/"),
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', 6, 64) }

func (c *httpClient) doGet(ctx context.Context, path string, q url.Values, out any) error {
	u := c.baseURL + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("quickcommerce: GET %s -> %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *httpClient) locQuery(loc Location) url.Values {
	q := url.Values{}
	q.Set("lat", ftoa(loc.Lat))
	q.Set("lon", ftoa(loc.Lon))
	return q
}

type rawProduct struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Brand      string   `json:"brand"`
	Available  bool     `json:"available"`
	MRP        float64  `json:"mrp"`
	OfferPrice float64  `json:"offer_price"`
	Quantity   string   `json:"quantity"`
	Rating     float64  `json:"rating"`
	Inventory  int      `json:"inventory"`
	Images     []string `json:"images"`
	DeepLink   string   `json:"deeplink"`
}

func (r rawProduct) toProduct() Product {
	return Product{
		ID: r.ID, Name: r.Name, Brand: r.Brand, Available: r.Available,
		PricePaise: util.RupeesToPaise(r.OfferPrice), MRPPaise: util.RupeesToPaise(r.MRP),
		Quantity: r.Quantity, Rating: r.Rating, Inventory: r.Inventory,
		Images: r.Images, DeepLink: r.DeepLink,
	}
}

func (c *httpClient) Search(ctx context.Context, query string, loc Location, platform string) (*SearchResult, error) {
	q := c.locQuery(loc)
	q.Set("q", query)
	q.Set("platform", platform)
	var raw struct {
		CreditsRemaining int `json:"credits_remaining"`
		Data             struct {
			Query        string       `json:"query"`
			Platform     string       `json:"platform"`
			TotalResults int          `json:"total_results"`
			Products     []rawProduct `json:"products"`
		} `json:"data"`
	}
	if err := c.doGet(ctx, "/search", q, &raw); err != nil {
		return nil, err
	}
	out := &SearchResult{Query: raw.Data.Query, Platform: raw.Data.Platform, TotalResults: raw.Data.TotalResults, CreditsRemaining: raw.CreditsRemaining}
	for _, p := range raw.Data.Products {
		out.Products = append(out.Products, p.toProduct())
	}
	return out, nil
}

func (c *httpClient) GetItem(ctx context.Context, itemID, platform string, loc Location) (*ItemDetail, error) {
	q := c.locQuery(loc)
	q.Set("id", itemID)
	q.Set("platform", platform)
	var raw struct {
		Data struct {
			ID        string  `json:"id"`
			Price     float64 `json:"price"`
			MRP       float64 `json:"mrp"`
			Available bool    `json:"available"`
			Stock     int     `json:"stock"`
		} `json:"data"`
	}
	if err := c.doGet(ctx, "/item", q, &raw); err != nil {
		return nil, err
	}
	return &ItemDetail{
		ID: raw.Data.ID, PricePaise: util.RupeesToPaise(raw.Data.Price),
		MRPPaise: util.RupeesToPaise(raw.Data.MRP), Available: raw.Data.Available, Stock: raw.Data.Stock,
	}, nil
}

func (c *httpClient) ETA(ctx context.Context, platform string, loc Location) (*ETAResult, error) {
	q := c.locQuery(loc)
	q.Set("platform", platform)
	var raw struct {
		Data struct {
			Platform    string `json:"platform"`
			ETA         int    `json:"eta"`
			Serviceable bool   `json:"serviceable"`
		} `json:"data"`
	}
	if err := c.doGet(ctx, "/eta", q, &raw); err != nil {
		return nil, err
	}
	return &ETAResult{Platform: raw.Data.Platform, ETAMinutes: raw.Data.ETA, Serviceable: raw.Data.Serviceable}, nil
}

func (c *httpClient) GroupSearch(ctx context.Context, query string, loc Location, platforms []string) (*GroupSearchResult, error) {
	q := c.locQuery(loc)
	q.Set("q", query)
	q.Set("platforms", strings.Join(platforms, ","))
	var raw struct {
		CreditsRemaining int `json:"credits_remaining"`
		Data             struct {
			Query     string                  `json:"query"`
			Platforms map[string][]rawProduct `json:"platforms"`
		} `json:"data"`
	}
	if err := c.doGet(ctx, "/groupsearch", q, &raw); err != nil {
		return nil, err
	}
	out := &GroupSearchResult{Query: raw.Data.Query, ByPlatform: make(map[string][]Product), CreditsRemaining: raw.CreditsRemaining}
	for plat, prods := range raw.Data.Platforms {
		for _, p := range prods {
			out.ByPlatform[plat] = append(out.ByPlatform[plat], p.toProduct())
		}
	}
	return out, nil
}

func (c *httpClient) GroupETA(ctx context.Context, loc Location, platforms []string) (*GroupETAResult, error) {
	q := c.locQuery(loc)
	q.Set("platforms", strings.Join(platforms, ","))
	var raw struct {
		CreditsRemaining int `json:"credits_remaining"`
		Data             struct {
			Platforms map[string]struct {
				ETA         int  `json:"eta"`
				Serviceable bool `json:"serviceable"`
			} `json:"platforms"`
		} `json:"data"`
	}
	if err := c.doGet(ctx, "/groupeta", q, &raw); err != nil {
		return nil, err
	}
	out := &GroupETAResult{ByPlatform: make(map[string]ETAResult), CreditsRemaining: raw.CreditsRemaining}
	for plat, e := range raw.Data.Platforms {
		out.ByPlatform[plat] = ETAResult{Platform: plat, ETAMinutes: e.ETA, Serviceable: e.Serviceable}
	}
	return out, nil
}

func (c *httpClient) Credits(ctx context.Context) (*Credits, error) {
	var raw struct {
		CreditsRemaining int `json:"credits_remaining"`
		Data             struct {
			CreditsRemaining int `json:"credits_remaining"`
		} `json:"data"`
	}
	if err := c.doGet(ctx, "/credits", nil, &raw); err != nil {
		return nil, err
	}
	rem := raw.CreditsRemaining
	if rem == 0 {
		rem = raw.Data.CreditsRemaining
	}
	return &Credits{Remaining: rem}, nil
}

func (c *httpClient) ListPlatforms(ctx context.Context) ([]string, error) {
	var raw struct {
		Platforms []string `json:"platforms"`
		Data      struct {
			Platforms []string `json:"platforms"`
		} `json:"data"`
	}
	if err := c.doGet(ctx, "/supported-platforms", nil, &raw); err != nil {
		return nil, err
	}
	if len(raw.Platforms) > 0 {
		return raw.Platforms, nil
	}
	return raw.Data.Platforms, nil
}
