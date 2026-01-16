// Package main is the entry point for the gh-action-readme CLI tool.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/helpers"
	"github.com/ivuorinen/gh-action-readme/internal/wizard"
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

// InputReader interface for reading user input (enables testing).
type InputReader interface {
	ReadLine() (string, error)
}

// StdinReader reads from actual stdin.
type StdinReader struct{}

func (r *StdinReader) ReadLine() (string, error) {
	var response string
	_, err := fmt.Scanln(&response)

	return strings.TrimSpace(response), err
}

// TestInputReader allows injecting test responses for testing.
type TestInputReader struct {
	responses []string
	index     int
}

func (r *TestInputReader) ReadLine() (string, error) {
	if r.index >= len(r.responses) {
		return "", errors.New("no more test responses")
	}
	response := r.responses[r.index]
	r.index++

	return response, nil
}

// Helper functions to reduce duplication.

func createOutputManager(quiet bool) *internal.ColoredOutput {
	return internal.NewColoredOutput(quiet)
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

// resolveExportFormat converts a format string to wizard.ExportFormat.
func resolveExportFormat(format string) wizard.ExportFormat {
	switch format {
	case appconstants.OutputFormatJSON:
		return wizard.FormatJSON
	case appconstants.OutputFormatTOML:
		return wizard.FormatTOML
	default:
		return wizard.FormatYAML
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
			verbose, _ := cmd.Flags().GetBool(appconstants.ConfigKeyVerbose)
			if verbose {
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

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate action.yml files and optionally autofill missing fields.",
		Run:   wrapHandlerWithErrorHandling(validateHandler),
	}
}

func newSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "schema",
		Short: "Show the action.yml schema info.",
		Run:   schemaHandler,
	}
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

func genHandler(cmd *cobra.Command, args []string) error {
	// Ensure globalConfig is initialized
	if globalConfig == nil {
		globalConfig = internal.DefaultAppConfig()
	}

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

func validateHandler(_ *cobra.Command, _ []string) error {
	// Ensure globalConfig is initialized
	if globalConfig == nil {
		globalConfig = internal.DefaultAppConfig()
	}

	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return fmt.Errorf("unable to determine current directory: %w", err)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(
		currentDir,
		true,
		globalConfig.IgnoredDirectories,
		"validation",
	) // Recursive for validation
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToDiscoverActionFiles, err)
	}

	// Validate the discovered files
	if err := generator.ValidateFiles(actionFiles); err != nil {
		return fmt.Errorf("validation failed for %d files: %w", len(actionFiles), err)
	}

	generator.Output.Success("\nAll validations passed successfully!")

	return nil
}

func schemaHandler(_ *cobra.Command, _ []string) {
	output := internal.NewColoredOutput(globalConfig.Quiet)
	if globalConfig.Verbose {
		output.Info("Using schema: %s", globalConfig.Schema)
	}
	output.Printf("Schema: schemas/action.schema.json (replaceable, editable)")
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
		Run: func(_ *cobra.Command, _ []string) {
			output := internal.NewColoredOutput(globalConfig.Quiet)
			path, err := internal.GetConfigPath()
			if err != nil {
				output.Error("Error getting config path: %v", err)

				return
			}
			output.Info("Configuration file location: %s", path)
			if globalConfig.Verbose {
				output.Info("Current config: %+v", globalConfig)
			}
		},
	}

	// Add subcommands
	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize default configuration file",
		Run:   wrapHandlerWithErrorHandling(configInitHandler),
	})

	initCmd := &cobra.Command{
		Use:   "wizard",
		Short: "Interactive configuration wizard",
		Long:  "Launch an interactive wizard to set up your configuration step by step",
		Run:   wrapHandlerWithErrorHandling(configWizardHandler),
	}
	initCmd.Flags().String(appconstants.FlagFormat, "yaml", "Export format: yaml, json, toml")
	initCmd.Flags().String(appconstants.FlagOutput, "", "Output path (default: XDG config directory)")
	cmd.AddCommand(initCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run:   configShowHandler,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "themes",
		Short: "List available themes",
		Run:   configThemesHandler,
	})

	return cmd
}

func configInitHandler(_ *cobra.Command, _ []string) error {
	// Ensure globalConfig is initialized
	if globalConfig == nil {
		globalConfig = internal.DefaultAppConfig()
	}

	output := createOutputManager(globalConfig.Quiet)

	// Check if config already exists
	configPath, err := internal.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		output.Warning("Configuration file already exists at: %s", configPath)
		output.Info("Use 'gh-action-readme config show' to view current configuration")

		return nil
	}

	// Create default config
	if err := internal.WriteDefaultConfig(); err != nil {
		return fmt.Errorf("failed to write default configuration: %w", err)
	}

	output.Success("Created default configuration at: %s", configPath)
	output.Info("Edit this file to customize your settings")

	return nil
}

func configShowHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)

	output.Bold("Current Configuration:")
	output.Printf("Theme: %s\n", globalConfig.Theme)
	output.Printf("Output Format: %s\n", globalConfig.OutputFormat)
	output.Printf("Output Directory: %s\n", globalConfig.OutputDir)
	output.Printf("Template: %s\n", globalConfig.Template)
	output.Printf("Schema: %s\n", globalConfig.Schema)
	output.Printf("Verbose: %t\n", globalConfig.Verbose)
	output.Printf("Quiet: %t\n", globalConfig.Quiet)
}

func configThemesHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)

	output.Bold("Available Themes:")
	themes := []struct {
		name string
		desc string
	}{
		{appconstants.ThemeDefault, "Original simple template"},
		{appconstants.ThemeGitHub, "GitHub-style with badges and collapsible sections"},
		{appconstants.ThemeGitLab, "GitLab-focused with CI/CD examples"},
		{appconstants.ThemeMinimal, "Clean and concise documentation"},
		{appconstants.ThemeProfessional, "Comprehensive with troubleshooting and ToC"},
	}

	for _, theme := range themes {
		if theme.name == globalConfig.Theme {
			output.Success("• %s - %s (current)", theme.name, theme.desc)
		} else {
			output.Printf("• %s - %s\n", theme.name, theme.desc)
		}
	}

	output.Info("\nUse --theme flag or set 'theme' in config file to change theme")
}

func newDepsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deps",
		Short: "Dependency management commands",
		Long:  "Analyze and manage GitHub Action dependencies",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all dependencies in action files",
		Run:   wrapHandlerWithErrorHandling(depsListHandler),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "security",
		Short: "Analyze dependency security (pinned vs floating versions)",
		Run:   wrapHandlerWithErrorHandling(depsSecurityHandler),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "outdated",
		Short: "Check for outdated dependencies",
		Run:   wrapHandlerWithErrorHandling(depsOutdatedHandler),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "graph",
		Short: "Generate dependency graph",
		Run:   depsGraphHandler,
	})

	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade dependencies with interactive or CI mode",
		Long:  "Upgrade dependencies to latest versions. Use --ci for automated pinned updates.",
		Run:   wrapHandlerWithErrorHandling(depsUpgradeHandler),
	}
	upgradeCmd.Flags().Bool(appconstants.FlagCI, false, "CI/CD mode: automatically pin all updates to commit SHAs")
	upgradeCmd.Flags().Bool(appconstants.InputAll, false, "Update all outdated dependencies without prompts")
	upgradeCmd.Flags().Bool(appconstants.InputDryRun, false, "Show what would be updated without making changes")
	cmd.AddCommand(upgradeCmd)

	pinCmd := &cobra.Command{
		Use:   appconstants.CommandPin,
		Short: "Pin floating versions to specific commits",
		Long:  "Convert floating versions (like @v4) to pinned commit SHAs with version comments.",
		Run:   wrapHandlerWithErrorHandling(depsUpgradeHandler), // Uses same handler with different flags
	}
	pinCmd.Flags().Bool(appconstants.InputAll, false, "Pin all floating dependencies")
	pinCmd.Flags().Bool(appconstants.InputDryRun, false, "Show what would be pinned without making changes")
	cmd.AddCommand(pinCmd)

	return cmd
}

func newCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Cache management commands",
		Long:  "Manage the XDG-compliant dependency cache",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear the dependency cache",
		Run:   wrapHandlerWithErrorHandling(cacheClearHandler),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "stats",
		Short: "Show cache statistics",
		Run:   wrapHandlerWithErrorHandling(cacheStatsHandler),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show cache directory path",
		Run:   wrapHandlerWithErrorHandling(cachePathHandler),
	})

	return cmd
}

func depsListHandler(_ *cobra.Command, _ []string) error {
	// Ensure globalConfig is initialized
	if globalConfig == nil {
		globalConfig = internal.DefaultAppConfig()
	}

	output := createOutputManager(globalConfig.Quiet)
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return wrapError(appconstants.ErrErrorGettingCurrentDir, err)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(
		currentDir,
		true,
		globalConfig.IgnoredDirectories,
		"dependency listing",
	)
	if err := handleNoFilesFoundError(err, output); err != nil {
		return err
	}

	analyzer := createAnalyzer(generator, output)
	totalDeps := analyzeDependencies(output, actionFiles, analyzer)

	if totalDeps > 0 {
		output.Bold("\nTotal dependencies: %d", totalDeps)
	}

	return nil
}

// analyzeDependencies analyzes and displays dependencies.
func analyzeDependencies(output *internal.ColoredOutput, actionFiles []string, analyzer *dependencies.Analyzer) int {
	totalDeps := 0
	output.Bold("Dependencies found in action files:")

	// Create progress bar for multiple files
	progressMgr := internal.NewProgressBarManager(output.IsQuiet())

	progressMgr.ProcessWithProgressBar(
		"Analyzing dependencies",
		actionFiles,
		func(actionFile string, bar *progressbar.ProgressBar) {
			if bar == nil {
				output.Info("\n📄 %s", actionFile)
			}
			totalDeps += analyzeActionFileDeps(output, actionFile, analyzer)
		},
	)

	return totalDeps
}

// analyzeActionFileDeps analyzes dependencies in a single action file.
func analyzeActionFileDeps(output *internal.ColoredOutput, actionFile string, analyzer *dependencies.Analyzer) int {
	if analyzer == nil {
		output.Printf("  • Cannot analyze (no GitHub token)\n")

		return 0
	}

	deps, err := analyzer.AnalyzeActionFile(actionFile)
	if err != nil {
		output.Warning("  ⚠️  Error analyzing: %v", err)

		return 0
	}

	if len(deps) == 0 {
		output.Printf("  • No dependencies (not a composite action)\n")

		return 0
	}

	for _, dep := range deps {
		if dep.IsPinned {
			output.Success("  🔒 %s @ %s - %s", dep.Name, dep.Version, dep.Description)
		} else {
			output.Warning("  📌 %s @ %s - %s", dep.Name, dep.Version, dep.Description)
		}
	}

	return len(deps)
}

func depsSecurityHandler(_ *cobra.Command, _ []string) error {
	// Ensure globalConfig is initialized
	if globalConfig == nil {
		globalConfig = internal.DefaultAppConfig()
	}

	output := createOutputManager(globalConfig.Quiet)

	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(
		currentDir,
		true,
		globalConfig.IgnoredDirectories,
		"security analysis",
	)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToDiscoverActionFiles, err)
	}

	analyzer := createAnalyzer(generator, output)
	if analyzer == nil {
		output.Warning(
			"⚠️  Analyzer disabled: GitHub token not configured. " +
				"Use GITHUB_TOKEN or GH_README_GITHUB_TOKEN environment variable.",
		)

		return nil // Analyzer can be nil if token isn't configured, gracefully handle
	}

	pinnedCount, floatingDeps := analyzeSecurityDeps(output, actionFiles, analyzer)
	displaySecuritySummary(output, currentDir, pinnedCount, floatingDeps)

	return nil
}

// analyzeSecurityDeps analyzes dependencies for security issues.
func analyzeSecurityDeps(
	output *internal.ColoredOutput,
	actionFiles []string,
	analyzer *dependencies.Analyzer,
) (int, []struct {
	file string
	dep  dependencies.Dependency
}) {
	pinnedCount := 0
	var floatingDeps []struct {
		file string
		dep  dependencies.Dependency
	}

	output.Bold("Security Analysis of GitHub Action Dependencies:")

	// Create progress bar for multiple files
	progressMgr := internal.NewProgressBarManager(output.IsQuiet())

	progressMgr.ProcessWithProgressBar(
		"Security analysis",
		actionFiles,
		func(actionFile string, _ *progressbar.ProgressBar) {
			deps, err := analyzer.AnalyzeActionFile(actionFile)
			if err != nil {
				return
			}

			for _, dep := range deps {
				if dep.IsPinned {
					pinnedCount++
				} else {
					floatingDeps = append(floatingDeps, struct {
						file string
						dep  dependencies.Dependency
					}{actionFile, dep})
				}
			}
		},
	)

	return pinnedCount, floatingDeps
}

// displaySecuritySummary shows security analysis results.
func displaySecuritySummary(output *internal.ColoredOutput, currentDir string, pinnedCount int, floatingDeps []struct {
	file string
	dep  dependencies.Dependency
}) {
	output.Success("\n🔒 Pinned versions: %d (Recommended for security)", pinnedCount)
	floatingCount := len(floatingDeps)

	if floatingCount > 0 {
		output.Warning("📌 Floating versions: %d (Consider pinning)", floatingCount)
		displayFloatingDeps(output, currentDir, floatingDeps)
		output.Info("\nRecommendation: Pin dependencies to specific commits or semantic versions for better security.")
	} else if pinnedCount > 0 {
		output.Info("\n✅ All dependencies are properly pinned!")
	}
}

// displayFloatingDeps shows floating dependencies details.
func displayFloatingDeps(output *internal.ColoredOutput, currentDir string, floatingDeps []struct {
	file string
	dep  dependencies.Dependency
}) {
	output.Bold("\nFloating dependencies that should be pinned:")
	for _, fd := range floatingDeps {
		relPath, _ := filepath.Rel(currentDir, fd.file)
		output.Warning("  • %s @ %s", fd.dep.Name, fd.dep.Version)
		output.Printf("    in %s\n", relPath)
	}
}

func depsOutdatedHandler(_ *cobra.Command, _ []string) error {
	// Ensure globalConfig is initialized
	if globalConfig == nil {
		globalConfig = internal.DefaultAppConfig()
	}

	output := createOutputManager(globalConfig.Quiet)
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return wrapError(appconstants.ErrErrorGettingCurrentDir, err)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(
		currentDir,
		true,
		globalConfig.IgnoredDirectories,
		"outdated dependency analysis",
	)
	if err := handleNoFilesFoundError(err, output); err != nil {
		return err
	}

	analyzer := createAnalyzer(generator, output)
	if !validateGitHubToken(output) {
		return nil // Not an error, just no token available
	}

	if analyzer == nil {
		return nil // Analyzer can be nil if token isn't configured, gracefully handle
	}

	allOutdated := checkAllOutdated(output, actionFiles, analyzer)
	displayOutdatedResults(output, allOutdated)

	return nil
}

// validateGitHubToken checks if GitHub token is available.
func validateGitHubToken(output *internal.ColoredOutput) bool {
	if globalConfig.GitHubToken == "" {
		contextualErr := apperrors.New(appconstants.ErrCodeGitHubAuth, "GitHub token not found").
			WithSuggestions(apperrors.GetSuggestions(appconstants.ErrCodeGitHubAuth, map[string]string{})...).
			WithHelpURL(apperrors.GetHelpURL(appconstants.ErrCodeGitHubAuth))

		output.Warning("⚠️  %s", contextualErr.Error())

		return false
	}

	return true
}

// checkAllOutdated checks all action files for outdated dependencies.
func checkAllOutdated(
	output *internal.ColoredOutput,
	actionFiles []string,
	analyzer *dependencies.Analyzer,
) []dependencies.OutdatedDependency {
	output.Bold("Checking for outdated dependencies...")
	var allOutdated []dependencies.OutdatedDependency

	for _, actionFile := range actionFiles {
		deps, err := analyzer.AnalyzeActionFile(actionFile)
		if err != nil {
			output.Warning(appconstants.ErrErrorAnalyzing, actionFile, err)

			continue
		}

		outdated, err := analyzer.CheckOutdated(deps)
		if err != nil {
			output.Warning(appconstants.ErrErrorCheckingOutdated, actionFile, err)

			continue
		}

		allOutdated = append(allOutdated, outdated...)
	}

	return allOutdated
}

// displayOutdatedResults shows outdated dependency results.
func displayOutdatedResults(output *internal.ColoredOutput, allOutdated []dependencies.OutdatedDependency) {
	if len(allOutdated) == 0 {
		output.Success("✅ All dependencies are up to date!")

		return
	}

	output.Warning("Found %d outdated dependencies:", len(allOutdated))
	for _, outdated := range allOutdated {
		output.Printf("  • %s: %s → %s (%s update)",
			outdated.Current.Name,
			outdated.Current.Version,
			outdated.LatestVersion,
			outdated.UpdateType)
		if outdated.IsSecurityUpdate {
			output.Warning("    🔒 Potential security update")
		}
	}

	output.Info("\nRun 'gh-action-readme deps upgrade' to update dependencies")
}

func depsUpgradeHandler(cmd *cobra.Command, _ []string) error {
	// Ensure globalConfig is initialized
	if globalConfig == nil {
		globalConfig = internal.DefaultAppConfig()
	}

	output := createOutputManager(globalConfig.Quiet)
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return wrapError(appconstants.ErrErrorGettingCurrentDir, err)
	}

	// Setup and validation
	analyzer, actionFiles, err := setupDepsUpgrade(output, currentDir, nil)
	if err != nil {
		// setupDepsUpgrade returns descriptive errors, so just pass them through
		return err
	}

	// Parse flags and show mode
	ciMode, _ := cmd.Flags().GetBool(appconstants.FlagCI)
	allFlag, _ := cmd.Flags().GetBool(appconstants.InputAll)
	dryRun, _ := cmd.Flags().GetBool(appconstants.InputDryRun)
	isPinCmd := cmd.Use == appconstants.CommandPin

	showUpgradeMode(output, ciMode, isPinCmd)

	// Collect all updates
	allUpdates := collectAllUpdates(output, analyzer, actionFiles)
	if len(allUpdates) == 0 {
		output.Success("✅ No updates needed - all dependencies are current and pinned!")

		return nil
	}

	// Show and apply updates
	showPendingUpdates(output, allUpdates, currentDir)
	if !dryRun {
		if err := applyUpdates(output, analyzer, allUpdates, ciMode || allFlag, nil); err != nil {
			return err
		}
	} else {
		output.Info("\n🔍 Dry run complete - no changes made")
	}

	return nil
}

// setupDepsUpgrade handles initial setup and validation for dependency upgrades.
// The config parameter allows injection for testing (pass nil to use globalConfig).
func setupDepsUpgrade(
	_ *internal.ColoredOutput,
	currentDir string,
	config *internal.AppConfig,
) (*dependencies.Analyzer, []string, error) {
	// Default to globalConfig if not provided (backward compatible)
	if config == nil {
		if globalConfig == nil {
			globalConfig = internal.DefaultAppConfig()
		}
		config = globalConfig
	}

	generator := internal.NewGenerator(config)
	actionFiles, err := generator.DiscoverActionFiles(currentDir, true, config.IgnoredDirectories)
	if err != nil {
		return nil, nil, fmt.Errorf("error discovering action files: %w", err)
	}

	if len(actionFiles) == 0 {
		return nil, nil, errors.New(appconstants.ErrNoActionFilesFound)
	}

	analyzer, err := generator.CreateDependencyAnalyzer()
	if err != nil {
		return nil, nil, fmt.Errorf("could not create dependency analyzer: %w", err)
	}

	if config.GitHubToken == "" {
		return nil, nil, errors.New("no GitHub token found, set GITHUB_TOKEN environment variable")
	}

	return analyzer, actionFiles, nil
}

// showUpgradeMode displays the current upgrade mode to the user.
func showUpgradeMode(output *internal.ColoredOutput, ciMode, isPinCmd bool) {
	switch {
	case ciMode:
		output.Bold("🤖 CI/CD Mode: Automated dependency updates with pinned commit SHAs")
	case isPinCmd:
		output.Bold("📌 Pinning floating dependencies to commit SHAs")
	default:
		output.Bold("🔄 Interactive dependency upgrade")
	}
}

// collectAllUpdates gathers all available updates from action files.
func collectAllUpdates(
	output *internal.ColoredOutput,
	analyzer *dependencies.Analyzer,
	actionFiles []string,
) []dependencies.PinnedUpdate {
	var allUpdates []dependencies.PinnedUpdate

	for _, actionFile := range actionFiles {
		deps, err := analyzer.AnalyzeActionFile(actionFile)
		if err != nil {
			output.Warning(appconstants.ErrErrorAnalyzing, actionFile, err)

			continue
		}

		outdated, err := analyzer.CheckOutdated(deps)
		if err != nil {
			output.Warning(appconstants.ErrErrorCheckingOutdated, actionFile, err)

			continue
		}

		for _, outdatedDep := range outdated {
			update, err := analyzer.GeneratePinnedUpdate(
				actionFile,
				outdatedDep.Current,
				outdatedDep.LatestVersion,
				outdatedDep.LatestSHA,
			)
			if err != nil {
				output.Warning("Error generating update for %s: %v", outdatedDep.Current.Name, err)

				continue
			}
			allUpdates = append(allUpdates, *update)
		}
	}

	return allUpdates
}

// showPendingUpdates displays what updates will be applied.
func showPendingUpdates(
	output *internal.ColoredOutput,
	allUpdates []dependencies.PinnedUpdate,
	currentDir string,
) {
	output.Info("Found %d dependencies to update:", len(allUpdates))
	for _, update := range allUpdates {
		relPath, _ := filepath.Rel(currentDir, update.FilePath)
		output.Printf("  • %s (%s update)", update.OldUses, update.UpdateType)
		output.Printf("    → %s", update.NewUses)
		output.Printf("    in %s", relPath)
	}
}

// applyUpdates applies the collected updates either automatically or interactively.
// The reader parameter allows injection of input for testing (pass nil to use stdin).
func applyUpdates(
	output *internal.ColoredOutput,
	analyzer *dependencies.Analyzer,
	allUpdates []dependencies.PinnedUpdate,
	automatic bool,
	reader InputReader,
) error {
	// Default to stdin if not provided
	if reader == nil {
		reader = &StdinReader{}
	}

	if automatic {
		output.Info("\n🚀 Applying updates...")
		if err := analyzer.ApplyPinnedUpdates(allUpdates); err != nil {
			return fmt.Errorf(appconstants.ErrFailedToApplyUpdatesWrapped, err)
		}
		output.Success("✅ Successfully updated %d dependencies with pinned commit SHAs", len(allUpdates))
	} else {
		// Interactive mode
		output.Info("\n❓ This will modify your action.yml files. Continue? (y/N): ")
		response, err := reader.ReadLine()
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != appconstants.InputYes {
			output.Info("Canceled")

			return nil
		}

		output.Info("🚀 Applying updates...")
		if err := analyzer.ApplyPinnedUpdates(allUpdates); err != nil {
			return fmt.Errorf(appconstants.ErrFailedToApplyUpdatesWrapped, err)
		}
		output.Success("✅ Successfully updated %d dependencies", len(allUpdates))
	}

	return nil
}

func depsGraphHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)
	output.Bold("Dependency Graph:")
	output.Info("Generating visual dependency graph...")
	output.Printf("This feature is not yet implemented\n")
}

func cacheClearHandler(_ *cobra.Command, _ []string) error {
	// Ensure globalConfig is initialized
	if globalConfig == nil {
		globalConfig = internal.DefaultAppConfig()
	}

	output := createOutputManager(globalConfig.Quiet)
	output.Info("Clearing dependency cache...")

	// Create a cache instance
	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
	if err != nil {
		return wrapError(appconstants.ErrFailedToAccessCache, err)
	}

	if err := cacheInstance.Clear(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	output.Success("Cache cleared successfully")

	return nil
}

func cacheStatsHandler(_ *cobra.Command, _ []string) error {
	// Ensure globalConfig is initialized
	if globalConfig == nil {
		globalConfig = internal.DefaultAppConfig()
	}

	output := createOutputManager(globalConfig.Quiet)

	// Create a cache instance
	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
	if err != nil {
		return wrapError(appconstants.ErrFailedToAccessCache, err)
	}

	stats := cacheInstance.Stats()

	output.Bold("Cache Statistics:")
	output.Printf("Cache location: %s\n", stats[appconstants.CacheStatsKeyDir])
	output.Printf("Total entries: %d\n", stats["total_entries"])
	output.Printf("Expired entries: %d\n", stats["expired_count"])

	// Format size nicely
	totalSize, ok := stats["total_size"].(int64)
	if !ok {
		totalSize = 0
	}
	sizeStr := formatSize(totalSize)
	output.Printf("Total size: %s\n", sizeStr)

	return nil
}

func cachePathHandler(_ *cobra.Command, _ []string) error {
	// Ensure globalConfig is initialized
	if globalConfig == nil {
		globalConfig = internal.DefaultAppConfig()
	}

	output := createOutputManager(globalConfig.Quiet)

	// Create a cache instance
	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
	if err != nil {
		return wrapError(appconstants.ErrFailedToAccessCache, err)
	}

	stats := cacheInstance.Stats()
	cachePath, ok := stats[appconstants.CacheStatsKeyDir].(string)
	if !ok {
		cachePath = appconstants.ScopeUnknown
	}

	output.Bold("Cache Directory:")
	output.Printf("%s\n", cachePath)

	// Check if directory exists
	if _, err := os.Stat(cachePath); err == nil {
		output.Success("Directory exists")
	} else if os.IsNotExist(err) {
		output.Warning("Directory does not exist (will be created on first use)")
	}

	return nil
}

func configWizardHandler(cmd *cobra.Command, _ []string) error {
	// Ensure globalConfig is initialized
	if globalConfig == nil {
		globalConfig = internal.DefaultAppConfig()
	}

	output := createOutputManager(globalConfig.Quiet)

	// Create and run the wizard
	configWizard := wizard.NewConfigWizard(output)
	config, err := configWizard.Run()
	if err != nil {
		return fmt.Errorf("wizard failed: %w", err)
	}

	// Get export format and output path
	format, _ := cmd.Flags().GetString(appconstants.FlagFormat)
	outputPath, _ := cmd.Flags().GetString(appconstants.FlagOutput)

	// Create exporter and export configuration
	exporter := wizard.NewConfigExporter(output)

	// Use default output path if not specified
	if outputPath == "" {
		exportFormat := resolveExportFormat(format)
		defaultPath, err := exporter.GetDefaultOutputPath(exportFormat)
		if err != nil {
			return fmt.Errorf("failed to get default output path: %w", err)
		}
		outputPath = defaultPath
	}

	// Export the configuration
	exportFormat := resolveExportFormat(format)

	if err := exporter.ExportConfig(config, exportFormat, outputPath); err != nil {
		return fmt.Errorf("failed to export configuration: %w", err)
	}

	output.Info("\n🎉 Configuration wizard completed successfully!")
	output.Info("You can now use 'gh-action-readme gen' to generate documentation.")

	return nil
}
