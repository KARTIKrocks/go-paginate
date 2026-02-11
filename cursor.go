package paginate

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// CursorPaginator provides cursor-based pagination.
// Instances are safe to read concurrently. Use With* methods to create
// modified copies for thread-safe updates.
type CursorPaginator struct {
	Cursor  string `json:"cursor,omitempty"`
	Limit   int    `json:"limit"`
	Forward bool   `json:"forward"` // true for next, false for previous
}

// CursorData holds the data encoded in a cursor.
// This structure is base64-encoded and can optionally be signed for security.
// The type parameter T controls the type of Value, enabling type-safe round-trips.
type CursorData[T any] struct {
	ID        string    `json:"id,omitempty"`
	Value     T         `json:"v,omitempty"`
	Timestamp time.Time `json:"ts,omitzero"`
	Offset    int       `json:"o,omitempty"`
}

// NewCursor creates a new cursor paginator with default values.
func NewCursor() *CursorPaginator {
	return &CursorPaginator{
		Limit:   DefaultPageSize,
		Forward: true,
	}
}

// NewCursorWithLimit creates a cursor paginator with a specific limit.
func NewCursorWithLimit(limit int) *CursorPaginator {
	return NewCursor().WithLimit(limit)
}

// WithLimit returns a new cursor paginator with the specified limit.
// This method is thread-safe as it returns a new instance.
func (c *CursorPaginator) WithLimit(limit int) *CursorPaginator {
	clone := c.Clone()
	if limit < MinPageSize {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}
	clone.Limit = limit
	return clone
}

// WithCursor returns a new cursor paginator with the specified cursor.
// This method is thread-safe as it returns a new instance.
func (c *CursorPaginator) WithCursor(cursor string) *CursorPaginator {
	clone := c.Clone()
	clone.Cursor = cursor
	return clone
}

// WithForward returns a new cursor paginator with the specified direction.
// This method is thread-safe as it returns a new instance.
func (c *CursorPaginator) WithForward(forward bool) *CursorPaginator {
	clone := c.Clone()
	clone.Forward = forward
	return clone
}

// Clone creates a copy of the cursor paginator.
func (c *CursorPaginator) Clone() *CursorPaginator {
	return &CursorPaginator{
		Cursor:  c.Cursor,
		Limit:   c.Limit,
		Forward: c.Forward,
	}
}

// HasCursor returns true if a cursor is set.
func (c *CursorPaginator) HasCursor() bool {
	return c.Cursor != ""
}

// Decode decodes the cursor into CursorData[any].
// Returns nil if no cursor is set, or an error if the cursor is invalid.
func (c *CursorPaginator) Decode() (*CursorData[any], error) {
	if c.Cursor == "" {
		return nil, nil
	}
	return DecodeCursor[any](c.Cursor)
}

// Encode encodes cursor data and returns a base64 cursor string.
// This is a convenience method that delegates to the package-level EncodeCursor.
func (c *CursorPaginator) Encode(data CursorData[any]) (string, error) {
	return EncodeCursor(&data)
}

// Validate validates the cursor paginator parameters.
func (c *CursorPaginator) Validate() error {
	if c.Limit < MinPageSize || c.Limit > MaxPageSize {
		return ErrInvalidPageSize
	}
	if c.Cursor != "" {
		if _, err := c.Decode(); err != nil {
			return err
		}
	}
	return nil
}

// QueryParams returns URL query parameters for the cursor paginator.
func (c *CursorPaginator) QueryParams() url.Values {
	params := url.Values{}
	if c.Cursor != "" {
		if c.Forward {
			params.Set("after", c.Cursor)
		} else {
			params.Set("before", c.Cursor)
		}
	}
	params.Set("limit", strconv.Itoa(c.Limit))
	return params
}

// CursorFromRequest parses cursor pagination from HTTP request.
func CursorFromRequest(r *http.Request) *CursorPaginator {
	return CursorFromQuery(r.URL.Query())
}

// CursorFromQuery parses cursor pagination from URL query values.
// Supports multiple query parameter formats:
//   - cursor + limit (generic)
//   - after/before + limit (directional)
//   - first/last (GraphQL-style)
func CursorFromQuery(q url.Values) *CursorPaginator {
	c := NewCursor()

	// Generic cursor parameter
	if cursor := q.Get("cursor"); cursor != "" {
		c = c.WithCursor(cursor)
	}

	// Support "after" and "before" cursors (more explicit)
	if after := q.Get("after"); after != "" {
		c = c.WithCursor(after).WithForward(true)
	}
	if before := q.Get("before"); before != "" {
		c = c.WithCursor(before).WithForward(false)
	}

	// Standard limit parameter
	if limitStr := q.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			c = c.WithLimit(limit)
		}
	}

	// GraphQL-style first/last parameters
	if firstStr := q.Get("first"); firstStr != "" {
		if first, err := strconv.Atoi(firstStr); err == nil && first > 0 {
			c = c.WithLimit(first).WithForward(true)
		}
	}
	if lastStr := q.Get("last"); lastStr != "" {
		if last, err := strconv.Atoi(lastStr); err == nil && last > 0 {
			c = c.WithLimit(last).WithForward(false)
		}
	}

	return c
}

// EncodeCursor encodes cursor data to a base64 string.
// Returns an empty string and nil error if data is nil.
// Returns an error if the data cannot be marshaled to JSON.
func EncodeCursor[T any](data *CursorData[T]) (string, error) {
	if data == nil {
		return "", nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// DecodeCursor decodes a base64 cursor string to cursor data.
// Returns an error if the cursor is malformed.
func DecodeCursor[T any](cursor string) (*CursorData[T], error) {
	if cursor == "" {
		return nil, nil
	}

	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	var data CursorData[T]
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, ErrInvalidCursor
	}

	return &data, nil
}

// NewCursorFromID creates a cursor from an ID.
func NewCursorFromID(id string) (string, error) {
	return EncodeCursor(&CursorData[any]{ID: id})
}

// NewCursorFromValue creates a cursor from a typed value.
// Note: The value should be JSON-serializable.
func NewCursorFromValue[T any](value T) (string, error) {
	return EncodeCursor(&CursorData[T]{Value: value})
}

// NewCursorFromTimestamp creates a cursor from a timestamp and ID.
// This is useful for time-based pagination with tie-breaking.
func NewCursorFromTimestamp(ts time.Time, id string) (string, error) {
	return EncodeCursor(&CursorData[any]{Timestamp: ts, ID: id})
}

// NewCursorFromOffset creates a cursor from an offset.
// This allows using cursor-style APIs with offset-based backends.
func NewCursorFromOffset(offset int) (string, error) {
	return EncodeCursor(&CursorData[any]{Offset: offset})
}
