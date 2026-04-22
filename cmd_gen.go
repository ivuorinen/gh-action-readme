// Package main provides the gen command for generating README documentation.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/helpers"
)

func newGenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen [directory_or_file]",
		Short: "Generate README.md and/or HTML for GitHub Action files.",
		Long: `Generate documentation for GitHub Actions.

Examples:
	gh-action-readme gen                               # Current directory
	gh-action-readme gen testdata/example-action/     # Specific directory
	gh-action-readme gen testdata/action.yml          # Specific file
	gh-action-readme gen -f html testdata/action/     # HTML format
	gh-action-readme gen -f html --output custom.html testdata/action/
	gh-action-readme gen --output docs/action1.html testdata/action1/`,
		Args: cobra.MaximumNArgs(1),
		Run:  wrapHandlerWithErrorHandling(genHandler),
	}

	cmd.Flags().StringP(
		appconstants.FlagOutputFormat,
		"f",
		appconstants.OutputFormatMarkdown,
		"output format: md, html, json, asciidoc",
	)
	cmd.Flags().StringP(appconstants.FlagOutputDir, "o", ".", "output directory")
	cmd.Flags().StringP(appconstants.FlagOutput, "", "", "custom output filename (overrides default naming)")
	cmd.Flags().StringP(appconstants.ConfigKeyTheme, "t", "", "template theme: github, gitlab, minimal, professional")
	cmd.Flags().BoolP(appconstants.FlagRecursive, "r", false, "search for action.yml files recursively")
	cmd.Flags().StringSlice(
		appconstants.FlagIgnoreDirs,
		[]string{},
		"directories to ignore during recursive discovery (comma-separated)",
	)

	return cmd
}

func genHandler(cmd *cobra.Command, args []string) error {
	// Resolve and validate target path
	absTargetPath, info, err := resolveAndValidateTargetPath(args)
	if err != nil {
		return err
	}

	var workingDir string
	var actionFiles []string

	if info.IsDir() {
		// Target is a directory
		workingDir = absTargetPath
		generator := internal.NewGenerator(globalConfig) // Temporary generator for discovery
		recursive, _ := cmd.Flags().GetBool(appconstants.FlagRecursive)

		// Get ignored directories from CLI flag or use config defaults
		ignoredDirs, _ := cmd.Flags().GetStringSlice(appconstants.FlagIgnoreDirs)
		if len(ignoredDirs) == 0 {
			ignoredDirs = globalConfig.IgnoredDirectories
		}

		actionFiles, err = generator.DiscoverActionFilesWithValidation(
			workingDir,
			recursive,
			ignoredDirs,
			"documentation generation",
		)
		if err != nil {
			return fmt.Errorf(appconstants.ErrFailedToDiscoverActionFiles, err)
		}
	} else {
		// Target is a file - validate it's an action file
		lowerPath := strings.ToLower(absTargetPath)
		if !strings.HasSuffix(lowerPath, ".yml") && !strings.HasSuffix(lowerPath, ".yaml") {
			return fmt.Errorf("file must be a YAML file (.yml or .yaml): %s", absTargetPath)
		}
		workingDir = filepath.Dir(absTargetPath)
		actionFiles = []string{absTargetPath}
	}

	repoRoot := helpers.FindGitRepoRoot(workingDir)
	config, err := loadGenConfig(repoRoot, workingDir)
	if err != nil {
		return err
	}
	applyGlobalFlags(config)
	applyCommandFlags(cmd, config)

	generator := internal.NewGenerator(config)
	logConfigInfo(generator, config, repoRoot)

	return processActionFiles(generator, actionFiles)
}

// resolveAndValidateTargetPath resolves the target path from arguments or current directory,
// validates it exists, and returns the absolute path and file info.
func resolveAndValidateTargetPath(args []string) (string, os.FileInfo, error) {
	// Determine target path from arguments or current directory
	var targetPath string
	if len(args) > 0 {
		targetPath = args[0]
	} else {
		var err error
		targetPath, err = helpers.GetCurrentDir()
		if err != nil {
			return "", nil, wrapError(appconstants.ErrErrorGettingCurrentDir, err)
		}
	}

	// Resolve target path to absolute path
	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", nil, fmt.Errorf("error resolving path %s: %w", targetPath, err)
	}

	// Check if target exists
	info, err := os.Stat(absTargetPath)
	if err != nil {
		return "", nil, fmt.Errorf("path does not exist: %s", targetPath)
	}

	return absTargetPath, info, nil
}

// loadGenConfig loads multi-level configuration using ConfigurationLoader.
func loadGenConfig(repoRoot, currentDir string) (*internal.AppConfig, error) {
	loader := internal.NewConfigurationLoader()
	config, err := loader.LoadConfiguration(configFile, repoRoot, currentDir)
	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %w", err)
	}

	// Validate the loaded configuration
	if err := loader.ValidateConfiguration(config); err != nil {
		return nil, fmt.Errorf("configuration validation error: %w", err)
	}

	return config, nil
}

// applyGlobalFlags applies global verbose/quiet flags.
func applyGlobalFlags(config *internal.AppConfig) {
	if verbose {
		config.Verbose = true
	}
	if quiet {
		config.Quiet = true
		config.Verbose = false
	}
}

// applyCommandFlags applies command-specific flags.
func applyCommandFlags(cmd *cobra.Command, config *internal.AppConfig) {
	outputFormat, _ := cmd.Flags().GetString(appconstants.FlagOutputFormat)
	outputDir, _ := cmd.Flags().GetString(appconstants.FlagOutputDir)
	outputFilename, _ := cmd.Flags().GetString(appconstants.FlagOutput)
	theme, _ := cmd.Flags().GetString(appconstants.ConfigKeyTheme)

	if outputFormat != appconstants.OutputFormatMarkdown {
		config.OutputFormat = outputFormat
	}
	if outputDir != "." {
		config.OutputDir = outputDir
	}
	if outputFilename != "" {
		config.OutputFilename = outputFilename
	}
	if theme != "" {
		config.Theme = theme
	}
}

// logConfigInfo logs configuration details if verbose.
func logConfigInfo(generator *internal.Generator, config *internal.AppConfig, repoRoot string) {
	if config.Verbose {
		generator.Output.Info("Using effective config: %+v", config)
		if repoRoot != "" {
			generator.Output.Info("Repository root: %s", repoRoot)
		}
	}
}

// processActionFiles processes discovered files.
func processActionFiles(generator *internal.Generator, actionFiles []string) error {
	if err := generator.ProcessBatch(actionFiles); err != nil {
		return fmt.Errorf("error during generation: %w", err)
	}

	return nil
}
