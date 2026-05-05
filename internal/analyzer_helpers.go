package internal

import (
	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
)

// CreateAnalyzer creates a dependency analyzer with standardized error handling.
// Returns nil if creation fails (error already logged to output).
func CreateAnalyzer(generator *Generator, output MessageLogger) *dependencies.Analyzer {
	analyzer, err := generator.CreateDependencyAnalyzer()
	if err != nil {
		output.Warning(appconstants.ErrCouldNotCreateDependencyAnalyzer, err)

		return nil
	}

	return analyzer
}
