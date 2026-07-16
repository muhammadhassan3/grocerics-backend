package test

import (
	"net/http/httptest"
	"testing"

	"grocerics-backend/internal/query"

	"github.com/gin-gonic/gin"
)

func newCtx(rawQuery string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.URL.RawQuery = rawQuery
	return c
}

func TestPageFromContext(t *testing.T) {
	cases := []struct {
		query    string
		wantNum  int
		wantSize int
	}{
		{"", query.DefaultPage, query.DefaultPageSize},
		{"page=1", 1, query.DefaultPageSize},
		{"page=5&page_size=50", 5, 50},
		{"page=0", query.DefaultPage, query.DefaultPageSize},
		{"page=-3", query.DefaultPage, query.DefaultPageSize},
		{"page=abc", query.DefaultPage, query.DefaultPageSize},
		{"page_size=0", query.DefaultPage, query.DefaultPageSize},
		{"page_size=-1", query.DefaultPage, query.DefaultPageSize},
		{"page_size=abc", query.DefaultPage, query.DefaultPageSize},
		{"page_size=999", query.DefaultPage, query.MaxPageSize},
		{"page_size=100", query.DefaultPage, query.MaxPageSize},
		{"page_size=101", query.DefaultPage, query.MaxPageSize},
	}
	for _, c := range cases {
		t.Run(c.query, func(t *testing.T) {
			p := query.PageFromContext(newCtx(c.query))
			if p.Number != c.wantNum || p.Size != c.wantSize {
				t.Fatalf("got {%d,%d}, want {%d,%d}", p.Number, p.Size, c.wantNum, c.wantSize)
			}
		})
	}
}

func TestSortFromContext(t *testing.T) {
	allow := []string{"created_at", "name", "users.email"}
	fallback := query.Sort{Column: "created_at", Direction: "desc"}

	cases := []struct {
		name    string
		query   string
		wantCol string
		wantDir string
	}{
		{"allowed col + asc", "sort=name&order=asc", "name", "asc"},
		{"allowed col + desc", "sort=name&order=desc", "name", "desc"},
		{"qualified col allowed", "sort=users.email&order=asc", "users.email", "asc"},
		{"disallowed col falls back", "sort=password_hash&order=asc", "created_at", "asc"},
		{"injection attempt falls back", "sort=name;DROP TABLE users--&order=asc", "created_at", "asc"},
		{"empty sort uses fallback col", "order=asc", "created_at", "asc"},
		{"invalid direction falls back", "sort=name&order=sideways", "name", "desc"},
		{"missing direction falls back", "sort=name", "name", "desc"},
		{"all defaults", "", "created_at", "desc"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := query.SortFromContext(newCtx(c.query), allow, fallback)
			if s.Column != c.wantCol || s.Direction != c.wantDir {
				t.Fatalf("got {%q,%q}, want {%q,%q}", s.Column, s.Direction, c.wantCol, c.wantDir)
			}
		})
	}
}

func TestSortFromContext_EmptyAllowlist(t *testing.T) {
	fallback := query.Sort{Column: "created_at", Direction: "desc"}
	s := query.SortFromContext(newCtx("sort=name&order=asc"), nil, fallback)
	if s.Column != "created_at" || s.Direction != "asc" {
		t.Fatalf("got {%q,%q}, want {created_at,asc}", s.Column, s.Direction)
	}
}

func TestBuildMeta(t *testing.T) {
	p := query.Page{Number: 1, Size: 20}
	cases := []struct {
		total     int64
		wantPages int
	}{
		{0, 0},
		{1, 1},
		{19, 1},
		{20, 1},
		{21, 2},
		{100, 5},
		{101, 6},
	}
	for _, c := range cases {
		m := query.BuildMeta(c.total, p)
		if m.Pages != c.wantPages || m.Total != c.total || m.Page != p.Number || m.PerPage != p.Size {
			t.Errorf("total=%d: got Meta%+v, want Pages=%d", c.total, m, c.wantPages)
		}
	}
}

func TestPage_OffsetLimit(t *testing.T) {
	cases := []struct {
		page       query.Page
		wantOffset int
		wantLimit  int
	}{
		{query.Page{Number: 1, Size: 20}, 0, 20},
		{query.Page{Number: 2, Size: 20}, 20, 20},
		{query.Page{Number: 5, Size: 7}, 28, 7},
	}
	for _, c := range cases {
		if c.page.Offset() != c.wantOffset || c.page.Limit() != c.wantLimit {
			t.Errorf("page %+v: got offset=%d limit=%d, want %d/%d",
				c.page, c.page.Offset(), c.page.Limit(), c.wantOffset, c.wantLimit)
		}
	}
}

func TestSort_OrderClause(t *testing.T) {
	got := query.Sort{Column: "name", Direction: "asc"}.OrderClause()
	want := "name asc, id desc"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
