// Package internal contains the core generator functionality.
package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/schollz/progressbar/v3"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/git"
)

// newCacheFunc is the function used to create a cache instance.
// It is a variable so tests can replace it to simulate cache-creation failures.
var newCacheFunc = cache.NewCache

// Generator orchestrates the documentation generation process.
// It uses focused interfaces to reduce coupling and improve testability.
type Generator struct {
	Config   *AppConfig
	Output   CompleteOutput
	Progress ProgressManager
}

// isUnitTestEnvironment detects if we're running unit tests (not integration tests).
// Integration tests compile and exec a production binary; unit tests run inside a *.test binary.
func isUnitTestEnvironment() bool {
	// All Go test binaries have the ".test" suffix (or ".test.exe" on Windows).
	if strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe") {
		return true
	}

	// Explicit opt-in for environments where argv[0] is non-standard.
	return os.Getenv("UNIT_TEST_MODE") != ""
}

// NewGenerator creates a new generator instance with the provided configuration.
// This constructor maintains backward compatibility by using concrete implementations.
// In unit test environments, it automatically uses NullOutput to suppress output.
// If config is nil, it uses DefaultAppConfig() to prevent panics.
func NewGenerator(config *AppConfig) *Generator {
	// Handle nil config gracefully
	if config == nil {
		config = DefaultAppConfig()
	}

	// Use null output in unit test environments to keep tests clean
	// Integration tests need real output to verify CLI behavior
	if isUnitTestEnvironment() {
		return NewGeneratorWithDependencies(
			config,
			NewNullOutput(),
			NewNullProgressManager(),
		)
	}

	return NewGeneratorWithDependencies(
		config,
		NewColoredOutput(config.Quiet),
		NewProgressBarManager(config.Quiet),
	)
}

// NewGeneratorWithDependencies creates a new generator with dependency injection.
// This constructor allows for better testability and flexibility by accepting interfaces.
func NewGeneratorWithDependencies(
	config *AppConfig,
	output CompleteOutput,
	progress ProgressManager,
) *Generator {
	return &Generator{
		Config:   config,
		Output:   output,
		Progress: progress,
	}
}

// CreateDependencyAnalyzer creates a dependency analyzer with GitHub client and cache.
func (g *Generator) CreateDependencyAnalyzer() (*dependencies.Analyzer, error) {
	// Get git info
	repoRoot, err := git.FindRepositoryRoot(".")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	gitInfo, err := git.DetectRepository(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to detect repository info: %w", err)
	}

	// Create GitHub client if token is available
	var githubClient *github.Client
	if token := GetGitHubToken(g.Config); token != "" {
		clientWrapper, err := NewGitHubClient(token)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub client: %w", err)
		}
		githubClient = clientWrapper.Client
	}

	// Create cache
	depCache, err := newCacheFunc(cache.DefaultConfig())
	if err != nil {
		// Continue without cache
		depCache = nil
	}

	// Create cache adapter
	var cacheAdapter dependencies.DependencyCache
	if depCache != nil {
		cacheAdapter = dependencies.NewCacheAdapter(depCache)
	} else {
		cacheAdapter = dependencies.NewNoOpCache()
	}

	return dependencies.NewAnalyzer(githubClient, *gitInfo, cacheAdapter), nil
}

// GenerateFromFile processes a single action.yml file and generates documentation.
func (g *Generator) GenerateFromFile(actionPath string) error {
	if g.Config.Verbose {
		g.Output.Progress("Processing file: %s", actionPath)
	}

	action, err := g.parseAndValidateAction(actionPath)
	if err != nil {
		return err
	}

	outputDir := g.determineOutputDir(actionPath)

	return g.generateByFormat(action, outputDir, actionPath)
}

// DiscoverActionFiles finds action.yml and action.yaml files in the given directory
// using the centralized parser function and adds verbose logging.
func (g *Generator) DiscoverActionFiles(dir string, recursive bool, ignoredDirs []string) ([]string, error) {
	actionFiles, err := DiscoverActionFiles(dir, recursive, ignoredDirs)
	if err != nil {
		return nil, err
	}

	// Add verbose logging
	if g.Config.Verbose {
		for _, file := range actionFiles {
			if recursive {
				g.Output.Info("Discovered action file: %s", file)
			} else {
				g.Output.Info("Found action file: %s", file)
			}
		}
	}

	return actionFiles, nil
}

// DiscoverActionFilesWithValidation discovers action files with centralized error handling and validation.
// This function consolidates the duplicated file discovery logic across the codebase.
func (g *Generator) DiscoverActionFilesWithValidation(
	dir string,
	recursive bool,
	ignoredDirs []string,
	context string,
) ([]string, error) {
	// Discover action files
	actionFiles, err := g.DiscoverActionFiles(dir, recursive, ignoredDirs)
	if err != nil {
		g.Output.ErrorWithContext(
			appconstants.ErrCodeFileNotFound,
			"failed to discover action files for "+context,
			map[string]string{
				"directory":                  dir,
				"recursive":                  strconv.FormatBool(recursive),
				"context":                    context,
				appconstants.ContextKeyError: err.Error(),
			},
		)

		return nil, err
	}

	// Check if any files were found
	if len(actionFiles) == 0 {
		contextMsg := "no GitHub Action files found for " + context
		g.Output.ErrorWithContext(
			appconstants.ErrCodeNoActionFiles,
			contextMsg,
			map[string]string{
				"directory":  dir,
				"recursive":  strconv.FormatBool(recursive),
				"context":    context,
				"suggestion": "Please run this command in a directory containing GitHub Action files (action.yml or action.yaml)",
			},
		)

		return nil, fmt.Errorf("no action files found in directory: %s", dir)
	}

	return actionFiles, nil
}

// ProcessBatch processes multiple action.yml files.
func (g *Generator) ProcessBatch(paths []string) error {
	if len(paths) == 0 {
		return errors.New("no action files to process")
	}

	// A single explicit --output filename cannot disambiguate multiple inputs:
	// every file would resolve to the same path and silently overwrite the prior
	// one. Fail fast and point the user at --output-dir instead.
	if len(paths) > 1 && g.Config.OutputFilename != "" {
		return fmt.Errorf(
			"--output filename cannot be used with %d action files; "+
				"use --output-dir to write one file per action",
			len(paths),
		)
	}

	bar := g.Progress.CreateProgressBarForFiles("Processing files", paths)
	parseErrors, successCount := g.processFiles(paths, bar)
	g.Progress.FinishProgressBarWithNewline(bar)
	g.reportResults(successCount, parseErrors)

	if len(parseErrors) > 0 {
		return fmt.Errorf("encountered %d errors during batch processing", len(parseErrors))
	}

	return nil
}

// ValidateFiles validates multiple action.yml files and reports results.
func (g *Generator) ValidateFiles(paths []string) error {
	if len(paths) == 0 {
		return errors.New("no action files to validate")
	}

	bar := g.Progress.CreateProgressBarForFiles("Validating files", paths)
	allResults, parseErrors := g.validateFiles(paths, bar)
	g.Progress.FinishProgressBarWithNewline(bar)

	if !g.Config.Quiet {
		g.reportValidationResults(allResults, parseErrors)
	}

	// Count validation failures (files with missing required fields)
	validationFailures := 0
	for _, result := range allResults {
		// Each result starts with "file: <path>" so check if there are actual missing fields beyond that
		if len(result.MissingFields) > 1 {
			validationFailures++
		}
	}

	if len(parseErrors) > 0 || validationFailures > 0 {
		totalFailures := len(parseErrors) + validationFailures

		return fmt.Errorf("validation failed for %d files", totalFailures)
	}

	return nil
}

// resolveTemplatePathForFormat determines the correct template path
// based on the configured theme or custom template path.
// If a theme is specified, it takes precedence over the template path.
func (g *Generator) resolveTemplatePathForFormat() (string, error) {
	if g.Config.Theme != "" {
		if err := g.validateTheme(); err != nil {
			return "", err
		}

		return resolveThemeTemplate(g.Config.Theme), nil
	}

	return g.Config.Template, nil
}

// validateTheme rejects a non-empty --theme that is not a known theme. An unknown
// theme must be an explicit error rather than a silent fall-through to the default
// template — a typo'd --theme would otherwise produce default output with no
// warning. The valid-theme list is derived from the canonical appconstants source
// so it cannot drift from the themes the generator actually accepts.
func (g *Generator) validateTheme() error {
	if g.Config.Theme != "" && resolveThemeTemplate(g.Config.Theme) == "" {
		return fmt.Errorf(
			"unknown theme %q; valid themes: %s",
			g.Config.Theme, strings.Join(appconstants.GetSupportedThemes(), ", "),
		)
	}

	return nil
}

// renderTemplateForAction builds template data and renders it using the specified options.
// It finds the repository root for git information, builds comprehensive template data,
// and renders the template. Returns the rendered content or an error.
func (g *Generator) renderTemplateForAction(
	action *ActionYML,
	outputDir string,
	actionPath string,
	opts TemplateOptions,
) (string, error) {
	// Find repository root for git information
	repoRoot, _ := git.FindRepositoryRoot(outputDir)

	// Build comprehensive template data
	templateData := BuildTemplateData(action, g.Config, repoRoot, actionPath)

	// Render template with data
	content, err := RenderReadme(templateData, opts)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return content, nil
}

// generateSimpleFormat is a helper for generating simple text-based formats (Markdown, AsciiDoc).
// It consolidates the common pattern of template rendering, file writing, and success messaging.
func (g *Generator) generateSimpleFormat(
	action *ActionYML,
	outputDir, actionPath string,
	format, defaultFilename, successMsg string,
) error {
	templatePath, err := g.resolveTemplatePathForFormat()
	if err != nil {
		return err
	}

	opts := TemplateOptions{
		TemplatePath: templatePath,
		Format:       format,
	}

	content, err := g.renderTemplateForAction(action, outputDir, actionPath, opts)
	if err != nil {
		return fmt.Errorf("failed to render %s template: %w", format, err)
	}

	outputPath, err := g.resolveOutputPath(outputDir, defaultFilename)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToResolveOutputPath, err)
	}
	if err := os.WriteFile(outputPath, []byte(content), appconstants.FilePermDefault); err != nil {
		// #nosec G306 -- output file permissions
		return fmt.Errorf("failed to write %s to %s: %w", format, outputPath, err)
	}

	g.Output.Success("%s: %s", successMsg, outputPath)

	return nil
}

// generateMarkdown creates a README.md file using the template.
func (g *Generator) generateMarkdown(action *ActionYML, outputDir, actionPath string) error {
	return g.generateSimpleFormat(
		action, outputDir, actionPath,
		appconstants.OutputFormatMarkdown, appconstants.ReadmeMarkdown, "Generated README.md",
	)
}

// generateHTML creates an HTML file using the template and optional header/footer.
func (g *Generator) generateHTML(action *ActionYML, outputDir, actionPath string) error {
	templatePath, err := g.resolveTemplatePathForFormat()
	if err != nil {
		return err
	}

	opts := TemplateOptions{
		TemplatePath: templatePath,
		HeaderPath:   g.Config.Header,
		FooterPath:   g.Config.Footer,
		Format:       "html",
	}

	content, err := g.renderTemplateForAction(action, outputDir, actionPath, opts)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	// Use HTMLWriter for consistent HTML output
	writer := &HTMLWriter{
		Header: "", // Header/footer are handled by template options
		Footer: "",
	}

	// Sanitize the action name for use as a filename: a name containing "/" (valid
	// in action.yml) would otherwise resolve into a non-existent subdirectory and
	// fail os.Create with ENOENT. Fall back to a stable default for empty names.
	safeName := strings.ReplaceAll(action.Name, "/", "-")
	safeName = strings.ReplaceAll(safeName, string(filepath.Separator), "-")
	if strings.TrimSpace(safeName) == "" {
		safeName = "action"
	}
	defaultFilename := safeName + ".html"
	outputPath, err := g.resolveOutputPath(outputDir, defaultFilename)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToResolveOutputPath, err)
	}
	if err := writer.Write(content, outputPath); err != nil {
		return fmt.Errorf("failed to write HTML to %s: %w", outputPath, err)
	}

	g.Output.Success("Generated HTML: %s", outputPath)

	return nil
}

// generateJSON creates a JSON file with structured documentation data.
func (g *Generator) generateJSON(action *ActionYML, outputDir string) error {
	writer := NewJSONWriter(g.Config)

	outputPath, err := g.resolveOutputPath(outputDir, appconstants.ActionDocsJSON)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToResolveOutputPath, err)
	}
	if err := writer.Write(action, outputPath); err != nil {
		return fmt.Errorf("failed to write JSON to %s: %w", outputPath, err)
	}

	g.Output.Success("Generated JSON: %s", outputPath)

	return nil
}

// generateASCIIDoc creates an AsciiDoc file using the template.
func (g *Generator) generateASCIIDoc(action *ActionYML, outputDir, actionPath string) error {
	return g.generateSimpleFormat(
		action, outputDir, actionPath,
		"asciidoc", appconstants.ReadmeASCIIDoc, "Generated AsciiDoc",
	)
}

// processFiles processes each file and tracks results.
func (g *Generator) processFiles(paths []string, bar *progressbar.ProgressBar) ([]string, int) {
	var parseErrors []string
	successCount := 0

	for _, path := range paths {
		if err := g.GenerateFromFile(path); err != nil {
			errorMsg := fmt.Sprintf("failed to process %s: %v", path, err)
			parseErrors = append(parseErrors, errorMsg)
			if g.Config.Verbose {
				g.Output.Error("%s", errorMsg)
			}
		} else {
			successCount++
		}

		g.Progress.UpdateProgressBar(bar)
	}

	return parseErrors, successCount
}

// reportResults displays processing summary.
func (g *Generator) reportResults(successCount int, parseErrors []string) {
	if g.Config.Quiet {
		return
	}

	g.Output.Bold("\nProcessing complete: %d successful, %d failed", successCount, len(parseErrors))

	if len(parseErrors) > 0 && g.Config.Verbose {
		g.Output.Error("\nErrors encountered:")
		for _, errMsg := range parseErrors {
			g.Output.Printf("  - %s\n", errMsg)
		}
	}
}

// parseAndValidateAction parses and validates an action.yml file.
func (g *Generator) parseAndValidateAction(actionPath string) (*ActionYML, error) {
	action, err := ParseActionYML(actionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse action file %s: %w", actionPath, err)
	}

	validationResult := ValidateActionYML(action)
	if len(validationResult.MissingFields) > 0 {
		// Check for critical validation errors that cannot be fixed with defaults
		for _, field := range validationResult.MissingFields {
			// All core required fields should cause validation failure
			if field == appconstants.FieldName || field == appconstants.FieldDescription ||
				field == appconstants.FieldRuns || field == appconstants.FieldRunsUsing {
				// Required fields missing - cannot be fixed with defaults, must fail
				return nil, fmt.Errorf(
					"action file %s has invalid configuration, missing required field(s): %v",
					actionPath,
					validationResult.MissingFields,
				)
			}
		}

		if g.Config.Verbose {
			g.Output.Warning("Missing fields in %s: %v", actionPath, validationResult.MissingFields)
		}
		FillMissing(action, g.Config.Defaults)
		if g.Config.Verbose {
			g.Output.Info("Applied default values for missing fields")
		}
	}

	return action, nil
}

// determineOutputDir calculates the output directory for generated files.
func (g *Generator) determineOutputDir(actionPath string) string {
	if g.Config.OutputDir == "" || g.Config.OutputDir == "." {
		return filepath.Dir(actionPath)
	}

	return g.Config.OutputDir
}

// resolveOutputPath resolves the final output path and validates it prevents path traversal.
// Returns an error if the resolved path would escape the outputDir.
func (g *Generator) resolveOutputPath(outputDir, defaultFilename string) (string, error) {
	// Determine the filename to use
	filename := defaultFilename
	if g.Config.OutputFilename != "" {
		filename = g.Config.OutputFilename
	}

	// Reject paths containing .. components (path traversal attempt)
	if strings.Contains(filename, "..") {
		return "", fmt.Errorf(appconstants.ErrPathTraversal, filename, outputDir)
	}

	// Handle absolute paths - allow them as-is (user's explicit choice)
	if filepath.IsAbs(filename) {
		cleaned := filepath.Clean(filename)
		if cleaned != filename {
			return "", fmt.Errorf("absolute path contains extraneous components: %s", filename)
		}

		return cleaned, nil
	}

	// For relative paths, join with output directory
	finalPath := filepath.Join(outputDir, filename)

	// Validate the final path stays within outputDir
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf(appconstants.ErrInvalidOutputPath, err)
	}

	absFinalPath, err := filepath.Abs(finalPath)
	if err != nil {
		return "", fmt.Errorf(appconstants.ErrInvalidOutputPath, err)
	}

	// Check if final path is within output directory using filepath.Rel
	relPath, err := filepath.Rel(absOutputDir, absFinalPath)
	if err != nil {
		return "", fmt.Errorf(appconstants.ErrInvalidOutputPath, err)
	}

	// If relative path starts with "..", it's outside the output directory
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf(appconstants.ErrPathTraversal, filename, outputDir)
	}

	return absFinalPath, nil
}

// generateByFormat generates documentation in the specified format.
func (g *Generator) generateByFormat(action *ActionYML, outputDir, actionPath string) error {
	// Validate the theme once, up front, so an invalid --theme fails identically
	// for every output format (the JSON path does not call resolveTemplatePathForFormat).
	if err := g.validateTheme(); err != nil {
		return err
	}

	switch g.Config.OutputFormat {
	case appconstants.OutputFormatMarkdown:
		return g.generateMarkdown(action, outputDir, actionPath)
	case appconstants.OutputFormatHTML:
		return g.generateHTML(action, outputDir, actionPath)
	case appconstants.OutputFormatJSON:
		return g.generateJSON(action, outputDir)
	case appconstants.OutputFormatASCIIDoc:
		return g.generateASCIIDoc(action, outputDir, actionPath)
	default:
		return fmt.Errorf("unsupported output format: %s", g.Config.OutputFormat)
	}
}

// validateFiles processes each file for validation.
func (g *Generator) validateFiles(paths []string, bar *progressbar.ProgressBar) ([]ValidationResult, []string) {
	allResults := make([]ValidationResult, 0, len(paths))
	var parseErrors []string

	for _, path := range paths {
		if g.Config.Verbose && bar == nil {
			g.Output.Progress("Validating: %s", path)
		}

		action, err := ParseActionYML(path)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to parse %s: %v", path, err)
			parseErrors = append(parseErrors, errorMsg)

			continue
		}

		result := ValidateActionYML(action)
		result.MissingFields = append([]string{"file: " + path}, result.MissingFields...)
		allResults = append(allResults, result)

		g.Progress.UpdateProgressBar(bar)
	}

	return allResults, parseErrors
}

// reportValidationResults provides a summary of validation results.
func (g *Generator) reportValidationResults(results []ValidationResult, parseErrors []string) {
	totalFiles := len(results) + len(parseErrors)
	validFiles, totalIssues := g.countValidationStats(results)

	g.showValidationSummary(totalFiles, validFiles, totalIssues, len(results), len(parseErrors))
	g.showDetailedIssues(results, totalIssues)
	g.showParseErrors(parseErrors)
}

// countValidationStats counts valid files and total issues from results.
func (g *Generator) countValidationStats(results []ValidationResult) (validFiles, totalIssues int) {
	for _, result := range results {
		if len(result.MissingFields) == 1 { // Only contains file path
			validFiles++
		} else {
			totalIssues += len(result.MissingFields) - 1 // Subtract file path entry
		}
	}

	return validFiles, totalIssues
}

// showValidationSummary displays the summary statistics.
func (g *Generator) showValidationSummary(totalFiles, validFiles, totalIssues, resultCount, errorCount int) {
	g.Output.Bold("\nValidation Summary for %d files:", totalFiles)
	g.Output.Printf("%s", "="+strings.Repeat("=", 35)+"\n")

	g.Output.Success("Valid files: %d", validFiles)
	if resultCount-validFiles > 0 {
		g.Output.Warning("Files with issues: %d", resultCount-validFiles)
	}
	if errorCount > 0 {
		g.Output.Error("Parse errors: %d", errorCount)
	}
	if totalIssues > 0 {
		g.Output.Info("Total validation issues: %d", totalIssues)
	}
}

// showDetailedIssues displays detailed validation issues and suggestions.
func (g *Generator) showDetailedIssues(results []ValidationResult, totalIssues int) {
	if totalIssues == 0 && !g.Config.Verbose {
		return
	}

	g.Output.Bold("\nDetailed Issues & Suggestions:")
	g.Output.Printf("%s", "-"+strings.Repeat("-", 35)+"\n")

	for _, result := range results {
		if len(result.MissingFields) > 1 || len(result.Warnings) > 0 {
			g.showFileIssues(result)
		}
	}
}

// showFileIssues displays issues for a specific file.
func (g *Generator) showFileIssues(result ValidationResult) {
	filename := result.MissingFields[0][6:] // Remove "file: " prefix
	g.Output.Info("📁 File: %s", filename)

	// Show missing fields
	for _, field := range result.MissingFields[1:] {
		g.Output.Error("  ❌ Missing required field: %s", field)
	}

	// Show warnings
	for _, warning := range result.Warnings {
		g.Output.Warning("  ⚠️  Missing recommended field: %s", warning)
	}

	// Show suggestions
	if len(result.Suggestions) > 0 {
		g.Output.Info("  💡 Suggestions:")
		for _, suggestion := range result.Suggestions {
			g.Output.Printf("     • %s\n", suggestion)
		}
	}
	g.Output.Printf("\n")
}

// showParseErrors displays parse errors if any exist.
func (g *Generator) showParseErrors(parseErrors []string) {
	if len(parseErrors) == 0 {
		return
	}

	g.Output.Bold("\nParse Errors:")
	g.Output.Printf("%s", "-"+strings.Repeat("-", 15)+"\n")
	for _, errMsg := range parseErrors {
		g.Output.Error("  - %s", errMsg)
	}
}
