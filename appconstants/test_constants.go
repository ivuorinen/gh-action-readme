package appconstants

// This file contains constants used exclusively for testing.
// These are separated from production constants to:
// - Reduce API surface pollution in the main constants file
// - Make it clear which constants are test-only
// - Improve code organization and maintainability
//
// Note: These constants must remain exported so they can be used by
// test files in other packages (e.g., internal/*_test.go, main_test.go).

// Test assertion message format templates.
const (
	// TestMsgExitCode is the format for exit code mismatch assertions.
	TestMsgExitCode = "expected exit code %d, got %d"

	// TestMsgStdout is the format for standard output logging.
	TestMsgStdout = "stdout: %s"

	// TestMsgStderr is the format for standard error logging.
	TestMsgStderr = "stderr: %s"
)

// Test fixture path constants.
const (
	// JavaScript action fixtures.
	TestFixtureJavaScriptSimple = "actions/javascript/simple.yml"

	// Composite action fixtures.
	TestFixtureCompositeBasic    = "actions/composite/basic.yml"
	TestFixtureCompositeWithDeps = "actions/composite/with-dependencies.yml"

	// Docker action fixtures.
	TestFixtureDockerBasic = "actions/docker/basic.yml"

	// Invalid action fixtures.
	TestFixtureInvalidMissingDescription = "actions/invalid/missing-description.yml"
	TestFixtureInvalidInvalidUsing       = "actions/invalid/invalid-using.yml"

	// Minimal/other fixtures.
	TestFixtureMinimalAction       = "minimal-action.yml"
	TestFixtureProfessionalConfig  = "professional-config.yml"
	TestFixtureTestCompositeAction = "test-composite-action.yml"
	TestFixtureMyNewAction         = "my-new-action.yml"
)

// Test file path constants.
const (
	TestPathConfigYML       = "config.yml"
	TestPathCustomConfigYML = "custom-config.yml"
	TestPathNonexistentYML  = "nonexistent.yml"
)

// Test directory path constants.
const (
	TestDirSubdir           = "subdir"
	TestDirActions          = "actions"
	TestDirActionsDeploy    = "actions/deploy"
	TestDirActionsTest      = "actions/test"
	TestDirActionsComposite = "actions/composite"
	TestDirActionsDocker    = "actions/docker"
	TestDirNested           = "nested"
	TestDirNestedDeep       = "nested/deep"

	// Config directories.
	TestDirConfigGhActionReadme = ".config/gh-action-readme"
	TestDirDotConfig            = ".config"
	TestDirCacheGhActionReadme  = ".cache/gh-action-readme"
)

// (Test file permission constants removed - use production constants from appconstants/constants.go)

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

// Cache test constants for reducing string duplication.
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
	TestOutputDir  = "/tmp/output"
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
