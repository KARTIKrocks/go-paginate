# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-02-11

### Breaking Changes

- `CursorData` is now generic: `CursorData[T any]` — all existing `CursorData{...}` literals must become `CursorData[any]{...}`
- `EncodeCursor` is now generic: `EncodeCursor[T any](data *CursorData[T]) (string, error)`
- `DecodeCursor` is now generic: `DecodeCursor[T any](cursor string) (*CursorData[T], error)`
- `NewCursorFromValue` is now generic: `NewCursorFromValue[T any](value T) (string, error)`
- `CursorPaginator.Encode` now takes `CursorData[any]` instead of `CursorData`
- `CursorPaginator.Decode` now returns `*CursorData[any]` instead of `*CursorData`
- Module path changed to `github.com/KARTIKrocks/go-paginate/v2`

### Added

- Type-safe cursor values via generic `CursorData[T]` — eliminates JSON round-trip data loss (e.g. `int` no longer decodes as `float64`)
- `NewCursorFromValue[T any]` for creating cursors with typed values

### Migration from v1

Replace all `CursorData{...}` with `CursorData[any]{...}` for a drop-in upgrade. Use concrete types like `CursorData[int]` or `CursorData[string]` where type-safe round-trips are needed.

## [1.0.0] - 2026-02-11

### Added

- Initial release of go-paginate
- Offset-based pagination with `Paginator`
- Cursor-based pagination with `CursorPaginator`
- Range-based pagination (HTTP Range header style)
- GraphQL-style connections
- Multiple response types: `Page`, `CursorPage`, `Connection`, `RangeResponse`
- HTTP request parsing: `FromRequest`, `CursorFromRequest`, `RangeFromRequest`
- SQL clause generation for PostgreSQL and MySQL
- RFC 5988 Link header support
- Thread-safe With\* methods for immutable updates
- Overflow-safe offset calculations using int64
- Comprehensive validation with detailed error messages
- Generic types for type-safe responses
- Zero external dependencies
- ~94% test coverage
- Complete documentation and examples
- CI/CD with GitHub Actions
- golangci-lint integration

### Features

#### Offset Pagination

- Traditional page-based navigation
- SQL query helpers (LIMIT/OFFSET)
- Total pages calculation
- Navigation helpers (HasNext, HasPrevious)
- Query parameter parsing from HTTP requests
- Support for common parameter names (limit, per_page, page_size)

#### Cursor Pagination

- Forward and backward navigation
- Multiple cursor creation helpers (ID, timestamp, offset, value)
- Base64 encoding/decoding
- GraphQL-style parameters support (first, last, after, before)
- Efficient for large datasets

#### Range Pagination

- HTTP Range header parsing and generation
- Content-Range header support
- Partial content responses
- Conversion to/from offset pagination

#### GraphQL Connections

- Relay-style connections
- Edge and node structure
- PageInfo with cursors
- Total count support

#### Response Types

- Generic response types for type safety
- Helper methods (Empty, Count, Nodes)
- Link header generation
- Content-Range header generation

### Security

- Input validation for all parameters
- Safe defaults for invalid input
- Protection against integer overflow
- Bounds checking for page size limits

### Performance

- No external dependencies
- Minimal allocations
- Efficient cursor encoding
- Optimized for concurrent use

## [Unreleased]

### Planned

- Cursor signing/verification for tamper detection
- Optional cursor encryption
- Metrics and observability hooks
- Custom validator support
- Additional SQL dialect helpers
- HTTP middleware for automatic pagination

---

## Version History

- **v2.0.0** - Generic `CursorData[T]` for type-safe cursor values
- **v1.0.0** - Initial stable release with comprehensive pagination support

[2.0.0]: https://github.com/KARTIKrocks/go-paginate/releases/tag/v2.0.0
[1.0.0]: https://github.com/KARTIKrocks/go-paginate/releases/tag/v1.0.0
[Unreleased]: https://github.com/KARTIKrocks/go-paginate/compare/v2.0.0...HEAD
