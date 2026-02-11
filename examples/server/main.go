package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/KARTIKrocks/go-paginate"
)

// User represents a user in our system
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Mock data store
var users = []User{
	{ID: "1", Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now().Add(-100 * time.Hour)},
	{ID: "2", Name: "Bob", Email: "bob@example.com", CreatedAt: time.Now().Add(-90 * time.Hour)},
	{ID: "3", Name: "Charlie", Email: "charlie@example.com", CreatedAt: time.Now().Add(-80 * time.Hour)},
	{ID: "4", Name: "Diana", Email: "diana@example.com", CreatedAt: time.Now().Add(-70 * time.Hour)},
	{ID: "5", Name: "Eve", Email: "eve@example.com", CreatedAt: time.Now().Add(-60 * time.Hour)},
	{ID: "6", Name: "Frank", Email: "frank@example.com", CreatedAt: time.Now().Add(-50 * time.Hour)},
	{ID: "7", Name: "Grace", Email: "grace@example.com", CreatedAt: time.Now().Add(-40 * time.Hour)},
	{ID: "8", Name: "Henry", Email: "henry@example.com", CreatedAt: time.Now().Add(-30 * time.Hour)},
	{ID: "9", Name: "Iris", Email: "iris@example.com", CreatedAt: time.Now().Add(-20 * time.Hour)},
	{ID: "10", Name: "Jack", Email: "jack@example.com", CreatedAt: time.Now().Add(-10 * time.Hour)},
}

func main() {
	http.HandleFunc("/users/offset", handleOffsetPagination)
	http.HandleFunc("/users/cursor", handleCursorPagination)
	http.HandleFunc("/users/range", handleRangePagination)
	http.HandleFunc("/users/graphql", handleGraphQLConnection)

	fmt.Println("Server starting on :8080")
	fmt.Println("Try these endpoints:")
	fmt.Println("  - http://localhost:8080/users/offset?page=1&page_size=3")
	fmt.Println("  - http://localhost:8080/users/cursor?limit=3")
	fmt.Println("  - http://localhost:8080/users/range (with Range: items=0-2)")
	fmt.Println("  - http://localhost:8080/users/graphql?first=3")

	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// handleOffsetPagination demonstrates offset-based pagination
func handleOffsetPagination(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	p := paginate.FromRequest(r)

	// Validate
	if err := p.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get paginated data
	total := int64(len(users))
	start := p.Offset()
	end := start + int64(p.Limit())

	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}

	pageUsers := users[start:end]

	// Create response
	response := paginate.NewPage(pageUsers, total, p)

	// Add Link header
	baseURL := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	links := paginate.BuildLinkHeader(baseURL, p, total)
	w.Header().Set("Link", links.String())

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

// handleCursorPagination demonstrates cursor-based pagination
func handleCursorPagination(w http.ResponseWriter, r *http.Request) {
	// Parse cursor parameters
	c := paginate.CursorFromRequest(r)

	// Validate
	if err := c.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Decode cursor if present
	var startIdx int
	if c.HasCursor() {
		cursorData, err := c.DecodeCursor()
		if err != nil {
			http.Error(w, "invalid cursor", http.StatusBadRequest)
			return
		}

		// Find start index from cursor
		for i, u := range users {
			if u.ID == cursorData.ID {
				if c.Forward {
					startIdx = i + 1
				} else {
					startIdx = i - c.Limit
					if startIdx < 0 {
						startIdx = 0
					}
				}
				break
			}
		}
	}

	// Get items
	endIdx := startIdx + c.Limit
	if endIdx > len(users) {
		endIdx = len(users)
	}

	pageUsers := users[startIdx:endIdx]

	// Create cursors
	var nextCursor, prevCursor string
	if endIdx < len(users) {
		lastUser := pageUsers[len(pageUsers)-1]
		nextCursor = paginate.NewCursorFromID(lastUser.ID)
	}
	if startIdx > 0 && len(pageUsers) > 0 {
		firstUser := pageUsers[0]
		prevCursor = paginate.NewCursorFromID(firstUser.ID)
	}

	// Create response
	response := paginate.NewCursorPage(pageUsers, c.Limit, nextCursor, prevCursor, nextCursor != "")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

// handleRangePagination demonstrates range-based pagination
func handleRangePagination(w http.ResponseWriter, r *http.Request) {
	// Parse Range header
	rng, err := paginate.RangeFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Default range if not provided
	if rng == nil {
		rng = paginate.NewRange(0, 2) // First 3 items
	}

	// Validate
	if err := rng.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get ranged data
	total := int64(len(users))
	start := rng.Start
	end := rng.End + 1 // Range is inclusive, slice is exclusive

	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}

	rangeUsers := users[start:end]

	// Create response
	response := paginate.NewRangeResponse(rangeUsers, rng, total)

	// Set headers
	w.Header().Set("Content-Range", response.ContentRange())
	w.Header().Set("Accept-Ranges", "items")
	w.Header().Set("Content-Type", "application/json")

	// Set status code
	if start == 0 && end >= total {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusPartialContent)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

// handleGraphQLConnection demonstrates GraphQL-style connections
func handleGraphQLConnection(w http.ResponseWriter, r *http.Request) {
	// Parse cursor parameters (supports GraphQL-style first/last/after/before)
	c := paginate.CursorFromRequest(r)

	// Validate
	if err := c.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Simple implementation - just take first N items
	limit := c.Limit
	if limit > len(users) {
		limit = len(users)
	}

	pageUsers := users[:limit]

	// Create connection
	conn := paginate.NewConnection(
		pageUsers,
		func(u User) string {
			return paginate.NewCursorFromID(u.ID)
		},
		false,              // hasPrev
		limit < len(users), // hasNext
		int64(len(users)),  // total
	)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(conn); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
