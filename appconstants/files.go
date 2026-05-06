// Package appconstants provides common constants used throughout the application.
package appconstants

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

	// ActionFilenamePrefix is the filename prefix for GitHub Action files.
	ActionFilenamePrefix = "action"
)

// File permission constants.
const (
	// FilePermDefault is the default file permission for created files and tests.
	FilePermDefault = 0600
	// FilePermDir is the directory permission.
	FilePermDir = 0750
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
	// ConfigYAML is the config.yaml filename.
	ConfigYAML = "config.yaml"
)

// Configuration file constants.
const (
	// ConfigFileName is the primary configuration file name.
	ConfigFileName = "config"
	// ConfigFileExtYAML is the configuration file extension.
	ConfigFileExtYAML = ActionFileExtYAML
	// ConfigFileNameFull is the full configuration file name.
	ConfigFileNameFull = ConfigFileName + ConfigFileExtYAML
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

// defaultIgnoredDirectories lists directories to ignore during file discovery.
var defaultIgnoredDirectories = []string{
	DirGit, DirGitLab, DirSVN, // VCS; keep .github searchable for .github/actions
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

// Environment variable names.
const (
	// EnvGitHubToken is the tool-specific GitHub token environment variable.
	EnvGitHubToken = "GH_README_GITHUB_TOKEN" // #nosec G101 -- environment variable name, not a credential
	// EnvGitHubTokenStandard is the standard GitHub token environment variable.
	EnvGitHubTokenStandard = "GITHUB_TOKEN" // #nosec G101 -- environment variable name, not a credential
)

// File operation constants.
const (
	// BackupExtension is the file backup extension.
	BackupExtension = ".backup"
	// UsesFieldPrefix is the YAML uses field prefix.
	UsesFieldPrefix = "uses: "
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
