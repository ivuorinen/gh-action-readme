package internal

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// ConfigurationLoader handles loading and merging configuration from multiple sources.
type ConfigurationLoader struct {
	// sources tracks which sources are enabled
	sources map[appconstants.ConfigurationSource]bool
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
	}
}

// NewConfigurationLoaderWithOptions creates a configuration loader with custom options.
func NewConfigurationLoaderWithOptions(opts ConfigurationOptions) *ConfigurationLoader {
	loader := &ConfigurationLoader{
		sources: make(map[appconstants.ConfigurationSource]bool),
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
	validFormats := appconstants.GetSupportedOutputFormats()
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

// loadConfigStep is a generic helper for loading and merging configuration from a specific source.
func (cl *ConfigurationLoader) loadConfigStep(
	config *AppConfig,
	sourceName appconstants.ConfigurationSource,
	dirPath string,
	loadFunc func(string) (*AppConfig, error),
	errorFormat string,
	mergeTokens bool,
) error {
	if !cl.sources[sourceName] || dirPath == "" {
		return nil
	}

	loadedConfig, err := loadFunc(dirPath)
	if err != nil {
		return fmt.Errorf(errorFormat, err)
	}
	cl.mergeConfigs(config, loadedConfig, mergeTokens)

	return nil
}

// loadRepoConfigStep loads repository root configuration.
func (cl *ConfigurationLoader) loadRepoConfigStep(config *AppConfig, repoRoot string) error {
	return cl.loadConfigStep(
		config,
		appconstants.SourceRepoConfig,
		repoRoot,
		cl.loadRepoConfig,
		appconstants.ErrFailedToLoadRepoConfig,
		false, // No tokens in repo config
	)
}

// loadActionConfigStep loads action-specific configuration.
func (cl *ConfigurationLoader) loadActionConfigStep(config *AppConfig, actionDir string) error {
	return cl.loadConfigStep(
		config,
		appconstants.SourceActionConfig,
		actionDir,
		cl.loadActionConfig,
		appconstants.ErrFailedToLoadActionConfig,
		false, // No tokens in action config
	)
}

// loadEnvironmentStep applies environment variable overrides.
func (cl *ConfigurationLoader) loadEnvironmentStep(config *AppConfig) {
	if cl.sources[appconstants.SourceEnvironment] {
		cl.applyEnvironmentOverrides(config)
	}
}

// loadGlobalConfig initializes and loads the global configuration using Viper.
func (cl *ConfigurationLoader) loadGlobalConfig(configFile string) (*AppConfig, error) {
	v, err := initializeViperInstance()
	if err != nil {
		return nil, err
	}

	return loadAndUnmarshalConfig(configFile, v)
}

// loadRepoConfig loads repository-level configuration from hidden config files.
func (cl *ConfigurationLoader) loadRepoConfig(repoRoot string) (*AppConfig, error) {
	return loadRepoConfigInternal(repoRoot)
}

// loadActionConfig loads action-level configuration from config.yaml.
func (cl *ConfigurationLoader) loadActionConfig(actionDir string) (*AppConfig, error) {
	return loadActionConfigInternal(actionDir)
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
	if token := loadGitHubTokenFromEnv(); token != "" {
		config.GitHubToken = token
	}
}

// mergeConfigs merges a source config into a destination config.
func (cl *ConfigurationLoader) mergeConfigs(dst *AppConfig, src *AppConfig, allowTokens bool) {
	MergeConfigs(dst, src, allowTokens)
}

// validateTheme validates that a theme exists and is supported.
func (cl *ConfigurationLoader) validateTheme(theme string) error {
	if theme == "" {
		return errors.New("theme cannot be empty")
	}

	// Check if it's a built-in theme
	if containsString(appconstants.GetSupportedThemes(), theme) {
		return nil
	}

	// Check if it's a custom template path
	if filepath.IsAbs(theme) || strings.Contains(theme, "/") {
		// Assume it's a custom template path - we can't easily validate without filesystem access
		return nil
	}

	return fmt.Errorf("unsupported theme '%s', must be one of: %s",
		theme, strings.Join(appconstants.GetSupportedThemes(), ", "))
}
