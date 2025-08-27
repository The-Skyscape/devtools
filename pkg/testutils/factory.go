package testutils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Factory provides methods to generate test data
type Factory struct {
	counter int
}

// NewFactory creates a new test data factory
func NewFactory() *Factory {
	return &Factory{counter: 1}
}

// String generates a unique string with optional prefix
func (f *Factory) String(prefix string) string {
	result := fmt.Sprintf("%s%d", prefix, f.counter)
	f.counter++
	return result
}

// Email generates a unique test email
func (f *Factory) Email() string {
	return f.String("user") + "@test.com"
}

// Name generates a test name
func (f *Factory) Name() string {
	names := []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry"}
	return names[f.counter%len(names)] + fmt.Sprintf(" Test%d", f.counter)
}

// Token generates a random token
func (f *Factory) Token() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// URL generates a test URL
func (f *Factory) URL() string {
	return fmt.Sprintf("https://test%d.example.com", f.counter)
}

// ID generates a test UUID-like ID
func (f *Factory) ID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Time generates a test time
func (f *Factory) Time() time.Time {
	return time.Now().Add(-time.Duration(f.counter) * time.Hour)
}

// TimePtr generates a pointer to a test time
func (f *Factory) TimePtr() *time.Time {
	t := f.Time()
	return &t
}

// Int generates an incrementing integer
func (f *Factory) Int() int {
	result := f.counter
	f.counter++
	return result
}

// Float generates a test float value
func (f *Factory) Float(base float64) float64 {
	return base * float64(f.counter)
}

// Bool alternates between true and false
func (f *Factory) Bool() bool {
	return f.counter%2 == 0
}