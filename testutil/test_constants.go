package testutil

// This file contains test-only constants moved from appconstants.
// These constants are exported for use across test files in different packages.

// Test cache constants for reducing string duplication.
const (
	CacheTestKey   = "test-key"
	CacheTestValue = "test-value"
)

// Error handler test constants for reducing string duplication.
const (
	UnknownErrorMsg = "unknown error"
	HelloWorldStr   = "hello world"
)

// Validation component test constants for reducing string duplication.
const (
	TestItemName = "test-item"
)

// Generator test constants for reducing string duplication.
const (
	TestActionName = "Test Action"
	TestActionDesc = "Test Description"
)

// Validation test file identifiers for reducing string duplication.
const (
	ValidationTestFile1 = "file: action1.yml"
	ValidationTestFile2 = "file: action2.yml"
	ValidationTestFile3 = "file: action.yml"
)

// GitHub Actions runner names for reducing string duplication.
const (
	RunnerUbuntuLatest  = "ubuntu-latest"
	RunnerWindowsLatest = "windows-latest"
	RunnerMacosLatest   = "macos-latest"
)

// Test assertion message format templates for reducing string duplication.
const (
	TestMsgExitCode = "expected exit code %d, got %d"
	TestMsgStdout   = "stdout: %s"
	TestMsgStderr   = "stderr: %s"
)

// Test fixture path constants for reducing string duplication.
const (
	TestFixtureJavaScriptSimple          = "actions/javascript/simple.yml"
	TestFixtureCompositeBasic            = "actions/composite/basic.yml"
	TestFixtureCompositeWithDeps         = "actions/composite/with-dependencies.yml"
	TestFixtureDockerBasic               = "actions/docker/basic.yml"
	TestFixtureInvalidMissingDescription = "actions/invalid/missing-description.yml"
	TestFixtureInvalidInvalidUsing       = "actions/invalid/invalid-using.yml"
	TestFixtureMinimalAction             = "minimal-action.yml"
	TestFixtureTestCompositeAction       = "test-composite-action.yml"
	TestFixtureMyNewAction               = "my-new-action.yml"
)

// Test file path constants for reducing string duplication.
const (
	TestPathConfigYML = "config.yml"
)

// Test directory path constants for reducing string duplication.
const (
	TestDirSubdir               = "subdir"
	TestDirDotConfig            = ".config"
	TestDirConfigGhActionReadme = ".config/gh-action-readme"
)

// Test YAML content for parser tests.
const (
	TestYAMLRoot        = "name: root"
	TestYAMLNodeModules = "name: node_modules"
	TestYAMLVendor      = "name: vendor"
	TestYAMLGit         = "name: git"
	TestYAMLSrc         = "name: src"
	TestYAMLNested      = "name: nested"
	TestYAMLSub         = "name: sub"
)

// Test YAML template strings for parser tests.
const (
	TestActionFilePattern = "action-*.yml"
	TestPermissionsHeader = "# permissions:\n"
	TestActionNameLine    = "name: Test Action\n"
	TestDescriptionLine   = "description: Test\n"
	TestRunsLine          = "runs:\n"
	TestCompositeUsing    = "  using: composite\n"
	TestStepsEmpty        = "  steps: []\n"
	TestErrorFormat       = "ParseActionYML() error = %v"
	TestContentsRead      = "#   contents: read\n"
)

// Test path constants for template tests.
const (
	TestRepoActionPath      = "/repo/action.yml"
	TestRepoBuildActionPath = "/repo/build/action.yml"
	TestVersionV123         = "@v1.2.3"
)
