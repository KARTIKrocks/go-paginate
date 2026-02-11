package paginate

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Default pagination values.
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 1000
	MinPageSize     = 1
)

// Paginator represents offset-based pagination parameters.
// Instances are safe to read concurrently. Use With* methods to create
// modified copies for thread-safe updates.
type Paginator struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// New creates a new Paginator with default values.
func New() *Paginator {
	return &Paginator{
		Page:     DefaultPage,
		PageSize: DefaultPageSize,
	}
}

// NewWithSize creates a new Paginator with a specific page size.
func NewWithSize(pageSize int) *Paginator {
	p := New()
	return p.WithPageSize(pageSize)
}

// NewFromValues creates a Paginator from page and pageSize values.
func NewFromValues(page, pageSize int) *Paginator {
	return New().WithPage(page).WithPageSize(pageSize)
}

// WithPage returns a new paginator with the specified page number.
// This method is thread-safe as it returns a new instance.
func (p *Paginator) WithPage(page int) *Paginator {
	clone := p.Clone()
	if page < 1 {
		page = DefaultPage
	}
	clone.Page = page
	return clone
}

// WithPageSize returns a new paginator with the specified page size.
// This method is thread-safe as it returns a new instance.
func (p *Paginator) WithPageSize(size int) *Paginator {
	clone := p.Clone()
	if size < MinPageSize {
		size = DefaultPageSize
	}
	if size > MaxPageSize {
		size = MaxPageSize
	}
	clone.PageSize = size
	return clone
}

// Offset returns the offset for SQL queries.
// Uses int64 to prevent overflow with large page numbers.
func (p *Paginator) Offset() int64 {
	return int64(p.Page-1) * int64(p.PageSize)
}

// Limit returns the limit for SQL queries.
func (p *Paginator) Limit() int {
	return p.PageSize
}

// Validate validates the pagination parameters.
func (p *Paginator) Validate() error {
	if p.Page < 1 {
		return fmt.Errorf("%w: got %d", ErrInvalidPage, p.Page)
	}
	if p.PageSize < MinPageSize || p.PageSize > MaxPageSize {
		return fmt.Errorf("%w: got %d, allowed range [%d, %d]",
			ErrInvalidPageSize, p.PageSize, MinPageSize, MaxPageSize)
	}
	return nil
}

// SQLClause returns SQL LIMIT OFFSET clause (PostgreSQL style).
func (p *Paginator) SQLClause() string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", p.Limit(), p.Offset())
}

// SQLClauseMySQL returns MySQL-style LIMIT clause.
func (p *Paginator) SQLClauseMySQL() string {
	return fmt.Sprintf("LIMIT %d, %d", p.Offset(), p.Limit())
}

// HasPrevious returns true if there's a previous page.
func (p *Paginator) HasPrevious() bool {
	return p.Page > 1
}

// PreviousPage returns the previous page number.
// Returns 1 if already on the first page.
func (p *Paginator) PreviousPage() int {
	if p.Page <= 1 {
		return 1
	}
	return p.Page - 1
}

// NextPage returns the next page number.
func (p *Paginator) NextPage() int {
	return p.Page + 1
}

// TotalPages calculates total pages from total count.
// Returns 0 if total is 0 or negative.
func (p *Paginator) TotalPages(total int64) int {
	if total <= 0 || p.PageSize <= 0 {
		return 0
	}

	pages := total / int64(p.PageSize)
	if total%int64(p.PageSize) > 0 {
		pages++
	}

	// Prevent overflow when converting to int
	const maxInt = int64(^uint(0) >> 1)
	if pages > maxInt {
		return int(maxInt)
	}

	return int(pages)
}

// HasNext returns true if there's a next page.
func (p *Paginator) HasNext(total int64) bool {
	return p.Page < p.TotalPages(total)
}

// IsLastPage returns true if this is the last page.
func (p *Paginator) IsLastPage(total int64) bool {
	totalPages := p.TotalPages(total)
	return totalPages > 0 && p.Page >= totalPages
}

// IsFirstPage returns true if this is the first page.
func (p *Paginator) IsFirstPage() bool {
	return p.Page == 1
}

// IsEmpty returns true if the current page would be empty given the total count.
func (p *Paginator) IsEmpty(total int64) bool {
	return p.Offset() >= total
}

// Clone creates a copy of the paginator.
func (p *Paginator) Clone() *Paginator {
	return &Paginator{
		Page:     p.Page,
		PageSize: p.PageSize,
	}
}

// Clamp adjusts the page number to be within valid range based on total count.
// Returns a new paginator instance.
func (p *Paginator) Clamp(total int64) *Paginator {
	maxPage := p.TotalPages(total)
	if maxPage == 0 {
		maxPage = 1
	}
	if p.Page > maxPage {
		return p.WithPage(maxPage)
	}
	return p
}

// Items returns the range of item indices for this page [start, end).
// Note: end is exclusive.
func (p *Paginator) Items() (start, end int64) {
	start = p.Offset()
	end = start + int64(p.PageSize)
	return
}

// QueryParams returns URL query parameters.
func (p *Paginator) QueryParams() url.Values {
	params := url.Values{}
	params.Set("page", strconv.Itoa(p.Page))
	params.Set("page_size", strconv.Itoa(p.PageSize))
	return params
}

// QueryString returns URL query string.
func (p *Paginator) QueryString() string {
	return p.QueryParams().Encode()
}

// FromRequest parses pagination from HTTP request.
// Returns a paginator with validated default values.
func FromRequest(r *http.Request) *Paginator {
	return FromQuery(r.URL.Query())
}

// FromQuery parses pagination from URL query values.
// Invalid values are ignored and defaults are used instead.
func FromQuery(q url.Values) *Paginator {
	p := New()

	if pageStr := q.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			p = p.WithPage(page)
		}
	}

	if sizeStr := q.Get("page_size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			p = p.WithPageSize(size)
		}
	}

	// Support common alternatives
	if sizeStr := q.Get("limit"); sizeStr != "" && q.Get("page_size") == "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			p = p.WithPageSize(size)
		}
	}

	if sizeStr := q.Get("per_page"); sizeStr != "" && q.Get("page_size") == "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			p = p.WithPageSize(size)
		}
	}

	return p
}

// FromMap parses pagination from a map (useful for JSON APIs).
// Invalid values are ignored and defaults are used instead.
func FromMap(m map[string]any) *Paginator {
	p := New()

	if v, ok := m["page"]; ok {
		if page := extractInt(v); page > 0 {
			p = p.WithPage(page)
		}
	}

	if v, ok := m["page_size"]; ok {
		if size := extractInt(v); size > 0 {
			p = p.WithPageSize(size)
		}
	}

	return p
}

// extractInt is a helper to extract int from various types.
func extractInt(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case int32:
		return int(val)
	case float64:
		return int(val)
	case float32:
		return int(val)
	case string:
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return 0
}
