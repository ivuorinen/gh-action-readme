package internal

import (
	"bytes"
	htmltemplate "html/template"
	"io"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	"github.com/google/go-github/v74/github"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/internal/validation"
	templatesembed "github.com/ivuorinen/gh-action-readme/templates_embed"
)

// TemplateOptions defines options for rendering templates.
type TemplateOptions struct {
	TemplatePath string
	HeaderPath   string
	FooterPath   string
	Format       string // md or html
}

// TemplateData represents all data available to templates.
type TemplateData struct {
	// Action Data
	*ActionYML

	// Git Repository Information
	Git git.RepoInfo `json:"git"`

	// Configuration
	Config *AppConfig `json:"config"`

	// Computed Values
	UsesStatement string `json:"uses_statement"`

	// Path information for subdirectory extraction
	ActionPath string `json:"action_path,omitempty"`
	RepoRoot   string `json:"repo_root,omitempty"`

	// Dependencies (populated by dependency analysis)
	Dependencies []dependencies.Dependency `json:"dependencies,omitempty"`
}

// templateFuncs returns a map of custom template functions.
func templateFuncs() texttemplate.FuncMap {
	return texttemplate.FuncMap{
		"lower":         strings.ToLower,
		"upper":         strings.ToUpper,
		"replace":       strings.ReplaceAll,
		"join":          strings.Join,
		"gitOrg":        getGitOrg,
		"gitRepo":       getGitRepo,
		"gitUsesString": getGitUsesString,
		"actionVersion": getActionVersion,
	}
}

// getFieldWithFallback extracts a field from TemplateData with Git-then-Config fallback logic.
func getFieldWithFallback(data any, gitGetter, configGetter func(*TemplateData) string, defaultValue string) string {
	if td, ok := data.(*TemplateData); ok {
		if gitValue := gitGetter(td); gitValue != "" {
			return gitValue
		}
		// configGetter dereferences td.Config; guard against a manually
		// constructed TemplateData with a nil Config (consistent with getActionVersion).
		if td.Config != nil {
			if configValue := configGetter(td); configValue != "" {
				return configValue
			}
		}
	}

	return defaultValue
}

// getGitOrg returns the Git organization from template data.
func getGitOrg(data any) string {
	return getFieldWithFallback(data,
		func(td *TemplateData) string { return td.Git.Organization },
		func(td *TemplateData) string { return td.Config.Organization },
		appconstants.DefaultOrgPlaceholder)
}

// getGitRepo returns the Git repository name from template data.
func getGitRepo(data any) string {
	return getFieldWithFallback(data,
		func(td *TemplateData) string { return td.Git.Repository },
		func(td *TemplateData) string { return td.Config.Repository },
		appconstants.DefaultRepoPlaceholder)
}

// getGitUsesString returns a complete uses string for the action.
func getGitUsesString(data any) string {
	td, ok := data.(*TemplateData)
	if !ok {
		return appconstants.DefaultUsesPlaceholder
	}

	org := strings.TrimSpace(getGitOrg(data))
	repo := strings.TrimSpace(getGitRepo(data))

	if !isValidOrgRepo(org, repo) {
		return appconstants.DefaultUsesPlaceholder
	}

	version := formatVersion(getActionVersion(data))

	return buildUsesString(td, org, repo, version)
}

// isValidOrgRepo checks if org and repo are valid.
func isValidOrgRepo(org, repo string) bool {
	return org != "" && repo != "" &&
		org != appconstants.DefaultOrgPlaceholder &&
		repo != appconstants.DefaultRepoPlaceholder
}

// formatVersion ensures version has proper @ prefix.
func formatVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return appconstants.VersionRefV1
	}
	if !strings.HasPrefix(version, "@") {
		return "@" + version
	}

	return version
}

// buildUsesString constructs the uses string with optional subdirectory path.
func buildUsesString(td *TemplateData, org, repo, version string) string {
	// Use the validation package's FormatUsesStatement for consistency
	if org == "" || repo == "" {
		return appconstants.DefaultUsesPlaceholder
	}

	// For monorepo actions in subdirectories, extract the actual directory path
	subdir := extractActionSubdirectory(td.ActionPath, td.RepoRoot)

	if subdir != "" {
		// Action is in a subdirectory: org/repo/subdir@version
		return validation.FormatUsesStatement(org, repo+"/"+subdir, version)
	}

	// Action is at repo root: org/repo@version
	return validation.FormatUsesStatement(org, repo, version)
}

// resolveSymlinkedAbs returns the absolute, symlink-resolved form of path. It
// falls back to the plain absolute path when symlinks cannot be evaluated (e.g.
// the path does not exist), and returns "" only when even filepath.Abs fails.
func resolveSymlinkedAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	if resolved, rerr := filepath.EvalSymlinks(abs); rerr == nil {
		return resolved
	}

	return abs
}

// extractActionSubdirectory extracts the subdirectory path for an action relative to repo root.
// For monorepo actions (e.g., org/repo/subdir/action.yml), returns "subdir".
// For repo-root actions (e.g., org/repo/action.yml), returns empty string.
// Returns empty string if paths cannot be determined.
func extractActionSubdirectory(actionPath, repoRoot string) string {
	// Validate inputs
	if actionPath == "" || repoRoot == "" {
		return ""
	}

	// Get absolute, symlink-resolved paths for reliable comparison. Resolving
	// symlinks on both sides keeps filepath.Rel from producing a spurious ".."
	// (and silently dropping the subdir) when the action path is reached through
	// a symlink that diverges from the real repo-root tree.
	absActionPath := resolveSymlinkedAbs(actionPath)
	absRepoRoot := resolveSymlinkedAbs(repoRoot)
	if absActionPath == "" || absRepoRoot == "" {
		return ""
	}

	// Get the directory containing action.yml
	actionDir := filepath.Dir(absActionPath)

	// Calculate relative path from repo root to action directory
	relPath, err := filepath.Rel(absRepoRoot, actionDir)
	if err != nil {
		return ""
	}

	// If relative path is "." or empty, action is at repo root
	if relPath == "." || relPath == "" {
		return ""
	}

	// If relative path starts with "..", action is outside repo (shouldn't happen)
	if strings.HasPrefix(relPath, "..") {
		return ""
	}

	// Return the subdirectory path (e.g., "actions/csharp-build")
	return relPath
}

// getActionVersion returns the action version from template data.
// Priority: 1) Config.Version (explicit override), 2) Default branch (if enabled), 3) "v1" (fallback).
func getActionVersion(data any) string {
	td, ok := data.(*TemplateData)
	if !ok || td.Config == nil {
		return appconstants.VersionTagV1
	}

	// Priority 1: Explicit version override
	if td.Config.Version != "" {
		return td.Config.Version
	}

	// Priority 2: Use default branch if enabled and available
	if td.Config.UseDefaultBranch && td.Git.DefaultBranch != "" {
		return td.Git.DefaultBranch
	}

	// Priority 3: Fallback
	return appconstants.VersionTagV1
}

// BuildTemplateData constructs comprehensive template data from action and configuration.
func BuildTemplateData(action *ActionYML, config *AppConfig, repoRoot, actionPath string) *TemplateData {
	// Guard against a nil config: this is an exported entry point and the
	// template funcs dereference Config unconditionally.
	if config == nil {
		config = DefaultAppConfig()
	}

	data := &TemplateData{
		ActionYML:  action,
		Config:     config,
		ActionPath: actionPath,
		RepoRoot:   repoRoot,
	}

	// Populate Git information
	if repoRoot != "" {
		if info, err := git.DetectRepository(repoRoot); err == nil {
			data.Git = *info
		}
	}

	// Override with configuration values if available
	if config.Organization != "" {
		data.Git.Organization = config.Organization
	}
	if config.Repository != "" {
		data.Git.Repository = config.Repository
	}

	// Build uses statement
	data.UsesStatement = getGitUsesString(data)

	// Add dependency analysis if enabled
	if config.AnalyzeDependencies && actionPath != "" {
		data.Dependencies = analyzeDependencies(actionPath, config, data.Git)
	}

	return data
}

// analyzeDependencies performs dependency analysis on the action file.
func analyzeDependencies(actionPath string, config *AppConfig, gitInfo git.RepoInfo) []dependencies.Dependency {
	// Create GitHub client if we have a token
	var client *GitHubClient
	if token := GetGitHubToken(config); token != "" {
		var err error
		client, err = NewGitHubClient(token)
		if err != nil {
			// Log error but continue with no client (graceful degradation)
			client = nil
		}
	}

	// Create high-performance cache
	var depCache dependencies.DependencyCache
	if cacheInstance, err := cache.NewCache(cache.DefaultConfig()); err == nil {
		depCache = dependencies.NewCacheAdapter(cacheInstance)
	} else {
		// Fallback to no-op cache if cache creation fails
		depCache = dependencies.NewNoOpCache()
	}

	// Create dependency analyzer
	var githubClient *github.Client
	if client != nil {
		githubClient = client.Client
	}

	analyzer := dependencies.NewAnalyzer(githubClient, gitInfo, depCache)

	// Analyze dependencies
	deps, err := analyzer.AnalyzeActionFile(actionPath)
	if err != nil {
		// Log error but don't fail - return empty dependencies
		return []dependencies.Dependency{}
	}

	return deps
}

// executableTemplate is a common interface for html/template and text/template.
type executableTemplate interface {
	Execute(io.Writer, any) error
}

// parseReadmeTemplate parses raw template content into an executable template.
// HTML format uses html/template for automatic XSS-safe escaping.
func parseReadmeTemplate(content []byte, format string) (executableTemplate, error) {
	if format == appconstants.OutputFormatHTML {
		funcs := htmltemplate.FuncMap(templateFuncs())

		return htmltemplate.New(appconstants.TemplateNameReadme).Funcs(funcs).Parse(string(content))
	}

	return texttemplate.New(appconstants.TemplateNameReadme).Funcs(templateFuncs()).Parse(string(content))
}

// RenderReadme renders a README using a Go template and the parsed action.yml data.
func RenderReadme(action any, opts TemplateOptions) (string, error) {
	tmplContent, err := templatesembed.ReadTemplate(opts.TemplatePath)
	if err != nil {
		return "", err
	}

	tmpl, err := parseReadmeTemplate(tmplContent, opts.Format)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	if opts.Format == appconstants.OutputFormatHTML && opts.HeaderPath != "" {
		if h, e := templatesembed.ReadTemplate(opts.HeaderPath); e == nil {
			buf.Write(h)
		}
	}

	if err := tmpl.Execute(&buf, action); err != nil {
		return "", err
	}

	if opts.Format == appconstants.OutputFormatHTML && opts.FooterPath != "" {
		if f, e := templatesembed.ReadTemplate(opts.FooterPath); e == nil {
			buf.Write(f)
		}
	}

	return buf.String(), nil
}
