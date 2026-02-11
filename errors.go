package paginate

import "errors"

// Sentinel errors for pagination operations.
// These can be checked using errors.Is() for proper error handling.
var (
	// ErrInvalidPage indicates the page number is invalid (< 1).
	ErrInvalidPage = errors.New("paginate: page must be >= 1")

	// ErrInvalidPageSize indicates the page size is outside allowed bounds.
	ErrInvalidPageSize = errors.New("paginate: page_size must be between min and max allowed values")

	// ErrInvalidCursor indicates the cursor is malformed or has been tampered with.
	ErrInvalidCursor = errors.New("paginate: cursor is malformed or invalid")

	// ErrInvalidOffset indicates the offset value is invalid (< 0).
	ErrInvalidOffset = errors.New("paginate: offset must be >= 0")

	// ErrInvalidRange indicates the range parameters are invalid.
	ErrInvalidRange = errors.New("paginate: invalid range parameters")
)
