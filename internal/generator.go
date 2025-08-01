// Package internal contains the core generator functionality.
package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/schollz/progressbar/v3"

	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/git"
)

// Output format constants.
const (
	OutputFormatHTML     = "html"
	OutputFormatMD       = "md"
	OutputFormatJSON     = "json"
	OutputFormatASCIIDoc = "asciidoc"
)

// Generator orchestrates the documentation generation process.
type Generator struct {
	Config *AppConfig
	Output *ColoredOutput
}

// NewGenerator creates a new generator instance with the provided configuration.
func NewGenerator(config *AppConfig) *Generator {
	return &Generator{
		Config: config,
		Output: NewColoredOutput(config.Quiet),
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
	if g.Config.GitHubToken != "" {
		clientWrapper, err := NewGitHubClient(g.Config.GitHubToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub client: %w", err)
		}
		githubClient = clientWrapper.Client
	}

	// Create cache
	depCache, err := cache.NewCache(cache.DefaultConfig())
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

// parseAndValidateAction parses and validates an action.yml file.
func (g *Generator) parseAndValidateAction(actionPath string) (*ActionYML, error) {
	action, err := ParseActionYML(actionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse action file %s: %w", actionPath, err)
	}

	validationResult := ValidateActionYML(action)
	if len(validationResult.MissingFields) > 0 {
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
	if g.Config.OutputDir == "." {
		return filepath.Dir(actionPath)
	}
	return g.Config.OutputDir
}

// generateByFormat generates documentation in the specified format.
func (g *Generator) generateByFormat(action *ActionYML, outputDir, actionPath string) error {
	switch g.Config.OutputFormat {
	case "md":
		return g.generateMarkdown(action, outputDir, actionPath)
	case OutputFormatHTML:
		return g.generateHTML(action, outputDir)
	case OutputFormatJSON:
		return g.generateJSON(action, outputDir)
	case OutputFormatASCIIDoc:
		return g.generateASCIIDoc(action, outputDir)
	default:
		return fmt.Errorf("unsupported output format: %s", g.Config.OutputFormat)
	}
}

// generateMarkdown creates a README.md file using the template.
func (g *Generator) generateMarkdown(action *ActionYML, outputDir, actionPath string) error {
	// Use theme-based template if theme is specified, otherwise use explicit template path
	templatePath := g.Config.Template
	if g.Config.Theme != "" && g.Config.Theme != "default" {
		templatePath = resolveThemeTemplate(g.Config.Theme)
	}

	opts := TemplateOptions{
		TemplatePath: templatePath,
		Format:       "md",
	}

	// Find repository root for git information
	repoRoot, _ := git.FindRepositoryRoot(outputDir)

	// Build comprehensive template data
	templateData := BuildTemplateData(action, g.Config, repoRoot, actionPath)

	content, err := RenderReadme(templateData, opts)
	if err != nil {
		return fmt.Errorf("failed to render markdown template: %w", err)
	}

	outputPath := filepath.Join(outputDir, "README.md")
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write README.md to %s: %w", outputPath, err)
	}

	g.Output.Success("Generated README.md: %s", outputPath)
	return nil
}

// generateHTML creates an HTML file using the template and optional header/footer.
func (g *Generator) generateHTML(action *ActionYML, outputDir string) error {
	opts := TemplateOptions{
		TemplatePath: g.Config.Template,
		HeaderPath:   g.Config.Header,
		FooterPath:   g.Config.Footer,
		Format:       "html",
	}

	content, err := RenderReadme(action, opts)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	// Use HTMLWriter for consistent HTML output
	writer := &HTMLWriter{
		Header: "", // Header/footer are handled by template options
		Footer: "",
	}

	outputPath := filepath.Join(outputDir, action.Name+".html")
	if err := writer.Write(content, outputPath); err != nil {
		return fmt.Errorf("failed to write HTML to %s: %w", outputPath, err)
	}

	g.Output.Success("Generated HTML: %s", outputPath)
	return nil
}

// generateJSON creates a JSON file with structured documentation data.
func (g *Generator) generateJSON(action *ActionYML, outputDir string) error {
	writer := NewJSONWriter(g.Config)

	outputPath := filepath.Join(outputDir, "action-docs.json")
	if err := writer.Write(action, outputPath); err != nil {
		return fmt.Errorf("failed to write JSON to %s: %w", outputPath, err)
	}

	g.Output.Success("Generated JSON: %s", outputPath)
	return nil
}

// generateASCIIDoc creates an AsciiDoc file using the template.
func (g *Generator) generateASCIIDoc(action *ActionYML, outputDir string) error {
	// Use AsciiDoc template
	templatePath := resolveTemplatePath("templates/themes/asciidoc/readme.adoc")

	opts := TemplateOptions{
		TemplatePath: templatePath,
		Format:       "asciidoc",
	}

	content, err := RenderReadme(action, opts)
	if err != nil {
		return fmt.Errorf("failed to render AsciiDoc template: %w", err)
	}

	outputPath := filepath.Join(outputDir, "README.adoc")
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write AsciiDoc to %s: %w", outputPath, err)
	}

	g.Output.Success("Generated AsciiDoc: %s", outputPath)
	return nil
}

// DiscoverActionFiles finds action.yml and action.yaml files in the given directory
// using the centralized parser function and adds verbose logging.
func (g *Generator) DiscoverActionFiles(dir string, recursive bool) ([]string, error) {
	actionFiles, err := DiscoverActionFiles(dir, recursive)
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

// ProcessBatch processes multiple action.yml files.
func (g *Generator) ProcessBatch(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no action files to process")
	}

	bar := g.createProgressBar("Processing files", paths)
	errors, successCount := g.processFiles(paths, bar)
	g.finishProgressBar(bar)
	g.reportResults(successCount, errors)

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during batch processing", len(errors))
	}
	return nil
}

// processFiles processes each file and tracks results.
func (g *Generator) processFiles(paths []string, bar *progressbar.ProgressBar) ([]string, int) {
	var errors []string
	successCount := 0

	for _, path := range paths {
		if err := g.GenerateFromFile(path); err != nil {
			errorMsg := fmt.Sprintf("failed to process %s: %v", path, err)
			errors = append(errors, errorMsg)
			if g.Config.Verbose {
				g.Output.Error("%s", errorMsg)
			}
		} else {
			successCount++
		}

		if bar != nil {
			_ = bar.Add(1)
		}
	}
	return errors, successCount
}

// reportResults displays processing summary.
func (g *Generator) reportResults(successCount int, errors []string) {
	if g.Config.Quiet {
		return
	}

	g.Output.Bold("\nProcessing complete: %d successful, %d failed", successCount, len(errors))

	if len(errors) > 0 && g.Config.Verbose {
		g.Output.Error("\nErrors encountered:")
		for _, errMsg := range errors {
			g.Output.Printf("  - %s\n", errMsg)
		}
	}
}

// ValidateFiles validates multiple action.yml files and reports results.
func (g *Generator) ValidateFiles(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no action files to validate")
	}

	bar := g.createProgressBar("Validating files", paths)
	allResults, errors := g.validateFiles(paths, bar)
	g.finishProgressBar(bar)

	if !g.Config.Quiet {
		g.reportValidationResults(allResults, errors)
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed for %d files", len(errors))
	}
	return nil
}

// createProgressBar creates a progress bar with the specified description.
func (g *Generator) createProgressBar(description string, paths []string) *progressbar.ProgressBar {
	if len(paths) <= 1 || g.Config.Quiet {
		return nil
	}
	return progressbar.NewOptions(len(paths),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
}

// finishProgressBar completes the progress bar display.
func (g *Generator) finishProgressBar(bar *progressbar.ProgressBar) {
	if bar != nil {
		fmt.Println()
	}
}

// validateFiles processes each file for validation.
func (g *Generator) validateFiles(paths []string, bar *progressbar.ProgressBar) ([]ValidationResult, []string) {
	allResults := make([]ValidationResult, 0, len(paths))
	var errors []string

	for _, path := range paths {
		if g.Config.Verbose && bar == nil {
			g.Output.Progress("Validating: %s", path)
		}

		action, err := ParseActionYML(path)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to parse %s: %v", path, err)
			errors = append(errors, errorMsg)
			continue
		}

		result := ValidateActionYML(action)
		result.MissingFields = append([]string{fmt.Sprintf("file: %s", path)}, result.MissingFields...)
		allResults = append(allResults, result)

		if bar != nil {
			_ = bar.Add(1)
		}
	}
	return allResults, errors
}

// reportValidationResults provides a summary of validation results.
func (g *Generator) reportValidationResults(results []ValidationResult, errors []string) {
	totalFiles := len(results) + len(errors)
	validFiles, totalIssues := g.countValidationStats(results)

	g.showValidationSummary(totalFiles, validFiles, totalIssues, len(results), len(errors))
	g.showDetailedIssues(results, totalIssues)
	g.showParseErrors(errors)
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
	g.Output.Printf("=" + strings.Repeat("=", 35) + "\n")

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
	g.Output.Printf("-" + strings.Repeat("-", 35) + "\n")

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
func (g *Generator) showParseErrors(errors []string) {
	if len(errors) == 0 {
		return
	}

	g.Output.Bold("\nParse Errors:")
	g.Output.Printf("-" + strings.Repeat("-", 15) + "\n")
	for _, errMsg := range errors {
		g.Output.Error("  - %s", errMsg)
	}
}
