// Package helpers provides helper functions used across the application.
package helpers

import (
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
)

// CreateAnalyzer creates a dependency analyzer with standardized error handling.
// Returns nil if creation fails (error already logged to output).
func CreateAnalyzer(generator *internal.Generator, output *internal.ColoredOutput) *dependencies.Analyzer {
	analyzer, err := generator.CreateDependencyAnalyzer()
	if err != nil {
		output.Warning("Could not create dependency analyzer: %v", err)
		return nil
	}
	return analyzer
}

// CreateAnalyzerOrExit creates a dependency analyzer or exits on failure.
func CreateAnalyzerOrExit(generator *internal.Generator, output *internal.ColoredOutput) *dependencies.Analyzer {
	analyzer := CreateAnalyzer(generator, output)
	if analyzer == nil {
		// Error already logged, just exit
		return nil
	}
	return analyzer
}
