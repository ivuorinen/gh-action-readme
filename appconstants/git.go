// Package appconstants provides common constants used throughout the application.
package appconstants

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
	// GitObjectTypeTag is the GitHub Git ref object type for an annotated tag.
	GitObjectTypeTag = "tag"
)

// GitHub Actions runner constants.
const (
	// RunnerUbuntuLatest is the latest Ubuntu runner.
	RunnerUbuntuLatest = "ubuntu-latest"
	// RunnerWindowsLatest is the latest Windows runner.
	RunnerWindowsLatest = "windows-latest"
	// RunnerMacosLatest is the latest macOS runner.
	RunnerMacosLatest = "macos-latest"
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

// GitHub URL constants.
const (
	// GitHubBaseURL is the base GitHub URL.
	GitHubBaseURL = "https://github.com"
	// MarketplaceBaseURL is the GitHub Marketplace base URL.
	MarketplaceBaseURL = "https://github.com/marketplace/actions/"
)
