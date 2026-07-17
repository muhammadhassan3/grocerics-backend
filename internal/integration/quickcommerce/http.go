package quickcommerce

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"grocerics-backend/internal/util"

	"go.uber.org/zap"
)

const defaultBaseURL = "https://api.quickcommerceapi.com/v1"

type httpClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
	record  func(RawCall)
}

func NewHTTPClient(cfg Config) Client {
	base := cfg.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	return &httpClient{
		apiKey:  cfg.APIKey,
		baseURL: strings.TrimRight(base, "/"),
		http:    &http.Client{Timeout: 20 * time.Second},
		record:  cfg.Record,
	}
}

func (c *httpClient) emit(call RawCall) {
	if c.record == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			zap.S().Warnw("qc: raw response recorder panicked", "endpoint", call.Endpoint, "panic", r)
		}
	}()
	c.record(call)
}

func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', 6, 64) }

func (c *httpClient) doGet(ctx context.Context, path string, q url.Values, out any) (err error) {
	u := c.baseURL + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	start := time.Now()
	call := RawCall{Endpoint: path, Params: q}
	defer func() {
		if err != nil {
			call.Err = err.Error()
		}
		call.DurationMs = int(time.Since(start).Milliseconds())
		c.emit(call)
	}()

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
	body, err := io.ReadAll(resp.Body)
	call.StatusCode, call.Body = resp.StatusCode, body
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("quickcommerce: GET %s -> %d", path, resp.StatusCode)
	}
	return json.Unmarshal(body, out)
}

func (c *httpClient) locQuery(loc Location) url.Values {
	q := url.Values{}
	q.Set("lat", ftoa(loc.Lat))
	q.Set("lon", ftoa(loc.Lon))
	if loc.Pincode != "" {
		q.Set("pincode", loc.Pincode)
	}
	return q
}

type rawProduct struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Brand      string        `json:"brand"`
	Available  flexBool      `json:"available"`
	MRP        flexFloat     `json:"mrp"`
	OfferPrice flexFloat     `json:"offer_price"`
	Quantity   string        `json:"quantity"`
	Rating     float64       `json:"rating"`
	Inventory  flexInventory `json:"inventory"`
	Images     []string      `json:"images"`
	DeepLink   string        `json:"deeplink"`
}

func (r rawProduct) toProduct() Product {
	avail, inv, label := resolveStock(bool(r.Available), r.Inventory)
	return Product{
		ID: r.ID, Name: r.Name, Brand: r.Brand, Available: avail,
		PricePaise: util.RupeesToPaise(float64(r.OfferPrice)),
		MRPPaise:   util.RupeesToPaise(float64(r.MRP)),
		Quantity:   r.Quantity, Multipack: parseMultipack(r.Quantity), Rating: r.Rating,
		Inventory: inv, StockLabel: label, Images: r.Images, DeepLink: r.DeepLink,
	}
}

func resolveStock(rowAvail bool, fi flexInventory) (bool, *int, string) {
	if !fi.present {
		return rowAvail, nil, ""
	}
	if fi.Count != nil {
		return *fi.Count > 0, fi.Count, fi.Label
	}
	return fi.Available, nil, fi.Label
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

type rawItem struct {
	ItemID    string        `json:"item_id"`
	Name      string        `json:"name"`
	Quantity  string        `json:"quantity"`
	Available flexBool      `json:"available"`
	Price     flexFloat     `json:"price"`
	MRP       flexFloat     `json:"mrp"`
	Inventory flexInventory `json:"inventory"`
	DeepLink  string        `json:"deeplink"`
}

func (c *httpClient) GetItem(ctx context.Context, itemID, platform string, loc Location) (*ItemDetail, error) {
	q := c.locQuery(loc)
	q.Set("item_id", itemID)
	q.Set("platform", platform)
	var raw struct {
		Data struct {
			Items []rawItem `json:"items"`
		} `json:"data"`
	}
	if err := c.doGet(ctx, "/item", q, &raw); err != nil {
		return nil, err
	}
	if len(raw.Data.Items) == 0 {
		return nil, fmt.Errorf("quickcommerce: item %s not found on %s", itemID, platform)
	}
	it := raw.Data.Items[0]
	avail, inv, label := resolveStock(bool(it.Available), it.Inventory)
	return &ItemDetail{
		ID: it.ItemID, Name: it.Name, Quantity: it.Quantity, Available: avail,
		PricePaise: util.RupeesToPaise(float64(it.Price)), MRPPaise: util.RupeesToPaise(float64(it.MRP)),
		Inventory: inv, StockLabel: label, DeepLink: it.DeepLink,
	}, nil
}

func (c *httpClient) ETA(ctx context.Context, platform string, loc Location) (*ETAResult, error) {
	q := c.locQuery(loc)
	q.Set("platform", platform)
	var raw struct {
		Data struct {
			ETA  string `json:"eta"`
			Open bool   `json:"open"`
		} `json:"data"`
	}
	if err := c.doGet(ctx, "/eta", q, &raw); err != nil {
		return nil, err
	}
	return &ETAResult{Platform: platform, ETAMinutes: firstInt(raw.Data.ETA), ETAText: raw.Data.ETA, Serviceable: raw.Data.Open}, nil
}

func (c *httpClient) GroupSearch(ctx context.Context, query string, loc Location, platforms []string) (*GroupSearchResult, error) {
	q := c.locQuery(loc)
	q.Set("q", query)
	q.Set("platforms", strings.Join(platforms, ","))
	var raw struct {
		CreditsRemaining int `json:"credits_remaining"`
		Data             struct {
			Query   string                  `json:"query"`
			Results map[string][]rawProduct `json:"results"`
		} `json:"data"`
	}
	if err := c.doGet(ctx, "/groupsearch", q, &raw); err != nil {
		return nil, err
	}
	out := &GroupSearchResult{Query: raw.Data.Query, ByPlatform: make(map[string][]Product), CreditsRemaining: raw.CreditsRemaining}
	for plat, prods := range raw.Data.Results {
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
			Results []struct {
				Platform string `json:"platform"`
				ETA      string `json:"eta"`
				Open     bool   `json:"open"`
			} `json:"results"`
		} `json:"data"`
	}
	if err := c.doGet(ctx, "/groupeta", q, &raw); err != nil {
		return nil, err
	}
	out := &GroupETAResult{ByPlatform: make(map[string]ETAResult), CreditsRemaining: raw.CreditsRemaining}
	for _, e := range raw.Data.Results {
		out.ByPlatform[e.Platform] = ETAResult{Platform: e.Platform, ETAMinutes: firstInt(e.ETA), ETAText: e.ETA, Serviceable: e.Open}
	}
	return out, nil
}

func (c *httpClient) Credits(ctx context.Context) (*Credits, error) {
	var raw struct {
		CreditsRemaining int `json:"credits_remaining"`
		Summary          struct {
			TotalAvailable int `json:"total_available"`
		} `json:"summary"`
	}
	if err := c.doGet(ctx, "/credits", nil, &raw); err != nil {
		return nil, err
	}
	rem := raw.CreditsRemaining
	if rem == 0 {
		rem = raw.Summary.TotalAvailable
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
