package testutil

// This file contains test-only constants moved from appconstants.
// These constants are exported for use across test files in different packages.

// Test cache constants for reducing string duplication.
const (
	CacheTestKey    = "test-key"
	CacheTestValue  = "test-value"
	CacheTestKey1   = "key1"
	CacheTestKey2   = "key2"
	CacheTestValue1 = "value1"
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

// Wizard test constants for reducing string duplication.
const (
	ErrOutputDirMismatch = "OutputDir = %q, want %q"
)

// Generator test constants for reducing string duplication.
const (
	TestActionName = "Test Action"
	TestActionDesc = "Test Description"
)

// GitHub authentication test constants for reducing string duplication.
const (
	TestTokenValue = "test-token"
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
	TestFixtureJavaScriptSimple            = "actions/javascript/simple.yml"
	TestFixtureCompositeBasic              = "actions/composite/basic.yml"
	TestFixtureCompositeWithDeps           = "actions/composite/with-dependencies.yml"
	TestFixtureCompositeMultipleNamedSteps = "actions/composite/with-multiple-named-steps.yml"
	TestFixtureCompositeWithShellStep      = "actions/composite/with-shell-step.yml"
	TestFixtureDockerBasic                 = "actions/docker/basic.yml"
	TestFixtureInvalidMissingDescription   = "actions/invalid/missing-description.yml"
	TestFixtureInvalidInvalidUsing         = "actions/invalid/invalid-using.yml"
	TestFixtureMinimalAction               = "minimal-action.yml"
	TestFixtureTestCompositeAction         = "test-composite-action.yml"
	TestFixtureMyNewAction                 = "my-new-action.yml"
	TestFixtureActionWithCheckoutV3        = "dependencies/action-with-checkout-v3.yml"
	TestFixtureActionWithCheckoutV4        = "dependencies/action-with-checkout-v4.yml"
	TestFixtureSimpleCheckout              = "dependencies/simple-test-checkout.yml"
)

// Dependency update test constants for reducing string duplication in updater_test.go.
const (
	// Actions checkout references for dependency update tests.
	TestCheckoutV4OldUses  = "actions/checkout@v4"
	TestCheckoutPinnedV417 = "actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v4.1.7"
	TestCheckoutPinnedV411 = "actions/checkout@abc123 # v4.1.1"

	// Version string for dependency tests.
	TestVersionV417 = "v4.1.7"
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

// Test error message formats for testutil tests.
const (
	TestErrUnexpected     = "unexpected error: %v"
	TestErrNonEmptyAction = "expected non-empty action content"
	TestErrStatusCode     = "expected status 200, got %d"
)

// Validation test constants.
const (
	TestVersionSemantic = "v1.2.3"
	TestVersionPlain    = "1.2.3"
	TestCaseNameEmpty   = "empty string"
	TestBranchMain      = "main"
)

// Wizard test constants.
const (
	WizardInputYes           = "y\n"
	WizardInputNo            = "n\n"
	WizardInputYesNewline    = "y\ny\n"
	WizardInputThreeNewlines = "\n\n\n"
	WizardInputEnterToken    = "Enter token"
	WizardPromptContinue     = "Continue?"
	WizardOrgTest            = "testorg"
	WizardRepoTest           = "testrepo"
	WizardPromptEnter        = "Enter value"
)

// Test directories and paths for wizard tests.
const (
	TestDirDocs   = "./docs"
	TestDirOutput = "./output"
)

// Test file names for multiple action scenarios.
const (
	TestFileAction1 = "action1.yml"
	TestFileAction2 = "action2.yml"
)

// Test action references.
const (
	TestActionCheckout   = "actions/checkout"
	TestActionCheckoutV4 = "actions/checkout@v4"
)

// Test assertion and error message formats.
const (
	TestMsgThemeFormat           = "Theme = %q, want %q"
	TestMsgAnalyzeDepsTrue       = "AnalyzeDependencies should be true"
	TestMsgNoGitHubToken         = "returns error when no GitHub token"
	TestMsgGitNotInstalled       = "git not installed"
	TestErrPathTraversal         = "path traversal"
	TestInvalidYAMLPrefix        = "invalid: [yaml"
	TestLangJavaScriptTypeScript = "JavaScript/TypeScript"
)
