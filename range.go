package paginate

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

// Range represents range-based pagination (similar to HTTP Range header).
// This is useful for APIs that want to support byte-range-like pagination.
type Range struct {
	Start int64
	End   int64
	Unit  string
}

// NewRange creates a new range with the default "items" unit.
func NewRange(start, end int64) *Range {
	return &Range{
		Start: start,
		End:   end,
		Unit:  "items",
	}
}

// NewRangeWithUnit creates a new range with a custom unit.
func NewRangeWithUnit(start, end int64, unit string) *Range {
	return &Range{
		Start: start,
		End:   end,
		Unit:  unit,
	}
}

// Size returns the number of items in the range.
func (r *Range) Size() int64 {
	if r.End < r.Start {
		return 0
	}
	return r.End - r.Start + 1
}

// Validate validates the range parameters.
func (r *Range) Validate() error {
	if r.Start < 0 {
		return ErrInvalidOffset
	}
	if r.End < r.Start {
		return ErrInvalidRange
	}
	return nil
}

// SQLClause returns SQL LIMIT OFFSET clause from range.
func (r *Range) SQLClause() string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", r.Size(), r.Start)
}

// Header returns the Range header value.
// Example: "items=0-24"
func (r *Range) Header() string {
	return fmt.Sprintf("%s=%d-%d", r.Unit, r.Start, r.End)
}

// ContentRangeHeader returns the Content-Range header value.
// Set total to -1 if the total is unknown.
// Example: "items 0-24/100" or "items 0-24/*"
func (r *Range) ContentRangeHeader(total int64) string {
	if total < 0 {
		return fmt.Sprintf("%s %d-%d/*", r.Unit, r.Start, r.End)
	}
	return fmt.Sprintf("%s %d-%d/%d", r.Unit, r.Start, r.End, total)
}

// RangeResponse represents a range-based pagination response.
type RangeResponse[T any] struct {
	Items []T    `json:"items"`
	Start int64  `json:"start"`
	End   int64  `json:"end"`
	Total int64  `json:"total"`
	Unit  string `json:"unit"`
}

// NewRangeResponse creates a new range response.
// The actual end is calculated based on the number of items returned.
func NewRangeResponse[T any](items []T, r *Range, total int64) *RangeResponse[T] {
	actualEnd := r.Start
	if len(items) > 0 {
		actualEnd = r.Start + int64(len(items)) - 1
	}

	return &RangeResponse[T]{
		Items: items,
		Start: r.Start,
		End:   actualEnd,
		Total: total,
		Unit:  r.Unit,
	}
}

// ContentRange returns the Content-Range header value.
func (r *RangeResponse[T]) ContentRange() string {
	if len(r.Items) == 0 {
		return fmt.Sprintf("%s */%d", r.Unit, r.Total)
	}
	return fmt.Sprintf("%s %d-%d/%d", r.Unit, r.Start, r.End, r.Total)
}

// HasMore returns true if there are more items after this range.
func (r *RangeResponse[T]) HasMore() bool {
	return r.End < r.Total-1
}

// Empty returns true if the response has no items.
func (r *RangeResponse[T]) Empty() bool {
	return len(r.Items) == 0
}

// Count returns the number of items in the response.
func (r *RangeResponse[T]) Count() int {
	return len(r.Items)
}

// Regular expression for parsing Range headers.
// Matches patterns like "items=0-24" or "bytes=100-199"
var rangeRegex = regexp.MustCompile(`^(\w+)=(\d+)-(\d*)$`)

// ParseRangeHeader parses the Range header value.
// Supports formats like "items=0-24" or "items=100-"
// If the end is omitted, it defaults to start + DefaultPageSize - 1.
func ParseRangeHeader(header string) (*Range, error) {
	if header == "" {
		return nil, nil
	}

	matches := rangeRegex.FindStringSubmatch(header)
	if matches == nil {
		return nil, ErrInvalidRange
	}

	unit := matches[1]
	start, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return nil, ErrInvalidOffset
	}

	var end int64
	if matches[3] != "" {
		end, err = strconv.ParseInt(matches[3], 10, 64)
		if err != nil {
			return nil, ErrInvalidRange
		}
	} else {
		// Open-ended range: use default page size
		end = start + int64(DefaultPageSize) - 1
	}

	rng := &Range{
		Start: start,
		End:   end,
		Unit:  unit,
	}

	return rng, rng.Validate()
}

// RangeFromRequest parses range from HTTP request Range header.
func RangeFromRequest(r *http.Request) (*Range, error) {
	return ParseRangeHeader(r.Header.Get("Range"))
}

// RangeFromOffsetLimit creates a range from offset and limit values.
func RangeFromOffsetLimit(offset, limit int) *Range {
	start := int64(offset)
	end := start + int64(limit) - 1
	if limit <= 0 {
		end = start
	}
	return &Range{
		Start: start,
		End:   end,
		Unit:  "items",
	}
}

// ToPaginator converts a range to an offset-based paginator (approximate).
// This is useful for backends that use offset pagination but need to support
// range-based APIs.
func (r *Range) ToPaginator() *Paginator {
	pageSize := int(r.Size())
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	page := int(r.Start/int64(pageSize)) + 1
	return NewFromValues(page, pageSize)
}
