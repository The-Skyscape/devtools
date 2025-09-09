package testutils

import (
	"path/filepath"
	"testing"
	
	"github.com/The-Skyscape/devtools/pkg/authentication"
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/database/local"
)

// TestWorkspace provides a complete test environment with database and data directory
type TestWorkspace struct {
	DataDir string
	DB      *database.DynamicDB
	Auth    *authentication.Collection
	t       *testing.T
}

// SetupTestWorkspace creates a complete test workspace with temp directories
func SetupTestWorkspace(t *testing.T) *TestWorkspace {
	// Create temp directory for all test data
	dataDir := t.TempDir()
	
	// Create database in the temp directory
	dbPath := filepath.Join(dataDir, "test.db")
	db := local.Database(dbPath)
	
	// Initialize authentication
	auth := authentication.Manage(db)
	
	return &TestWorkspace{
		DataDir: dataDir,
		DB:      db,
		Auth:    auth,
		t:       t,
	}
}

// CreateTestUser creates a test user in the workspace
func (w *TestWorkspace) CreateTestUser(email string) *authentication.User {
	user := &authentication.User{
		Email:    email,
		Name:     "Test User",
	}
	
	// Extract handle from email
	if len(email) > 12 && email[len(email)-12:] == "@example.com" {
		user.Handle = email[:len(email)-12]
	} else {
		user.Handle = email
	}
	
	// Set a test password
	if err := user.SetupPassword("testpassword123"); err != nil {
		w.t.Fatalf("Failed to set password: %v", err)
	}
	
	created, err := w.Auth.Users.Insert(user)
	if err != nil {
		w.t.Fatalf("Failed to create test user: %v", err)
	}
	
	return created
}

// Cleanup cleans up the test workspace
func (w *TestWorkspace) Cleanup() {
	// The temp directory will be cleaned up automatically by t.TempDir()
	// Just close any open connections if needed
}

// GetRepoPath returns a path for test repositories
func (w *TestWorkspace) GetRepoPath(repoName string) string {
	return filepath.Join(w.DataDir, "repos", repoName)
}

// GetFilePath returns a path for test files
func (w *TestWorkspace) GetFilePath(filename string) string {
	return filepath.Join(w.DataDir, "files", filename)
}