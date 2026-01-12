// Package appconstants provides common constants used throughout the application.
package appconstants

import "time"

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
	// FilePermDefault is the default file permission for created files and tests.
	FilePermDefault = 0600
)

// ErrorCode represents a category of error for providing specific help.
type ErrorCode string

// Error code constants for application error handling.
const (
	// ErrCodeFileNotFound represents file not found errors.
	ErrCodeFileNotFound ErrorCode = "FILE_NOT_FOUND"
	// ErrCodePermission represents permission denied errors.
	ErrCodePermission ErrorCode = "PERMISSION_DENIED"
	// ErrCodeInvalidYAML represents invalid YAML syntax errors.
	ErrCodeInvalidYAML ErrorCode = "INVALID_YAML"
	// ErrCodeInvalidAction represents invalid action file errors.
	ErrCodeInvalidAction ErrorCode = "INVALID_ACTION"
	// ErrCodeNoActionFiles represents no action files found errors.
	ErrCodeNoActionFiles ErrorCode = "NO_ACTION_FILES"
	// ErrCodeGitHubAPI represents GitHub API errors.
	ErrCodeGitHubAPI ErrorCode = "GITHUB_API_ERROR"
	// ErrCodeGitHubRateLimit represents GitHub API rate limit errors.
	ErrCodeGitHubRateLimit ErrorCode = "GITHUB_RATE_LIMIT"
	// ErrCodeGitHubAuth represents GitHub authentication errors.
	ErrCodeGitHubAuth ErrorCode = "GITHUB_AUTH_ERROR"
	// ErrCodeConfiguration represents configuration errors.
	ErrCodeConfiguration ErrorCode = "CONFIG_ERROR"
	// ErrCodeValidation represents validation errors.
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
	// ErrCodeTemplateRender represents template rendering errors.
	ErrCodeTemplateRender ErrorCode = "TEMPLATE_ERROR"
	// ErrCodeFileWrite represents file write errors.
	ErrCodeFileWrite ErrorCode = "FILE_WRITE_ERROR"
	// ErrCodeDependencyAnalysis represents dependency analysis errors.
	ErrCodeDependencyAnalysis ErrorCode = "DEPENDENCY_ERROR"
	// ErrCodeCacheAccess represents cache access errors.
	ErrCodeCacheAccess ErrorCode = "CACHE_ERROR"
	// ErrCodeUnknown represents unknown error types.
	ErrCodeUnknown ErrorCode = "UNKNOWN_ERROR"
)

// Error detection pattern constants.
const (
	// ErrorPatternFileNotFound is the error pattern for file not found errors.
	ErrorPatternFileNotFound = "no such file or directory"
	// ErrorPatternPermission is the error pattern for permission denied errors.
	ErrorPatternPermission = "permission denied"
)

// Exit code constants.
const (
	// ExitCodeError is the exit code for errors.
	ExitCodeError = 1
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

// supportedThemes lists all available theme names (unexported to prevent modification).
var supportedThemes = []string{
	ThemeDefault,
	ThemeGitHub,
	ThemeGitLab,
	ThemeMinimal,
	ThemeProfessional,
}

// GetSupportedThemes returns a copy of the supported theme names.
// Returns a new slice to prevent external modification of the internal list.
func GetSupportedThemes() []string {
	themes := make([]string, len(supportedThemes))
	copy(themes, supportedThemes)

	return themes
}

// Template placeholder constants for Git repository information.
const (
	// DefaultOrgPlaceholder is the default organization placeholder.
	DefaultOrgPlaceholder = "your-org"
	// DefaultRepoPlaceholder is the default repository placeholder.
	DefaultRepoPlaceholder = "your-repo"
	// DefaultUsesPlaceholder is the default uses statement placeholder.
	DefaultUsesPlaceholder = "your-org/your-action@v1"
)

// Environment variable names.
const (
	// EnvGitHubToken is the tool-specific GitHub token environment variable.
	EnvGitHubToken = "GH_README_GITHUB_TOKEN" // #nosec G101 -- environment variable name, not a credential
	// EnvGitHubTokenStandard is the standard GitHub token environment variable.
	EnvGitHubTokenStandard = "GITHUB_TOKEN" // #nosec G101 -- environment variable name, not a credential
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

// configSearchPaths defines the order in which config files are searched (unexported to prevent modification).
var configSearchPaths = []string{
	ConfigFilePatternHidden,
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

// defaultIgnoredDirectories lists directories to ignore during file discovery.
var defaultIgnoredDirectories = []string{
	DirGit, DirGitHub, DirGitLab, DirSVN, // VCS
	DirNodeModules, DirBowerComponents, // JavaScript
	DirVendor,                                       // Go/PHP
	DirVenvDot, DirVenv, DirEnv, DirTox, DirPycache, // Python
	DirDist, DirBuild, DirTarget, DirOut, // Build outputs
	DirIdea, DirVscode, // IDEs
	DirCache, DirTmpDot, DirTmp, // Cache/temp
}

// GetDefaultIgnoredDirectories returns a copy of the default ignored directory names.
// Returns a new slice to prevent external modification of the internal list.
func GetDefaultIgnoredDirectories() []string {
	dirs := make([]string, len(defaultIgnoredDirectories))
	copy(dirs, defaultIgnoredDirectories)

	return dirs
}

// Output format constants.
const (
	// OutputFormatMarkdown is the Markdown output format.
	OutputFormatMarkdown = "md"
	// OutputFormatHTML is the HTML output format.
	OutputFormatHTML = "html"
	// OutputFormatJSON is the JSON output format.
	OutputFormatJSON = "json"
	// OutputFormatYAML is the YAML output format.
	OutputFormatYAML = "yaml"
	// OutputFormatTOML is the TOML output format.
	OutputFormatTOML = "toml"
	// OutputFormatASCIIDoc is the AsciiDoc output format.
	OutputFormatASCIIDoc = "asciidoc"
)

// Common file names.
const (
	// ReadmeMarkdown is the standard README markdown filename.
	ReadmeMarkdown = "README.md"
	// ReadmeASCIIDoc is the AsciiDoc README filename.
	ReadmeASCIIDoc = "README.adoc"
	// ActionDocsJSON is the JSON action docs filename.
	ActionDocsJSON = "action-docs.json"
	// CacheJSON is the cache file name.
	CacheJSON = "cache.json"
	// PackageJSON is the npm package.json filename.
	PackageJSON = "package.json"
	// TemplateReadme is the readme template filename.
	TemplateReadme = "readme.tmpl"
	// TemplateNameReadme is the template name used in template.New().
	TemplateNameReadme = "readme"
	// ConfigYAML is the config.yaml filename.
	ConfigYAML = "config.yaml"
)

// Directory and path constants.
const (
	// DirGit is the .git directory name.
	DirGit = ".git"
	// DirTemplates is the templates directory.
	DirTemplates = "templates/"
	// DirTestdata is the testdata directory.
	DirTestdata = "testdata"
	// DirYAMLFixtures is the yaml-fixtures directory.
	DirYAMLFixtures = "yaml-fixtures"
	// PathEtcConfig is the etc config directory path.
	PathEtcConfig = "/etc/gh-action-readme"
	// PathXDGConfig is the XDG config path pattern.
	PathXDGConfig = "gh-action-readme/config.yaml"
	// AppName is the application name.
	AppName = "gh-action-readme"
	// EnvPrefix is the environment variable prefix.
	EnvPrefix = "GH_ACTION_README"
)

// Directory names commonly ignored during file discovery.
// These constants are used to exclude build artifacts, dependencies,
// version control, and temporary files from action file discovery.
const (
	// Version Control System directories
	// DirGit = ".git" (already defined above in "Directory and path constants").
	DirGitHub = ".github"
	DirGitLab = ".gitlab"
	DirSVN    = ".svn"

	// JavaScript/Node.js dependencies.
	DirNodeModules     = "node_modules"
	DirBowerComponents = "bower_components"

	// Package manager vendor directories.
	DirVendor = "vendor"

	// Python virtual environments and cache.
	DirVenv    = "venv"
	DirVenvDot = ".venv"
	DirEnv     = "env"
	DirTox     = ".tox"
	DirPycache = "__pycache__"

	// Build output directories.
	DirDist   = "dist"
	DirBuild  = "build"
	DirTarget = "target"
	DirOut    = "out"

	// IDE configuration directories.
	DirIdea   = ".idea"
	DirVscode = ".vscode"

	// Cache and temporary directories.
	DirCache  = ".cache"
	DirTmp    = "tmp"
	DirTmpDot = ".tmp"
)

// Git constants.
const (
	// GitCommand is the git command name.
	GitCommand = "git"
	// GitDefaultBranch is the default git branch name.
	GitDefaultBranch = "main"
	// GitShowRef is the git show-ref command.
	GitShowRef = "show-ref"
	// GitVerify is the git --verify flag.
	GitVerify = "--verify"
	// GitQuiet is the git --quiet flag.
	GitQuiet = "--quiet"
	// GitConfigURL is the git config url pattern.
	GitConfigURL = "url = "
)

// Action type constants.
const (
	// ActionTypeComposite is the composite action type.
	ActionTypeComposite = "composite"
	// ActionTypeJavaScript is the JavaScript action type.
	ActionTypeJavaScript = "javascript"
	// ActionTypeDocker is the Docker action type.
	ActionTypeDocker = "docker"
	// ActionTypeInvalid is the invalid action type for testing.
	ActionTypeInvalid = "invalid"
	// ActionTypeMinimal is the minimal action type for testing.
	ActionTypeMinimal = "minimal"
)

// Programming language identifier constants.
const (
	// LangJavaScriptTypeScript is the JavaScript/TypeScript language identifier.
	LangJavaScriptTypeScript = "JavaScript/TypeScript"
	// LangGo is the Go language identifier.
	LangGo = "Go"
	// LangPython is the Python programming language identifier.
	LangPython = "Python"
)

// Update type constants for version comparison.
const (
	// UpdateTypeNone indicates no update is needed.
	UpdateTypeNone = "none"
	// UpdateTypeMajor indicates a major version update.
	UpdateTypeMajor = "major"
	// UpdateTypeMinor indicates a minor version update.
	UpdateTypeMinor = "minor"
	// UpdateTypePatch indicates a patch version update.
	UpdateTypePatch = "patch"
)

// Timeout constants for API operations.
const (
	// APICallTimeout is the timeout for API calls.
	APICallTimeout = 10 * time.Second
	// CacheDefaultTTL is the default cache time-to-live.
	CacheDefaultTTL = 1 * time.Hour
)

// GitHub URL constants.
const (
	// GitHubBaseURL is the base GitHub URL.
	GitHubBaseURL = "https://github.com"
	// MarketplaceBaseURL is the GitHub Marketplace base URL.
	MarketplaceBaseURL = "https://github.com/marketplace/actions/"
)

// Version validation constants.
const (
	// FullSHALength is the full commit SHA length.
	FullSHALength = 40
	// MinSHALength is the minimum commit SHA length.
	MinSHALength = 7
	// VersionPartsCount is the number of parts in semantic versioning.
	VersionPartsCount = 3
)

// Path prefix constants.
const (
	// DockerPrefix is the Docker image prefix.
	DockerPrefix = "docker://"
	// LocalPathPrefix is the local path prefix.
	LocalPathPrefix = "./"
	// LocalPathUpPrefix is the parent directory path prefix.
	LocalPathUpPrefix = "../"
)

// File operation constants.
const (
	// BackupExtension is the file backup extension.
	BackupExtension = ".backup"
	// UsesFieldPrefix is the YAML uses field prefix.
	UsesFieldPrefix = "uses: "
)

// Cache key prefix constants.
const (
	// CacheKeyLatest is the cache key prefix for latest versions.
	CacheKeyLatest = "latest:"
	// CacheKeyRepo is the cache key prefix for repository data.
	CacheKeyRepo = "repo:"
)

// Miscellaneous analysis constants.
const (
	// ScriptLineEstimate is the estimated lines per script step.
	ScriptLineEstimate = 10
)

// Scope level constants.
const (
	// ScopeGlobal is the global scope.
	ScopeGlobal = "global"
	// ScopeUnknown is the unknown scope.
	ScopeUnknown = "unknown"
)

// User input constants.
const (
	// InputYes is the yes confirmation input.
	InputYes = "yes"
	// InputAll is the all input option.
	InputAll = "all"
	// InputDryRun is the dry-run input option.
	InputDryRun = "dry-run"
)

// YAML format string constants for test fixtures and action generation.
const (
	// YAMLFieldName is the YAML name field format.
	YAMLFieldName = "name: %s\n"
	// YAMLFieldDescription is the YAML description field format.
	YAMLFieldDescription = "description: %s\n"
	// YAMLFieldRuns is the YAML runs field.
	YAMLFieldRuns = "runs:\n"
	// JSONCloseBrace is the JSON closing brace with newline.
	JSONCloseBrace = "  },\n"
)

// UI and display constants.
const (
	// SymbolArrow is the arrow symbol for UI.
	SymbolArrow = "►"
	// FormatKeyValue is the key-value format string.
	FormatKeyValue = "%s: %s"
	// FormatDetailKeyValue is the detailed key-value format string.
	FormatDetailKeyValue = "  %s: %s"
	// FormatPrompt is the prompt format string.
	FormatPrompt = "%s: "
	// FormatPromptDefault is the prompt with default format string.
	FormatPromptDefault = "%s [%s]: "
	// FormatEnvVar is the environment variable format string.
	FormatEnvVar = "%s = %q\n"
)

// CLI flag and command names.
const (
	// FlagFormat is the format flag name.
	FlagFormat = "format"
	// FlagOutputDir is the output-dir flag name.
	FlagOutputDir = "output-dir"
	// FlagOutputFormat is the output-format flag name.
	FlagOutputFormat = "output-format"
	// FlagOutput is the output flag name.
	FlagOutput = "output"
	// FlagRecursive is the recursive flag name.
	FlagRecursive = "recursive"
	// FlagIgnoreDirs is the ignore-dirs flag name.
	FlagIgnoreDirs = "ignore-dirs"
	// FlagCI is the CI mode flag name.
	FlagCI = "ci"

	// CommandPin is the pin command name.
	CommandPin = "pin"

	// CacheStatsKeyDir is the cache stats key for directory.
	CacheStatsKeyDir = "cache_dir"
)

// Field names for validation.
const (
	// FieldName is the name field.
	FieldName = "name"
	// FieldDescription is the description field.
	FieldDescription = "description"
	// FieldRuns is the runs field.
	FieldRuns = "runs"
	// FieldRunsUsing is the runs.using field.
	FieldRunsUsing = "runs.using"
)

// Error patterns for error handling.
const (
	// ErrorPatternYAML is the yaml error pattern.
	ErrorPatternYAML = "yaml"
	// ErrorPatternGitHub is the github error pattern.
	ErrorPatternGitHub = "github"
	// ErrorPatternConfig is the config error pattern.
	ErrorPatternConfig = "config"
)

// Regex patterns.
const (
	// RegexGitSHA is the regex pattern for git SHA.
	RegexGitSHA = "^[a-f0-9]{7,40}$"
)

// Token prefixes for validation.
const (
	// TokenPrefixGitHubPersonal is the GitHub personal access token prefix.
	TokenPrefixGitHubPersonal = "ghp_" // #nosec G101 -- token prefix pattern, not a credential
	// TokenPrefixGitHubPAT is the GitHub PAT prefix.
	TokenPrefixGitHubPAT = "github_pat_" // #nosec G101 -- token prefix pattern, not a credential
	// TokenFallback is the fallback token value.
	TokenFallback = "fallback-token" // #nosec G101 -- test value, not a credential
)

// Section markers for output.
const (
	// SectionDetails is the details section marker.
	SectionDetails = "\nDetails:"
	// SectionSuggestions is the suggestions section marker.
	SectionSuggestions = "\nSuggestions:"
)

// URL patterns.
const (
	// URLPatternGitHubRepo is the GitHub repository URL pattern.
	URLPatternGitHubRepo = "%s/%s"
)

// Common error messages.
const (
	// ErrFailedToLoadActionConfig is the failed to load action config error.
	ErrFailedToLoadActionConfig = "failed to load action config: %w"
	// ErrFailedToLoadRepoConfig is the failed to load repo config error.
	ErrFailedToLoadRepoConfig = "failed to load repo config: %w"
	// ErrFailedToLoadGlobalConfig is the failed to load global config error.
	ErrFailedToLoadGlobalConfig = "failed to load global config: %w"
	// ErrFailedToReadConfigFile is the failed to read config file error.
	ErrFailedToReadConfigFile = "failed to read config file: %w"
	// ErrFailedToUnmarshalConfig is the failed to unmarshal config error.
	ErrFailedToUnmarshalConfig = "failed to unmarshal config: %w"
	// ErrFailedToGetXDGConfigDir is the failed to get XDG config directory error.
	ErrFailedToGetXDGConfigDir = "failed to get XDG config directory: %w"
	// ErrFailedToGetXDGConfigFile is the failed to get XDG config file path error.
	ErrFailedToGetXDGConfigFile = "failed to get XDG config file path: %w"
	// ErrFailedToCreateRateLimiter is the failed to create rate limiter error.
	ErrFailedToCreateRateLimiter = "failed to create rate limiter: %w"
	// ErrFailedToGetCurrentDir is the failed to get current directory error.
	ErrFailedToGetCurrentDir = "failed to get current directory: %w"
	// ErrCouldNotCreateDependencyAnalyzer is the could not create dependency analyzer error.
	ErrCouldNotCreateDependencyAnalyzer = "Could not create dependency analyzer: %v"
	// ErrErrorAnalyzing is the error analyzing error.
	ErrErrorAnalyzing = "Error analyzing %s: %v"
	// ErrErrorCheckingOutdated is the error checking outdated error.
	ErrErrorCheckingOutdated = "Error checking outdated for %s: %v"
	// ErrErrorGettingCurrentDir is the error getting current directory error.
	ErrErrorGettingCurrentDir = "Error getting current directory: %v"
	// ErrFailedToApplyUpdates is the failed to apply updates error.
	ErrFailedToApplyUpdates = "Failed to apply updates: %v"
	// ErrFailedToAccessCache is the failed to access cache error.
	ErrFailedToAccessCache = "Failed to access cache: %v"
	// ErrNoActionFilesFound is the no action files found error.
	ErrNoActionFilesFound = "no action files found"
	// ErrFailedToGetCurrentFilePath is the failed to get current file path error.
	ErrFailedToGetCurrentFilePath = "failed to get current file path"
	// ErrFailedToLoadActionFixture is the failed to load action fixture error.
	ErrFailedToLoadActionFixture = "failed to load action fixture %s: %v"
	// ErrFailedToApplyUpdatesWrapped is the failed to apply updates error with wrapping.
	ErrFailedToApplyUpdatesWrapped = "failed to apply updates: %w"
	// ErrFailedToDiscoverActionFiles is the failed to discover action files error with wrapping.
	ErrFailedToDiscoverActionFiles = "failed to discover action files: %w"
	// ErrPathTraversal is the path traversal attempt error.
	ErrPathTraversal = "path traversal detected: output path '%s' attempts to escape output directory '%s'"
	// ErrInvalidOutputPath is the invalid output path error.
	ErrInvalidOutputPath = "invalid output path: %w"
	// ErrFailedToResolveOutputPath is the failed to resolve output path error with wrapping.
	ErrFailedToResolveOutputPath = "failed to resolve output path: %w"
)

// Common message templates.
const (
	// MsgConfigHeader is the config file header.
	MsgConfigHeader = "# gh-action-readme configuration file\n"
	// MsgConfigWizardHeader is the config wizard header.
	MsgConfigWizardHeader = "# Generated by the interactive configuration wizard\n\n"
	// MsgConfigurationExportedTo is the configuration exported to success message.
	MsgConfigurationExportedTo = "Configuration exported to: %s"
)

// Test command names - used across multiple test files.
const (
	TestCmdGen      = "gen"
	TestCmdConfig   = "config"
	TestCmdValidate = "validate"
	TestCmdDeps     = "deps"
	TestCmdShow     = "show"
	TestCmdList     = "list"
)

// Test file paths and names - used across multiple test files.
const (
	TestTmpDir               = "/tmp"
	TestTmpActionFile        = "/tmp/action.yml"
	TestErrorScenarioOldDeps = "error-scenarios/action-with-old-deps.yml"
	TestErrorScenarioMissing = "error-scenarios/missing-required-fields.yml"
	TestErrorScenarioInvalid = "error-scenarios/invalid-yaml-syntax.yml"
)

// TestMinimalAction is the minimal action YAML content for testing.
const TestMinimalAction = "name: Test\ndescription: Test\nruns:\n  using: composite\n  steps: []"

// TestScenarioNoDeps is the common test scenario description for actions with no dependencies.
const TestScenarioNoDeps = "handles action with no dependencies"

// Test messages and error strings - used in output tests.
const (
	TestMsgFileNotFound        = "File not found"
	TestMsgInvalidYAML         = "Invalid YAML"
	TestMsgQuietSuppressOutput = "quiet mode suppresses output"
	TestMsgNoOutputInQuiet     = "Expected no output in quiet mode, got %q"
	TestMsgVerifyPermissions   = "Verify permissions"
	TestMsgSuggestions         = "Suggestions"
	TestMsgDetails             = "Details"
	TestMsgCheckFilePath       = "Check the file path"
	TestMsgTryAgain            = "Try again"
	TestMsgProcessingStarted   = "Processing started"
	TestMsgOperationCompleted  = "Operation completed"
	TestMsgOutputMissingEmoji  = "Output missing error emoji: %q"
)

// Test scenario names - used in output tests.
const (
	TestScenarioColorEnabled  = "with color enabled"
	TestScenarioColorDisabled = "with color disabled"
	TestScenarioQuietEnabled  = "quiet mode enabled"
	TestScenarioQuietDisabled = "quiet mode disabled"
)

// Test URLs and paths - used in output tests.
const (
	TestURLHelp = "https://example.com/help"
	TestKeyFile = "file"
	TestKeyPath = "path"
)

// Test wizard inputs and prompts - used in wizard tests.
const (
	TestWizardInputYes       = "y\n"
	TestWizardInputNo        = "n\n"
	TestWizardInputYesYes    = "y\ny\n"
	TestWizardInputTwo       = "2\n"
	TestWizardInputTripleNL  = "\n\n\n"
	TestWizardInputDoubleNL  = "\n\n"
	TestWizardPromptContinue = "Continue?"
	TestWizardPromptEnter    = "Enter value"
)

// Test repository and organization names - used in wizard tests.
const (
	TestOrgName  = "testorg"
	TestRepoName = "testrepo"
	TestValue    = "test"
	TestVersion  = "v1.0.0"
	TestDocsPath = "./docs"
)

// Test assertion messages - used in wizard tests.
const (
	TestAssertTheme = "Theme = %q, want %q"
)

// Test dependency actions - used in updater tests.
const (
	TestActionCheckoutV4       = "actions/checkout@v4"
	TestActionCheckoutPinned   = "actions/checkout@abc123 # v4.1.1"
	TestActionCheckoutFullSHA  = "actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v4.1.7"
	TestActionCheckoutSHA      = "692973e3d937129bcbf40652eb9f2f61becf3332"
	TestActionCheckoutVersion  = "v4.1.7"
	TestCacheKey               = "test-key"
	TestUpdateTypePatch        = "patch"
	TestDepsSimpleCheckoutFile = "dependencies/simple-test-checkout.yml"
)

// Test paths and output - used in generator tests.
const (
	TestOutputPath = "/tmp/output"
)

// Test HTML content - used in html tests.
const (
	TestHTMLNewContent        = "New content"
	TestHTMLClosingTag        = "\n</html>"
	TestMsgFailedToReadOutput = "Failed to read output file: %v"
)

// Test detector messages - used in detector tests.
const (
	TestMsgFailedToCreateAction = "Failed to create action.yml: %v"
	TestPermRead                = "read"
	TestPermWrite               = "write"
	TestPermContents            = "contents"
)

// Test repository names - used in configuration tests.
const (
	TestRepoTestOrgTestRepo = "test-org/test-repo"
)

// File permissions (additional).
const (
	// FilePermDir is the directory permission.
	FilePermDir = 0750
)

// Integration test directory and file names - used in integration tests.
const (
	TestDirDotGitHub       = ".github"
	TestFileGitIgnore      = ".gitignore"
	TestFileGHActionReadme = "gh-action-readme.yml"
	TestBinaryName         = "gh-action-readme"
)

// Integration test CLI flags - used in integration tests.
const (
	TestFlagOutputFormat = "--output-format"
	TestFlagRecursive    = "--recursive"
	TestFlagTheme        = "--theme"
	TestFlagVerbose      = "--verbose"
)

// Integration test output messages - used in integration tests.
const (
	TestMsgCurrentConfig     = "Current Configuration"
	TestMsgDependenciesFound = "Dependencies found"
)

// Integration test file patterns - used in integration tests.
const (
	TestPatternHTML      = "*.html"
	TestPatternREADME    = "README*.md"
	TestPatternREADMEAll = "**/README*.md"
)

// Config test constants - used in config tests.
const (
	TestFileGHReadmeYAML = ".ghreadme.yaml"
	TestFileConfigYAML   = "config.yaml"
	TestTokenConfig      = "config-token"
	TestTokenStd         = "std-token"
	TestFileCustomConfig = "custom-config.yml"
)

// Dependency analyzer test constants - used in analyzer tests.
const (
	TestActionCheckoutV3   = "actions/checkout@v3"
	TestActionCheckoutName = "actions/checkout"
	TestVersionV4_1_1      = "v4.1.1"
	TestVersionV4_0_0      = "v4.0.0"
	TestSHAForTesting      = "8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e"
)

// String returns a string representation of a ConfigurationSource.
func (s ConfigurationSource) String() string {
	switch s {
	case SourceDefaults:
		return ConfigKeyDefaults
	case SourceGlobal:
		return ScopeGlobal
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
		return ScopeUnknown
	}
}
