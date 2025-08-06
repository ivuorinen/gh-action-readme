// Package main is the entry point for the gh-action-readme CLI tool.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/errors"
	"github.com/ivuorinen/gh-action-readme/internal/helpers"
	"github.com/ivuorinen/gh-action-readme/internal/wizard"
)

const (
	// Export format constants.
	formatJSON = "json"
	formatTOML = "toml"
	formatYAML = "yaml"
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
	case formatJSON:
		return wizard.FormatJSON
	case formatTOML:
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

func main() {
	rootCmd := &cobra.Command{
		Use:   "gh-action-readme",
		Short: "Auto-generate beautiful README and HTML documentation for GitHub Actions.",
		Long: `gh-action-readme is a CLI tool for parsing one or many action.yml files and ` +
			`generating informative, modern, and customizable documentation.`,
		PersistentPreRun: initConfig,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default: XDG config directory)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output (overrides verbose)")

	rootCmd.AddCommand(newGenCmd())
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newSchemaCmd())
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  "Print the version number and build information",
		Run: func(cmd *cobra.Command, _ []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
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

func initConfig(_ *cobra.Command, _ []string) {
	var err error

	// Use ConfigurationLoader for loading global configuration
	loader := internal.NewConfigurationLoader()
	globalConfig, err = loader.LoadGlobalConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	// Override with command line flags
	if verbose {
		globalConfig.Verbose = true
	}
	if quiet {
		globalConfig.Quiet = true
		globalConfig.Verbose = false // quiet overrides verbose
	}
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
		Run:  genHandler,
	}

	cmd.Flags().StringP("output-format", "f", "md", "output format: md, html, json, asciidoc")
	cmd.Flags().StringP("output-dir", "o", ".", "output directory")
	cmd.Flags().StringP("output", "", "", "custom output filename (overrides default naming)")
	cmd.Flags().StringP("theme", "t", "", "template theme: github, gitlab, minimal, professional")
	cmd.Flags().BoolP("recursive", "r", false, "search for action.yml files recursively")

	return cmd
}

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate action.yml files and optionally autofill missing fields.",
		Run:   validateHandler,
	}
}

func newSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "schema",
		Short: "Show the action.yml schema info.",
		Run:   schemaHandler,
	}
}

func genHandler(cmd *cobra.Command, args []string) {
	output := createOutputManager(globalConfig.Quiet)

	// Determine target path from arguments or current directory
	var targetPath string
	if len(args) > 0 {
		targetPath = args[0]
	} else {
		var err error
		targetPath, err = helpers.GetCurrentDir()
		if err != nil {
			output.Error("Error getting current directory: %v", err)
			os.Exit(1)
		}
	}

	// Resolve target path to absolute path
	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		output.Error("Error resolving path %s: %v", targetPath, err)
		os.Exit(1)
	}

	// Check if target exists
	info, err := os.Stat(absTargetPath)
	if err != nil {
		output.Error("Path does not exist: %s", targetPath)
		os.Exit(1)
	}

	var workingDir string
	var actionFiles []string

	if info.IsDir() {
		// Target is a directory
		workingDir = absTargetPath
		generator := internal.NewGenerator(globalConfig) // Temporary generator for discovery
		recursive, _ := cmd.Flags().GetBool("recursive")
		actionFiles, err = generator.DiscoverActionFilesWithValidation(
			workingDir,
			recursive,
			"documentation generation",
		)
		if err != nil {
			os.Exit(1)
		}
	} else {
		// Target is a file - validate it's an action file
		lowerPath := strings.ToLower(absTargetPath)
		if !strings.HasSuffix(lowerPath, ".yml") && !strings.HasSuffix(lowerPath, ".yaml") {
			output.Error("File must be a YAML file (.yml or .yaml): %s", targetPath)
			os.Exit(1)
		}
		workingDir = filepath.Dir(absTargetPath)
		actionFiles = []string{absTargetPath}
	}

	repoRoot := helpers.FindGitRepoRoot(workingDir)
	config := loadGenConfig(repoRoot, workingDir)
	applyGlobalFlags(config)
	applyCommandFlags(cmd, config)

	generator := internal.NewGenerator(config)
	logConfigInfo(generator, config, repoRoot)

	processActionFiles(generator, actionFiles)
}

// loadGenConfig loads multi-level configuration using ConfigurationLoader.
func loadGenConfig(repoRoot, currentDir string) *internal.AppConfig {
	loader := internal.NewConfigurationLoader()
	config, err := loader.LoadConfiguration(configFile, repoRoot, currentDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate the loaded configuration
	if err := loader.ValidateConfiguration(config); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration validation error: %v\n", err)
		os.Exit(1)
	}

	return config
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
	outputFormat, _ := cmd.Flags().GetString("output-format")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	outputFilename, _ := cmd.Flags().GetString("output")
	theme, _ := cmd.Flags().GetString("theme")

	if outputFormat != "md" {
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
func processActionFiles(generator *internal.Generator, actionFiles []string) {
	if err := generator.ProcessBatch(actionFiles); err != nil {
		generator.Output.Error("Error during generation: %v", err)
		os.Exit(1)
	}
}

func validateHandler(_ *cobra.Command, _ []string) {
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		_, errorHandler := setupOutputAndErrorHandling()
		errorHandler.HandleSimpleError("Unable to determine current directory", err)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(
		currentDir,
		true,
		"validation",
	) // Recursive for validation
	if err != nil {
		os.Exit(1)
	}

	// Validate the discovered files
	if err := generator.ValidateFiles(actionFiles); err != nil {
		generator.Output.ErrorWithContext(
			errors.ErrCodeValidation,
			"validation failed",
			map[string]string{
				"files_count":            fmt.Sprintf("%d", len(actionFiles)),
				internal.ContextKeyError: err.Error(),
			},
		)
		os.Exit(1)
	}

	generator.Output.Success("\nAll validations passed successfully!")
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
		Run:   configInitHandler,
	})

	initCmd := &cobra.Command{
		Use:   "wizard",
		Short: "Interactive configuration wizard",
		Long:  "Launch an interactive wizard to set up your configuration step by step",
		Run:   configWizardHandler,
	}
	initCmd.Flags().String("format", "yaml", "Export format: yaml, json, toml")
	initCmd.Flags().String("output", "", "Output path (default: XDG config directory)")
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

func configInitHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)

	// Check if config already exists
	configPath, err := internal.GetConfigPath()
	if err != nil {
		output.Error("Failed to get config path: %v", err)
		os.Exit(1)
	}

	if _, err := os.Stat(configPath); err == nil {
		output.Warning("Configuration file already exists at: %s", configPath)
		output.Info("Use 'gh-action-readme config show' to view current configuration")
		return
	}

	// Create default config
	if err := internal.WriteDefaultConfig(); err != nil {
		output.Error("Failed to write default configuration: %v", err)
		os.Exit(1)
	}

	output.Success("Created default configuration at: %s", configPath)
	output.Info("Edit this file to customize your settings")
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
		{internal.ThemeDefault, "Original simple template"},
		{internal.ThemeGitHub, "GitHub-style with badges and collapsible sections"},
		{internal.ThemeGitLab, "GitLab-focused with CI/CD examples"},
		{internal.ThemeMinimal, "Clean and concise documentation"},
		{internal.ThemeProfessional, "Comprehensive with troubleshooting and ToC"},
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
		Run:   depsListHandler,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "security",
		Short: "Analyze dependency security (pinned vs floating versions)",
		Run:   depsSecurityHandler,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "outdated",
		Short: "Check for outdated dependencies",
		Run:   depsOutdatedHandler,
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
		Run:   depsUpgradeHandler,
	}
	upgradeCmd.Flags().Bool("ci", false, "CI/CD mode: automatically pin all updates to commit SHAs")
	upgradeCmd.Flags().Bool("all", false, "Update all outdated dependencies without prompts")
	upgradeCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")
	cmd.AddCommand(upgradeCmd)

	pinCmd := &cobra.Command{
		Use:   "pin",
		Short: "Pin floating versions to specific commits",
		Long:  "Convert floating versions (like @v4) to pinned commit SHAs with version comments.",
		Run:   depsUpgradeHandler, // Uses same handler with different flags
	}
	pinCmd.Flags().Bool("all", false, "Pin all floating dependencies")
	pinCmd.Flags().Bool("dry-run", false, "Show what would be pinned without making changes")
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
		Run:   cacheClearHandler,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "stats",
		Short: "Show cache statistics",
		Run:   cacheStatsHandler,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show cache directory path",
		Run:   cachePathHandler,
	})

	return cmd
}

func depsListHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		output.Error("Error getting current directory: %v", err)
		os.Exit(1)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(currentDir, true, "dependency listing")
	if err != nil {
		// For deps list, we can continue if no files found (show warning instead of error)
		output.Warning("No action files found")
		return
	}

	analyzer := createAnalyzer(generator, output)
	totalDeps := analyzeDependencies(output, actionFiles, analyzer)

	if totalDeps > 0 {
		output.Bold("\nTotal dependencies: %d", totalDeps)
	}
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

func depsSecurityHandler(_ *cobra.Command, _ []string) {
	output, errorHandler := setupOutputAndErrorHandling()

	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		errorHandler.HandleSimpleError("Failed to get current directory", err)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(currentDir, true, "security analysis")
	if err != nil {
		os.Exit(1)
	}

	analyzer := createAnalyzer(generator, output)
	if analyzer == nil {
		return
	}

	pinnedCount, floatingDeps := analyzeSecurityDeps(output, actionFiles, analyzer)
	displaySecuritySummary(output, currentDir, pinnedCount, floatingDeps)
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

func depsOutdatedHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		output.Error("Error getting current directory: %v", err)
		os.Exit(1)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(currentDir, true, "outdated dependency analysis")
	if err != nil {
		// For deps outdated, we can continue if no files found (show warning instead of error)
		output.Warning("No action files found")
		return
	}

	analyzer := createAnalyzer(generator, output)
	if analyzer == nil {
		return
	}

	if !validateGitHubToken(output) {
		return
	}

	allOutdated := checkAllOutdated(output, actionFiles, analyzer)
	displayOutdatedResults(output, allOutdated)
}

// validateGitHubToken checks if GitHub token is available.
func validateGitHubToken(output *internal.ColoredOutput) bool {
	if globalConfig.GitHubToken == "" {
		contextualErr := errors.New(errors.ErrCodeGitHubAuth, "GitHub token not found").
			WithSuggestions(errors.GetSuggestions(errors.ErrCodeGitHubAuth, map[string]string{})...).
			WithHelpURL(errors.GetHelpURL(errors.ErrCodeGitHubAuth))

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
			output.Warning("Error analyzing %s: %v", actionFile, err)
			continue
		}

		outdated, err := analyzer.CheckOutdated(deps)
		if err != nil {
			output.Warning("Error checking outdated for %s: %v", actionFile, err)
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

func depsUpgradeHandler(cmd *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		output.Error("Error getting current directory: %v", err)
		os.Exit(1)
	}

	// Setup and validation
	analyzer, actionFiles := setupDepsUpgrade(output, currentDir)
	if analyzer == nil || len(actionFiles) == 0 {
		return
	}

	// Parse flags and show mode
	ciMode, _ := cmd.Flags().GetBool("ci")
	allFlag, _ := cmd.Flags().GetBool("all")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	isPinCmd := cmd.Use == "pin"

	showUpgradeMode(output, ciMode, isPinCmd)

	// Collect all updates
	allUpdates := collectAllUpdates(output, analyzer, actionFiles)
	if len(allUpdates) == 0 {
		output.Success("✅ No updates needed - all dependencies are current and pinned!")
		return
	}

	// Show and apply updates
	showPendingUpdates(output, allUpdates, currentDir)
	if !dryRun {
		applyUpdates(output, analyzer, allUpdates, ciMode || allFlag)
	} else {
		output.Info("\n🔍 Dry run complete - no changes made")
	}
}

// setupDepsUpgrade handles initial setup and validation for dependency upgrades.
func setupDepsUpgrade(output *internal.ColoredOutput, currentDir string) (*dependencies.Analyzer, []string) {
	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFiles(currentDir, true)
	if err != nil {
		output.Error("Error discovering action files: %v", err)
		os.Exit(1)
	}

	if len(actionFiles) == 0 {
		output.Warning("No action files found")
		return nil, nil
	}

	analyzer, err := generator.CreateDependencyAnalyzer()
	if err != nil {
		output.Warning("Could not create dependency analyzer: %v", err)
		return nil, nil
	}

	if globalConfig.GitHubToken == "" {
		output.Warning("No GitHub token found. Set GITHUB_TOKEN environment variable")
		return nil, nil
	}

	return analyzer, actionFiles
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
			output.Warning("Error analyzing %s: %v", actionFile, err)
			continue
		}

		outdated, err := analyzer.CheckOutdated(deps)
		if err != nil {
			output.Warning("Error checking outdated for %s: %v", actionFile, err)
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
func applyUpdates(
	output *internal.ColoredOutput,
	analyzer *dependencies.Analyzer,
	allUpdates []dependencies.PinnedUpdate,
	automatic bool,
) {
	if automatic {
		output.Info("\n🚀 Applying updates...")
		if err := analyzer.ApplyPinnedUpdates(allUpdates); err != nil {
			output.Error("Failed to apply updates: %v", err)
			os.Exit(1)
		}
		output.Success("✅ Successfully updated %d dependencies with pinned commit SHAs", len(allUpdates))
	} else {
		// Interactive mode
		output.Info("\n❓ This will modify your action.yml files. Continue? (y/N): ")
		var response string
		_, _ = fmt.Scanln(&response) // User input, scan error not critical
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			output.Info("Canceled")
			return
		}

		output.Info("🚀 Applying updates...")
		if err := analyzer.ApplyPinnedUpdates(allUpdates); err != nil {
			output.Error("Failed to apply updates: %v", err)
			os.Exit(1)
		}
		output.Success("✅ Successfully updated %d dependencies", len(allUpdates))
	}
}

func depsGraphHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)
	output.Bold("Dependency Graph:")
	output.Info("Generating visual dependency graph...")
	output.Printf("This feature is not yet implemented\n")
}

func cacheClearHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)
	output.Info("Clearing dependency cache...")

	// Create a cache instance
	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
	if err != nil {
		output.Error("Failed to access cache: %v", err)
		os.Exit(1)
	}

	if err := cacheInstance.Clear(); err != nil {
		output.Error("Failed to clear cache: %v", err)
		os.Exit(1)
	}

	output.Success("Cache cleared successfully")
}

func cacheStatsHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)

	// Create a cache instance
	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
	if err != nil {
		output.Error("Failed to access cache: %v", err)
		os.Exit(1)
	}

	stats := cacheInstance.Stats()

	output.Bold("Cache Statistics:")
	output.Printf("Cache location: %s\n", stats["cache_dir"])
	output.Printf("Total entries: %d\n", stats["total_entries"])
	output.Printf("Expired entries: %d\n", stats["expired_count"])

	// Format size nicely
	totalSize, ok := stats["total_size"].(int64)
	if !ok {
		totalSize = 0
	}
	sizeStr := formatSize(totalSize)
	output.Printf("Total size: %s\n", sizeStr)
}

func cachePathHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)

	// Create a cache instance
	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
	if err != nil {
		output.Error("Failed to access cache: %v", err)
		os.Exit(1)
	}

	stats := cacheInstance.Stats()
	cachePath, ok := stats["cache_dir"].(string)
	if !ok {
		cachePath = "unknown"
	}

	output.Bold("Cache Directory:")
	output.Printf("%s\n", cachePath)

	// Check if directory exists
	if _, err := os.Stat(cachePath); err == nil {
		output.Success("Directory exists")
	} else if os.IsNotExist(err) {
		output.Warning("Directory does not exist (will be created on first use)")
	}
}

func configWizardHandler(cmd *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)

	// Create and run the wizard
	configWizard := wizard.NewConfigWizard(output)
	config, err := configWizard.Run()
	if err != nil {
		output.Error("Wizard failed: %v", err)
		os.Exit(1)
	}

	// Get export format and output path
	format, _ := cmd.Flags().GetString("format")
	outputPath, _ := cmd.Flags().GetString("output")

	// Create exporter and export configuration
	exporter := wizard.NewConfigExporter(output)

	// Use default output path if not specified
	if outputPath == "" {
		exportFormat := resolveExportFormat(format)
		defaultPath, err := exporter.GetDefaultOutputPath(exportFormat)
		if err != nil {
			output.Error("Failed to get default output path: %v", err)
			os.Exit(1)
		}
		outputPath = defaultPath
	}

	// Export the configuration
	exportFormat := resolveExportFormat(format)

	if err := exporter.ExportConfig(config, exportFormat, outputPath); err != nil {
		output.Error("Failed to export configuration: %v", err)
		os.Exit(1)
	}

	output.Info("\n🎉 Configuration wizard completed successfully!")
	output.Info("You can now use 'gh-action-readme gen' to generate documentation.")
}
