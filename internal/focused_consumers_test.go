package internal

import (
	"errors"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// compositeOutputWriterForTest wraps testutil mocks to satisfy OutputWriter interface.
type compositeOutputWriterForTest struct {
	*testutil.MessageLoggerMock
	*testutil.ProgressReporterMock
	*testutil.OutputConfigMock
}

// errorManagerForTest wraps testutil mocks to satisfy ErrorManager interface.
type errorManagerForTest struct {
	*testutil.ErrorReporterMock
	*testutil.ErrorFormatterMock
}

// FormatContextualError implements ErrorManager interface.
func (e *errorManagerForTest) FormatContextualError(err *apperrors.ContextualError) string {
	if err != nil {
		return e.ErrorFormatterMock.FormatContextualError(err)
	}

	return ""
}

// ErrorWithSuggestions implements ErrorManager interface.
func (e *errorManagerForTest) ErrorWithSuggestions(err *apperrors.ContextualError) {
	e.ErrorReporterMock.ErrorWithSuggestions(err)
}

// TestNewCompositeOutputWriter tests the composite output writer constructor.
func TestNewCompositeOutputWriter(t *testing.T) {
	t.Parallel()

	writer := &compositeOutputWriterForTest{
		MessageLoggerMock:    &testutil.MessageLoggerMock{},
		ProgressReporterMock: &testutil.ProgressReporterMock{},
		OutputConfigMock:     &testutil.OutputConfigMock{},
	}
	cow := NewCompositeOutputWriter(writer)

	if cow == nil {
		t.Fatal("NewCompositeOutputWriter() returned nil")
	}

	if cow.writer != writer {
		t.Error("NewCompositeOutputWriter() did not set writer correctly")
	}
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

			logger := &testutil.MessageLoggerMock{}
			progress := &testutil.ProgressReporterMock{}
			writer := &compositeOutputWriterForTest{
				MessageLoggerMock:    logger,
				ProgressReporterMock: progress,
				OutputConfigMock:     &testutil.OutputConfigMock{QuietMode: tt.isQuiet},
			}
			cow := NewCompositeOutputWriter(writer)

			cow.ProcessWithOutput(tt.items)

			totalMessages := len(logger.InfoCalls) + len(progress.ProgressCalls) + len(logger.SuccessCalls)
			if totalMessages != tt.wantMessages {
				t.Errorf("ProcessWithOutput() produced %d messages, want %d",
					totalMessages, tt.wantMessages)
			}

			hasInfo := len(logger.InfoCalls) > 0
			hasProgress := len(progress.ProgressCalls) > 0
			hasSuccess := len(logger.SuccessCalls) > 0

			if hasInfo != tt.wantInfo {
				t.Errorf("ProcessWithOutput() hasInfo = %v, want %v", hasInfo, tt.wantInfo)
			}
			if hasProgress != tt.wantProgress {
				t.Errorf("ProcessWithOutput() hasProgress = %v, want %v", hasProgress, tt.wantProgress)
			}
			if hasSuccess != tt.wantSuccess {
				t.Errorf("ProcessWithOutput() hasSuccess = %v, want %v", hasSuccess, tt.wantSuccess)
			}
		})
	}
}

// TestNewValidationComponent tests the validation component constructor.
func TestNewValidationComponent(t *testing.T) {
	t.Parallel()

	errorManager := &errorManagerForTest{
		ErrorReporterMock:  &testutil.ErrorReporterMock{},
		ErrorFormatterMock: &testutil.ErrorFormatterMock{},
	}
	logger := &testutil.MessageLoggerMock{}

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

// getErrorCallType returns the type of error call that was made.
func getErrorCallType(reporter *testutil.ErrorReporterMock) string {
	switch {
	case len(reporter.ErrorWithSuggestionsCalls) > 0:
		return "ErrorWithSuggestions"
	case len(reporter.ErrorCalls) > 0:
		return "Error"
	case len(reporter.ErrorWithSimpleFixCalls) > 0:
		return "ErrorWithSimpleFix"
	case len(reporter.ErrorWithContextCalls) > 0:
		return "ErrorWithContext"
	default:
		return ""
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
			item:              testutil.TestItemName,
			isValid:           true,
			err:               nil,
			wantLoggerCalls:   1,
			wantErrorCalls:    0,
			wantErrorCallType: "",
		},
		{
			name:              "invalid with contextual error",
			item:              testutil.TestItemName,
			isValid:           false,
			err:               apperrors.New(appconstants.ErrCodeValidation, "validation failed"),
			wantLoggerCalls:   0,
			wantErrorCalls:    1,
			wantErrorCallType: "ErrorWithSuggestions",
		},
		{
			name:              "invalid with regular error",
			item:              testutil.TestItemName,
			isValid:           false,
			err:               errors.New("regular error"),
			wantLoggerCalls:   0,
			wantErrorCalls:    1,
			wantErrorCallType: "Error",
		},
		{
			name:              "invalid without error",
			item:              testutil.TestItemName,
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

			errorReporter := &testutil.ErrorReporterMock{}
			errorManager := &errorManagerForTest{
				ErrorReporterMock:  errorReporter,
				ErrorFormatterMock: &testutil.ErrorFormatterMock{},
			}
			logger := &testutil.MessageLoggerMock{}
			vc := NewValidationComponent(errorManager, logger)

			vc.ValidateAndReport(tt.item, tt.isValid, tt.err)

			totalLoggerCalls := len(
				logger.InfoCalls,
			) + len(
				logger.SuccessCalls,
			) + len(
				logger.WarningCalls,
			) + len(
				logger.BoldCalls,
			) + len(
				logger.PrintfCalls,
			)
			if totalLoggerCalls != tt.wantLoggerCalls {
				t.Errorf("ValidateAndReport() logger calls = %d, want %d",
					totalLoggerCalls, tt.wantLoggerCalls)
			}

			totalErrorCalls := len(
				errorReporter.ErrorCalls,
			) + len(
				errorReporter.ErrorWithSuggestionsCalls,
			) + len(
				errorReporter.ErrorWithContextCalls,
			) + len(
				errorReporter.ErrorWithSimpleFixCalls,
			)
			if totalErrorCalls != tt.wantErrorCalls {
				t.Errorf("ValidateAndReport() error calls = %d, want %d",
					totalErrorCalls, tt.wantErrorCalls)
			}

			if tt.wantErrorCallType != "" {
				actualCallType := getErrorCallType(errorReporter)
				if actualCallType != tt.wantErrorCallType {
					t.Errorf("ValidateAndReport() error call type = %s, want %s",
						actualCallType, tt.wantErrorCallType)
				}
			}
		})
	}
}
