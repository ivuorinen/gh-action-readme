// Package internal contains the internal implementation of gh-action-readme.
package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v74/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/internal/validation"
	templatesembed "github.com/ivuorinen/gh-action-readme/templates_embed"
)

// AppConfig represents the application configuration that can be used at multiple levels.
type AppConfig struct {
	// GitHub API (Global Only - Security)
	GitHubToken string `mapstructure:"github_token" yaml:"github_token,omitempty"` // Only in global config

	// Repository Information (auto-detected, overridable)
	Organization     string `mapstructure:"organization"       yaml:"organization,omitempty"`
	Repository       string `mapstructure:"repository"         yaml:"repository,omitempty"`
	Version          string `mapstructure:"version"            yaml:"version,omitempty"`
	UseDefaultBranch bool   `mapstructure:"use_default_branch" yaml:"use_default_branch"`

	// Template Settings
	Theme          string `mapstructure:"theme"           yaml:"theme"`
	OutputFormat   string `mapstructure:"output_format"   yaml:"output_format"`
	OutputDir      string `mapstructure:"output_dir"      yaml:"output_dir"`
	OutputFilename string `mapstructure:"output_filename" yaml:"output_filename,omitempty"`

	// Legacy template fields (backward compatibility)
	Template string `mapstructure:"template" yaml:"template,omitempty"`
	Header   string `mapstructure:"header"   yaml:"header,omitempty"`
	Footer   string `mapstructure:"footer"   yaml:"footer,omitempty"`
	Schema   string `mapstructure:"schema"   yaml:"schema,omitempty"`

	// Workflow Requirements
	Permissions map[string]string `mapstructure:"permissions" yaml:"permissions,omitempty"`
	RunsOn      []string          `mapstructure:"runs_on"     yaml:"runs_on,omitempty"`

	// Features
	AnalyzeDependencies bool `mapstructure:"analyze_dependencies" yaml:"analyze_dependencies"`
	ShowSecurityInfo    bool `mapstructure:"show_security_info"   yaml:"show_security_info"`

	// Custom Template Variables
	Variables map[string]string `mapstructure:"variables" yaml:"variables,omitempty"`

	// Repository-specific overrides (Global config only)
	RepoOverrides map[string]AppConfig `mapstructure:"repo_overrides" yaml:"repo_overrides,omitempty"`

	// Behavior
	Verbose            bool     `mapstructure:"verbose"             yaml:"verbose"`
	Quiet              bool     `mapstructure:"quiet"               yaml:"quiet"`
	IgnoredDirectories []string `mapstructure:"ignored_directories" yaml:"ignored_directories,omitempty"`

	// Default values for action.yml files (legacy)
	Defaults DefaultValues `mapstructure:"defaults" yaml:"defaults,omitempty"`
}

// DefaultValues stores configurable default values for all fields (legacy support).
type DefaultValues struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Runs        map[string]any `yaml:"runs"`
	Branding    Branding       `yaml:"branding"`
}

// GitHubClient wraps the GitHub API client with rate limiting.
type GitHubClient struct {
	Client *github.Client
	Token  string
}

// GetGitHubToken returns the GitHub token from environment variables or config.
func GetGitHubToken(config *AppConfig) string {
	// Priority 1 & 2: Environment variables
	if token := loadGitHubTokenFromEnv(); token != "" {
		return token
	}

	// Priority 3: Global config only (never repo/action configs)
	if config.GitHubToken != "" {
		return config.GitHubToken
	}

	return "" // Graceful degradation
}

// NewGitHubClient creates a new GitHub API client with rate limiting.
func NewGitHubClient(token string) (*GitHubClient, error) {
	var client *github.Client

	if token != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)

		// Add rate limiting with proper error handling
		rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)
		if err != nil {
			return nil, fmt.Errorf(appconstants.ErrFailedToCreateRateLimiter, err)
		}

		client = github.NewClient(rateLimiter)
	} else {
		// For no token, use basic rate limiter
		rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(nil)
		if err != nil {
			return nil, fmt.Errorf(appconstants.ErrFailedToCreateRateLimiter, err)
		}
		client = github.NewClient(rateLimiter)
	}

	return &GitHubClient{
		Client: client,
		Token:  token,
	}, nil
}

// FillMissing applies defaults for missing fields in ActionYML (legacy support).
func FillMissing(action *ActionYML, defs DefaultValues) {
	if action.Name == "" {
		action.Name = defs.Name
	}
	if action.Description == "" {
		action.Description = defs.Description
	}
	if len(action.Runs) == 0 && len(defs.Runs) > 0 {
		action.Runs = defs.Runs
	}
	if action.Branding == nil && defs.Branding.Icon != "" {
		action.Branding = &defs.Branding
	}
}

// resolveTemplatePath resolves a template path, preferring embedded templates.
// For custom/absolute paths, falls back to filesystem.
func resolveTemplatePath(templatePath string) string {
	if filepath.IsAbs(templatePath) {
		return templatePath
	}

	// Check if template is available in embedded filesystem first
	if templatesembed.IsEmbeddedTemplateAvailable(templatePath) {
		// Return a special marker to indicate this should use embedded templates
		// The actual template loading will handle embedded vs filesystem
		return templatePath
	}

	// Fallback to filesystem resolution for custom templates
	// Check if template exists in current directory
	if _, err := os.Stat(templatePath); err == nil {
		return templatePath
	}

	binaryDir, err := validation.GetBinaryDir()
	if err != nil {
		// Fallback to current working directory if we can't determine binary location
		return templatePath
	}

	resolvedPath := filepath.Join(binaryDir, templatePath)

	// Check if the resolved path exists, if not, try relative to current directory as fallback
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return templatePath
	}

	return resolvedPath
}

// resolveAllTemplatePaths resolves all template-related paths in the config.
func resolveAllTemplatePaths(config *AppConfig) {
	config.Template = resolveTemplatePath(config.Template)
	config.Header = resolveTemplatePath(config.Header)
	config.Footer = resolveTemplatePath(config.Footer)
	config.Schema = resolveTemplatePath(config.Schema)
}

// themeTemplates maps theme names to their template paths.
var themeTemplates = map[string]string{
	appconstants.ThemeDefault:      appconstants.TemplatePathDefault,
	appconstants.ThemeGitHub:       appconstants.TemplatePathGitHub,
	appconstants.ThemeGitLab:       appconstants.TemplatePathGitLab,
	appconstants.ThemeMinimal:      appconstants.TemplatePathMinimal,
	appconstants.ThemeProfessional: appconstants.TemplatePathProfessional,
}

// resolveThemeTemplate resolves the template path based on the selected theme.
func resolveThemeTemplate(theme string) string {
	if path, ok := themeTemplates[theme]; ok {
		return resolveTemplatePath(path)
	}

	return ""
}

// DefaultAppConfig returns the default application configuration.
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		// Repository Information (will be auto-detected)
		Organization:     "",
		Repository:       "",
		Version:          "",
		UseDefaultBranch: true, // Use detected default branch (main/master) in usage examples

		// Template Settings
		Theme:        appconstants.ThemeDefault, // default, github, gitlab, minimal, professional
		OutputFormat: appconstants.OutputFormatMarkdown,
		OutputDir:    ".",

		// Legacy template fields (backward compatibility)
		Template: resolveTemplatePath("templates/readme.tmpl"),
		Header:   resolveTemplatePath("templates/header.tmpl"),
		Footer:   resolveTemplatePath("templates/footer.tmpl"),
		Schema:   resolveTemplatePath("schemas/schema.json"),

		// Workflow Requirements
		Permissions: map[string]string{},
		RunsOn:      []string{appconstants.RunnerUbuntuLatest},

		// Features
		AnalyzeDependencies: false,
		ShowSecurityInfo:    false,

		// Custom Template Variables
		Variables: map[string]string{},

		// Repository-specific overrides (empty by default)
		RepoOverrides: map[string]AppConfig{},

		// Behavior
		Verbose:            false,
		Quiet:              false,
		IgnoredDirectories: appconstants.GetDefaultIgnoredDirectories(),

		// Default values for action.yml files (legacy)
		Defaults: DefaultValues{
			Name:        appconstants.LabelGitHubAction,
			Description: "A reusable GitHub Action.",
			Runs:        map[string]any{},
			Branding: Branding{
				Icon:  appconstants.ActivityWorkflowType,
				Color: "blue",
			},
		},
	}
}

// MergeConfigs merges a source config into a destination config, excluding security-sensitive fields.
func MergeConfigs(dst *AppConfig, src *AppConfig, allowTokens bool) {
	mergeStringFields(dst, src)
	mergeMapFields(dst, src)
	mergeSliceFields(dst, src)
	mergeBooleanFields(dst, src)
	mergeSecurityFields(dst, src, allowTokens)
}

// mergeStringFields merges simple string fields from src to dst if non-empty.
func mergeStringFields(dst *AppConfig, src *AppConfig) {
	stringFields := []struct {
		dst *string
		src string
	}{
		{&dst.Organization, src.Organization},
		{&dst.Repository, src.Repository},
		{&dst.Version, src.Version},
		{&dst.Theme, src.Theme},
		{&dst.OutputFormat, src.OutputFormat},
		{&dst.OutputDir, src.OutputDir},
		{&dst.Template, src.Template},
		{&dst.Header, src.Header},
		{&dst.Footer, src.Footer},
		{&dst.Schema, src.Schema},
	}

	for _, field := range stringFields {
		if field.src != "" {
			*field.dst = field.src
		}
	}
}

// mergeStringMap is a generic helper that merges a source map into a destination map.
func mergeStringMap(src map[string]string, dst *map[string]string) {
	if len(src) == 0 {
		return
	}
	if *dst == nil {
		*dst = make(map[string]string)
	}
	for k, v := range src {
		(*dst)[k] = v
	}
}

// mergeMapFields merges map fields from src to dst if non-empty.
func mergeMapFields(dst *AppConfig, src *AppConfig) {
	mergeStringMap(src.Permissions, &dst.Permissions)
	mergeStringMap(src.Variables, &dst.Variables)
}

// mergeSliceFields merges slice fields from src to dst if non-empty.
// copySliceIfNotEmpty copies src slice to dst if src is not empty.
func copySliceIfNotEmpty(dst *[]string, src []string) {
	if len(src) > 0 {
		*dst = make([]string, len(src))
		copy(*dst, src)
	}
}

func mergeSliceFields(dst *AppConfig, src *AppConfig) {
	copySliceIfNotEmpty(&dst.RunsOn, src.RunsOn)
	copySliceIfNotEmpty(&dst.IgnoredDirectories, src.IgnoredDirectories)
}

// mergeBooleanFields merges boolean fields from src to dst if true.
func mergeBooleanFields(dst *AppConfig, src *AppConfig) {
	if src.AnalyzeDependencies {
		dst.AnalyzeDependencies = src.AnalyzeDependencies
	}
	if src.ShowSecurityInfo {
		dst.ShowSecurityInfo = src.ShowSecurityInfo
	}
	if src.Verbose {
		dst.Verbose = src.Verbose
	}
	if src.Quiet {
		dst.Quiet = src.Quiet
	}
	if src.UseDefaultBranch {
		dst.UseDefaultBranch = src.UseDefaultBranch
	}
}

// mergeSecurityFields merges security-sensitive fields if allowed.
func mergeSecurityFields(dst *AppConfig, src *AppConfig, allowTokens bool) {
	if allowTokens && src.GitHubToken != "" {
		dst.GitHubToken = src.GitHubToken
	}

	if allowTokens && len(src.RepoOverrides) > 0 {
		if dst.RepoOverrides == nil {
			dst.RepoOverrides = make(map[string]AppConfig)
		}
		for k, v := range src.RepoOverrides {
			dst.RepoOverrides[k] = v
		}
	}
}

// LoadRepoConfig loads repository-level configuration from hidden config files.
func LoadRepoConfig(repoRoot string) (*AppConfig, error) {
	return loadRepoConfigInternal(repoRoot)
}

// loadRepoConfigInternal is the shared internal implementation for repo config loading.
func loadRepoConfigInternal(repoRoot string) (*AppConfig, error) {
	configPath, found := findFirstExistingConfig(repoRoot, appconstants.GetConfigSearchPaths())
	if found {
		return loadConfigFromViper(configPath)
	}

	return &AppConfig{}, nil
}

// findFirstExistingConfig searches for the first existing config file
// from a list of config names within a base directory.
// Returns the full path to the first existing config file, or empty string if none exist.
func findFirstExistingConfig(basePath string, configNames []string) (string, bool) {
	for _, name := range configNames {
		path := filepath.Join(basePath, name)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}

	return "", false
}

// LoadActionConfig loads action-level configuration from config.yaml.
func LoadActionConfig(actionDir string) (*AppConfig, error) {
	return loadActionConfigInternal(actionDir)
}

// loadActionConfigInternal is the shared internal implementation for action config loading.
func loadActionConfigInternal(actionDir string) (*AppConfig, error) {
	configPath := filepath.Join(actionDir, appconstants.ConfigYAML)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &AppConfig{}, nil
	}

	return loadConfigFromViper(configPath)
}

// DetectRepositoryName detects the repository name from git remote URL.
func DetectRepositoryName(repoRoot string) string {
	if repoRoot == "" {
		return ""
	}

	info, err := git.DetectRepository(repoRoot)
	if err != nil {
		return ""
	}

	return info.GetRepositoryName()
}

// loadAndMergeConfig is a helper that loads config from a directory and merges it.
// Returns nil if dir is empty (no-op). Returns error if loading fails.
func loadAndMergeConfig(
	config *AppConfig,
	dir string,
	loadFunc func(string) (*AppConfig, error),
	errorFormat string,
	allowTokens bool,
) error {
	if dir == "" {
		return nil
	}

	loadedConfig, err := loadFunc(dir)
	if err != nil {
		return fmt.Errorf(errorFormat, err)
	}

	MergeConfigs(config, loadedConfig, allowTokens)

	return nil
}

// LoadConfiguration loads configuration with multi-level hierarchy.
func LoadConfiguration(configFile, repoRoot, actionDir string) (*AppConfig, error) {
	// 1. Start with defaults
	config := DefaultAppConfig()

	// 2. Load global config
	globalConfig, err := InitConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf(appconstants.ErrFailedToLoadGlobalConfig, err)
	}
	MergeConfigs(config, globalConfig, true) // Allow tokens for global config

	// 3. Apply repo-specific overrides from global config
	repoName := DetectRepositoryName(repoRoot)
	if repoName != "" {
		if repoOverride, exists := globalConfig.RepoOverrides[repoName]; exists {
			MergeConfigs(config, &repoOverride, false) // No tokens in overrides
		}
	}

	// 4. Load repository root ghreadme.yaml
	if err := loadAndMergeConfig(config, repoRoot, LoadRepoConfig,
		appconstants.ErrFailedToLoadRepoConfig, false); err != nil {
		return nil, err
	}

	// 5. Load action-specific config.yaml
	if err := loadAndMergeConfig(config, actionDir, LoadActionConfig,
		appconstants.ErrFailedToLoadActionConfig, false); err != nil {
		return nil, err
	}

	// 6. Apply environment variable overrides for GitHub token
	// Check environment variables directly with higher priority
	if token := loadGitHubTokenFromEnv(); token != "" {
		config.GitHubToken = token
	}

	return config, nil
}

// InitConfig initializes the global configuration using Viper with XDG compliance.
func InitConfig(configFile string) (*AppConfig, error) {
	v, err := initializeViperInstance()
	if err != nil {
		return nil, err
	}

	return loadAndUnmarshalConfig(configFile, v)
}

// WriteDefaultConfig writes a default configuration file to the XDG config directory.
func WriteDefaultConfig() error {
	configFile, err := xdg.ConfigFile(appconstants.PathXDGConfig)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToGetXDGConfigFile, err)
	}

	// Ensure the directory exists
	configFileDir := filepath.Dir(configFile)
	// #nosec G301 -- config directory permissions
	if err := os.MkdirAll(configFileDir, appconstants.FilePermDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType(appconstants.OutputFormatYAML)

	// Set default values
	defaults := DefaultAppConfig()
	v.Set(appconstants.ConfigKeyTheme, defaults.Theme)
	v.Set(appconstants.ConfigKeyOutputFormat, defaults.OutputFormat)
	v.Set(appconstants.ConfigKeyOutputDir, defaults.OutputDir)
	v.Set(appconstants.ConfigKeyAnalyzeDependencies, defaults.AnalyzeDependencies)
	v.Set(appconstants.ConfigKeyShowSecurityInfo, defaults.ShowSecurityInfo)
	v.Set(appconstants.ConfigKeyVerbose, defaults.Verbose)
	v.Set(appconstants.ConfigKeyQuiet, defaults.Quiet)
	v.Set(appconstants.ConfigKeyTemplate, defaults.Template)
	v.Set(appconstants.ConfigKeyHeader, defaults.Header)
	v.Set(appconstants.ConfigKeyFooter, defaults.Footer)
	v.Set(appconstants.ConfigKeySchema, defaults.Schema)
	v.Set(appconstants.ConfigKeyDefaults, defaults.Defaults)

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the configuration file.
func GetConfigPath() (string, error) {
	configDir, err := xdg.ConfigFile(appconstants.PathXDGConfig)
	if err != nil {
		return "", fmt.Errorf(appconstants.ErrFailedToGetXDGConfigFile, err)
	}

	return configDir, nil
}
