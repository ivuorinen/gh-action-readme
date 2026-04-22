package testutil

import (
	"sync"
)

// ErrorFormatterMock tracks error formatter calls for testing.
type ErrorFormatterMock struct {
	mu                         sync.Mutex
	FormatContextualErrorCalls []string
}

// FormatContextualError captures contextual error formatting calls.
func (m *ErrorFormatterMock) FormatContextualError(err error) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err != nil {
		formatted := err.Error()
		m.FormatContextualErrorCalls = append(m.FormatContextualErrorCalls, formatted)

		return formatted
	}

	return ""
}
