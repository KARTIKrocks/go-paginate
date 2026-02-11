# go-paginate

[![Go Reference](https://pkg.go.dev/badge/github.com/KARTIKrocks/go-paginate.svg)](https://pkg.go.dev/github.com/KARTIKrocks/go-paginate)
[![Go Report Card](https://goreportcard.com/badge/github.com/KARTIKrocks/go-paginate)](https://goreportcard.com/report/github.com/KARTIKrocks/go-paginate)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive, production-ready pagination library for Go supporting multiple pagination strategies.

## Features

✅ **Multiple Pagination Strategies**

- Offset-based pagination (traditional page numbers)
- Cursor-based pagination (efficient, consistent results)
- Range-based pagination (HTTP Range header style)
- GraphQL connections (Relay-style)

✅ **Production Ready**

- Thread-safe with immutable setters
- Overflow-safe calculations
- Comprehensive validation
- Zero external dependencies
- ~94% test coverage

✅ **Easy Integration**

- HTTP request parsing
- SQL query generation
- Response formatting
- RESTful Link headers

✅ **Developer Friendly**

- Clean, fluent API
- Extensive documentation
- Type-safe with generics
- Framework agnostic

## Installation

```bash
go get github.com/KARTIKrocks/go-paginate
```

**Requirements:** Go 1.24+

## Quick Start

### Offset Pagination

```go
import "github.com/KARTIKrocks/go-paginate"

// Parse from HTTP request
func handleUsers(w http.ResponseWriter, r *http.Request) {
    p := paginate.FromRequest(r)

    // Validate
    if err := p.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Use with SQL
    users := []User{}
    db.Limit(p.Limit()).Offset(p.Offset()).Find(&users)

    // Get total count
    var total int64
    db.Model(&User{}).Count(&total)

    // Create response
    response := paginate.NewPage(users, total, p)
    json.NewEncoder(w).Encode(response)
}
```

### Cursor Pagination

```go
// Parse cursor from request
c := paginate.CursorFromRequest(r)

// Create cursor from last item
var users []User
db.Where("id > ?", lastID).Limit(c.Limit).Find(&users)

var nextCursor string
if len(users) == c.Limit {
    lastUser := users[len(users)-1]
    nextCursor = paginate.NewCursorFromID(lastUser.ID)
}

// Create response
response := paginate.NewCursorPageSimple(users, c.Limit, nextCursor)
json.NewEncoder(w).Encode(response)
```

## Usage Examples

### Basic Offset Pagination

```go
// Create paginator
p := paginate.New().
    WithPage(2).
    WithPageSize(25)

// Get offset and limit for SQL
offset := p.Offset() // 25
limit := p.Limit()   // 25

// Check navigation
if p.HasNext(totalCount) {
    nextPage := p.NextPage() // 3
}

// Generate SQL (PostgreSQL)
query := fmt.Sprintf("SELECT * FROM users %s", p.SQLClause())
// SELECT * FROM users LIMIT 25 OFFSET 25

// Generate SQL (MySQL)
query := fmt.Sprintf("SELECT * FROM users %s", p.SQLClauseMySQL())
// SELECT * FROM users LIMIT 25, 25
```

### Parsing from HTTP Request

```go
// URL: /users?page=2&page_size=50
p := paginate.FromRequest(r)

// Also supports common alternatives
// ?page=2&limit=50
// ?page=2&per_page=50
```

### Creating Responses

```go
// Offset pagination response
page := paginate.NewPage(items, totalCount, p)
// {
//   "items": [...],
//   "total": 100,
//   "page": 2,
//   "page_size": 25,
//   "total_pages": 4,
//   "has_prev": true,
//   "has_next": true
// }

// Cursor pagination response
cursorPage := paginate.NewCursorPage(items, 20, nextCursor, prevCursor, hasMore)
// {
//   "items": [...],
//   "next_cursor": "eyJpZCI6IjEyMyJ9",
//   "prev_cursor": "eyJpZCI6Ijk4In0",
//   "has_more": true,
//   "limit": 20
// }
```

### GraphQL Connections

```go
conn := paginate.NewConnection(
    items,
    func(item User) string {
        return paginate.NewCursorFromID(item.ID)
    },
    hasPrev,
    hasNext,
    totalCount,
)

// {
//   "edges": [
//     {"node": {...}, "cursor": "..."},
//     ...
//   ],
//   "page_info": {
//     "has_previous_page": false,
//     "has_next_page": true,
//     "start_cursor": "...",
//     "end_cursor": "..."
//   },
//   "total_count": 100
// }
```

### HTTP Link Headers

```go
// RFC 5988 compliant Link headers
links := paginate.BuildLinkHeader("https://api.example.com/users", p, totalCount)
w.Header().Set("Link", links.String())
// Link: <https://api.example.com/users?page=1&page_size=20>; rel="first",
//       <https://api.example.com/users?page=1&page_size=20>; rel="prev",
//       <https://api.example.com/users?page=3&page_size=20>; rel="next",
//       <https://api.example.com/users?page=5&page_size=20>; rel="last"
```

### Range-based Pagination

```go
// Parse Range header
rng, err := paginate.RangeFromRequest(r)
// Range: items=0-24

// Use with SQL
query := fmt.Sprintf("SELECT * FROM users %s", rng.SQLClause())

// Create response
response := paginate.NewRangeResponse(items, rng, totalCount)
w.Header().Set("Content-Range", response.ContentRange())
// Content-Range: items 0-24/100
```

## Advanced Usage

### Custom Validation

```go
p := paginate.FromRequest(r)

// Custom max page size for specific endpoints
if p.PageSize > 100 {
    p = p.WithPageSize(100)
}

// Ensure page is within bounds
p = p.Clamp(totalCount)
```

### Thread-Safe Usage

```go
// Base paginator
base := paginate.New()

// Safe to use concurrently
go func() {
    p1 := base.WithPage(1) // New instance
    // Use p1...
}()

go func() {
    p2 := base.WithPage(2) // Different new instance
    // Use p2...
}()
```

### Cursor Types

```go
// Simple ID cursor
cursor := paginate.NewCursorFromID("user_123")

// Timestamp-based cursor (for time-ordered data)
cursor := paginate.NewCursorFromTimestamp(time.Now(), "user_123")

// Offset-based cursor (cursor API with offset backend)
cursor := paginate.NewCursorFromOffset(100)

// Decode cursor
data, err := paginate.DecodeCursor(cursor)
if err != nil {
    // Handle invalid cursor
}
fmt.Println(data.ID, data.Timestamp, data.Offset)
```

### Working with ORMs

#### GORM

```go
func ListUsers(db *gorm.DB, p *paginate.Paginator) (*paginate.Page[User], error) {
    var users []User
    var total int64

    // Get total count
    if err := db.Model(&User{}).Count(&total).Error; err != nil {
        return nil, err
    }

    // Get page of results
    err := db.Offset(int(p.Offset())).
        Limit(p.Limit()).
        Find(&users).Error

    if err != nil {
        return nil, err
    }

    return paginate.NewPage(users, total, p), nil
}
```

#### sqlx

```go
func ListUsers(db *sqlx.DB, p *paginate.Paginator) (*paginate.Page[User], error) {
    var users []User

    query := fmt.Sprintf(`
        SELECT * FROM users
        ORDER BY created_at DESC
        %s
    `, p.SQLClause())

    err := db.Select(&users, query)
    if err != nil {
        return nil, err
    }

    var total int64
    db.Get(&total, "SELECT COUNT(*) FROM users")

    return paginate.NewPage(users, total, p), nil
}
```

## Configuration

### Constants

```go
const (
    DefaultPage     = 1     // Default page number
    DefaultPageSize = 20    // Default items per page
    MaxPageSize     = 1000  // Maximum allowed page size
    MinPageSize     = 1     // Minimum allowed page size
)
```

These can be referenced but not modified. If you need different limits, validate and clamp manually:

```go
p := paginate.FromRequest(r)
if p.PageSize > 100 {
    p = p.WithPageSize(100)
}
```

## API Reference

### Offset Pagination

#### Creating Paginators

- `New() *Paginator` - Create with defaults
- `NewWithSize(pageSize int) *Paginator` - Create with custom page size
- `NewFromValues(page, pageSize int) *Paginator` - Create with both values
- `FromRequest(r *http.Request) *Paginator` - Parse from HTTP request
- `FromQuery(q url.Values) *Paginator` - Parse from query values
- `FromMap(m map[string]any) *Paginator` - Parse from map (for JSON)

#### Paginator Methods

- `WithPage(page int) *Paginator` - Return new instance with page
- `WithPageSize(size int) *Paginator` - Return new instance with page size
- `Offset() int64` - Get SQL offset
- `Limit() int` - Get SQL limit
- `HasNext(total int64) bool` - Check if next page exists
- `HasPrevious() bool` - Check if previous page exists
- `TotalPages(total int64) int` - Calculate total pages
- `Validate() error` - Validate parameters
- `Clone() *Paginator` - Create a copy
- `Clamp(total int64) *Paginator` - Adjust page to valid range

### Cursor Pagination

#### Creating Cursor Paginators

- `NewCursor() *CursorPaginator` - Create with defaults
- `NewCursorWithLimit(limit int) *CursorPaginator` - Create with custom limit
- `CursorFromRequest(r *http.Request) *CursorPaginator` - Parse from request
- `CursorFromQuery(q url.Values) *CursorPaginator` - Parse from query values

#### Cursor Helper Functions

- `EncodeCursor(data *CursorData) string` - Encode cursor data
- `DecodeCursor(cursor string) (*CursorData, error)` - Decode cursor
- `NewCursorFromID(id string) string` - Create cursor from ID
- `NewCursorFromTimestamp(ts time.Time, id string) string` - Create from timestamp
- `NewCursorFromOffset(offset int) string` - Create from offset

## Error Handling

```go
var (
    ErrInvalidPage     error // Page < 1
    ErrInvalidPageSize error // Page size out of bounds
    ErrInvalidCursor   error // Malformed cursor
    ErrInvalidOffset   error // Offset < 0
    ErrInvalidRange    error // Invalid range parameters
)
```

Use `errors.Is()` for checking:

```go
if errors.Is(err, paginate.ErrInvalidCursor) {
    // Handle invalid cursor
}
```

## Best Practices

### 1. Always Validate Input

```go
p := paginate.FromRequest(r)
if err := p.Validate(); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

### 2. Use Appropriate Pagination Strategy

- **Offset**: Simple, good for small datasets, supports jumping to pages
- **Cursor**: Large datasets, real-time data, consistent results
- **Range**: File-like access patterns, partial content requests

### 3. Set Reasonable Limits

```go
// Prevent abuse
if p.PageSize > 100 {
    p = p.WithPageSize(100)
}
```

### 4. Include Total Count When Needed

Offset pagination benefits from total count for UI (page numbers), but it can be expensive for large tables. Consider omitting for very large datasets:

```go
// With count (better UX, slower)
page := paginate.NewPage(items, totalCount, p)

// Without count (faster, less info)
// Just set total to -1 or 0
```

### 5. Index Your Database

```sql
-- For offset pagination
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- For cursor pagination (compound index for tie-breaking)
CREATE INDEX idx_users_created_id ON users(created_at DESC, id DESC);
```

## Performance Considerations

### Offset Pagination

- **Pros**: Simple, supports random access, easy to implement
- **Cons**: Slower for large offsets, inconsistent with concurrent writes
- **Best for**: Small to medium datasets, infrequent access to deep pages

### Cursor Pagination

- **Pros**: Consistent results, efficient for large datasets, better for real-time data
- **Cons**: No random access, more complex implementation
- **Best for**: Large datasets, infinite scroll, real-time feeds

### Range Pagination

- **Pros**: Standard HTTP semantics, good for downloads/exports
- **Cons**: Similar issues to offset pagination
- **Best for**: File-like resources, bulk exports

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:

- Tests pass (`go test ./...`)
- Code is formatted (`go fmt ./...`)
- Linting passes (`golangci-lint run`)

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Inspired by pagination patterns from:

- GraphQL Cursor Connections Specification
- RFC 5988 (Web Linking)
- Common REST API patterns

## Support

- 📖 [Documentation](https://pkg.go.dev/github.com/KARTIKrocks/go-paginate)
- 🐛 [Issue Tracker](https://github.com/KARTIKrocks/go-paginate/issues)
- 💬 [Discussions](https://github.com/KARTIKrocks/go-paginate/discussions)

---

Made with ❤️ for the Go community
