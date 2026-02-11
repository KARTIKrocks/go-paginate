// Package paginate provides comprehensive pagination utilities for Go applications.
//
// This package supports multiple pagination strategies to fit different use cases:
//
//   - Offset-based pagination: Traditional page numbers (page 1, 2, 3...)
//   - Cursor-based pagination: Efficient forward/backward navigation using opaque cursors
//   - Range-based pagination: HTTP Range header style pagination
//   - GraphQL connections: Relay-style connection pagination
//
// # Offset Pagination
//
// Use offset pagination for traditional page-based navigation:
//
//	p := paginate.New().
//		WithPage(2).
//		WithPageSize(25)
//
//	// Use with SQL
//	offset := p.Offset() // 25
//	limit := p.Limit()   // 25
//
//	// Create response
//	page := paginate.NewPage(items, totalCount, p)
//
// # Cursor Pagination
//
// Use cursor pagination for efficient, consistent results:
//
//	c := paginate.NewCursor().WithLimit(20)
//	cursor, err := c.Encode(paginate.CursorData[any]{
//		ID: "user_123",
//		Timestamp: time.Now(),
//	})
//
//	response := paginate.NewCursorPage(items, 20, nextCursor, "", hasMore)
//
// # HTTP Integration
//
// Parse pagination from HTTP requests:
//
//	// Offset pagination
//	p := paginate.FromRequest(r)
//
//	// Cursor pagination
//	c := paginate.CursorFromRequest(r)
//
//	// Range pagination
//	rng, err := paginate.RangeFromRequest(r)
//
// # Thread Safety
//
// Paginator instances are safe to read concurrently. The With* methods
// return new instances, making them safe for concurrent modifications:
//
//	p := paginate.New()
//	p1 := p.WithPage(1) // New instance
//	p2 := p.WithPage(2) // Different new instance
//
// # Validation
//
// Always validate pagination parameters from untrusted input:
//
//	p := paginate.FromRequest(r)
//	if err := p.Validate(); err != nil {
//		// Handle invalid pagination
//	}
//
// # Configuration
//
// Default values can be customized:
//
//	const (
//		DefaultPage     = 1
//		DefaultPageSize = 20
//		MaxPageSize     = 1000
//		MinPageSize     = 1
//	)
//
// These constants ensure safe pagination behavior and prevent abuse.
package paginate
