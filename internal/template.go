package internal

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/google/go-github/v57/github"

	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/internal/validation"
)

const (
	defaultOrgPlaceholder  = "your-org"
	defaultRepoPlaceholder = "your-repo"
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

	// Dependencies (populated by dependency analysis)
	Dependencies []dependencies.Dependency `json:"dependencies,omitempty"`
}

// templateFuncs returns a map of custom template functions.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
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

// getGitOrg returns the Git organization from template data.
func getGitOrg(data any) string {
	if td, ok := data.(*TemplateData); ok {
		if td.Git.Organization != "" {
			return td.Git.Organization
		}
		if td.Config.Organization != "" {
			return td.Config.Organization
		}
	}

	return defaultOrgPlaceholder
}

// getGitRepo returns the Git repository name from template data.
func getGitRepo(data any) string {
	if td, ok := data.(*TemplateData); ok {
		if td.Git.Repository != "" {
			return td.Git.Repository
		}
		if td.Config.Repository != "" {
			return td.Config.Repository
		}
	}

	return defaultRepoPlaceholder
}

// getGitUsesString returns a complete uses string for the action.
func getGitUsesString(data any) string {
	td, ok := data.(*TemplateData)
	if !ok {
		return "your-org/your-action@v1"
	}

	org := strings.TrimSpace(getGitOrg(data))
	repo := strings.TrimSpace(getGitRepo(data))

	if !isValidOrgRepo(org, repo) {
		return "your-org/your-action@v1"
	}

	version := formatVersion(getActionVersion(data))

	return buildUsesString(td, org, repo, version)
}

// isValidOrgRepo checks if org and repo are valid.
func isValidOrgRepo(org, repo string) bool {
	return org != "" && repo != "" && org != defaultOrgPlaceholder && repo != defaultRepoPlaceholder
}

// formatVersion ensures version has proper @ prefix.
func formatVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "@v1"
	}
	if !strings.HasPrefix(version, "@") {
		return "@" + version
	}

	return version
}

// buildUsesString constructs the uses string with optional action name.
func buildUsesString(td *TemplateData, org, repo, version string) string {
	if td.Name != "" {
		actionName := validation.SanitizeActionName(td.Name)
		if actionName != "" && actionName != repo {
			return fmt.Sprintf("%s/%s/%s%s", org, repo, actionName, version)
		}
	}

	return fmt.Sprintf("%s/%s%s", org, repo, version)
}

// getActionVersion returns the action version from template data.
func getActionVersion(data any) string {
	if td, ok := data.(*TemplateData); ok {
		if td.Config.Version != "" {
			return td.Config.Version
		}
	}

	return "v1"
}

// BuildTemplateData constructs comprehensive template data from action and configuration.
func BuildTemplateData(action *ActionYML, config *AppConfig, repoRoot, actionPath string) *TemplateData {
	data := &TemplateData{
		ActionYML: action,
		Config:    config,
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

// RenderReadme renders a README using a Go template and the parsed action.yml data.
func RenderReadme(action any, opts TemplateOptions) (string, error) {
	tmplContent, err := os.ReadFile(opts.TemplatePath)
	if err != nil {
		return "", err
	}
	var tmpl *template.Template
	if opts.Format == OutputFormatHTML {
		tmpl, err = template.New("readme").Funcs(templateFuncs()).Parse(string(tmplContent))
		if err != nil {
			return "", err
		}
		var head, foot string
		if opts.HeaderPath != "" {
			h, _ := os.ReadFile(opts.HeaderPath)
			head = string(h)
		}
		if opts.FooterPath != "" {
			f, _ := os.ReadFile(opts.FooterPath)
			foot = string(f)
		}
		// Wrap template output in header/footer
		buf := &bytes.Buffer{}
		buf.WriteString(head)
		if err := tmpl.Execute(buf, action); err != nil {
			return "", err
		}
		buf.WriteString(foot)

		return buf.String(), nil
	}

	tmpl, err = template.New("readme").Funcs(templateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, action); err != nil {
		return "", err
	}

	return buf.String(), nil
}
