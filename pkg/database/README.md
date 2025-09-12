# Database Package

The `database` package provides a type-safe, generic repository pattern for database operations with automatic UUID generation and timestamp management.

## Features

- 🔒 **Type-Safe Operations** - Generic collections with compile-time safety
- 🆔 **Automatic UUIDs** - Generated primary keys for distribution safety
- ⏰ **Timestamp Management** - Automatic CreatedAt/UpdatedAt handling
- 🔤 **PascalCase SQL** - Direct mapping between Go structs and SQL
- 🚀 **Zero Reflection at Runtime** - Only during initialization
- 🔄 **Connection Pooling** - Built-in concurrency handling

## Quick Start

```go
package models

import (
    "github.com/The-Skyscape/devtools/pkg/database"
    "github.com/The-Skyscape/devtools/pkg/database/engines/sqlite3"
)

// Define your model
type User struct {
    database.Model  // Embed for ID, CreatedAt, UpdatedAt
    Name  string
    Email string
    Active bool
}

// Implement Table() method
func (*User) Table() string { return "users" }

// Create collection at package level
var DB = sqlite3.Open("app.db")
var Users = database.Manage(DB, new(User))

// Use type-safe operations
func Example() {
    // Insert
    user, err := Users.Insert(&User{
        Name:  "Alice",
        Email: "alice@example.com",
    })
    
    // Get by ID
    user, err = Users.Get(id)
    
    // Search with PascalCase SQL
    activeUsers, err := Users.Search("WHERE Active = ? ORDER BY CreatedAt DESC", true)
    
    // Update
    user.Name = "Alice Smith"
    err = Users.Update(user)
    
    // Delete
    err = Users.Delete(user)
}
```

## CRITICAL: PascalCase SQL Convention

⚠️ **This package uses PascalCase for SQL field names, matching Go struct fields exactly.**

```go
// Go struct
type Order struct {
    database.Model
    UserID    string  // SQL field: UserID (NOT user_id)
    TotalCost float64 // SQL field: TotalCost (NOT total_cost)
}

// ✅ CORRECT - PascalCase fields
orders, err := Orders.Search("WHERE UserID = ? AND TotalCost > ?", userID, 100)

// ❌ WRONG - snake_case fields (will fail)
orders, err := Orders.Search("WHERE user_id = ? AND total_cost > ?", userID, 100)
```

### Why PascalCase?

This eliminates an entire category of field name transformation bugs. The slight unconventionality is worth the correctness guarantee.

## Repository Pattern

### Creating Collections

Collections are typically created at package initialization:

```go
// models/database.go
package models

import (
    "github.com/The-Skyscape/devtools/pkg/database"
    "github.com/The-Skyscape/devtools/pkg/database/engines/sqlite3"
)

var DB = sqlite3.Open("app.db")

// Register all collections
var (
    Users    = database.Manage(DB, new(User))
    Posts    = database.Manage(DB, new(Post))
    Comments = database.Manage(DB, new(Comment))
)
```

### CRUD Operations

```go
// Insert - automatically generates ID and timestamps
user, err := Users.Insert(&User{Name: "Bob"})
// user.ID is now set with UUID
// user.CreatedAt and UpdatedAt are set

// Get by ID
user, err := Users.Get(id)
if errors.Is(err, database.ErrNotFound) {
    // Handle not found
}

// Update - automatically updates UpdatedAt
user.Name = "Robert"
err := Users.Update(user)

// Delete
err := Users.Delete(user)

// Count
total := Users.Count("")  // Count all
active := Users.Count("WHERE Active = ?", true)

// First - returns first matching record
newest, err := Users.First("ORDER BY CreatedAt DESC")

// Search - returns all matching records
results, err := Users.Search("WHERE Age > ? ORDER BY Name", 18)
```

### Pagination

```go
// Get page of results with total count
page := 1
pageSize := 20
offset := (page - 1) * pageSize

items, total, err := Users.SearchPaginated(
    "WHERE Active = ? ORDER BY CreatedAt DESC",
    pageSize,
    offset,
    true,
)
```

### Complex Queries

```go
// Join example (manual)
type UserWithPosts struct {
    User
    PostCount int
}

iter := DB.Query(`
    SELECT u.*, COUNT(p.ID) as PostCount
    FROM users u
    LEFT JOIN posts p ON p.UserID = u.ID
    WHERE u.Active = ?
    GROUP BY u.ID
`, true)

var results []UserWithPosts
// ... scan results
```

## Model Definition

### Basic Model

```go
type Product struct {
    database.Model  // ALWAYS embed this
    Name        string
    Price       float64
    Stock       int
    CategoryID  string
}

func (*Product) Table() string { return "products" }
```

### Model with Relationships

```go
type Post struct {
    database.Model
    Title   string
    Content string
    UserID  string
}

func (*Post) Table() string { return "posts" }

// Relationship helper
func (p *Post) User() (*User, error) {
    return Users.Get(p.UserID)
}

// Reverse relationship
func (u *User) Posts() ([]*Post, error) {
    return Posts.Search("WHERE UserID = ? ORDER BY CreatedAt DESC", u.ID)
}
```

## Indexes

Create indexes for better performance:

```go
// Single column index
err := Users.Index("Email")

// Composite index
err := Posts.Index("UserID", "CreatedAt")

// Unique index
err := Users.UniqueIndex("Email")
```

## Thread Safety

Collections are initialized at package level during program startup (single-threaded context), eliminating race conditions:

```go
// This happens before main() in single-threaded context
var Users = database.Manage(DB, new(User))

// Safe to use concurrently in handlers
func handler(w http.ResponseWriter, r *http.Request) {
    user, err := Users.Get(id)  // Thread-safe
}
```

## Performance Characteristics

- **Template queries**: Parsed once per type at startup
- **No reflection**: During query execution
- **Connection pooling**: Handled by SQLite driver
- **Prepared statements**: Cached by SQLite
- **UUID generation**: ~1μs per insert
- **Typical operations**:
  - Insert: ~100μs
  - Get by ID: ~50μs
  - Search (10 records): ~200μs

## Testing

```go
func TestUserRepository(t *testing.T) {
    // Create in-memory database for testing
    db := sqlite3.Open(":memory:")
    
    // Create table
    db.Query(`CREATE TABLE users (...)`).Exec()
    
    // Create collection
    Users := database.Manage(db, new(User))
    
    // Test operations
    user, err := Users.Insert(&User{Name: "Test"})
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

## Migration Guide

### From GORM

```go
// GORM
db.Create(&user)
db.First(&user, id)
db.Model(&user).Updates(User{Name: "new"})
db.Delete(&user)

// DevTools Database
Users.Insert(&user)
user, err := Users.Get(id)
Users.Update(user)
Users.Delete(user)
```

### From sqlx

```go
// sqlx
db.Get(&user, "SELECT * FROM users WHERE id = ?", id)
db.Select(&users, "SELECT * FROM users WHERE active = ?", true)

// DevTools Database
user, err := Users.Get(id)
users, err := Users.Search("WHERE Active = ?", true)
```

## Best Practices

1. **Always embed database.Model** - Ensures consistency
2. **Use PascalCase in SQL** - Matches Go struct fields
3. **Create collections at package level** - Thread-safe initialization
4. **Check for ErrNotFound** - Use errors.Is()
5. **Let the framework handle IDs** - UUIDs are generated automatically
6. **Use transactions for complex operations** - Wrap in DB transaction

## Design Decisions

### Why UUIDs?
- Globally unique without coordination
- Safe for distributed systems
- No sequence bottlenecks
- Simplifies data migration

### Why PascalCase SQL?
- Direct mapping to Go structs
- Eliminates transformation bugs
- Consistency over convention

### Why Model Embedding?
- Every entity needs ID and timestamps
- Ensures consistency
- Reduces boilerplate

## License

MIT