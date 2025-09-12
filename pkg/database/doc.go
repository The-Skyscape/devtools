// Package database provides a type-safe, generic repository pattern for database operations
// with support for SQLite and future database engines.
//
// # Design Philosophy
//
// This package embraces Go's type system and generics to provide:
//   - Type-safe CRUD operations without code generation
//   - Zero reflection at query time (only at initialization)
//   - Automatic UUID generation and timestamp management
//   - PascalCase SQL fields matching Go struct fields
//
// # Core Concepts
//
// Repository Pattern:
// The Collection[T] type provides a generic repository for any entity type.
// Each collection handles CRUD operations with full type safety.
//
// Model Embedding:
// All models embed the base Model struct to inherit ID, CreatedAt, and UpdatedAt fields.
// This ensures consistency across all entities in the system.
//
// # Quick Start
//
//	type User struct {
//	    database.Model
//	    Name  string
//	    Email string
//	}
//
//	func (*User) Table() string { return "users" }
//
//	// Register the collection
//	var Users = database.Manage(db, new(User))
//
//	// Use type-safe operations
//	user, err := Users.Get(id)
//	users, err := Users.Search("WHERE CreatedAt > ?", since)
//
// # SQL Convention
//
// CRITICAL: This package uses PascalCase for SQL field names, matching Go struct fields exactly.
// This eliminates field name transformation bugs and maintains consistency.
//
//	// Go struct field
//	type Order struct {
//	    UserID    string  // SQL: UserID (NOT user_id)
//	    TotalCost float64 // SQL: TotalCost (NOT total_cost)
//	}
//
//	// SQL queries use PascalCase
//	orders, err := Orders.Search("WHERE UserID = ? AND TotalCost > ?", userID, 100)
//
// This is an intentional design choice that prioritizes correctness over convention.
//
// # ID Strategy
//
// The package uses string UUIDs as primary keys by default:
//   - Globally unique without coordination
//   - No sequence bottlenecks
//   - Safe for distributed systems
//
// While this may seem opinionated, it eliminates entire categories of bugs related to
// ID collisions and simplifies data migration between systems.
//
// # Thread Safety
//
// Collections are typically initialized at package level during program startup:
//
//	var Users = database.Manage(db, new(User))  // Single-threaded init
//
// This happens before main() in a single-threaded context, eliminating race conditions.
// The underlying SQLite connection pool handles concurrent access safely.
//
// # Performance Characteristics
//
//   - Template queries are parsed once per type at startup
//   - No reflection during query execution
//   - Connection pooling handled by database driver
//   - Prepared statements cached by SQLite
//
// # Design Decisions
//
// Why PascalCase SQL:
// Go struct fields are PascalCase. Using the same in SQL eliminates transformation bugs.
// The slight unconventionality is worth the elimination of an entire bug category.
//
// Why UUIDs:
// Sequential IDs require coordination and create bottlenecks. UUIDs are globally unique
// without coordination, making them ideal for distributed systems and migrations.
//
// Why Model Embedding:
// Every entity needs ID and timestamps. Embedding ensures consistency and reduces boilerplate.
// The slight coupling is worth the massive reduction in repetitive code.
package database