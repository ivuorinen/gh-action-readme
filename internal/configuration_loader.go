package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

// ConfigurationSource represents different sources of configuration.
type ConfigurationSource int

// Configuration source priority order (lowest to highest priority).
const (
	// SourceDefaults represents default configuration values.
	SourceDefaults ConfigurationSource = iota
	SourceGlobal
	SourceRepoOverride
	SourceRepoConfig
	SourceActionConfig
	SourceEnvironment
	SourceCLIFlags
)

// ConfigurationLoader handles loading and merging configuration from multiple sources.
type ConfigurationLoader struct {
	// sources tracks which sources are enabled
	sources map[ConfigurationSource]bool
	// viper instance for global configuration
	viper *viper.Viper
}

// ConfigurationOptions configures how configuration loading behaves.
type ConfigurationOptions struct {
	// ConfigFile specifies a custom global config file path
	ConfigFile string
	// AllowTokens controls whether security-sensitive fields can be loaded
	AllowTokens bool
	// EnabledSources controls which configuration sources are used
	EnabledSources []ConfigurationSource
}

// NewConfigurationLoader creates a new configuration loader with default options.
func NewConfigurationLoader() *ConfigurationLoader {
	return &ConfigurationLoader{
		sources: map[ConfigurationSource]bool{
			SourceDefaults:     true,
			SourceGlobal:       true,
			SourceRepoOverride: true,
			SourceRepoConfig:   true,
			SourceActionConfig: true,
			SourceEnvironment:  true,
			SourceCLIFlags:     false, // CLI flags are applied separately
		},
		viper: viper.New(),
	}
}

// NewConfigurationLoaderWithOptions creates a configuration loader with custom options.
func NewConfigurationLoaderWithOptions(opts ConfigurationOptions) *ConfigurationLoader {
	loader := &ConfigurationLoader{
		sources: make(map[ConfigurationSource]bool),
		viper:   viper.New(),
	}

	// Set default sources if none specified
	if len(opts.EnabledSources) == 0 {
		opts.EnabledSources = []ConfigurationSource{
			SourceDefaults, SourceGlobal, SourceRepoOverride,
			SourceRepoConfig, SourceActionConfig, SourceEnvironment,
		}
	}

	// Configure enabled sources
	for _, source := range opts.EnabledSources {
		loader.sources[source] = true
	}

	return loader
}

// LoadConfiguration loads configuration with multi-level hierarchy.
func (cl *ConfigurationLoader) LoadConfiguration(configFile, repoRoot, actionDir string) (*AppConfig, error) {
	config := &AppConfig{}

	cl.loadDefaultsStep(config)

	if err := cl.loadGlobalStep(config, configFile); err != nil {
		return nil, err
	}

	cl.loadRepoOverrideStep(config, repoRoot)

	if err := cl.loadRepoConfigStep(config, repoRoot); err != nil {
		return nil, err
	}

	if err := cl.loadActionConfigStep(config, actionDir); err != nil {
		return nil, err
	}

	cl.loadEnvironmentStep(config)

	return config, nil
}

// loadDefaultsStep loads default configuration values.
func (cl *ConfigurationLoader) loadDefaultsStep(config *AppConfig) {
	if cl.sources[SourceDefaults] {
		defaults := DefaultAppConfig()
		*config = *defaults
	}
}

// loadGlobalStep loads global configuration.
func (cl *ConfigurationLoader) loadGlobalStep(config *AppConfig, configFile string) error {
	if !cl.sources[SourceGlobal] {
		return nil
	}

	globalConfig, err := cl.loadGlobalConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}
	cl.mergeConfigs(config, globalConfig, true) // Allow tokens for global config

	return nil
}

// loadRepoOverrideStep applies repo-specific overrides from global config.
func (cl *ConfigurationLoader) loadRepoOverrideStep(config *AppConfig, repoRoot string) {
	if !cl.sources[SourceRepoOverride] || repoRoot == "" {
		return
	}

	cl.applyRepoOverrides(config, repoRoot)
}

// loadRepoConfigStep loads repository root configuration.
func (cl *ConfigurationLoader) loadRepoConfigStep(config *AppConfig, repoRoot string) error {
	if !cl.sources[SourceRepoConfig] || repoRoot == "" {
		return nil
	}

	repoConfig, err := cl.loadRepoConfig(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to load repo config: %w", err)
	}
	cl.mergeConfigs(config, repoConfig, false) // No tokens in repo config

	return nil
}

// loadActionConfigStep loads action-specific configuration.
func (cl *ConfigurationLoader) loadActionConfigStep(config *AppConfig, actionDir string) error {
	if !cl.sources[SourceActionConfig] || actionDir == "" {
		return nil
	}

	actionConfig, err := cl.loadActionConfig(actionDir)
	if err != nil {
		return fmt.Errorf("failed to load action config: %w", err)
	}
	cl.mergeConfigs(config, actionConfig, false) // No tokens in action config

	return nil
}

// loadEnvironmentStep applies environment variable overrides.
func (cl *ConfigurationLoader) loadEnvironmentStep(config *AppConfig) {
	if cl.sources[SourceEnvironment] {
		cl.applyEnvironmentOverrides(config)
	}
}

// LoadGlobalConfig loads only the global configuration.
func (cl *ConfigurationLoader) LoadGlobalConfig(configFile string) (*AppConfig, error) {
	return cl.loadGlobalConfig(configFile)
}

// ValidateConfiguration validates a configuration for consistency and required values.
func (cl *ConfigurationLoader) ValidateConfiguration(config *AppConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Validate output format
	validFormats := []string{"md", "html", "json", "asciidoc"}
	if !containsString(validFormats, config.OutputFormat) {
		return fmt.Errorf("invalid output format '%s', must be one of: %s",
			config.OutputFormat, strings.Join(validFormats, ", "))
	}

	// Validate theme (if set)
	if config.Theme != "" {
		if err := cl.validateTheme(config.Theme); err != nil {
			return fmt.Errorf("invalid theme: %w", err)
		}
	}

	// Validate output directory
	if config.OutputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	// Validate mutually exclusive flags
	if config.Verbose && config.Quiet {
		return fmt.Errorf("verbose and quiet flags are mutually exclusive")
	}

	return nil
}

// loadGlobalConfig initializes and loads the global configuration using Viper.
func (cl *ConfigurationLoader) loadGlobalConfig(configFile string) (*AppConfig, error) {
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
	cl.setViperDefaults(v)

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

// loadRepoConfig loads repository-level configuration from hidden config files.
func (cl *ConfigurationLoader) loadRepoConfig(repoRoot string) (*AppConfig, error) {
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
			return cl.loadConfigFromFile(configPath)
		}
	}

	// No config found, return empty config
	return &AppConfig{}, nil
}

// loadActionConfig loads action-level configuration from config.yaml.
func (cl *ConfigurationLoader) loadActionConfig(actionDir string) (*AppConfig, error) {
	configPath := filepath.Join(actionDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &AppConfig{}, nil // No action config is fine
	}

	return cl.loadConfigFromFile(configPath)
}

// loadConfigFromFile loads configuration from a specific file.
func (cl *ConfigurationLoader) loadConfigFromFile(configPath string) (*AppConfig, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", configPath, err)
	}

	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// applyRepoOverrides applies repository-specific overrides from global config.
func (cl *ConfigurationLoader) applyRepoOverrides(config *AppConfig, repoRoot string) {
	repoName := DetectRepositoryName(repoRoot)
	if repoName == "" {
		return // No repository detected
	}

	if config.RepoOverrides == nil {
		return // No overrides configured
	}

	if repoOverride, exists := config.RepoOverrides[repoName]; exists {
		cl.mergeConfigs(config, &repoOverride, false) // No tokens in overrides
	}
}

// applyEnvironmentOverrides applies environment variable overrides.
func (cl *ConfigurationLoader) applyEnvironmentOverrides(config *AppConfig) {
	// Check environment variables directly with higher priority
	if token := os.Getenv(EnvGitHubToken); token != "" {
		config.GitHubToken = token
	} else if token := os.Getenv(EnvGitHubTokenStandard); token != "" {
		config.GitHubToken = token
	}
}

// mergeConfigs merges a source config into a destination config.
func (cl *ConfigurationLoader) mergeConfigs(dst *AppConfig, src *AppConfig, allowTokens bool) {
	MergeConfigs(dst, src, allowTokens)
}

// setViperDefaults sets default values in viper.
func (cl *ConfigurationLoader) setViperDefaults(v *viper.Viper) {
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
}

// validateTheme validates that a theme exists and is supported.
func (cl *ConfigurationLoader) validateTheme(theme string) error {
	if theme == "" {
		return fmt.Errorf("theme cannot be empty")
	}

	// Check if it's a built-in theme
	supportedThemes := []string{"default", "github", "gitlab", "minimal", "professional"}
	if containsString(supportedThemes, theme) {
		return nil
	}

	// Check if it's a custom template path
	if filepath.IsAbs(theme) || strings.Contains(theme, "/") {
		// Assume it's a custom template path - we can't easily validate without filesystem access
		return nil
	}

	return fmt.Errorf("unsupported theme '%s', must be one of: %s",
		theme, strings.Join(supportedThemes, ", "))
}

// containsString checks if a slice contains a string.
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}

// GetConfigurationSources returns the currently enabled configuration sources.
func (cl *ConfigurationLoader) GetConfigurationSources() []ConfigurationSource {
	var sources []ConfigurationSource
	for source, enabled := range cl.sources {
		if enabled {
			sources = append(sources, source)
		}
	}

	return sources
}

// EnableSource enables a specific configuration source.
func (cl *ConfigurationLoader) EnableSource(source ConfigurationSource) {
	cl.sources[source] = true
}

// DisableSource disables a specific configuration source.
func (cl *ConfigurationLoader) DisableSource(source ConfigurationSource) {
	cl.sources[source] = false
}

// String returns a string representation of a ConfigurationSource.
func (s ConfigurationSource) String() string {
	switch s {
	case SourceDefaults:
		return "defaults"
	case SourceGlobal:
		return "global"
	case SourceRepoOverride:
		return "repo-override"
	case SourceRepoConfig:
		return "repo-config"
	case SourceActionConfig:
		return "action-config"
	case SourceEnvironment:
		return "environment"
	case SourceCLIFlags:
		return "cli-flags"
	default:
		return "unknown"
	}
}
