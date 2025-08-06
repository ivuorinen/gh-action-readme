package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"

	"github.com/ivuorinen/gh-action-readme/internal/errors"
)

// ColoredOutput provides methods for colored terminal output.
// It implements all the focused interfaces for backward compatibility.
type ColoredOutput struct {
	NoColor bool
	Quiet   bool
}

// Compile-time interface checks.
var (
	_ MessageLogger    = (*ColoredOutput)(nil)
	_ ErrorReporter    = (*ColoredOutput)(nil)
	_ ErrorFormatter   = (*ColoredOutput)(nil)
	_ ProgressReporter = (*ColoredOutput)(nil)
	_ OutputConfig     = (*ColoredOutput)(nil)
	_ CompleteOutput   = (*ColoredOutput)(nil)
)

// NewColoredOutput creates a new colored output instance.
func NewColoredOutput(quiet bool) *ColoredOutput {
	return &ColoredOutput{
		NoColor: color.NoColor || os.Getenv("NO_COLOR") != "",
		Quiet:   quiet,
	}
}

// IsQuiet returns whether the output is in quiet mode.
func (co *ColoredOutput) IsQuiet() bool {
	return co.Quiet
}

// Success prints a success message in green.
func (co *ColoredOutput) Success(format string, args ...any) {
	if co.Quiet {
		return
	}
	if co.NoColor {
		fmt.Printf("✅ "+format+"\n", args...)
	} else {
		color.Green("✅ "+format, args...)
	}
}

// Error prints an error message in red to stderr.
func (co *ColoredOutput) Error(format string, args ...any) {
	if co.NoColor {
		fmt.Fprintf(os.Stderr, "❌ "+format+"\n", args...)
	} else {
		_, _ = color.New(color.FgRed).Fprintf(os.Stderr, "❌ "+format+"\n", args...)
	}
}

// Warning prints a warning message in yellow.
func (co *ColoredOutput) Warning(format string, args ...any) {
	if co.Quiet {
		return
	}
	if co.NoColor {
		fmt.Printf("⚠️  "+format+"\n", args...)
	} else {
		color.Yellow("⚠️  "+format, args...)
	}
}

// Info prints an info message in blue.
func (co *ColoredOutput) Info(format string, args ...any) {
	if co.Quiet {
		return
	}
	if co.NoColor {
		fmt.Printf("ℹ️  "+format+"\n", args...)
	} else {
		color.Blue("ℹ️  "+format, args...)
	}
}

// Progress prints a progress message in cyan.
func (co *ColoredOutput) Progress(format string, args ...any) {
	if co.Quiet {
		return
	}
	if co.NoColor {
		fmt.Printf("🔄 "+format+"\n", args...)
	} else {
		color.Cyan("🔄 "+format, args...)
	}
}

// Bold prints text in bold.
func (co *ColoredOutput) Bold(format string, args ...any) {
	if co.Quiet {
		return
	}
	if co.NoColor {
		fmt.Printf(format+"\n", args...)
	} else {
		_, _ = color.New(color.Bold).Printf(format+"\n", args...)
	}
}

// Printf prints without color formatting (respects quiet mode).
func (co *ColoredOutput) Printf(format string, args ...any) {
	if co.Quiet {
		return
	}
	fmt.Printf(format, args...)
}

// Fprintf prints to specified writer without color formatting.
func (co *ColoredOutput) Fprintf(w *os.File, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

// ErrorWithSuggestions prints a ContextualError with suggestions and help.
func (co *ColoredOutput) ErrorWithSuggestions(err *errors.ContextualError) {
	if err == nil {
		return
	}

	// Print main error message
	if co.NoColor {
		fmt.Fprintf(os.Stderr, "❌ %s\n", err.Error())
	} else {
		color.Red("❌ %s", err.Error())
	}
}

// ErrorWithContext creates and prints a contextual error with suggestions.
func (co *ColoredOutput) ErrorWithContext(
	code errors.ErrorCode,
	message string,
	context map[string]string,
) {
	suggestions := errors.GetSuggestions(code, context)
	helpURL := errors.GetHelpURL(code)

	contextualErr := errors.New(code, message).
		WithSuggestions(suggestions...).
		WithHelpURL(helpURL)

	if len(context) > 0 {
		contextualErr = contextualErr.WithDetails(context)
	}

	co.ErrorWithSuggestions(contextualErr)
}

// ErrorWithSimpleFix prints an error with a simple suggestion.
func (co *ColoredOutput) ErrorWithSimpleFix(message, suggestion string) {
	contextualErr := errors.New(errors.ErrCodeUnknown, message).
		WithSuggestions(suggestion)

	co.ErrorWithSuggestions(contextualErr)
}

// FormatContextualError formats a ContextualError for display.
func (co *ColoredOutput) FormatContextualError(err *errors.ContextualError) string {
	if err == nil {
		return ""
	}

	var parts []string

	// Add main error message
	parts = append(parts, co.formatMainError(err))

	// Add details section
	if len(err.Details) > 0 {
		parts = append(parts, co.formatDetailsSection(err.Details)...)
	}

	// Add suggestions section
	if len(err.Suggestions) > 0 {
		parts = append(parts, co.formatSuggestionsSection(err.Suggestions)...)
	}

	// Add help URL section
	if err.HelpURL != "" {
		parts = append(parts, co.formatHelpURLSection(err.HelpURL))
	}

	return strings.Join(parts, "\n")
}

// formatMainError formats the main error message with code.
func (co *ColoredOutput) formatMainError(err *errors.ContextualError) string {
	mainMsg := fmt.Sprintf("%s [%s]", err.Error(), err.Code)
	if co.NoColor {
		return "❌ " + mainMsg
	}

	return color.RedString("❌ ") + mainMsg
}

// formatDetailsSection formats the details section.
func (co *ColoredOutput) formatDetailsSection(details map[string]string) []string {
	var parts []string

	if co.NoColor {
		parts = append(parts, "\nDetails:")
	} else {
		parts = append(parts, color.New(color.Bold).Sprint("\nDetails:"))
	}

	for key, value := range details {
		if co.NoColor {
			parts = append(parts, fmt.Sprintf("  %s: %s", key, value))
		} else {
			parts = append(parts, fmt.Sprintf("  %s: %s",
				color.CyanString(key),
				color.WhiteString(value)))
		}
	}

	return parts
}

// formatSuggestionsSection formats the suggestions section.
func (co *ColoredOutput) formatSuggestionsSection(suggestions []string) []string {
	var parts []string

	if co.NoColor {
		parts = append(parts, "\nSuggestions:")
	} else {
		parts = append(parts, color.New(color.Bold).Sprint("\nSuggestions:"))
	}

	for _, suggestion := range suggestions {
		if co.NoColor {
			parts = append(parts, fmt.Sprintf("  • %s", suggestion))
		} else {
			parts = append(parts, fmt.Sprintf("  %s %s",
				color.YellowString("•"),
				color.WhiteString(suggestion)))
		}
	}

	return parts
}

// formatHelpURLSection formats the help URL section.
func (co *ColoredOutput) formatHelpURLSection(helpURL string) string {
	if co.NoColor {
		return fmt.Sprintf("\nFor more help: %s", helpURL)
	}

	return fmt.Sprintf("\n%s: %s",
		color.New(color.Bold).Sprint("For more help"),
		color.BlueString(helpURL))
}
