package paginate

import (
	"net/http"
	"testing"
)

func TestNewRange(t *testing.T) {
	r := NewRange(0, 24)

	if r.Start != 0 {
		t.Errorf("Expected start 0, got %d", r.Start)
	}
	if r.End != 24 {
		t.Errorf("Expected end 24, got %d", r.End)
	}
	if r.Unit != "items" {
		t.Errorf("Expected unit 'items', got '%s'", r.Unit)
	}
}

func TestNewRangeWithUnit(t *testing.T) {
	r := NewRangeWithUnit(100, 199, "bytes")

	if r.Unit != "bytes" {
		t.Errorf("Expected unit 'bytes', got '%s'", r.Unit)
	}
}

func TestRangeSize(t *testing.T) {
	tests := []struct {
		name     string
		start    int64
		end      int64
		expected int64
	}{
		{"Normal range", 0, 24, 25},
		{"Single item", 5, 5, 1},
		{"Large range", 0, 999, 1000},
		{"Invalid range", 10, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRange(tt.start, tt.end)
			if size := r.Size(); size != tt.expected {
				t.Errorf("Expected size %d, got %d", tt.expected, size)
			}
		})
	}
}

func TestRangeValidate(t *testing.T) {
	tests := []struct {
		name      string
		start     int64
		end       int64
		wantError bool
	}{
		{"Valid range", 0, 24, false},
		{"Single item", 5, 5, false},
		{"Negative start", -1, 10, true},
		{"End before start", 10, 5, true},
		{"Both zero", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRange(tt.start, tt.end)
			err := r.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Expected error=%v, got error=%v", tt.wantError, err)
			}
		})
	}
}

func TestRangeSQLClause(t *testing.T) {
	r := NewRange(40, 59)
	expected := "LIMIT 20 OFFSET 40"

	if clause := r.SQLClause(); clause != expected {
		t.Errorf("Expected '%s', got '%s'", expected, clause)
	}
}

func TestRangeHeader(t *testing.T) {
	r := NewRange(0, 24)
	expected := "items=0-24"

	if header := r.Header(); header != expected {
		t.Errorf("Expected '%s', got '%s'", expected, header)
	}
}

func TestContentRangeHeader(t *testing.T) {
	tests := []struct {
		name     string
		start    int64
		end      int64
		total    int64
		expected string
	}{
		{"With total", 0, 24, 100, "items 0-24/100"},
		{"Unknown total", 0, 24, -1, "items 0-24/*"},
		{"Last range", 75, 99, 100, "items 75-99/100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRange(tt.start, tt.end)
			if header := r.ContentRangeHeader(tt.total); header != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, header)
			}
		})
	}
}

func TestParseRangeHeader(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantStart int64
		wantEnd   int64
		wantUnit  string
		wantError bool
	}{
		{"Valid range", "items=0-24", 0, 24, "items", false},
		{"Bytes range", "bytes=100-199", 100, 199, "bytes", false},
		{"Open ended", "items=50-", 50, 69, "items", false}, // 50 + DefaultPageSize - 1
		{"Single digit", "items=0-0", 0, 0, "items", false},
		{"Invalid format", "invalid", 0, 0, "", true},
		{"No equals", "items0-24", 0, 0, "", true},
		{"No dash", "items=024", 0, 0, "", true},
		{"Empty", "", 0, 0, "", false}, // Returns nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ParseRangeHeader(tt.header)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.header == "" {
				if r != nil {
					t.Error("Expected nil range for empty header")
				}
				return
			}

			if r.Start != tt.wantStart {
				t.Errorf("Expected start %d, got %d", tt.wantStart, r.Start)
			}
			if r.End != tt.wantEnd {
				t.Errorf("Expected end %d, got %d", tt.wantEnd, r.End)
			}
			if r.Unit != tt.wantUnit {
				t.Errorf("Expected unit '%s', got '%s'", tt.wantUnit, r.Unit)
			}
		})
	}
}

func TestRangeFromRequest(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantStart int64
		wantEnd   int64
		wantError bool
	}{
		{"With range", "items=0-24", 0, 24, false},
		{"No range", "", 0, 0, false}, // Returns nil
		{"Invalid range", "invalid-format", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			if tt.header != "" {
				req.Header.Set("Range", tt.header)
			}

			r, err := RangeFromRequest(req)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.header == "" {
				if r != nil {
					t.Error("Expected nil for no Range header")
				}
				return
			}

			if r.Start != tt.wantStart {
				t.Errorf("Expected start %d, got %d", tt.wantStart, r.Start)
			}
			if r.End != tt.wantEnd {
				t.Errorf("Expected end %d, got %d", tt.wantEnd, r.End)
			}
		})
	}
}

func TestRangeFromOffsetLimit(t *testing.T) {
	tests := []struct {
		name      string
		offset    int
		limit     int
		wantStart int64
		wantEnd   int64
	}{
		{"Normal", 0, 20, 0, 19},
		{"With offset", 40, 20, 40, 59},
		{"Large offset", 1000, 50, 1000, 1049},
		{"Zero limit", 10, 0, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RangeFromOffsetLimit(tt.offset, tt.limit)

			if r.Start != tt.wantStart {
				t.Errorf("Expected start %d, got %d", tt.wantStart, r.Start)
			}
			if r.End != tt.wantEnd {
				t.Errorf("Expected end %d, got %d", tt.wantEnd, r.End)
			}
		})
	}
}

func TestRangeToPaginator(t *testing.T) {
	tests := []struct {
		name         string
		start        int64
		end          int64
		expectedPage int
		expectedSize int
	}{
		{"First page", 0, 19, 1, 20},
		{"Second page", 20, 39, 2, 20},
		{"Third page", 40, 59, 3, 20},
		{"Custom size", 0, 49, 1, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRange(tt.start, tt.end)
			p := r.ToPaginator()

			if p.Page != tt.expectedPage {
				t.Errorf("Expected page %d, got %d", tt.expectedPage, p.Page)
			}
			if p.PageSize != tt.expectedSize {
				t.Errorf("Expected page size %d, got %d", tt.expectedSize, p.PageSize)
			}
		})
	}
}

func TestNewRangeResponse(t *testing.T) {
	items := []string{"a", "b", "c"}
	r := NewRange(10, 15)
	total := int64(100)

	resp := NewRangeResponse(items, r, total)

	if resp.Start != 10 {
		t.Errorf("Expected start 10, got %d", resp.Start)
	}
	if resp.End != 12 { // 10 + 3 items - 1
		t.Errorf("Expected end 12, got %d", resp.End)
	}
	if resp.Total != total {
		t.Errorf("Expected total %d, got %d", total, resp.Total)
	}
	if resp.Count() != 3 {
		t.Errorf("Expected count 3, got %d", resp.Count())
	}
}

func TestRangeResponseEmpty(t *testing.T) {
	r := NewRange(0, 10)
	resp := NewRangeResponse([]string{}, r, 0)

	if !resp.Empty() {
		t.Error("Empty response should return true for Empty()")
	}
}

func TestRangeResponseContentRange(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		start    int64
		end      int64
		total    int64
		expected string
	}{
		{"With items", []string{"a", "b", "c"}, 0, 10, 100, "items 0-2/100"},
		{"Empty", []string{}, 0, 10, 100, "items */100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRange(tt.start, tt.end)
			resp := NewRangeResponse(tt.items, r, tt.total)

			if cr := resp.ContentRange(); cr != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, cr)
			}
		})
	}
}

func TestRangeResponseHasMore(t *testing.T) {
	tests := []struct {
		name     string
		end      int64
		total    int64
		expected bool
	}{
		{"Has more", 24, 100, true},
		{"Last item", 99, 100, false},
		{"Beyond total", 100, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &RangeResponse[string]{
				Items: []string{"a"},
				End:   tt.end,
				Total: tt.total,
			}

			if hasMore := resp.HasMore(); hasMore != tt.expected {
				t.Errorf("Expected HasMore=%v, got %v", tt.expected, hasMore)
			}
		})
	}
}
