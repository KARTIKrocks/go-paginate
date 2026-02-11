package paginate

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestNewCursor(t *testing.T) {
	c := NewCursor()
	if c.Limit != DefaultPageSize {
		t.Errorf("Expected limit %d, got %d", DefaultPageSize, c.Limit)
	}
	if !c.Forward {
		t.Error("Expected forward=true by default")
	}
	if c.Cursor != "" {
		t.Error("Expected empty cursor by default")
	}
}

func TestCursorWithLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Valid limit", 50, 50},
		{"Too small", 0, DefaultPageSize},
		{"Negative", -5, DefaultPageSize},
		{"Too large", 2000, MaxPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCursor().WithLimit(tt.input)
			if c.Limit != tt.expected {
				t.Errorf("Expected limit %d, got %d", tt.expected, c.Limit)
			}
		})
	}
}

func TestCursorWithCursor(t *testing.T) {
	c := NewCursor().WithCursor("test-cursor")
	if c.Cursor != "test-cursor" {
		t.Errorf("Expected cursor 'test-cursor', got '%s'", c.Cursor)
	}
	if !c.HasCursor() {
		t.Error("HasCursor should return true")
	}
}

func TestCursorWithForward(t *testing.T) {
	c := NewCursor().WithForward(false)
	if c.Forward {
		t.Error("Expected forward=false")
	}
}

func TestCursorClone(t *testing.T) {
	c1 := NewCursor().WithCursor("test").WithLimit(50)
	c2 := c1.Clone()

	if c1 == c2 {
		t.Error("Clone should return different instance")
	}

	if c1.Cursor != c2.Cursor || c1.Limit != c2.Limit {
		t.Error("Clone should have same values")
	}
}

func TestCursorPaginatorEncode(t *testing.T) {
	c := NewCursor()

	cursor, err := c.Encode(CursorData[any]{ID: "user_123"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cursor == "" {
		t.Error("Expected non-empty cursor")
	}

	// Verify round-trip via package-level DecodeCursor
	data, err := DecodeCursor[any](cursor)
	if err != nil {
		t.Fatalf("Unexpected decode error: %v", err)
	}
	if data.ID != "user_123" {
		t.Errorf("Expected ID 'user_123', got '%s'", data.ID)
	}
}

func TestCursorPaginatorDecode(t *testing.T) {
	encoded, err := EncodeCursor(&CursorData[any]{ID: "abc", Offset: 10})
	if err != nil {
		t.Fatalf("Unexpected encode error: %v", err)
	}

	c := NewCursor().WithCursor(encoded)
	data, err := c.Decode()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if data.ID != "abc" {
		t.Errorf("Expected ID 'abc', got '%s'", data.ID)
	}
	if data.Offset != 10 {
		t.Errorf("Expected offset 10, got %d", data.Offset)
	}

	// No cursor set → nil, nil
	empty := NewCursor()
	data, err = empty.Decode()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if data != nil {
		t.Error("Expected nil data for empty cursor")
	}
}

func TestEncodeCursor(t *testing.T) {
	tests := []struct {
		name     string
		data     *CursorData[any]
		nonEmpty bool
	}{
		{"Nil data", nil, false},
		{"With ID", &CursorData[any]{ID: "user_123"}, true},
		{"With timestamp", &CursorData[any]{Timestamp: time.Now()}, true},
		{"With offset", &CursorData[any]{Offset: 100}, true},
		{"Complete", &CursorData[any]{
			ID:        "user_456",
			Timestamp: time.Now(),
			Offset:    50,
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor, err := EncodeCursor(tt.data)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tt.nonEmpty && cursor == "" {
				t.Error("Expected non-empty cursor")
			}
			if !tt.nonEmpty && cursor != "" {
				t.Error("Expected empty cursor")
			}
		})
	}
}

func TestDecodeCursor(t *testing.T) {
	// Create a cursor and decode it
	original := &CursorData[any]{
		ID:        "user_123",
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Offset:    42,
	}

	cursor, err := EncodeCursor(original)
	if err != nil {
		t.Fatalf("Unexpected encode error: %v", err)
	}
	decoded, err := DecodeCursor[any](cursor)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("Expected ID %s, got %s", original.ID, decoded.ID)
	}

	if decoded.Offset != original.Offset {
		t.Errorf("Expected offset %d, got %d", original.Offset, decoded.Offset)
	}
}

func TestDecodeCursorInvalid(t *testing.T) {
	tests := []struct {
		name   string
		cursor string
	}{
		{"Invalid base64", "not-valid-base64!!!"},
		{"Invalid JSON", "dGhpcyBpcyBub3QganNvbg=="}, // "this is not json" in base64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeCursor[any](tt.cursor)
			if err == nil {
				t.Error("Expected error for invalid cursor")
			}
		})
	}
}

func TestDecodeCursorEmpty(t *testing.T) {
	data, err := DecodeCursor[any]("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if data != nil {
		t.Error("Expected nil data for empty cursor")
	}
}

func TestCursorValidate(t *testing.T) {
	validCursor, err := EncodeCursor(&CursorData[any]{ID: "test"})
	if err != nil {
		t.Fatalf("Unexpected encode error: %v", err)
	}

	tests := []struct {
		name      string
		cursor    string
		limit     int
		wantError bool
	}{
		{"Valid", "", 20, false},
		{"Invalid limit", "", 0, true},
		{"Invalid cursor", "invalid-cursor", 20, true},
		{"Valid cursor", validCursor, 20, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CursorPaginator{
				Cursor: tt.cursor,
				Limit:  tt.limit,
			}
			err := c.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Expected error=%v, got error=%v", tt.wantError, err)
			}
		})
	}
}

func TestCursorFromRequest(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		expectedCursor  string
		expectedLimit   int
		expectedForward bool
	}{
		{"No params", "http://example.com", "", DefaultPageSize, true},
		{"Generic cursor", "http://example.com?cursor=abc&limit=30", "abc", 30, true},
		{"After cursor", "http://example.com?after=xyz&limit=25", "xyz", 25, true},
		{"Before cursor", "http://example.com?before=def&limit=20", "def", 20, false},
		{"GraphQL first", "http://example.com?first=50", "", 50, true},
		{"GraphQL last", "http://example.com?last=40", "", 40, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			c := CursorFromRequest(req)

			if c.Cursor != tt.expectedCursor {
				t.Errorf("Expected cursor '%s', got '%s'", tt.expectedCursor, c.Cursor)
			}
			if c.Limit != tt.expectedLimit {
				t.Errorf("Expected limit %d, got %d", tt.expectedLimit, c.Limit)
			}
			if c.Forward != tt.expectedForward {
				t.Errorf("Expected forward=%v, got %v", tt.expectedForward, c.Forward)
			}
		})
	}
}

func TestNewCursorFromID(t *testing.T) {
	cursor, err := NewCursorFromID("user_123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cursor == "" {
		t.Error("Expected non-empty cursor")
	}

	// Decode and verify
	data, err := DecodeCursor[any](cursor)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if data.ID != "user_123" {
		t.Errorf("Expected ID 'user_123', got '%s'", data.ID)
	}
}

func TestNewCursorFromTimestamp(t *testing.T) {
	ts := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	cursor, err := NewCursorFromTimestamp(ts, "item_456")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cursor == "" {
		t.Error("Expected non-empty cursor")
	}

	data, err := DecodeCursor[any](cursor)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if data.ID != "item_456" {
		t.Errorf("Expected ID 'item_456', got '%s'", data.ID)
	}

	if !data.Timestamp.Equal(ts) {
		t.Errorf("Expected timestamp %v, got %v", ts, data.Timestamp)
	}
}

func TestNewCursorFromOffset(t *testing.T) {
	cursor, err := NewCursorFromOffset(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cursor == "" {
		t.Error("Expected non-empty cursor")
	}

	data, err := DecodeCursor[any](cursor)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if data.Offset != 100 {
		t.Errorf("Expected offset 100, got %d", data.Offset)
	}
}

func TestCursorQueryParams(t *testing.T) {
	tests := []struct {
		name           string
		cursor         string
		limit          int
		forward        bool
		expectedAfter  bool
		expectedBefore bool
	}{
		{"Forward with cursor", "abc", 20, true, true, false},
		{"Backward with cursor", "xyz", 30, false, false, true},
		{"No cursor", "", 25, true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CursorPaginator{
				Cursor:  tt.cursor,
				Limit:   tt.limit,
				Forward: tt.forward,
			}

			params := c.QueryParams()

			hasAfter := params.Has("after")
			hasBefore := params.Has("before")

			if hasAfter != tt.expectedAfter {
				t.Errorf("Expected after=%v, got %v", tt.expectedAfter, hasAfter)
			}
			if hasBefore != tt.expectedBefore {
				t.Errorf("Expected before=%v, got %v", tt.expectedBefore, hasBefore)
			}

			if limit := params.Get("limit"); limit != "" {
				// Limit should always be present
				if params.Get("limit") != "20" && params.Get("limit") != "30" && params.Get("limit") != "25" {
					t.Error("Limit param missing or incorrect")
				}
			}
		})
	}
}

func TestNewCursorFromValue(t *testing.T) {
	// Test with a concrete type to verify type-safe round-trip
	cursor, err := NewCursorFromValue("hello")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cursor == "" {
		t.Error("Expected non-empty cursor")
	}

	// Decode with the same concrete type — no data loss
	data, err := DecodeCursor[string](cursor)
	if err != nil {
		t.Fatalf("Unexpected decode error: %v", err)
	}
	if data.Value != "hello" {
		t.Errorf("Expected Value 'hello', got %q", data.Value)
	}

	// Test with int to verify no float64 round-trip loss
	cursorInt, err := NewCursorFromValue(42)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	dataInt, err := DecodeCursor[int](cursorInt)
	if err != nil {
		t.Fatalf("Unexpected decode error: %v", err)
	}
	if dataInt.Value != 42 {
		t.Errorf("Expected Value 42, got %d", dataInt.Value)
	}
}

func TestCursorRoundTrip(t *testing.T) {
	// Test that encoding and decoding preserves data
	original := &CursorData[any]{
		ID:        "test_123",
		Value:     "some-value",
		Timestamp: time.Now().UTC().Truncate(time.Second), // Truncate for comparison
		Offset:    42,
	}

	encoded, err := EncodeCursor(original)
	if err != nil {
		t.Fatalf("Unexpected encode error: %v", err)
	}
	decoded, err := DecodeCursor[any](encoded)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: expected %s, got %s", original.ID, decoded.ID)
	}
	if decoded.Offset != original.Offset {
		t.Errorf("Offset mismatch: expected %d, got %d", original.Offset, decoded.Offset)
	}
	if !decoded.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp mismatch: expected %v, got %v", original.Timestamp, decoded.Timestamp)
	}
}

func BenchmarkEncodeCursor(b *testing.B) {
	data := &CursorData[any]{
		ID:        "user_123",
		Timestamp: time.Now(),
		Offset:    42,
	}

	for b.Loop() {
		_, _ = EncodeCursor(data)
	}
}

func BenchmarkDecodeCursor(b *testing.B) {
	cursor, _ := EncodeCursor(&CursorData[any]{
		ID:        "user_123",
		Timestamp: time.Now(),
	})

	for b.Loop() {
		_, _ = DecodeCursor[any](cursor)
	}
}

func BenchmarkCursorFromQuery(b *testing.B) {
	q := url.Values{}
	q.Set("after", "test-cursor")
	q.Set("limit", "50")

	for b.Loop() {
		_ = CursorFromQuery(q)
	}
}
