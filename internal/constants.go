// Package internal provides common constants used throughout the application.
package internal

// File extension constants.
const (
	// ActionFileExtYML is the primary action file extension.
	ActionFileExtYML = ".yml"
	// ActionFileExtYAML is the alternative action file extension.
	ActionFileExtYAML = ".yaml"

	// ActionFileNameYML is the primary action file name.
	ActionFileNameYML = "action.yml"
	// ActionFileNameYAML is the alternative action file name.
	ActionFileNameYAML = "action.yaml"
)

// File permission constants.
const (
	// FilePermDefault is the default file permission for created files.
	FilePermDefault = 0600
	// FilePermTest is the file permission used in tests.
	FilePermTest = 0600
)

// Configuration file constants.
const (
	// ConfigFileName is the primary configuration file name.
	ConfigFileName = "config"
	// ConfigFileExtYAML is the configuration file extension.
	ConfigFileExtYAML = ".yaml"
	// ConfigFileNameFull is the full configuration file name.
	ConfigFileNameFull = ConfigFileName + ConfigFileExtYAML
)

// Context key constants for maps and data structures.
const (
	// ContextKeyError is used as a key for error information in context maps.
	ContextKeyError = "error"
	// ContextKeyTheme is used as a key for theme information.
	ContextKeyTheme = "theme"
	// ContextKeyConfig is used as a key for configuration information.
	ContextKeyConfig = "config"
)

// Common string identifiers.
const (
	// ThemeGitHub is the GitHub theme identifier.
	ThemeGitHub = "github"
	// ThemeGitLab is the GitLab theme identifier.
	ThemeGitLab = "gitlab"
	// ThemeMinimal is the minimal theme identifier.
	ThemeMinimal = "minimal"
	// ThemeProfessional is the professional theme identifier.
	ThemeProfessional = "professional"
	// ThemeDefault is the default theme identifier.
	ThemeDefault = "default"
)

// Environment variable names.
const (
	// EnvGitHubToken is the tool-specific GitHub token environment variable.
	EnvGitHubToken = "GH_README_GITHUB_TOKEN" // #nosec G101 -- environment variable name, not a credential
	// EnvGitHubTokenStandard is the standard GitHub token environment variable.
	EnvGitHubTokenStandard = "GITHUB_TOKEN" // #nosec G101 -- environment variable name, not a credential
)

// Configuration keys and paths.
const (
	// ConfigKeyGitHubToken is the configuration key for GitHub token.
	ConfigKeyGitHubToken = "github_token"
	// ConfigKeyTheme is the configuration key for theme.
	ConfigKeyTheme = "theme"
	// ConfigKeyOutputFormat is the configuration key for output format.
	ConfigKeyOutputFormat = "output_format"
	// ConfigKeyOutputDir is the configuration key for output directory.
	ConfigKeyOutputDir = "output_dir"
	// ConfigKeyVerbose is the configuration key for verbose mode.
	ConfigKeyVerbose = "verbose"
	// ConfigKeyQuiet is the configuration key for quiet mode.
	ConfigKeyQuiet = "quiet"
	// ConfigKeyAnalyzeDependencies is the configuration key for dependency analysis.
	ConfigKeyAnalyzeDependencies = "analyze_dependencies"
	// ConfigKeyShowSecurityInfo is the configuration key for security info display.
	ConfigKeyShowSecurityInfo = "show_security_info"
)

// Template path constants.
const (
	// TemplatePathDefault is the default template path.
	TemplatePathDefault = "templates/readme.tmpl"
	// TemplatePathGitHub is the GitHub theme template path.
	TemplatePathGitHub = "templates/themes/github/readme.tmpl"
	// TemplatePathGitLab is the GitLab theme template path.
	TemplatePathGitLab = "templates/themes/gitlab/readme.tmpl"
	// TemplatePathMinimal is the minimal theme template path.
	TemplatePathMinimal = "templates/themes/minimal/readme.tmpl"
	// TemplatePathProfessional is the professional theme template path.
	TemplatePathProfessional = "templates/themes/professional/readme.tmpl"
)

// Config file search patterns.
const (
	// ConfigFilePatternHidden is the primary hidden config file pattern.
	ConfigFilePatternHidden = ".ghreadme.yaml"
	// ConfigFilePatternConfig is the secondary config directory pattern.
	ConfigFilePatternConfig = ".config/ghreadme.yaml"
	// ConfigFilePatternGitHub is the GitHub ecosystem config pattern.
	ConfigFilePatternGitHub = ".github/ghreadme.yaml"
)
