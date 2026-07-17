package quickcommerce

import (
	"context"
	"net/url"
)

type Client interface {
	Search(ctx context.Context, query string, loc Location, platform string) (*SearchResult, error)
	GetItem(ctx context.Context, itemID, platform string, loc Location) (*ItemDetail, error)
	ETA(ctx context.Context, platform string, loc Location) (*ETAResult, error)
	GroupSearch(ctx context.Context, query string, loc Location, platforms []string) (*GroupSearchResult, error)
	GroupETA(ctx context.Context, loc Location, platforms []string) (*GroupETAResult, error)
	Credits(ctx context.Context) (*Credits, error)
	ListPlatforms(ctx context.Context) ([]string, error)
}

type RawCall struct {
	Endpoint   string
	Params     url.Values
	StatusCode int
	Body       []byte
	Err        string
	DurationMs int
}

type Config struct {
	APIKey  string
	BaseURL string
	Record  func(RawCall)
}

func New(cfg Config) Client {
	if cfg.APIKey == "" {
		return NewMock()
	}
	return NewHTTPClient(cfg)
}
