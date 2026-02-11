package paginate

import (
	"fmt"
	"net/url"
)

// Page represents a paginated response using offset pagination.
type Page[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasPrev    bool  `json:"has_prev"`
	HasNext    bool  `json:"has_next"`
}

// NewPage creates a new paginated response.
func NewPage[T any](items []T, total int64, p *Paginator) *Page[T] {
	totalPages := p.TotalPages(total)

	return &Page[T]{
		Items:      items,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
		HasPrev:    p.HasPrevious(),
		HasNext:    p.HasNext(total),
	}
}

// Empty returns true if the page has no items.
func (p *Page[T]) Empty() bool {
	return len(p.Items) == 0
}

// Count returns the number of items in this page.
func (p *Page[T]) Count() int {
	return len(p.Items)
}

// CursorPage represents a paginated response using cursor pagination.
type CursorPage[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
}

// NewCursorPage creates a new cursor-paginated response.
func NewCursorPage[T any](items []T, limit int, nextCursor, prevCursor string, hasMore bool) *CursorPage[T] {
	return &CursorPage[T]{
		Items:      items,
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasMore:    hasMore,
		Limit:      limit,
	}
}

// NewCursorPageSimple creates a simple cursor page with just a next cursor.
// This is useful when you only need forward pagination.
func NewCursorPageSimple[T any](items []T, limit int, nextCursor string) *CursorPage[T] {
	return &CursorPage[T]{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
		Limit:      limit,
	}
}

// Empty returns true if the page has no items.
func (p *CursorPage[T]) Empty() bool {
	return len(p.Items) == 0
}

// Count returns the number of items in this page.
func (p *CursorPage[T]) Count() int {
	return len(p.Items)
}

// Edge represents a GraphQL-style edge containing a node and cursor.
type Edge[T any] struct {
	Node   T      `json:"node"`
	Cursor string `json:"cursor"`
}

// PageInfo represents GraphQL-style page info.
type PageInfo struct {
	HasPreviousPage bool   `json:"has_previous_page"`
	HasNextPage     bool   `json:"has_next_page"`
	StartCursor     string `json:"start_cursor,omitempty"`
	EndCursor       string `json:"end_cursor,omitempty"`
}

// Connection represents a GraphQL-style connection.
type Connection[T any] struct {
	Edges      []Edge[T] `json:"edges"`
	PageInfo   PageInfo  `json:"page_info"`
	TotalCount int64     `json:"total_count,omitempty"`
}

// NewConnection creates a GraphQL-style connection.
// The cursorFn is called for each item to generate its cursor.
// Set hasPrev and hasNext based on your pagination logic.
func NewConnection[T any](
	items []T,
	cursorFn func(T) string,
	hasPrev, hasNext bool,
	total int64,
) *Connection[T] {
	edges := make([]Edge[T], len(items))
	var startCursor, endCursor string

	for i, item := range items {
		cursor := cursorFn(item)
		edges[i] = Edge[T]{
			Node:   item,
			Cursor: cursor,
		}
		if i == 0 {
			startCursor = cursor
		}
		if i == len(items)-1 {
			endCursor = cursor
		}
	}

	return &Connection[T]{
		Edges: edges,
		PageInfo: PageInfo{
			HasPreviousPage: hasPrev,
			HasNextPage:     hasNext,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: total,
	}
}

// Empty returns true if the connection has no edges.
func (c *Connection[T]) Empty() bool {
	return len(c.Edges) == 0
}

// Nodes extracts and returns just the nodes from the connection edges.
func (c *Connection[T]) Nodes() []T {
	nodes := make([]T, len(c.Edges))
	for i, edge := range c.Edges {
		nodes[i] = edge.Node
	}
	return nodes
}

// Count returns the number of edges in the connection.
func (c *Connection[T]) Count() int {
	return len(c.Edges)
}

// LinkHeader represents pagination links for HTTP Link header (RFC 5988).
type LinkHeader struct {
	First string `json:"first,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
	Last  string `json:"last,omitempty"`
}

// BuildLinkHeader builds pagination links for a given base URL.
// This creates RFC 5988 compliant Link headers for RESTful APIs.
func BuildLinkHeader(baseURL string, p *Paginator, total int64) *LinkHeader {
	totalPages := p.TotalPages(total)
	if totalPages == 0 {
		return &LinkHeader{}
	}

	header := &LinkHeader{}

	// First page
	first := p.WithPage(1)
	header.First = buildURL(baseURL, first.QueryParams())

	// Last page
	last := p.WithPage(totalPages)
	header.Last = buildURL(baseURL, last.QueryParams())

	// Previous page
	if p.HasPrevious() {
		prev := p.WithPage(p.PreviousPage())
		header.Prev = buildURL(baseURL, prev.QueryParams())
	}

	// Next page
	if p.HasNext(total) {
		next := p.WithPage(p.NextPage())
		header.Next = buildURL(baseURL, next.QueryParams())
	}

	return header
}

// buildURL combines base URL with query parameters.
func buildURL(baseURL string, params url.Values) string {
	if len(params) == 0 {
		return baseURL
	}
	return baseURL + "?" + params.Encode()
}

// String returns the Link header string in RFC 5988 format.
// Example: <url>; rel="first", <url>; rel="next"
func (h *LinkHeader) String() string {
	var links []string

	if h.First != "" {
		links = append(links, fmt.Sprintf(`<%s>; rel="first"`, h.First))
	}
	if h.Prev != "" {
		links = append(links, fmt.Sprintf(`<%s>; rel="prev"`, h.Prev))
	}
	if h.Next != "" {
		links = append(links, fmt.Sprintf(`<%s>; rel="next"`, h.Next))
	}
	if h.Last != "" {
		links = append(links, fmt.Sprintf(`<%s>; rel="last"`, h.Last))
	}

	result := ""
	for i, link := range links {
		if i > 0 {
			result += ", "
		}
		result += link
	}

	return result
}

// SetHeader sets the Link header on an HTTP response.
func (h *LinkHeader) SetHeader(header func(key, value string)) {
	if linkStr := h.String(); linkStr != "" {
		header("Link", linkStr)
	}
}
