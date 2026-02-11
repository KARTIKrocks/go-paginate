# Contributing to go-paginate

Thank you for your interest in contributing to go-paginate! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on what's best for the community
- Show empathy towards others

## Getting Started

### Prerequisites

- Go 1.24 or higher
- Git
- golangci-lint (for linting)

### Setting Up Development Environment

1. Fork the repository on GitHub
2. Clone your fork:

   ```bash
   git clone https://github.com/KARTIKrocks/go-paginate.git
   cd go-paginate
   ```

3. Add the upstream repository:

   ```bash
   git remote add upstream https://github.com/KARTIKrocks/go-paginate.git
   ```

4. Install development tools:
   ```bash
   make install-tools
   ```

## Development Workflow

### Making Changes

1. Create a new branch:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes

3. Write or update tests for your changes

4. Ensure all tests pass:

   ```bash
   make test
   ```

5. Format your code:

   ```bash
   make fmt
   ```

6. Run the linter:

   ```bash
   make lint
   ```

7. Commit your changes:
   ```bash
   git commit -m "Add feature: description of your changes"
   ```

### Commit Message Guidelines

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters
- Reference issues and pull requests liberally after the first line

Examples:

```
Add cursor pagination support for GraphQL

Fix overflow in offset calculation for large page numbers

Update documentation for range-based pagination
```

### Pull Request Process

1. Update your fork with the latest upstream changes:

   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. Push your changes to your fork:

   ```bash
   git push origin feature/your-feature-name
   ```

3. Create a Pull Request from your fork to the main repository

4. Ensure all CI checks pass

5. Request review from maintainers

6. Address any feedback

## Testing Guidelines

### Writing Tests

- All new code should have tests
- Aim for >80% code coverage
- Test edge cases and error conditions
- Use table-driven tests when appropriate

Example:

```go
func TestNewPaginator(t *testing.T) {
    tests := []struct {
        name     string
        page     int
        pageSize int
        wantPage int
        wantSize int
    }{
        {"valid values", 2, 25, 2, 25},
        {"invalid page", 0, 25, 1, 25},
        {"invalid size", 2, 0, 2, 20},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewFromValues(tt.page, tt.pageSize)
            if p.Page != tt.wantPage {
                t.Errorf("got page %d, want %d", p.Page, tt.wantPage)
            }
            if p.PageSize != tt.wantSize {
                t.Errorf("got size %d, want %d", p.PageSize, tt.wantSize)
            }
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run benchmarks
make bench
```

## Code Style

### General Guidelines

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Keep functions small and focused
- Write clear comments for exported functions

### Documentation

- All exported types, functions, and constants must have comments
- Comments should be complete sentences
- Start comments with the name of the thing being described

Example:

```go
// Paginator represents offset-based pagination parameters.
// Instances are safe to read concurrently. Use With* methods to create
// modified copies for thread-safe updates.
type Paginator struct {
    Page     int
    PageSize int
}
```

### Error Handling

- Use sentinel errors defined in `errors.go`
- Wrap errors with context using `fmt.Errorf`
- Don't panic unless absolutely necessary

Example:

```go
if p.Page < 1 {
    return fmt.Errorf("%w: got %d", ErrInvalidPage, p.Page)
}
```

## Adding New Features

### Before Starting

1. Check existing issues and pull requests
2. Open an issue to discuss the feature
3. Wait for maintainer feedback
4. Get approval before starting work

### Feature Requirements

- Must have tests
- Must have documentation
- Must not break existing functionality
- Should follow existing patterns
- Should be backward compatible

## Reporting Bugs

### Before Reporting

1. Check if the bug has already been reported
2. Verify you're using the latest version
3. Try to reproduce with a minimal example

### Bug Report Should Include

- Go version
- go-paginate version
- Minimal code to reproduce
- Expected behavior
- Actual behavior
- Any error messages

## Improving Documentation

Documentation improvements are always welcome:

- Fix typos or unclear wording
- Add examples
- Improve API documentation
- Add guides for common use cases

## Questions?

- Open a [Discussion](https://github.com/KARTIKrocks/go-paginate/discussions)
- Ask in an issue
- Check existing documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing! 🎉
