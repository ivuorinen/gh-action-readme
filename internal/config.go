// Package internal contains the internal implementation of gh-action-readme.
package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v57/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"

	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/internal/validation"
)

// AppConfig represents the application configuration that can be used at multiple levels.
type AppConfig struct {
	// GitHub API (Global Only - Security)
	GitHubToken string `mapstructure:"github_token" yaml:"github_token,omitempty"` // Only in global config

	// Repository Information (auto-detected, overridable)
	Organization string `mapstructure:"organization" yaml:"organization,omitempty"`
	Repository   string `mapstructure:"repository"   yaml:"repository,omitempty"`
	Version      string `mapstructure:"version"      yaml:"version,omitempty"`

	// Template Settings
	Theme        string `mapstructure:"theme"         yaml:"theme"`
	OutputFormat string `mapstructure:"output_format" yaml:"output_format"`
	OutputDir    string `mapstructure:"output_dir"    yaml:"output_dir"`

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
	Verbose bool `mapstructure:"verbose" yaml:"verbose"`
	Quiet   bool `mapstructure:"quiet"   yaml:"quiet"`

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
	// Priority 1: Tool-specific env var
	if token := os.Getenv(EnvGitHubToken); token != "" {
		return token
	}

	// Priority 2: Standard GitHub env var
	if token := os.Getenv(EnvGitHubTokenStandard); token != "" {
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
			return nil, fmt.Errorf("failed to create rate limiter: %w", err)
		}

		client = github.NewClient(rateLimiter)
	} else {
		// For no token, use basic rate limiter
		rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create rate limiter: %w", err)
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

// resolveTemplatePath resolves a template path relative to the binary directory if it's not absolute.
func resolveTemplatePath(templatePath string) string {
	if filepath.IsAbs(templatePath) {
		return templatePath
	}

	// Check if template exists in current directory first (for tests)
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

// resolveThemeTemplate resolves the template path based on the selected theme.
func resolveThemeTemplate(theme string) string {
	var templatePath string

	switch theme {
	case ThemeDefault:
		templatePath = TemplatePathDefault
	case ThemeGitHub:
		templatePath = TemplatePathGitHub
	case ThemeGitLab:
		templatePath = TemplatePathGitLab
	case ThemeMinimal:
		templatePath = TemplatePathMinimal
	case ThemeProfessional:
		templatePath = TemplatePathProfessional
	case "":
		// Empty theme should return empty path
		return ""
	default:
		// Unknown theme should return empty path
		return ""
	}

	return resolveTemplatePath(templatePath)
}

// DefaultAppConfig returns the default application configuration.
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		// Repository Information (will be auto-detected)
		Organization: "",
		Repository:   "",
		Version:      "",

		// Template Settings
		Theme:        "default", // default, github, gitlab, minimal, professional
		OutputFormat: "md",
		OutputDir:    ".",

		// Legacy template fields (backward compatibility)
		Template: resolveTemplatePath("templates/readme.tmpl"),
		Header:   resolveTemplatePath("templates/header.tmpl"),
		Footer:   resolveTemplatePath("templates/footer.tmpl"),
		Schema:   resolveTemplatePath("schemas/schema.json"),

		// Workflow Requirements
		Permissions: map[string]string{},
		RunsOn:      []string{"ubuntu-latest"},

		// Features
		AnalyzeDependencies: false,
		ShowSecurityInfo:    false,

		// Custom Template Variables
		Variables: map[string]string{},

		// Repository-specific overrides (empty by default)
		RepoOverrides: map[string]AppConfig{},

		// Behavior
		Verbose: false,
		Quiet:   false,

		// Default values for action.yml files (legacy)
		Defaults: DefaultValues{
			Name:        "GitHub Action",
			Description: "A reusable GitHub Action.",
			Runs:        map[string]any{},
			Branding: Branding{
				Icon:  "activity",
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

// mergeMapFields merges map fields from src to dst if non-empty.
func mergeMapFields(dst *AppConfig, src *AppConfig) {
	if len(src.Permissions) > 0 {
		if dst.Permissions == nil {
			dst.Permissions = make(map[string]string)
		}
		for k, v := range src.Permissions {
			dst.Permissions[k] = v
		}
	}

	if len(src.Variables) > 0 {
		if dst.Variables == nil {
			dst.Variables = make(map[string]string)
		}
		for k, v := range src.Variables {
			dst.Variables[k] = v
		}
	}
}

// mergeSliceFields merges slice fields from src to dst if non-empty.
func mergeSliceFields(dst *AppConfig, src *AppConfig) {
	if len(src.RunsOn) > 0 {
		dst.RunsOn = make([]string, len(src.RunsOn))
		copy(dst.RunsOn, src.RunsOn)
	}
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
	// Hidden config file paths in priority order
	configPaths := []string{
		".ghreadme.yaml",        // Primary hidden config
		".config/ghreadme.yaml", // Secondary hidden config
		".github/ghreadme.yaml", // GitHub ecosystem standard
	}

	for _, configName := range configPaths {
		configPath := filepath.Join(repoRoot, configName)
		if _, err := os.Stat(configPath); err == nil {
			// Config file found, load it
			v := viper.New()
			v.SetConfigFile(configPath)
			v.SetConfigType("yaml")

			if err := v.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("failed to read repo config %s: %w", configPath, err)
			}

			var config AppConfig
			if err := v.Unmarshal(&config); err != nil {
				return nil, fmt.Errorf("failed to unmarshal repo config: %w", err)
			}

			return &config, nil
		}
	}

	// No config found, return empty config
	return &AppConfig{}, nil
}

// LoadActionConfig loads action-level configuration from config.yaml.
func LoadActionConfig(actionDir string) (*AppConfig, error) {
	configPath := filepath.Join(actionDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &AppConfig{}, nil // No action config is fine
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read action config %s: %w", configPath, err)
	}

	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal action config: %w", err)
	}

	return &config, nil
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

// LoadConfiguration loads configuration with multi-level hierarchy.
func LoadConfiguration(configFile, repoRoot, actionDir string) (*AppConfig, error) {
	// 1. Start with defaults
	config := DefaultAppConfig()

	// 2. Load global config
	globalConfig, err := InitConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
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
	if repoRoot != "" {
		repoConfig, err := LoadRepoConfig(repoRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to load repo config: %w", err)
		}
		MergeConfigs(config, repoConfig, false) // No tokens in repo config
	}

	// 5. Load action-specific config.yaml
	if actionDir != "" {
		actionConfig, err := LoadActionConfig(actionDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load action config: %w", err)
		}
		MergeConfigs(config, actionConfig, false) // No tokens in action config
	}

	// 6. Apply environment variable overrides for GitHub token
	// Check environment variables directly with higher priority
	if token := os.Getenv(EnvGitHubToken); token != "" {
		config.GitHubToken = token
	} else if token := os.Getenv(EnvGitHubTokenStandard); token != "" {
		config.GitHubToken = token
	}

	return config, nil
}

// InitConfig initializes the global configuration using Viper with XDG compliance.
func InitConfig(configFile string) (*AppConfig, error) {
	v := viper.New()

	// Set configuration file name and type
	v.SetConfigName(ConfigFileName)
	v.SetConfigType("yaml")

	// Add XDG-compliant configuration directory
	configDir, err := xdg.ConfigFile("gh-action-readme")
	if err != nil {
		return nil, fmt.Errorf("failed to get XDG config directory: %w", err)
	}
	v.AddConfigPath(filepath.Dir(configDir))

	// Add additional search paths
	v.AddConfigPath(".")                              // current directory
	v.AddConfigPath("$HOME/.config/gh-action-readme") // fallback
	v.AddConfigPath("/etc/gh-action-readme")          // system-wide

	// Set environment variable prefix
	v.SetEnvPrefix("GH_ACTION_README")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	// Set defaults
	defaults := DefaultAppConfig()
	v.SetDefault("organization", defaults.Organization)
	v.SetDefault("repository", defaults.Repository)
	v.SetDefault("version", defaults.Version)
	v.SetDefault("theme", defaults.Theme)
	v.SetDefault("output_format", defaults.OutputFormat)
	v.SetDefault("output_dir", defaults.OutputDir)
	v.SetDefault("template", defaults.Template)
	v.SetDefault("header", defaults.Header)
	v.SetDefault("footer", defaults.Footer)
	v.SetDefault("schema", defaults.Schema)
	v.SetDefault("analyze_dependencies", defaults.AnalyzeDependencies)
	v.SetDefault("show_security_info", defaults.ShowSecurityInfo)
	v.SetDefault("verbose", defaults.Verbose)
	v.SetDefault("quiet", defaults.Quiet)
	v.SetDefault("defaults.name", defaults.Defaults.Name)
	v.SetDefault("defaults.description", defaults.Defaults.Description)
	v.SetDefault("defaults.branding.icon", defaults.Defaults.Branding.Icon)
	v.SetDefault("defaults.branding.color", defaults.Defaults.Branding.Color)

	// Use specific config file if provided
	if configFile != "" {
		v.SetConfigFile(configFile)
	}

	// Read configuration
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is not an error - we'll use defaults and env vars
	}

	// Unmarshal configuration into struct
	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Resolve template paths relative to binary if they're not absolute
	config.Template = resolveTemplatePath(config.Template)
	config.Header = resolveTemplatePath(config.Header)
	config.Footer = resolveTemplatePath(config.Footer)
	config.Schema = resolveTemplatePath(config.Schema)

	return &config, nil
}

// WriteDefaultConfig writes a default configuration file to the XDG config directory.
func WriteDefaultConfig() error {
	configFile, err := xdg.ConfigFile("gh-action-readme/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to get XDG config file path: %w", err)
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(configFile), 0750); err != nil { // #nosec G301 -- config directory permissions
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	// Set default values
	defaults := DefaultAppConfig()
	v.Set("theme", defaults.Theme)
	v.Set("output_format", defaults.OutputFormat)
	v.Set("output_dir", defaults.OutputDir)
	v.Set("analyze_dependencies", defaults.AnalyzeDependencies)
	v.Set("show_security_info", defaults.ShowSecurityInfo)
	v.Set("verbose", defaults.Verbose)
	v.Set("quiet", defaults.Quiet)
	v.Set("template", defaults.Template)
	v.Set("header", defaults.Header)
	v.Set("footer", defaults.Footer)
	v.Set("schema", defaults.Schema)
	v.Set("defaults", defaults.Defaults)

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the configuration file.
func GetConfigPath() (string, error) {
	configDir, err := xdg.ConfigFile("gh-action-readme/config.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to get XDG config file path: %w", err)
	}
	return configDir, nil
}
