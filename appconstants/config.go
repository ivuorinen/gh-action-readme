// Package appconstants provides common constants used throughout the application.
package appconstants

// Context key constants for maps and data structures.
const (
	// ContextKeyError is used as a key for error information in context maps.
	ContextKeyError = "error"
	// ContextKeyConfig is used as a key for configuration information.
	ContextKeyConfig = "config"
)

// Configuration keys - organized by functional groups.
const (
	// Repository/Project Configuration
	// ConfigKeyOrganization is the organization config key.
	ConfigKeyOrganization = "organization"
	// ConfigKeyRepository is the repository config key.
	ConfigKeyRepository = "repository"
	// ConfigKeyVersion is the version config key.
	ConfigKeyVersion = "version"
	// ConfigKeyUseDefaultBranch is the configuration key for use default branch behavior.
	ConfigKeyUseDefaultBranch = "use_default_branch"
	// ConfigKeyRepoOverrides is the configuration key for per-repository overrides.
	ConfigKeyRepoOverrides = "repo_overrides"

	// Template Configuration
	// ConfigKeyTheme is the configuration key for theme.
	ConfigKeyTheme = "theme"
	// ConfigKeyTemplate is the template config key.
	ConfigKeyTemplate = "template"
	// ConfigKeyHeader is the header config key.
	ConfigKeyHeader = "header"
	// ConfigKeyFooter is the footer config key.
	ConfigKeyFooter = "footer"
	// ConfigKeySchema is the schema config key.
	ConfigKeySchema = "schema"

	// Output Configuration
	// ConfigKeyOutputFormat is the configuration key for output format.
	ConfigKeyOutputFormat = "output_format"
	// ConfigKeyOutputDir is the configuration key for output directory.
	ConfigKeyOutputDir = "output_dir"

	// Feature Flags
	// ConfigKeyAnalyzeDependencies is the configuration key for dependency analysis.
	ConfigKeyAnalyzeDependencies = "analyze_dependencies"
	// ConfigKeyShowSecurityInfo is the configuration key for security info display.
	ConfigKeyShowSecurityInfo = "show_security_info"

	// Behavior Flags
	// ConfigKeyVerbose is the configuration key for verbose mode.
	ConfigKeyVerbose = "verbose"
	// ConfigKeyQuiet is the configuration key for quiet mode.
	ConfigKeyQuiet = "quiet"
	// ConfigKeyIgnoredDirectories is the configuration key for ignored directories during discovery.
	ConfigKeyIgnoredDirectories = "ignored_directories"

	// GitHub Integration
	// ConfigKeyGitHubToken is the configuration key for GitHub token.
	ConfigKeyGitHubToken = "github_token"

	// Default Values Configuration
	// ConfigKeyDefaults is the defaults config key.
	ConfigKeyDefaults = "defaults"
	// ConfigKeyDefaultsName is the defaults.name config key.
	ConfigKeyDefaultsName = "defaults.name"
	// ConfigKeyDefaultsDescription is the defaults.description config key.
	ConfigKeyDefaultsDescription = "defaults.description"
	// ConfigKeyDefaultsBrandingIcon is the defaults.branding.icon config key.
	ConfigKeyDefaultsBrandingIcon = "defaults.branding.icon"
	// ConfigKeyDefaultsBrandingColor is the defaults.branding.color config key.
	ConfigKeyDefaultsBrandingColor = "defaults.branding.color"
)

// ConfigurationSource represents different sources of configuration.
type ConfigurationSource int

// Configuration source priority constants (lowest to highest priority).
const (
	// SourceDefaults represents default configuration values.
	SourceDefaults ConfigurationSource = iota
	// SourceGlobal represents global user configuration.
	SourceGlobal
	// SourceRepoOverride represents repository-specific overrides from global config.
	SourceRepoOverride
	// SourceRepoConfig represents repository-level configuration.
	SourceRepoConfig
	// SourceActionConfig represents action-specific configuration.
	SourceActionConfig
	// SourceEnvironment represents environment variable configuration.
	SourceEnvironment
	// SourceCLIFlags represents command-line flag configuration.
	SourceCLIFlags
)

// Config file search patterns.
const (
	// ConfigFilePatternHidden is the primary hidden config file pattern.
	ConfigFilePatternHidden = ".ghreadme.yaml"
	// ConfigFilePatternHiddenLegacy is the legacy hidden config file pattern.
	ConfigFilePatternHiddenLegacy = ".gh-action-readme.yml"
	// ConfigFilePatternHiddenLegacyYAML is the legacy hidden config YAML pattern.
	ConfigFilePatternHiddenLegacyYAML = ".gh-action-readme.yaml"
	// ConfigFilePatternConfig is the secondary config directory pattern.
	ConfigFilePatternConfig = ".config/gh-action-readme/config.yaml"
	// ConfigFilePatternGitHub is the GitHub ecosystem config pattern.
	ConfigFilePatternGitHub = ".github/gh-action-readme.yaml"
)

// configSearchPaths defines the order in which config files are searched (unexported to prevent modification).
var configSearchPaths = []string{
	ConfigFilePatternHidden,
	ConfigFilePatternHiddenLegacy,
	ConfigFilePatternHiddenLegacyYAML,
	ConfigFilePatternConfig,
	ConfigFilePatternGitHub,
}

// GetConfigSearchPaths returns a copy of the config search paths.
// Returns a new slice to prevent external modification of the internal list.
func GetConfigSearchPaths() []string {
	paths := make([]string, len(configSearchPaths))
	copy(paths, configSearchPaths)

	return paths
}
