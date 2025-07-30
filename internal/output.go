package internal

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// ColoredOutput provides methods for colored terminal output.
type ColoredOutput struct {
	NoColor bool
	Quiet   bool
}

// NewColoredOutput creates a new colored output instance.
func NewColoredOutput(quiet bool) *ColoredOutput {
	return &ColoredOutput{
		NoColor: color.NoColor || os.Getenv("NO_COLOR") != "",
		Quiet:   quiet,
	}
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
