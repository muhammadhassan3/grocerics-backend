package query

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Page struct {
	Number int
	Size   int
}

func (p Page) Offset() int { return (p.Number - 1) * p.Size }
func (p Page) Limit() int  { return p.Size }

func (p Page) Apply(db *gorm.DB) *gorm.DB { return db.Limit(p.Size).Offset(p.Offset()) }

type Sort struct {
	Column    string
	Direction string
}

// OrderClause appends ", id desc" as a stable secondary sort so paging
// across rows with identical primary-sort values never duplicates or
// skips. Every domain table has an id PK via BaseModel.
func (s Sort) OrderClause() string { return s.Column + " " + s.Direction + ", id desc" }

func (s Sort) Apply(db *gorm.DB) *gorm.DB { return db.Order(s.OrderClause()) }

type Meta struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Pages   int   `json:"pages"`
}

// PageFromContext reads ?page=N&page_size=M, also handles invalid values.
func PageFromContext(c *gin.Context) Page {
	n, _ := strconv.Atoi(c.Query("page"))
	if n < 1 {
		n = DefaultPage
	}
	s, _ := strconv.Atoi(c.Query("page_size"))
	if s < 1 {
		s = DefaultPageSize
	}
	if s > MaxPageSize {
		s = MaxPageSize
	}
	return Page{Number: n, Size: s}
}

// SortFromContext reads ?sort=col&order=asc|desc.
func SortFromContext(c *gin.Context, allow []string, fallback Sort) Sort {
	col := c.Query("sort")
	ok := false
	for _, a := range allow {
		if a == col {
			ok = true
			break
		}
	}
	if !ok {
		col = fallback.Column
	}
	dir := c.Query("order")
	if dir != "asc" && dir != "desc" {
		dir = fallback.Direction
	}
	return Sort{Column: col, Direction: dir}
}

// BuildMeta returns the pagination envelope for a list response.
// total=0 yields Pages=0 (frontends should check total before paging).
func BuildMeta(total int64, p Page) Meta {
	pages := int(total) / p.Size
	if int(total)%p.Size != 0 {
		pages++
	}
	return Meta{Total: total, Page: p.Number, PerPage: p.Size, Pages: pages}
}
