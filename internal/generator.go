// Package internal contains the core generator functionality.
package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

// actionNameReplacer maps path separators and Windows-reserved filename
// characters to '-' so an action name is safe to use as a file name on any OS.
var actionNameReplacer = strings.NewReplacer(
	"/", "-",
	"\\", "-",
	":", "-",
	"*", "-",
	"?", "-",
	"\"", "-",
	"<", "-",
	">", "-",
	"|", "-",
)

// safeActionFilename builds a filesystem-safe per-action filename ("<name><ext>")
// from the action name, mapping path separators and Windows-reserved characters
// to '-' and falling back to "action" for an empty result. Used so multiple
// actions written to one output directory do not collide.
func safeActionFilename(action *ActionYML, ext string) string {
	name := actionNameReplacer.Replace(strings.TrimSpace(action.Name))
	name = strings.Trim(name, ".- ")
	if name == "" {
		name = "action"
	}

	return name + ext
}

// Generator orchestrates the documentation generation process.
// It uses focused interfaces to reduce coupling and improve testability.
type Generator struct {
	Config   *AppConfig
	Output   CompleteOutput
	Progress ProgressManager

	// usedNames tracks output filenames already emitted within a single batch so
	// per-action names that sanitize to the same string (e.g. two actions named
	// "Build", or "foo/bar" vs "foo:bar") get a "-N" suffix instead of silently
	// overwriting each other in a shared --output-dir. Non-nil only during a batch.
	usedNames map[string]int

	// depRes holds the GitHub client and dependency cache shared across a batch,
	// so cache.json is read/rewritten once per run instead of once per action
	// file. Non-nil only during a batch when dependency analysis is enabled.
	depRes *depResources
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

	// Attach the cache directly: *cache.Cache implements DependencyCache. When
	// creation failed, leave it a nil interface — the analyzer's nil-guards treat
	// that as "caching disabled" (assigning the typed-nil pointer would defeat
	// those guards, so only assign when non-nil).
	var depCacheIface dependencies.DependencyCache
	if depCache != nil {
		depCacheIface = depCache
	}

	return dependencies.NewAnalyzer(githubClient, *gitInfo, depCacheIface), nil
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

	// Reset the per-batch filename map so collisions are disambiguated across this
	// run's actions (see disambiguateName).
	g.usedNames = make(map[string]int)

	// Build the shared GitHub client + dependency cache once for the whole batch
	// (only when dependency analysis is enabled) and release them once at the end,
	// instead of reconstructing the cache/client for every action file.
	if g.Config.AnalyzeDependencies {
		g.depRes = newDepResources(g.Config)
		defer func() {
			_ = g.depRes.Close()
			g.depRes = nil
		}()
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

// disambiguateName returns name unchanged the first time it is seen within a batch
// and appends "-2", "-3", … on subsequent collisions. Deterministic for a given
// processing order. A no-op (returns name) outside a batch, when usedNames is nil.
func (g *Generator) disambiguateName(name string) string {
	if g.usedNames == nil {
		return name
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	candidate := name
	for i := 2; g.usedNames[candidate] > 0; i++ {
		candidate = fmt.Sprintf("%s-%d%s", base, i, ext)
	}
	g.usedNames[candidate]++

	return candidate
}

// generateFromFile is the implementation behind GenerateFromFile. uniqueNames is
// set by batch processing when multiple actions share one output directory, so
// fixed-name formats (JSON) get per-action filenames instead of overwriting.
func (g *Generator) generateFromFile(actionPath string, uniqueNames bool) error {
	if g.Config.Verbose {
		g.Output.Progress("Processing file: %s", actionPath)
	}

	action, err := g.parseAndValidateAction(actionPath)
	if err != nil {
		return err
	}

	outputDir := g.determineOutputDir(actionPath)

	return g.generateByFormat(action, outputDir, actionPath, uniqueNames)
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

// repoRootForAction resolves the repository root from the ACTION file's location,
// not the output directory. Deriving it from the output dir silently degrades git
// info and the monorepo uses: path whenever --output-dir points outside the repo
// (e.g. a /tmp target); the action file is always inside the repo it documents.
func (g *Generator) repoRootForAction(actionPath string) string {
	repoRoot, err := git.FindRepositoryRoot(filepath.Dir(actionPath))
	if err != nil && g.Config.Verbose {
		g.Output.Info("could not determine repository root for %s: %v", actionPath, err)
	}

	return repoRoot
}

// renderTemplateForAction builds template data and renders it using the specified options.
// It finds the repository root for git information, builds comprehensive template data,
// and renders the template. Returns the rendered content or an error.
func (g *Generator) renderTemplateForAction(
	action *ActionYML,
	actionPath string,
	opts TemplateOptions,
) (string, error) {
	repoRoot := g.repoRootForAction(actionPath)

	// Build comprehensive template data, reusing the batch-shared cache+client
	// (g.depRes is nil outside a batch or when dependency analysis is disabled,
	// in which case buildTemplateData falls back to per-call resources).
	templateData := buildTemplateData(action, g.Config, repoRoot, actionPath, g.depRes)
	templateData.OutputDir = opts.OutputDir

	// Render template with data
	content, err := RenderReadme(templateData, opts)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return content, nil
}

// simpleFormatSpec describes the format-specific half of a generateSimpleFormat
// render: what is written, under which name, and what to report on success.
type simpleFormatSpec struct {
	Format          string
	DefaultFilename string
	SuccessMsg      string
	// TemplateOverride forces a specific template (used by AsciiDoc, whose template
	// is selected by output format, not by --theme). Empty falls back to the
	// theme-resolved template.
	TemplateOverride string
}

// generateSimpleFormat is a helper for generating simple text-based formats (Markdown, AsciiDoc).
// It consolidates the common pattern of template rendering, file writing, and success messaging.
func (g *Generator) generateSimpleFormat(
	action *ActionYML,
	outputDir, actionPath string,
	spec simpleFormatSpec,
	uniqueNames bool,
) error {
	templatePath := spec.TemplateOverride
	if templatePath == "" {
		var err error
		templatePath, err = g.resolveTemplatePathForFormat()
		if err != nil {
			return err
		}
	}

	// When multiple actions share one output directory, derive a per-action
	// filename (keeping the default's extension) so the fixed README name does not
	// overwrite earlier outputs.
	filename := spec.DefaultFilename
	if uniqueNames {
		filename = g.disambiguateName(safeActionFilename(action, filepath.Ext(spec.DefaultFilename)))
	}

	// Resolve the destination BEFORE rendering: links to repository files are
	// relative to the document, and --output may add directory depth or be absolute,
	// in which case outputDir is not where the file lands.
	outputPath, err := g.resolveOutputPath(outputDir, filename)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToResolveOutputPath, err)
	}

	header, footer := g.resolveHeaderFooter(spec.Format)
	opts := TemplateOptions{
		TemplatePath: templatePath,
		HeaderPath:   header,
		FooterPath:   footer,
		Format:       spec.Format,
		OutputDir:    filepath.Dir(outputPath),
	}

	content, err := g.renderTemplateForAction(action, actionPath, opts)
	if err != nil {
		return fmt.Errorf("failed to render %s template: %w", spec.Format, err)
	}

	if err := ensureParentDir(outputPath); err != nil {
		return err
	}
	if err := writeFileTightMode(outputPath, []byte(content), appconstants.FilePermDefault); err != nil {
		return fmt.Errorf("failed to write %s to %s: %w", spec.Format, outputPath, err)
	}

	g.Output.Success("%s: %s", spec.SuccessMsg, outputPath)

	return nil
}

// writeFileTightMode writes data to path and enforces perm even when path already
// exists. os.WriteFile (and OpenFile's O_CREATE mode) only apply perm when the file
// is created, so rewriting an existing 0644 output would otherwise stay
// world-readable and defeat the 0600 guarantee the markdown/HTML/JSON writers share.
func writeFileTightMode(path string, data []byte, perm os.FileMode) error {
	if err := os.WriteFile(path, data, perm); err != nil { // #nosec G306 -- output file permissions
		return err
	}

	return os.Chmod(path, perm) // #nosec G302 -- output file permissions
}

// resolveHeaderFooter picks the header and footer partials for a render, resolving
// each part independently so overriding one keeps the other. Priority per part:
//
//  1. an explicitly configured Config.Header / Config.Footer;
//  2. the selected theme's partials/{header,footer}.tmpl, when it ships one;
//  3. for HTML only, the built-in HTML partials — they are an HTML document's
//     <head>/<body> scaffolding and would be nonsense injected into Markdown.
//
// "Explicitly configured" means different from the built-in default: DefaultAppConfig
// seeds Header/Footer with the HTML partial paths, and WriteDefaultConfig has written
// those into user config files, so a bare equality check against the default is what
// separates a real user choice from the inherited default. This mirrors the same
// treatment in wizard/exporter.go.
func (g *Generator) resolveHeaderFooter(format string) (header, footer string) {
	defaults := DefaultAppConfig()

	header = g.resolvePartial(
		g.Config.Header, defaults.Header, appconstants.ThemePartialHeader, format,
	)
	footer = g.resolvePartial(
		g.Config.Footer, defaults.Footer, appconstants.ThemePartialFooter, format,
	)

	return header, footer
}

// resolvePartial implements the per-part precedence described on resolveHeaderFooter.
func (g *Generator) resolvePartial(configured, builtinDefault, partialName, format string) string {
	if configured != "" && configured != builtinDefault {
		return configured
	}

	if themePartial := resolveThemePartial(g.Config.Theme, partialName); themePartial != "" {
		return themePartial
	}

	if format == appconstants.OutputFormatHTML {
		return builtinDefault
	}

	return ""
}

// generateMarkdown creates a README.md file using the template.
func (g *Generator) generateMarkdown(action *ActionYML, outputDir, actionPath string, uniqueNames bool) error {
	return g.generateSimpleFormat(
		action, outputDir, actionPath,
		simpleFormatSpec{
			Format:          appconstants.OutputFormatMarkdown,
			DefaultFilename: appconstants.ReadmeMarkdown,
			SuccessMsg:      "Generated README.md",
		},
		uniqueNames,
	)
}

// generateHTML creates an HTML file using the template and optional header/footer.
func (g *Generator) generateHTML(action *ActionYML, outputDir, actionPath string, uniqueNames bool) error {
	templatePath, err := g.resolveTemplatePathForFormat()
	if err != nil {
		return err
	}

	// Per-action filename so multiple actions written to one output directory do
	// not collide (and to avoid path separators / Windows-reserved characters in
	// the action name resolving into a bad path). In a shared-dir batch, also
	// disambiguate names that sanitize to the same string.
	defaultFilename := safeActionFilename(action, ".html")
	if uniqueNames {
		defaultFilename = g.disambiguateName(defaultFilename)
	}

	// Resolve the destination BEFORE rendering — see generateSimpleFormat.
	outputPath, err := g.resolveOutputPath(outputDir, defaultFilename)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToResolveOutputPath, err)
	}

	header, footer := g.resolveHeaderFooter(appconstants.OutputFormatHTML)
	opts := TemplateOptions{
		TemplatePath: templatePath,
		HeaderPath:   header,
		FooterPath:   footer,
		Format:       appconstants.OutputFormatHTML,
		OutputDir:    filepath.Dir(outputPath),
	}

	content, err := g.renderTemplateForAction(action, actionPath, opts)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	// Use HTMLWriter for consistent HTML output
	writer := &HTMLWriter{
		Header: "", // Header/footer are handled by template options
		Footer: "",
	}

	if err := ensureParentDir(outputPath); err != nil {
		return err
	}
	if err := writer.Write(content, outputPath); err != nil {
		return fmt.Errorf("failed to write HTML to %s: %w", outputPath, err)
	}

	g.Output.Success("Generated HTML: %s", outputPath)

	return nil
}

// generateJSON creates a JSON file with structured documentation data. When
// uniqueNames is set (multiple actions sharing one output directory), the file is
// named per action instead of the fixed action-docs.json so the outputs do not
// overwrite each other.
func (g *Generator) generateJSON(action *ActionYML, outputDir, actionPath string, uniqueNames bool) error {
	writer := NewJSONWriter(g.Config)

	// Resolve the real uses: reference and repository URL from git so the JSON
	// output matches the markdown output instead of emitting your-org/@v1 stubs.
	td := buildTemplateData(action, g.Config, g.repoRootForAction(actionPath), actionPath, g.depRes)
	writer.usesStatement = td.UsesStatement
	writer.license = td.License
	writer.dependencies = td.Dependencies
	if td.Git.Organization != "" && td.Git.Repository != "" {
		writer.repoURL = fmt.Sprintf("%s/%s/%s", appconstants.GitHubBaseURL, td.Git.Organization, td.Git.Repository)
	}

	jsonFilename := appconstants.ActionDocsJSON
	if uniqueNames {
		jsonFilename = g.disambiguateName(safeActionFilename(action, ".json"))
	}

	outputPath, err := g.resolveOutputPath(outputDir, jsonFilename)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToResolveOutputPath, err)
	}
	if err := ensureParentDir(outputPath); err != nil {
		return err
	}
	if err := writer.Write(action, outputPath); err != nil {
		return fmt.Errorf("failed to write JSON to %s: %w", outputPath, err)
	}

	g.Output.Success("Generated JSON: %s", outputPath)

	return nil
}

// generateASCIIDoc creates an AsciiDoc file using the template.
func (g *Generator) generateASCIIDoc(action *ActionYML, outputDir, actionPath string, uniqueNames bool) error {
	// Honor an explicit --template (a custom .adoc); fall back to the bundled
	// AsciiDoc template. Without this the always-non-empty override below wins and
	// `-f asciidoc --template custom.adoc` silently renders the bundled template.
	override := g.Config.Template
	if override == "" {
		override = resolveTemplatePath(appconstants.TemplatePathASCIIDoc)
	}

	return g.generateSimpleFormat(
		action, outputDir, actionPath,
		simpleFormatSpec{
			Format:           "asciidoc",
			DefaultFilename:  appconstants.ReadmeASCIIDoc,
			SuccessMsg:       "Generated AsciiDoc",
			TemplateOverride: override,
		},
		uniqueNames,
	)
}

// processFiles processes each file and tracks results.
func (g *Generator) processFiles(paths []string, bar *progressbar.ProgressBar) ([]string, int) {
	var parseErrors []string
	successCount := 0

	// When several actions are written into one shared --output-dir, fixed-name
	// formats (JSON) would overwrite each other; request per-action filenames.
	// With a per-action output dir (OutputDir unset/"."), each file lands in its
	// own directory, so the stable default names are kept.
	sharedDir := g.Config.OutputDir != "" && g.Config.OutputDir != "."
	uniqueNames := len(paths) > 1 && sharedDir

	for _, path := range paths {
		if err := g.generateFromFile(path, uniqueNames); err != nil {
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

// coreRequiredFields are the action.yml fields that FillMissing cannot synthesize a
// usable value for; a file missing any of them is rejected rather than defaulted.
var coreRequiredFields = []string{
	appconstants.FieldName,
	appconstants.FieldDescription,
	appconstants.FieldRuns,
	appconstants.FieldRunsUsing,
}

// hasMissingCoreField reports whether any missing field is a core required field.
func hasMissingCoreField(missing []string) bool {
	for _, field := range missing {
		if slices.Contains(coreRequiredFields, field) {
			return true
		}
	}

	return false
}

// parseAndValidateAction parses and validates an action.yml file.
func (g *Generator) parseAndValidateAction(actionPath string) (*ActionYML, error) {
	action, warnings, err := ParseActionYMLWithWarnings(actionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse action file %s: %w", actionPath, err)
	}

	// Surface non-fatal parse warnings (e.g. a header-comment permission block that
	// could not be scanned) instead of silently dropping the affected section.
	for _, warning := range warnings {
		g.Output.Warning("%s", warning)
	}

	validationResult := ValidateActionYML(action)
	if len(validationResult.MissingFields) > 0 {
		// Core required fields cannot be fixed with defaults, so they must fail.
		if hasMissingCoreField(validationResult.MissingFields) {
			return nil, fmt.Errorf(
				"action file %s has invalid configuration, missing required field(s): %v",
				actionPath,
				validationResult.MissingFields,
			)
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

	// Reject paths containing a ".." path component (traversal attempt). Check
	// path components, not a raw substring, so a legitimate filename that merely
	// contains ".." (e.g. "report..final.md" or "v1..2.md") is not rejected.
	if slices.Contains(strings.Split(filepath.ToSlash(filename), "/"), "..") {
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

	// Outside the output directory iff the first path component is "..". Match the
	// whole "..") component, not a ".." prefix, so a legitimate directory whose name
	// merely starts with ".." (e.g. "..reports/readme.md") is not falsely rejected.
	if relPath == appconstants.PathParent ||
		strings.HasPrefix(filepath.ToSlash(relPath), appconstants.PathParent+"/") {
		return "", fmt.Errorf(appconstants.ErrPathTraversal, filename, outputDir)
	}

	return absFinalPath, nil
}

// ensureParentDir creates the parent directory of path so a subsequent write
// succeeds even when --output-dir (or a nested --output filename) points at a
// directory that does not exist yet. Called only after path-traversal validation.
func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	// #nosec G301 -- generated documentation directory permissions
	if err := os.MkdirAll(dir, appconstants.FilePermDir); err != nil {
		return fmt.Errorf("failed to create output directory %q: %w", dir, err)
	}

	return nil
}

// generateByFormat generates documentation in the specified format. uniqueNames
// requests per-action output filenames (used when multiple actions share one
// output directory) for formats that otherwise use a fixed default name.
func (g *Generator) generateByFormat(action *ActionYML, outputDir, actionPath string, uniqueNames bool) error {
	// Validate the theme once, up front, so an invalid --theme fails identically
	// for every output format (the JSON path does not call resolveTemplatePathForFormat).
	if err := g.validateTheme(); err != nil {
		return err
	}

	switch g.Config.OutputFormat {
	case appconstants.OutputFormatMarkdown:
		return g.generateMarkdown(action, outputDir, actionPath, uniqueNames)
	case appconstants.OutputFormatHTML:
		return g.generateHTML(action, outputDir, actionPath, uniqueNames)
	case appconstants.OutputFormatJSON:
		return g.generateJSON(action, outputDir, actionPath, uniqueNames)
	case appconstants.OutputFormatASCIIDoc:
		return g.generateASCIIDoc(action, outputDir, actionPath, uniqueNames)
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
