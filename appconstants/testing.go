// Package appconstants provides common constants used throughout the application.
package appconstants

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
