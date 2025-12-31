package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// ConfigurationLoader handles loading and merging configuration from multiple sources.
type ConfigurationLoader struct {
	// sources tracks which sources are enabled
	sources map[appconstants.ConfigurationSource]bool
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
	EnabledSources []appconstants.ConfigurationSource
}

// NewConfigurationLoader creates a new configuration loader with default options.
func NewConfigurationLoader() *ConfigurationLoader {
	return &ConfigurationLoader{
		sources: map[appconstants.ConfigurationSource]bool{
			appconstants.SourceDefaults:     true,
			appconstants.SourceGlobal:       true,
			appconstants.SourceRepoOverride: true,
			appconstants.SourceRepoConfig:   true,
			appconstants.SourceActionConfig: true,
			appconstants.SourceEnvironment:  true,
			appconstants.SourceCLIFlags:     false, // CLI flags are applied separately
		},
		viper: viper.New(),
	}
}

// NewConfigurationLoaderWithOptions creates a configuration loader with custom options.
func NewConfigurationLoaderWithOptions(opts ConfigurationOptions) *ConfigurationLoader {
	loader := &ConfigurationLoader{
		sources: make(map[appconstants.ConfigurationSource]bool),
		viper:   viper.New(),
	}

	// Set default sources if none specified
	if len(opts.EnabledSources) == 0 {
		opts.EnabledSources = []appconstants.ConfigurationSource{
			appconstants.SourceDefaults, appconstants.SourceGlobal, appconstants.SourceRepoOverride,
			appconstants.SourceRepoConfig, appconstants.SourceActionConfig, appconstants.SourceEnvironment,
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

// LoadGlobalConfig loads only the global configuration.
func (cl *ConfigurationLoader) LoadGlobalConfig(configFile string) (*AppConfig, error) {
	return cl.loadGlobalConfig(configFile)
}

// ValidateConfiguration validates a configuration for consistency and required values.
func (cl *ConfigurationLoader) ValidateConfiguration(config *AppConfig) error {
	if config == nil {
		return errors.New("configuration cannot be nil")
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
		return errors.New("output directory cannot be empty")
	}

	// Validate mutually exclusive flags
	if config.Verbose && config.Quiet {
		return errors.New("verbose and quiet flags are mutually exclusive")
	}

	return nil
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
func (cl *ConfigurationLoader) GetConfigurationSources() []appconstants.ConfigurationSource {
	var sources []appconstants.ConfigurationSource
	for source, enabled := range cl.sources {
		if enabled {
			sources = append(sources, source)
		}
	}

	return sources
}

// EnableSource enables a specific configuration source.
func (cl *ConfigurationLoader) EnableSource(source appconstants.ConfigurationSource) {
	cl.sources[source] = true
}

// DisableSource disables a specific configuration source.
func (cl *ConfigurationLoader) DisableSource(source appconstants.ConfigurationSource) {
	cl.sources[source] = false
}

// loadDefaultsStep loads default configuration values.
func (cl *ConfigurationLoader) loadDefaultsStep(config *AppConfig) {
	if cl.sources[appconstants.SourceDefaults] {
		defaults := DefaultAppConfig()
		*config = *defaults
	}
}

// loadGlobalStep loads global configuration.
func (cl *ConfigurationLoader) loadGlobalStep(config *AppConfig, configFile string) error {
	if !cl.sources[appconstants.SourceGlobal] {
		return nil
	}

	globalConfig, err := cl.loadGlobalConfig(configFile)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToLoadGlobalConfig, err)
	}
	cl.mergeConfigs(config, globalConfig, true) // Allow tokens for global config

	return nil
}

// loadRepoOverrideStep applies repo-specific overrides from global config.
func (cl *ConfigurationLoader) loadRepoOverrideStep(config *AppConfig, repoRoot string) {
	if !cl.sources[appconstants.SourceRepoOverride] || repoRoot == "" {
		return
	}

	cl.applyRepoOverrides(config, repoRoot)
}

// loadRepoConfigStep loads repository root configuration.
func (cl *ConfigurationLoader) loadRepoConfigStep(config *AppConfig, repoRoot string) error {
	if !cl.sources[appconstants.SourceRepoConfig] || repoRoot == "" {
		return nil
	}

	repoConfig, err := cl.loadRepoConfig(repoRoot)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToLoadRepoConfig, err)
	}
	cl.mergeConfigs(config, repoConfig, false) // No tokens in repo config

	return nil
}

// loadActionConfigStep loads action-specific configuration.
func (cl *ConfigurationLoader) loadActionConfigStep(config *AppConfig, actionDir string) error {
	if !cl.sources[appconstants.SourceActionConfig] || actionDir == "" {
		return nil
	}

	actionConfig, err := cl.loadActionConfig(actionDir)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToLoadActionConfig, err)
	}
	cl.mergeConfigs(config, actionConfig, false) // No tokens in action config

	return nil
}

// loadEnvironmentStep applies environment variable overrides.
func (cl *ConfigurationLoader) loadEnvironmentStep(config *AppConfig) {
	if cl.sources[appconstants.SourceEnvironment] {
		cl.applyEnvironmentOverrides(config)
	}
}

// loadGlobalConfig initializes and loads the global configuration using Viper.
func (cl *ConfigurationLoader) loadGlobalConfig(configFile string) (*AppConfig, error) {
	v := viper.New()

	// Set configuration file name and type
	v.SetConfigName(appconstants.ConfigFileName)
	v.SetConfigType(appconstants.OutputFormatYAML)

	// Add XDG-compliant configuration directory
	configDir, err := xdg.ConfigFile(appconstants.AppName)
	if err != nil {
		return nil, fmt.Errorf(appconstants.ErrFailedToGetXDGConfigDir, err)
	}
	v.AddConfigPath(filepath.Dir(configDir))

	// Add additional search paths
	v.AddConfigPath(".")                         // current directory
	v.AddConfigPath(appconstants.PathHomeConfig) // fallback
	v.AddConfigPath(appconstants.PathEtcConfig)  // system-wide

	// Set environment variable prefix
	v.SetEnvPrefix(appconstants.EnvPrefix)
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
			return nil, fmt.Errorf(appconstants.ErrFailedToReadConfigFile, err)
		}
		// Config file not found is not an error - we'll use defaults and env vars
	}

	// Unmarshal configuration into struct
	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf(appconstants.ErrFailedToUnmarshalConfig, err)
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
	configPath := filepath.Join(actionDir, appconstants.ConfigYAML)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &AppConfig{}, nil // No action config is fine
	}

	return cl.loadConfigFromFile(configPath)
}

// loadConfigFromFile loads configuration from a specific file.
func (cl *ConfigurationLoader) loadConfigFromFile(configPath string) (*AppConfig, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType(appconstants.OutputFormatYAML)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", configPath, err)
	}

	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf(appconstants.ErrFailedToUnmarshalConfig, err)
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
	if token := os.Getenv(appconstants.EnvGitHubToken); token != "" {
		config.GitHubToken = token
	} else if token := os.Getenv(appconstants.EnvGitHubTokenStandard); token != "" {
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
	v.SetDefault(appconstants.ConfigKeyOrganization, defaults.Organization)
	v.SetDefault(appconstants.ConfigKeyRepository, defaults.Repository)
	v.SetDefault(appconstants.ConfigKeyVersion, defaults.Version)
	v.SetDefault(appconstants.ConfigKeyTheme, defaults.Theme)
	v.SetDefault(appconstants.ConfigKeyOutputFormat, defaults.OutputFormat)
	v.SetDefault(appconstants.ConfigKeyOutputDir, defaults.OutputDir)
	v.SetDefault(appconstants.ConfigKeyTemplate, defaults.Template)
	v.SetDefault(appconstants.ConfigKeyHeader, defaults.Header)
	v.SetDefault(appconstants.ConfigKeyFooter, defaults.Footer)
	v.SetDefault(appconstants.ConfigKeySchema, defaults.Schema)
	v.SetDefault(appconstants.ConfigKeyAnalyzeDependencies, defaults.AnalyzeDependencies)
	v.SetDefault(appconstants.ConfigKeyShowSecurityInfo, defaults.ShowSecurityInfo)
	v.SetDefault(appconstants.ConfigKeyVerbose, defaults.Verbose)
	v.SetDefault(appconstants.ConfigKeyQuiet, defaults.Quiet)
	v.SetDefault(appconstants.ConfigKeyDefaultsName, defaults.Defaults.Name)
	v.SetDefault(appconstants.ConfigKeyDefaultsDescription, defaults.Defaults.Description)
	v.SetDefault(appconstants.ConfigKeyDefaultsBrandingIcon, defaults.Defaults.Branding.Icon)
	v.SetDefault(appconstants.ConfigKeyDefaultsBrandingColor, defaults.Defaults.Branding.Color)
}

// validateTheme validates that a theme exists and is supported.
func (cl *ConfigurationLoader) validateTheme(theme string) error {
	if theme == "" {
		return errors.New("theme cannot be empty")
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
