package paginate

import (
	"testing"
)

func TestNewPage(t *testing.T) {
	items := []string{"a", "b", "c"}
	p := NewFromValues(2, 10)
	total := int64(50)

	page := NewPage(items, total, p)

	if page.Total != total {
		t.Errorf("Expected total %d, got %d", total, page.Total)
	}
	if page.Page != 2 {
		t.Errorf("Expected page 2, got %d", page.Page)
	}
	if page.PageSize != 10 {
		t.Errorf("Expected page size 10, got %d", page.PageSize)
	}
	if page.TotalPages != 5 {
		t.Errorf("Expected 5 total pages, got %d", page.TotalPages)
	}
	if !page.HasPrev {
		t.Error("Expected HasPrev to be true")
	}
	if !page.HasNext {
		t.Error("Expected HasNext to be true")
	}
	if page.Count() != 3 {
		t.Errorf("Expected count 3, got %d", page.Count())
	}
}

func TestPageEmpty(t *testing.T) {
	emptyPage := NewPage([]string{}, 0, New())
	if !emptyPage.Empty() {
		t.Error("Empty page should return true for Empty()")
	}

	nonEmptyPage := NewPage([]string{"a"}, 1, New())
	if nonEmptyPage.Empty() {
		t.Error("Non-empty page should return false for Empty()")
	}
}

func TestNewCursorPage(t *testing.T) {
	items := []int{1, 2, 3}
	nextCursor := "next-cursor"
	prevCursor := "prev-cursor"

	page := NewCursorPage(items, 10, nextCursor, prevCursor, true)

	if page.NextCursor != nextCursor {
		t.Errorf("Expected next cursor %s, got %s", nextCursor, page.NextCursor)
	}
	if page.PrevCursor != prevCursor {
		t.Errorf("Expected prev cursor %s, got %s", prevCursor, page.PrevCursor)
	}
	if !page.HasMore {
		t.Error("Expected HasMore to be true")
	}
	if page.Limit != 10 {
		t.Errorf("Expected limit 10, got %d", page.Limit)
	}
	if page.Count() != 3 {
		t.Errorf("Expected count 3, got %d", page.Count())
	}
}

func TestNewCursorPageSimple(t *testing.T) {
	items := []int{1, 2, 3}
	nextCursor := "cursor-123"

	page := NewCursorPageSimple(items, 20, nextCursor)

	if page.NextCursor != nextCursor {
		t.Errorf("Expected cursor %s, got %s", nextCursor, page.NextCursor)
	}
	if !page.HasMore {
		t.Error("Expected HasMore to be true when cursor is set")
	}

	// Test without cursor
	pageNoCursor := NewCursorPageSimple(items, 20, "")
	if pageNoCursor.HasMore {
		t.Error("Expected HasMore to be false when cursor is empty")
	}
}

func TestCursorPageEmpty(t *testing.T) {
	emptyPage := NewCursorPageSimple([]int{}, 10, "")
	if !emptyPage.Empty() {
		t.Error("Empty cursor page should return true for Empty()")
	}
}

type testItem struct {
	ID   string
	Name string
}

func TestNewConnection(t *testing.T) {
	items := []testItem{
		{ID: "1", Name: "First"},
		{ID: "2", Name: "Second"},
		{ID: "3", Name: "Third"},
	}

	cursorFn := func(item testItem) string {
		return NewCursorFromID(item.ID)
	}

	conn := NewConnection(items, cursorFn, false, true, 100)

	if len(conn.Edges) != 3 {
		t.Errorf("Expected 3 edges, got %d", len(conn.Edges))
	}

	if conn.PageInfo.HasPreviousPage {
		t.Error("Expected HasPreviousPage to be false")
	}
	if !conn.PageInfo.HasNextPage {
		t.Error("Expected HasNextPage to be true")
	}
	if conn.TotalCount != 100 {
		t.Errorf("Expected total count 100, got %d", conn.TotalCount)
	}

	// Check cursors
	if conn.PageInfo.StartCursor == "" {
		t.Error("Expected non-empty start cursor")
	}
	if conn.PageInfo.EndCursor == "" {
		t.Error("Expected non-empty end cursor")
	}

	// Verify edges
	for i, edge := range conn.Edges {
		if edge.Node.ID != items[i].ID {
			t.Errorf("Edge %d: expected ID %s, got %s", i, items[i].ID, edge.Node.ID)
		}
		if edge.Cursor == "" {
			t.Errorf("Edge %d: expected non-empty cursor", i)
		}
	}
}

func TestConnectionEmpty(t *testing.T) {
	conn := NewConnection([]testItem{}, func(item testItem) string {
		return item.ID
	}, false, false, 0)

	if !conn.Empty() {
		t.Error("Empty connection should return true for Empty()")
	}
	if conn.Count() != 0 {
		t.Errorf("Expected count 0, got %d", conn.Count())
	}
}

func TestConnectionNodes(t *testing.T) {
	items := []testItem{
		{ID: "1", Name: "First"},
		{ID: "2", Name: "Second"},
	}

	conn := NewConnection(items, func(item testItem) string {
		return item.ID
	}, false, false, 2)

	nodes := conn.Nodes()

	if len(nodes) != len(items) {
		t.Errorf("Expected %d nodes, got %d", len(items), len(nodes))
	}

	for i, node := range nodes {
		if node.ID != items[i].ID {
			t.Errorf("Node %d: expected ID %s, got %s", i, items[i].ID, node.ID)
		}
	}
}

func TestBuildLinkHeader(t *testing.T) {
	p := NewFromValues(3, 20)
	baseURL := "https://api.example.com/users"
	total := int64(100)

	links := BuildLinkHeader(baseURL, p, total)

	if links.First == "" {
		t.Error("Expected non-empty First link")
	}
	if links.Last == "" {
		t.Error("Expected non-empty Last link")
	}
	if links.Prev == "" {
		t.Error("Expected non-empty Prev link (page 3 has prev)")
	}
	if links.Next == "" {
		t.Error("Expected non-empty Next link (page 3 has next)")
	}
}

func TestBuildLinkHeaderEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		total      int64
		expectPrev bool
		expectNext bool
	}{
		{"First page", 1, 100, false, true},
		{"Last page", 5, 100, true, false},
		{"Middle page", 3, 100, true, true},
		{"Only page", 1, 20, false, false},
		{"Empty results", 1, 0, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFromValues(tt.page, 20)
			links := BuildLinkHeader("https://example.com", p, tt.total)

			hasPrev := links.Prev != ""
			hasNext := links.Next != ""

			if hasPrev != tt.expectPrev {
				t.Errorf("Expected Prev=%v, got %v", tt.expectPrev, hasPrev)
			}
			if hasNext != tt.expectNext {
				t.Errorf("Expected Next=%v, got %v", tt.expectNext, hasNext)
			}
		})
	}
}

func TestLinkHeaderString(t *testing.T) {
	links := &LinkHeader{
		First: "https://example.com?page=1",
		Prev:  "https://example.com?page=2",
		Next:  "https://example.com?page=4",
		Last:  "https://example.com?page=10",
	}

	str := links.String()

	// Should contain all links
	if str == "" {
		t.Error("Expected non-empty Link header string")
	}

	// Check format (should have rel= parts)
	expectedParts := []string{
		`rel="first"`,
		`rel="prev"`,
		`rel="next"`,
		`rel="last"`,
	}

	for _, part := range expectedParts {
		if !contains(str, part) {
			t.Errorf("Expected Link header to contain %s", part)
		}
	}
}

func TestLinkHeaderStringEmpty(t *testing.T) {
	links := &LinkHeader{}
	str := links.String()

	if str != "" {
		t.Error("Expected empty string for empty LinkHeader")
	}
}

func TestLinkHeaderStringPartial(t *testing.T) {
	// Only Next link
	links := &LinkHeader{
		Next: "https://example.com?page=2",
	}

	str := links.String()

	if !contains(str, `rel="next"`) {
		t.Error("Expected Link header to contain next rel")
	}
	if contains(str, `rel="prev"`) {
		t.Error("Expected Link header to not contain prev rel")
	}
}

func TestPageCount(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	page := NewPage(items, 100, New())

	if page.Count() != 5 {
		t.Errorf("Expected count 5, got %d", page.Count())
	}
}

func TestConnectionCount(t *testing.T) {
	items := []testItem{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	}

	conn := NewConnection(items, func(item testItem) string {
		return item.ID
	}, false, false, 10)

	if conn.Count() != 3 {
		t.Errorf("Expected count 3, got %d", conn.Count())
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
