package database

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound indicates a database query returned no results.
// Use errors.Is(err, database.ErrNotFound) for checking.
var ErrNotFound = errors.New("not found")

// Collection provides type-safe database operations for a specific entity type.
// It uses Go generics to ensure compile-time type safety for all operations.
//
// The Collection pattern eliminates the need for code generation or reflection
// at query time, providing both safety and performance.
type Collection[E Entity] struct {
	DB   *DynamicDB
	Ent  E
	Type reflect.Type
}

// Manage creates a Collection for type-safe database operations on entity E.
//
// This should be called once at package initialization:
//
//	var Users = database.Manage(db, new(User))
//
// Thread Safety: Since this is called at package init (before main),
// there are no concurrency concerns. The registration happens once
// in a single-threaded context.
//
// The returned Collection provides all CRUD operations with full type safety.
func Manage[E Entity](db *DynamicDB, ent E) *Collection[E] {
	db.Register(ent)
	t := reflect.TypeOf(ent)
	db.Repos[ent.Table()] = &Collection[Entity]{db, ent, t}
	return &Collection[E]{db, ent, t}
}

// Count returns the number of entities matching the query.
//
// Examples:
//
//	total := Users.Count("")  // Count all users
//	active := Users.Count("WHERE Status = ?", "active")
//	recent := Users.Count("WHERE CreatedAt > ?", time.Now().AddDate(0, -1, 0))
//
// SQL Note: Use PascalCase for field names (Status, not status).
func (c *Collection[E]) Count(query string, args ...any) (count int) {
	countQuery := `SELECT COUNT(*) FROM ` + c.Ent.Table()
	if query != "" {
		countQuery += " " + query
	}
	c.DB.Query(countQuery, args...).Scan(&count)
	return count
}

// First returns the first entity matching the query
// Example: Users.First("WHERE Email = ?", email)
func (c *Collection[E]) First(query string, args ...any) (E, error) {
	// Add LIMIT 1 if not already in query
	if !strings.Contains(strings.ToUpper(query), "LIMIT") {
		query += " LIMIT 1"
	}
	results, err := c.Search(query, args...)
	if err != nil {
		var zero E
		return zero, err
	}
	if len(results) == 0 {
		var zero E
		return zero, fmt.Errorf("%w", ErrNotFound)
	}
	return results[0], nil
}

// New creates a new instance of the entity type
func (c *Collection[E]) New() E {
	ent := reflect.New(c.Type.Elem()).Interface().(E)
	ent.GetModel().SetDB(c.DB)
	return ent
}

// Get retrieves an entity by its UUID.
//
// Returns database.ErrNotFound if the entity doesn't exist:
//
//	user, err := Users.Get(id)
//	if errors.Is(err, database.ErrNotFound) {
//	    // Handle not found
//	}
//
// IDs are always strings (UUIDs) for consistency and distribution safety.
func (c *Collection[E]) Get(id string) (E, error) {
	ent := c.New()
	return ent, c.DB.Get(id, ent)
}

// Insert creates a new entity in the database.
//
// Automatic fields:
//   - ID: Generated UUID if not provided
//   - CreatedAt: Current timestamp
//   - UpdatedAt: Current timestamp
//
// Example:
//
//	user := &User{Name: "Alice", Email: "alice@example.com"}
//	user, err := Users.Insert(user)
//	// user.ID is now set
//
// Design Note: UUIDs are used for global uniqueness without coordination.
func (c *Collection[E]) Insert(ent E) (E, error) {
	ent.GetModel().SetDB(c.DB)
	if ent.GetModel().ID == "" {
		ent.GetModel().ID = uuid.NewString()
	}
	// Always set timestamps for new records
	if ent.GetModel().CreatedAt.IsZero() {
		ent.GetModel().CreatedAt = time.Now()
	}
	if ent.GetModel().UpdatedAt.IsZero() {
		ent.GetModel().UpdatedAt = time.Now()
	}
	return ent, c.DB.Insert(ent)
}

// Update modifies an existing entity in the database.
//
// Automatic updates:
//   - UpdatedAt: Set to current timestamp
//
// The entity must have a valid ID. Returns error if not found.
func (c *Collection[E]) Update(ent E) error {
	ent.GetModel().UpdatedAt = time.Now()
	return c.DB.Update(ent)
}

// Delete removes an entity from the database
func (c *Collection[E]) Delete(ent E) error {
	return c.DB.Delete(ent)
}

// Search queries the database and returns matching entities.
//
// Query format: SQL WHERE and ORDER BY clauses
//
// Examples:
//
//	// Get all users ordered by name
//	users, err := Users.Search("ORDER BY Name")
//
//	// Filter by status
//	active, err := Users.Search("WHERE Status = ? ORDER BY CreatedAt DESC", "active")
//
//	// Complex query
//	results, err := Users.Search(`
//	    WHERE CreatedAt > ? AND Status IN (?, ?)
//	    ORDER BY CreatedAt DESC
//	    LIMIT 10
//	`, since, "active", "pending")
//
// CRITICAL: Use PascalCase for SQL fields (UserID not user_id).
func (c *Collection[E]) Search(query string, args ...any) ([]E, error) {
	apps := []E{}
	return apps, Cursor(c.DB, c.Ent, query, args...).
		Iter(func(load func(Entity) error) error {
			app := c.New()
			if err := load(app); err != nil {
				return err
			}
			apps = append(apps, app)
			return nil
		})
}

// Index creates an index on the collection's table
// Example: Repositories.Index("UserID", "Visibility")
func (c *Collection[E]) Index(columns ...string) error {
	return c.DB.Index(c.Ent.Table(), columns...)
}

// UniqueIndex creates a unique index on the collection's table
// Example: Permissions.UniqueIndex("RepoID", "UserID")
func (c *Collection[E]) UniqueIndex(columns ...string) error {
	return c.DB.UniqueIndex(c.Ent.Table(), columns...)
}

// SearchPaginated performs a paginated search with total count
// Returns: items, total count, error
func (c *Collection[E]) SearchPaginated(query string, limit, offset int, args ...any) ([]E, int, error) {
	// Get total count first
	var total int
	countQuery := "SELECT COUNT(*) FROM " + c.Ent.Table()
	if query != "" {
		countQuery += " " + query
	}
	err := c.DB.Query(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	paginatedQuery := query + fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	items, err := c.Search(paginatedQuery, args...)
	return items, total, err
}

// AllPaginated returns all items with pagination
// Returns: items, total count, error
func (c *Collection[E]) AllPaginated(limit, offset int) ([]E, int, error) {
	return c.SearchPaginated("ORDER BY CreatedAt DESC", limit, offset)
}
