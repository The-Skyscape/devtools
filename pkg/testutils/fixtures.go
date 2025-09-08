package testutils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// FixtureManager manages test fixtures
type FixtureManager struct {
	basePath string
	cache    map[string][]byte
}

// NewFixtureManager creates a new fixture manager
func NewFixtureManager(basePath string) *FixtureManager {
	return &FixtureManager{
		basePath: basePath,
		cache:    make(map[string][]byte),
	}
}

// Load loads a fixture file by name
func (fm *FixtureManager) Load(name string) ([]byte, error) {
	// Check cache first
	if data, ok := fm.cache[name]; ok {
		return data, nil
	}
	
	path := filepath.Join(fm.basePath, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load fixture %s: %w", name, err)
	}
	
	fm.cache[name] = data
	return data, nil
}

// LoadString loads a fixture as a string
func (fm *FixtureManager) LoadString(name string) (string, error) {
	data, err := fm.Load(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// LoadJSON loads a fixture and unmarshals it as JSON
func (fm *FixtureManager) LoadJSON(name string, target interface{}) error {
	data, err := fm.Load(name)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, target)
}

// Save saves data as a fixture (useful for generating fixtures)
func (fm *FixtureManager) Save(name string, data []byte) error {
	path := filepath.Join(fm.basePath, name)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create fixture directory: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to save fixture %s: %w", name, err)
	}
	
	// Update cache
	fm.cache[name] = data
	return nil
}

// SaveJSON saves data as a JSON fixture
func (fm *FixtureManager) SaveJSON(name string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return fm.Save(name, jsonData)
}

// Clear clears the fixture cache
func (fm *FixtureManager) Clear() {
	fm.cache = make(map[string][]byte)
}

// TestFixture is a helper for loading fixtures in tests
func TestFixture(t *testing.T, name string) []byte {
	t.Helper()
	
	fm := NewFixtureManager("testdata")
	data, err := fm.Load(name)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", name, err)
	}
	
	return data
}

// TestFixtureString is a helper for loading string fixtures in tests
func TestFixtureString(t *testing.T, name string) string {
	t.Helper()
	return string(TestFixture(t, name))
}

// TestFixtureJSON is a helper for loading JSON fixtures in tests
func TestFixtureJSON(t *testing.T, name string, target interface{}) {
	t.Helper()
	
	data := TestFixture(t, name)
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("Failed to unmarshal fixture %s: %v", name, err)
	}
}

// GoldenFile manages golden file testing
type GoldenFile struct {
	path   string
	update bool
}

// NewGoldenFile creates a new golden file helper
func NewGoldenFile(path string, update bool) *GoldenFile {
	return &GoldenFile{
		path:   path,
		update: update,
	}
}

// Assert compares actual output with golden file
func (g *GoldenFile) Assert(t *testing.T, actual []byte) {
	t.Helper()
	
	if g.update {
		// Update golden file
		if err := os.WriteFile(g.path, actual, 0644); err != nil {
			t.Fatalf("Failed to update golden file: %v", err)
		}
		t.Logf("Updated golden file: %s", g.path)
		return
	}
	
	// Compare with golden file
	expected, err := os.ReadFile(g.path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create golden file if it doesn't exist
			if err := os.WriteFile(g.path, actual, 0644); err != nil {
				t.Fatalf("Failed to create golden file: %v", err)
			}
			t.Logf("Created golden file: %s", g.path)
			return
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}
	
	if string(expected) != string(actual) {
		t.Errorf("Golden file mismatch for %s\nExpected:\n%s\nActual:\n%s", 
			g.path, string(expected), string(actual))
	}
}

// AssertString compares actual string output with golden file
func (g *GoldenFile) AssertString(t *testing.T, actual string) {
	t.Helper()
	g.Assert(t, []byte(actual))
}

// TempFile creates a temporary file for testing
func TempFile(t *testing.T, content string) string {
	t.Helper()
	
	file, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	if _, err := io.WriteString(file, content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	
	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	
	// Clean up after test
	t.Cleanup(func() {
		os.Remove(file.Name())
	})
	
	return file.Name()
}

// TempDir creates a temporary directory for testing
func TempDir(t *testing.T) string {
	t.Helper()
	
	dir, err := os.MkdirTemp("", "testdir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	// Clean up after test
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	
	return dir
}