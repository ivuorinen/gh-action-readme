// Package appconstants provides common constants used throughout the application.
package appconstants

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
	// ErrorPatternYAML is the yaml error pattern.
	ErrorPatternYAML = "yaml"
	// ErrorPatternGitHub is the github error pattern.
	ErrorPatternGitHub = "github"
	// ErrorPatternConfig is the config error pattern.
	ErrorPatternConfig = "config"
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
	// ErrFailedToGetConfigPath is the failed to get config path error.
	ErrFailedToGetConfigPath = "failed to get config path: %w"
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
