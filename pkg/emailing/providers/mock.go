package providers

import (
	"fmt"

	"github.com/The-Skyscape/devtools/pkg/emailing"
	"github.com/The-Skyscape/devtools/pkg/security"
)

// MockProvider is a mock email provider for testing
type MockProvider struct {
	SentEmails []*emailing.Email
	ShouldFail bool
	FailError  error
}

// NewMockProvider creates a new mock provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		SentEmails: []*emailing.Email{},
	}
}

// Init implements the Provider interface (no-op for mock)
func (m *MockProvider) Init(vault *security.Collection) error {
	// Mock provider doesn't need vault initialization
	return nil
}

// Name implements the Provider interface
func (m *MockProvider) Name() string {
	return "mock"
}

// Send implements the Provider interface
func (m *MockProvider) Send(msg *emailing.Email) error {
	if m.ShouldFail {
		if m.FailError != nil {
			return m.FailError
		}
		return fmt.Errorf("mock provider: send failed")
	}

	m.SentEmails = append(m.SentEmails, msg)
	return nil
}

// GetSentEmails returns a copy of sent emails (thread-safe)
func (m *MockProvider) GetSentEmails() []*emailing.Email {
	return m.SentEmails
}

// Clear resets the mock provider
func (m *MockProvider) Clear() {
	m.SentEmails = m.SentEmails[:0]
	m.ShouldFail = false
	m.FailError = nil
}

// SetFailure configures the mock to fail with an error
func (m *MockProvider) SetFailure(fail bool, err error) {
	m.ShouldFail = fail
	m.FailError = err
}

// LastEmail returns the last sent email or nil
func (m *MockProvider) LastEmail() *emailing.Email {
	if len(m.SentEmails) == 0 {
		return nil
	}

	return m.SentEmails[len(m.SentEmails)-1]
}

// EmailCount returns the number of emails sent
func (m *MockProvider) EmailCount() int {
	return len(m.SentEmails)
}

// FindEmail searches for an email by recipient
func (m *MockProvider) FindEmail(to string) *emailing.Email {
	for i := range m.SentEmails {
		if m.SentEmails[i].ToAddr == to {
			email := m.SentEmails[i]
			return email
		}
	}
	return nil
}

// FindEmailBySubject searches for an email by subject
func (m *MockProvider) FindEmailBySubject(subject string) *emailing.Email {
	for i := range m.SentEmails {
		if m.SentEmails[i].Subject == subject {
			email := m.SentEmails[i]
			return email
		}
	}
	return nil
}
