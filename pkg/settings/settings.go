// Package settings provides a simple key-value settings management system
// that can be used by applications built with devtools.
package settings

import (
	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/database"
)

// Setting represents an application setting with key-value storage
type Setting struct {
	application.Model
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"` // user_preference, system, config, etc
}

// Table returns the database table name
func (*Setting) Table() string {
	return "settings"
}

// Manager provides methods for managing application settings
type Manager struct {
	repo *database.Collection[*Setting]
}

// NewManager creates a new settings manager with the given database
func NewManager(db *database.DynamicDB) *Manager {
	return &Manager{
		repo: database.Manage(db, new(Setting)),
	}
}

// Get retrieves a setting by key, returns empty string if not found
func (m *Manager) Get(key string) (string, error) {
	setting, err := m.repo.Find("WHERE Key = ?", key)
	if err != nil {
		return "", nil // Return empty string instead of error for missing keys
	}
	return setting.Value, nil
}

// Set creates or updates a setting with the given key and value
func (m *Manager) Set(key, value, settingType string) error {
	settings, err := m.repo.Search("WHERE Key = ? LIMIT 1", key)
	if err != nil {
		return err
	}
	
	if len(settings) > 0 {
		// Update existing
		settings[0].Value = value
		return m.repo.Update(settings[0])
	}
	
	// Create new
	_, err = m.repo.Insert(&Setting{
		Key:   key,
		Value: value,
		Type:  settingType,
	})
	return err
}

// GetBool retrieves a setting as a boolean value
func (m *Manager) GetBool(key string) bool {
	value, _ := m.Get(key)
	return value == "true" || value == "1"
}

// SetBool sets a boolean setting
func (m *Manager) SetBool(key string, value bool, settingType string) error {
	strValue := "false"
	if value {
		strValue = "true"
	}
	return m.Set(key, strValue, settingType)
}

// Delete removes a setting by key
func (m *Manager) Delete(key string) error {
	settings, err := m.repo.Search("WHERE Key = ? LIMIT 1", key)
	if err != nil {
		return err
	}
	
	if len(settings) > 0 {
		return m.repo.Delete(settings[0])
	}
	
	return nil
}

// GetByType retrieves all settings of a specific type
func (m *Manager) GetByType(settingType string) ([]*Setting, error) {
	return m.repo.Search("WHERE Type = ? ORDER BY Key", settingType)
}