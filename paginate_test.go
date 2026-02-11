package paginate

import (
	"math"
	"net/http"
	"net/url"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	p := New()
	if p.Page != DefaultPage {
		t.Errorf("Expected page %d, got %d", DefaultPage, p.Page)
	}
	if p.PageSize != DefaultPageSize {
		t.Errorf("Expected page size %d, got %d", DefaultPageSize, p.PageSize)
	}
}

func TestWithPage(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Valid page", 5, 5},
		{"Zero page", 0, DefaultPage},
		{"Negative page", -1, DefaultPage},
		{"Large page", 1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New().WithPage(tt.input)
			if p.Page != tt.expected {
				t.Errorf("Expected page %d, got %d", tt.expected, p.Page)
			}
		})
	}
}

func TestWithPageSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Valid size", 50, 50},
		{"Too small", 0, DefaultPageSize},
		{"Negative size", -5, DefaultPageSize},
		{"Too large", 2000, MaxPageSize},
		{"Max size", MaxPageSize, MaxPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New().WithPageSize(tt.input)
			if p.PageSize != tt.expected {
				t.Errorf("Expected page size %d, got %d", tt.expected, p.PageSize)
			}
		})
	}
}

func TestOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		expected int64
	}{
		{"First page", 1, 20, 0},
		{"Second page", 2, 20, 20},
		{"Third page", 3, 20, 40},
		{"Large page", 100, 50, 4950},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFromValues(tt.page, tt.pageSize)
			if offset := p.Offset(); offset != tt.expected {
				t.Errorf("Expected offset %d, got %d", tt.expected, offset)
			}
		})
	}
}

func TestOffsetOverflow(t *testing.T) {
	// Test that offset calculation doesn't overflow
	p := NewFromValues(math.MaxInt32/2, math.MaxInt32/2)
	offset := p.Offset()
	// Should not panic and should return a valid int64
	if offset < 0 {
		t.Error("Offset overflowed to negative")
	}
}

func TestTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		pageSize int
		expected int
	}{
		{"Exact division", 100, 20, 5},
		{"With remainder", 101, 20, 6},
		{"Less than page size", 15, 20, 1},
		{"Zero total", 0, 20, 0},
		{"Negative total", -10, 20, 0},
		{"Large total", 10000, 50, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewWithSize(tt.pageSize)
			if pages := p.TotalPages(tt.total); pages != tt.expected {
				t.Errorf("Expected %d pages, got %d", tt.expected, pages)
			}
		})
	}
}

func TestHasNext(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		total    int64
		expected bool
	}{
		{"Has next", 1, 20, 100, true},
		{"Last page", 5, 20, 100, false},
		{"Beyond last page", 10, 20, 100, false},
		{"Empty result", 1, 20, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFromValues(tt.page, tt.pageSize)
			if hasNext := p.HasNext(tt.total); hasNext != tt.expected {
				t.Errorf("Expected HasNext=%v, got %v", tt.expected, hasNext)
			}
		})
	}
}

func TestHasPrevious(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		expected bool
	}{
		{"First page", 1, false},
		{"Second page", 2, true},
		{"Later page", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFromValues(tt.page, 20)
			if hasPrev := p.HasPrevious(); hasPrev != tt.expected {
				t.Errorf("Expected HasPrevious=%v, got %v", tt.expected, hasPrev)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		pageSize  int
		wantError bool
	}{
		{"Valid", 1, 20, false},
		{"Invalid page", 0, 20, true},
		{"Invalid page size (too small)", 1, 0, true},
		{"Invalid page size (too large)", 1, 2000, true},
		{"Both valid edges", 1, 1, false},
		{"Max valid", 1, MaxPageSize, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Paginator{Page: tt.page, PageSize: tt.pageSize}
			err := p.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Expected error=%v, got error=%v", tt.wantError, err)
			}
		})
	}
}

func TestClone(t *testing.T) {
	p1 := NewFromValues(5, 50)
	p2 := p1.Clone()

	if p1 == p2 {
		t.Error("Clone should return a different instance")
	}

	if p1.Page != p2.Page || p1.PageSize != p2.PageSize {
		t.Error("Clone should have same values")
	}

	// Modify clone
	p2.Page = 10
	if p1.Page == p2.Page {
		t.Error("Modifying clone should not affect original")
	}
}

func TestThreadSafety(t *testing.T) {
	p := New()
	var wg sync.WaitGroup

	// Concurrent reads should be safe
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			// Create new instances with different pages
			newP := p.WithPage(page)
			_ = newP.Offset()
			_ = newP.HasNext(1000)
		}(i)
	}

	wg.Wait()
}

func TestFromRequest(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectedPage int
		expectedSize int
	}{
		{"No params", "http://example.com", DefaultPage, DefaultPageSize},
		{"With page", "http://example.com?page=5", 5, DefaultPageSize},
		{"With page_size", "http://example.com?page_size=50", DefaultPage, 50},
		{"Both params", "http://example.com?page=3&page_size=25", 3, 25},
		{"Invalid page", "http://example.com?page=abc", DefaultPage, DefaultPageSize},
		{"Negative values", "http://example.com?page=-1&page_size=-10", DefaultPage, DefaultPageSize},
		{"Limit param", "http://example.com?limit=30", DefaultPage, 30},
		{"Per page param", "http://example.com?per_page=40", DefaultPage, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			p := FromRequest(req)

			if p.Page != tt.expectedPage {
				t.Errorf("Expected page %d, got %d", tt.expectedPage, p.Page)
			}
			if p.PageSize != tt.expectedSize {
				t.Errorf("Expected page size %d, got %d", tt.expectedSize, p.PageSize)
			}
		})
	}
}

func TestFromMap(t *testing.T) {
	tests := []struct {
		name         string
		input        map[string]any
		expectedPage int
		expectedSize int
	}{
		{"Empty map", map[string]any{}, DefaultPage, DefaultPageSize},
		{"Int values", map[string]any{"page": 5, "page_size": 50}, 5, 50},
		{"Int64 values", map[string]any{"page": int64(3), "page_size": int64(25)}, 3, 25},
		{"Float values", map[string]any{"page": 2.0, "page_size": 30.0}, 2, 30},
		{"String values", map[string]any{"page": "4", "page_size": "40"}, 4, 40},
		{"Invalid strings", map[string]any{"page": "abc", "page_size": "xyz"}, DefaultPage, DefaultPageSize},
		{"Mixed valid", map[string]any{"page": 10, "page_size": "50"}, 10, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := FromMap(tt.input)

			if p.Page != tt.expectedPage {
				t.Errorf("Expected page %d, got %d", tt.expectedPage, p.Page)
			}
			if p.PageSize != tt.expectedSize {
				t.Errorf("Expected page size %d, got %d", tt.expectedSize, p.PageSize)
			}
		})
	}
}

func TestSQLClause(t *testing.T) {
	p := NewFromValues(3, 20)
	expected := "LIMIT 20 OFFSET 40"
	if clause := p.SQLClause(); clause != expected {
		t.Errorf("Expected '%s', got '%s'", expected, clause)
	}
}

func TestSQLClauseMySQL(t *testing.T) {
	p := NewFromValues(3, 20)
	expected := "LIMIT 40, 20"
	if clause := p.SQLClauseMySQL(); clause != expected {
		t.Errorf("Expected '%s', got '%s'", expected, clause)
	}
}

func TestQueryParams(t *testing.T) {
	p := NewFromValues(5, 50)
	params := p.QueryParams()

	if params.Get("page") != "5" {
		t.Errorf("Expected page=5, got %s", params.Get("page"))
	}
	if params.Get("page_size") != "50" {
		t.Errorf("Expected page_size=50, got %s", params.Get("page_size"))
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		total    int64
		expected int
	}{
		{"Within range", 5, 1000, 5},
		{"Beyond total", 100, 50, 3}, // 50 items / 20 per page = 3 pages
		{"Zero total", 5, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFromValues(tt.page, 20).Clamp(tt.total)
			if p.Page != tt.expected {
				t.Errorf("Expected page %d, got %d", tt.expected, p.Page)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		total    int64
		expected bool
	}{
		{"Not empty", 1, 20, 100, false},
		{"Last page not empty", 5, 20, 100, false},
		{"Beyond total", 10, 20, 50, true},
		{"Exactly at total", 6, 20, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFromValues(tt.page, tt.pageSize)
			if isEmpty := p.IsEmpty(tt.total); isEmpty != tt.expected {
				t.Errorf("Expected IsEmpty=%v, got %v", tt.expected, isEmpty)
			}
		})
	}
}

func TestItems(t *testing.T) {
	p := NewFromValues(3, 20)
	start, end := p.Items()

	if start != 40 {
		t.Errorf("Expected start=40, got %d", start)
	}
	if end != 60 {
		t.Errorf("Expected end=60, got %d", end)
	}
}

func TestIsFirstPage(t *testing.T) {
	if !New().IsFirstPage() {
		t.Error("Default paginator should be on first page")
	}
	if New().WithPage(2).IsFirstPage() {
		t.Error("Page 2 should not be first page")
	}
}

func TestIsLastPage(t *testing.T) {
	p := NewFromValues(5, 20)
	if !p.IsLastPage(100) {
		t.Error("Page 5 with 100 items should be last page")
	}
	if p.IsLastPage(200) {
		t.Error("Page 5 with 200 items should not be last page")
	}
}

func BenchmarkOffset(b *testing.B) {
	p := NewFromValues(100, 50)
	for i := 0; i < b.N; i++ {
		_ = p.Offset()
	}
}

func BenchmarkFromQuery(b *testing.B) {
	q := url.Values{}
	q.Set("page", "5")
	q.Set("page_size", "50")

	for i := 0; i < b.N; i++ {
		_ = FromQuery(q)
	}
}

func BenchmarkWithPage(b *testing.B) {
	p := New()
	for i := 0; i < b.N; i++ {
		_ = p.WithPage(i % 100)
	}
}
