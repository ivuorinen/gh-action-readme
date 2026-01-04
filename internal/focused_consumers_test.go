package internal

import (
	"errors"
	"os"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
)

// trackingOutputWriter is a test implementation of OutputWriter that tracks calls.
type trackingOutputWriter struct {
	messages []string
	isQuiet  bool
}

func (m *trackingOutputWriter) Info(_ string, _ ...any) {
	m.messages = append(m.messages, "INFO")
}

func (m *trackingOutputWriter) Success(_ string, _ ...any) {
	m.messages = append(m.messages, "SUCCESS")
}

func (m *trackingOutputWriter) Warning(_ string, _ ...any) {
	m.messages = append(m.messages, "WARNING")
}

func (m *trackingOutputWriter) Bold(_ string, _ ...any) {
	m.messages = append(m.messages, "BOLD")
}

func (m *trackingOutputWriter) Printf(_ string, _ ...any) {
	m.messages = append(m.messages, "PRINTF")
}

func (m *trackingOutputWriter) Fprintf(_ *os.File, _ string, _ ...any) {
	m.messages = append(m.messages, "FPRINTF")
}

func (m *trackingOutputWriter) Progress(_ string, _ ...any) {
	m.messages = append(m.messages, "PROGRESS")
}

func (m *trackingOutputWriter) IsQuiet() bool {
	return m.isQuiet
}

// trackingErrorManager is a test implementation of ErrorManager that tracks calls.
type trackingErrorManager struct {
	errorCalls []string
}

func (m *trackingErrorManager) Error(_ string, _ ...any) {
	m.errorCalls = append(m.errorCalls, "Error")
}

func (m *trackingErrorManager) ErrorWithSuggestions(_ *apperrors.ContextualError) {
	m.errorCalls = append(m.errorCalls, "ErrorWithSuggestions")
}

func (m *trackingErrorManager) ErrorWithContext(_ appconstants.ErrorCode, _ string, _ map[string]string) {
	m.errorCalls = append(m.errorCalls, "ErrorWithContext")
}

func (m *trackingErrorManager) ErrorWithSimpleFix(_, _ string) {
	m.errorCalls = append(m.errorCalls, "ErrorWithSimpleFix")
}

func (m *trackingErrorManager) FormatContextualError(_ *apperrors.ContextualError) string {
	return "formatted error"
}

// trackingMessageLogger is a test implementation of MessageLogger that tracks calls.
type trackingMessageLogger struct {
	messages []string
}

func (m *trackingMessageLogger) Info(_ string, _ ...any) {
	m.messages = append(m.messages, "INFO")
}

func (m *trackingMessageLogger) Success(_ string, _ ...any) {
	m.messages = append(m.messages, "SUCCESS")
}

func (m *trackingMessageLogger) Warning(_ string, _ ...any) {
	m.messages = append(m.messages, "WARNING")
}

func (m *trackingMessageLogger) Bold(_ string, _ ...any) {
	m.messages = append(m.messages, "BOLD")
}

func (m *trackingMessageLogger) Printf(_ string, _ ...any) {
	m.messages = append(m.messages, "PRINTF")
}

func (m *trackingMessageLogger) Fprintf(_ *os.File, _ string, _ ...any) {
	m.messages = append(m.messages, "FPRINTF")
}

// TestNewCompositeOutputWriter tests the composite output writer constructor.
func TestNewCompositeOutputWriter(t *testing.T) {
	t.Parallel()

	writer := &trackingOutputWriter{}
	cow := NewCompositeOutputWriter(writer)

	if cow == nil {
		t.Fatal("NewCompositeOutputWriter() returned nil")
	}

	if cow.writer != writer {
		t.Error("NewCompositeOutputWriter() did not set writer correctly")
	}
}

// countMessageTypes counts the presence of specific message types in a slice of messages.
func countMessageTypes(messages []string, types ...string) map[string]bool {
	results := make(map[string]bool)
	for _, msgType := range types {
		results[msgType] = false
	}

	for _, msg := range messages {
		for _, msgType := range types {
			if msg == msgType {
				results[msgType] = true
			}
		}
	}

	return results
}

// TestCompositeOutputWriterProcessWithOutput tests processing with output.
func TestCompositeOutputWriterProcessWithOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		isQuiet      bool
		items        []string
		wantMessages int
		wantInfo     bool
		wantProgress bool
		wantSuccess  bool
	}{
		{
			name:         "with items not quiet",
			isQuiet:      false,
			items:        []string{"item1", "item2", "item3"},
			wantMessages: 5, // 1 info + 3 progress + 1 success
			wantInfo:     true,
			wantProgress: true,
			wantSuccess:  true,
		},
		{
			name:         "with quiet mode",
			isQuiet:      true,
			items:        []string{"item1", "item2"},
			wantMessages: 0,
			wantInfo:     false,
			wantProgress: false,
			wantSuccess:  false,
		},
		{
			name:         "with empty items",
			isQuiet:      false,
			items:        []string{},
			wantMessages: 2, // 1 info + 1 success
			wantInfo:     true,
			wantProgress: false,
			wantSuccess:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := &trackingOutputWriter{isQuiet: tt.isQuiet}
			cow := NewCompositeOutputWriter(writer)

			cow.ProcessWithOutput(tt.items)

			if len(writer.messages) != tt.wantMessages {
				t.Errorf("ProcessWithOutput() produced %d messages, want %d",
					len(writer.messages), tt.wantMessages)
			}

			msgTypes := countMessageTypes(writer.messages, "INFO", "PROGRESS", "SUCCESS")

			if msgTypes["INFO"] != tt.wantInfo {
				t.Errorf("ProcessWithOutput() hasInfo = %v, want %v", msgTypes["INFO"], tt.wantInfo)
			}
			if msgTypes["PROGRESS"] != tt.wantProgress {
				t.Errorf("ProcessWithOutput() hasProgress = %v, want %v", msgTypes["PROGRESS"], tt.wantProgress)
			}
			if msgTypes["SUCCESS"] != tt.wantSuccess {
				t.Errorf("ProcessWithOutput() hasSuccess = %v, want %v", msgTypes["SUCCESS"], tt.wantSuccess)
			}
		})
	}
}

// TestNewValidationComponent tests the validation component constructor.
func TestNewValidationComponent(t *testing.T) {
	t.Parallel()

	errorManager := &trackingErrorManager{}
	logger := &trackingMessageLogger{}

	vc := NewValidationComponent(errorManager, logger)

	if vc == nil {
		t.Fatal("NewValidationComponent() returned nil")
	}

	if vc.errorManager != errorManager {
		t.Error("NewValidationComponent() did not set errorManager correctly")
	}

	if vc.logger != logger {
		t.Error("NewValidationComponent() did not set logger correctly")
	}
}

// TestValidationComponentValidateAndReport tests validation reporting.
func TestValidationComponentValidateAndReport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		item              string
		isValid           bool
		err               error
		wantLoggerCalls   int
		wantErrorCalls    int
		wantErrorCallType string
	}{
		{
			name:              "valid item",
			item:              appconstants.TestItemName,
			isValid:           true,
			err:               nil,
			wantLoggerCalls:   1,
			wantErrorCalls:    0,
			wantErrorCallType: "",
		},
		{
			name:              "invalid with contextual error",
			item:              appconstants.TestItemName,
			isValid:           false,
			err:               apperrors.New(appconstants.ErrCodeValidation, "validation failed"),
			wantLoggerCalls:   0,
			wantErrorCalls:    1,
			wantErrorCallType: "ErrorWithSuggestions",
		},
		{
			name:              "invalid with regular error",
			item:              appconstants.TestItemName,
			isValid:           false,
			err:               errors.New("regular error"),
			wantLoggerCalls:   0,
			wantErrorCalls:    1,
			wantErrorCallType: "Error",
		},
		{
			name:              "invalid without error",
			item:              appconstants.TestItemName,
			isValid:           false,
			err:               nil,
			wantLoggerCalls:   0,
			wantErrorCalls:    1,
			wantErrorCallType: "ErrorWithSimpleFix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errorManager := &trackingErrorManager{}
			logger := &trackingMessageLogger{}
			vc := NewValidationComponent(errorManager, logger)

			vc.ValidateAndReport(tt.item, tt.isValid, tt.err)

			if len(logger.messages) != tt.wantLoggerCalls {
				t.Errorf("ValidateAndReport() logger calls = %d, want %d",
					len(logger.messages), tt.wantLoggerCalls)
			}

			if len(errorManager.errorCalls) != tt.wantErrorCalls {
				t.Errorf("ValidateAndReport() error calls = %d, want %d",
					len(errorManager.errorCalls), tt.wantErrorCalls)
			}

			if tt.wantErrorCallType != "" && len(errorManager.errorCalls) > 0 {
				if errorManager.errorCalls[0] != tt.wantErrorCallType {
					t.Errorf("ValidateAndReport() error call type = %s, want %s",
						errorManager.errorCalls[0], tt.wantErrorCallType)
				}
			}
		})
	}
}
