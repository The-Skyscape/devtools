package providers

import (
	"fmt"
	"sync"

	"github.com/The-Skyscape/devtools/pkg/emailing"
)

// MockProvider is a mock email provider for testing
type MockProvider struct {
	mu         sync.Mutex
	SentEmails []emailing.Message
	ShouldFail bool
	FailError  error
}

// NewMockProvider creates a new mock provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		SentEmails: make([]emailing.Message, 0),
	}
}

// Send implements the Provider interface
func (m *MockProvider) Send(msg *emailing.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ShouldFail {
		if m.FailError != nil {
			return m.FailError
		}
		return fmt.Errorf("mock provider: send failed")
	}

	m.SentEmails = append(m.SentEmails, *msg)
	return nil
}

// GetName implements the Provider interface
func (m *MockProvider) GetName() string {
	return "mock"
}

// GetSentEmails returns a copy of sent emails (thread-safe)
func (m *MockProvider) GetSentEmails() []emailing.Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	emails := make([]emailing.Message, len(m.SentEmails))
	copy(emails, m.SentEmails)
	return emails
}

// Clear resets the mock provider
func (m *MockProvider) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SentEmails = m.SentEmails[:0]
	m.ShouldFail = false
	m.FailError = nil
}

// SetFailure configures the mock to fail with an error
func (m *MockProvider) SetFailure(fail bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ShouldFail = fail
	m.FailError = err
}

// LastEmail returns the last sent email or nil
func (m *MockProvider) LastEmail() *emailing.Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.SentEmails) == 0 {
		return nil
	}

	last := m.SentEmails[len(m.SentEmails)-1]
	return &last
}

// EmailCount returns the number of emails sent
func (m *MockProvider) EmailCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.SentEmails)
}

// FindEmail searches for an email by recipient
func (m *MockProvider) FindEmail(to string) *emailing.Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.SentEmails {
		if m.SentEmails[i].ToAddr == to {
			email := m.SentEmails[i]
			return &email
		}
	}
	return nil
}

// FindEmailBySubject searches for an email by subject
func (m *MockProvider) FindEmailBySubject(subject string) *emailing.Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.SentEmails {
		if m.SentEmails[i].Subject == subject {
			email := m.SentEmails[i]
			return &email
		}
	}
	return nil
}
