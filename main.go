// Package main is the entry point for the gh-action-readme CLI tool.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/helpers"
)

var (
	// Version information (set by GoReleaser).
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"

	// Application state.
	globalConfig *internal.AppConfig
	configFile   string
	verbose      bool
	quiet        bool
)

// Helper functions to reduce duplication.

func createOutputManager(isQuiet bool) *internal.ColoredOutput {
	return internal.NewColoredOutput(isQuiet)
}

// formatSize formats a byte size into a human-readable string.
func formatSize(totalSize int64) string {
	if totalSize == 0 {
		return "0 bytes"
	}

	const unit = 1024
	switch {
	case totalSize < unit:
		return fmt.Sprintf("%d bytes", totalSize)
	case totalSize < unit*unit:
		return fmt.Sprintf("%.2f KB", float64(totalSize)/unit)
	case totalSize < unit*unit*unit:
		return fmt.Sprintf("%.2f MB", float64(totalSize)/(unit*unit))
	default:
		return fmt.Sprintf("%.2f GB", float64(totalSize)/(unit*unit*unit))
	}
}

// createErrorHandler creates an error handler for the given output manager.
func createErrorHandler(output *internal.ColoredOutput) *internal.ErrorHandler {
	return internal.NewErrorHandler(output)
}

// setupOutputAndErrorHandling creates output manager and error handler for commands.
func setupOutputAndErrorHandling() (*internal.ColoredOutput, *internal.ErrorHandler) {
	output := createOutputManager(globalConfig.Quiet)
	errorHandler := createErrorHandler(output)

	return output, errorHandler
}

func createAnalyzer(generator *internal.Generator, output *internal.ColoredOutput) *dependencies.Analyzer {
	return helpers.CreateAnalyzer(generator, output)
}

// wrapHandlerWithErrorHandling converts error-returning handler to Cobra handler.
// This allows handlers to return errors for testing while maintaining Cobra compatibility.
func wrapHandlerWithErrorHandling(handler func(*cobra.Command, []string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		// Ensure globalConfig is initialized (important for testing)
		if globalConfig == nil {
			globalConfig = internal.DefaultAppConfig()
		}

		if err := handler(cmd, args); err != nil {
			output := createOutputManager(globalConfig.Quiet)
			output.Error(err.Error())
			os.Exit(1)
		}
	}
}

// wrapError wraps an error with a message constant.
// This is a helper to reduce duplication of the fmt.Errorf("%s: %w", msg, err) pattern.
func wrapError(msgConstant string, err error) error {
	return fmt.Errorf("%s: %w", msgConstant, err)
}

// handleNoFilesFoundError handles errors where no action files are found, showing a warning instead of failing.
// Returns nil if the error is about no files found (after showing warning), otherwise returns the original error.
func handleNoFilesFoundError(err error, output *internal.ColoredOutput) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), appconstants.ErrNoActionFilesFound) {
		output.Warning(appconstants.ErrNoActionFilesFound)

		return nil
	}

	return err
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "gh-action-readme",
		Short: "Auto-generate beautiful README and HTML documentation for GitHub Actions.",
		Long: `gh-action-readme is a CLI tool for parsing one or many action.yml files and ` +
			`generating informative, modern, and customizable documentation.`,
		PersistentPreRunE: initConfig,
	}

	// Global flags
	configDesc := "config file (default: XDG config directory)"
	rootCmd.PersistentFlags().StringVar(&configFile, appconstants.ContextKeyConfig, "", configDesc)
	rootCmd.PersistentFlags().BoolVarP(&verbose, appconstants.ConfigKeyVerbose, "v", false, "verbose output")
	quietDesc := "quiet output (overrides verbose)"
	rootCmd.PersistentFlags().BoolVarP(&quiet, appconstants.ConfigKeyQuiet, "q", false, quietDesc)

	rootCmd.AddCommand(newGenCmd())
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newSchemaCmd())
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  "Print the version number and build information",
		Run: func(cmd *cobra.Command, _ []string) {
			isVerbose, _ := cmd.Flags().GetBool(appconstants.ConfigKeyVerbose)
			if isVerbose {
				fmt.Printf("gh-action-readme version %s\n", version)
				fmt.Printf("  commit: %s\n", commit)
				fmt.Printf("  built at: %s\n", date)
				fmt.Printf("  built by: %s\n", builtBy)
			} else {
				fmt.Println(version)
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "about",
		Short: "About this tool",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("gh-action-readme: Generates README.md and HTML for GitHub Actions. MIT License.")
		},
	})
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newDepsCmd())
	rootCmd.AddCommand(newCacheCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initConfig(_ *cobra.Command, _ []string) error {
	var err error

	// Use ConfigurationLoader for loading global configuration
	loader := internal.NewConfigurationLoader()
	globalConfig, err = loader.LoadGlobalConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Override with command line flags
	if verbose {
		globalConfig.Verbose = true
	}
	if quiet {
		globalConfig.Quiet = true
		globalConfig.Verbose = false // quiet overrides verbose
	}

	return nil
}
