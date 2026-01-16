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

	// TestErrFileNotFound is used in error handler tests for file not found scenarios.
	TestErrFileNotFound = "file not found"

	// TestErrFileError is used in error handler tests for generic file errors.
	TestErrFileError = "file error"

	// TestErrPermissionDenied is used in error handler tests for permission errors.
	TestErrPermissionDenied = "permission denied"
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
	TestFixtureEmptyAction                 = "error-scenarios/empty-action.yml"
	TestFixtureGlobalConfig                = "configs/global/default.yml"
	TestFixtureProfessionalConfig          = "professional-config.yml"
	TestFixtureRepoConfig                  = "repo-config.yml"
	TestFixtureActionSimple                = "actions/simple/action.yml"
	TestFixtureActionMinimal               = "actions/minimal/action.yml"

	// Config test fixtures for configuration tests.
	TestConfigGlobalGitHubHTML        = "configs/global-github-html.yml"
	TestConfigGlobalDefaultMD         = "configs/global-default-md.yml"
	TestConfigGlobalGitHubHTMLVerbose = "configs/global-github-html-verbose.yml"
	TestConfigMinimalWithToken        = "configs/minimal-with-token.yml" // #nosec G101 -- fixture path

	// Error scenario fixtures for error handling tests.
	TestErrorInvalidYAMLBrackets     = "error-scenarios/invalid-yaml-brackets.yml"
	TestErrorInvalidYAMLBraces       = "error-scenarios/invalid-yaml-braces.yml"
	TestErrorInvalidYAMLTripleBraces = "error-scenarios/invalid-yaml-triple-braces.yml"

	// JSON fixture paths - located in testdata/yaml-fixtures/json-fixtures/.
	TestJSONPackageFull        = "json-fixtures/package-full.json"
	TestJSONPackageVersionOnly = "json-fixtures/package-version-only.json"

	// Permission test fixtures for parser tests.
	TestFixturePermissionsDashSingle     = "permissions/dash-format-single.yml"
	TestFixturePermissionsDashMultiple   = "permissions/dash-format-multiple.yml"
	TestFixturePermissionsObject         = "permissions/object-format.yml"
	TestFixturePermissionsInlineComments = "permissions/inline-comments.yml"
	TestFixturePermissionsMixed          = "permissions/mixed-format.yml"
	TestFixturePermissionsEmpty          = "permissions/empty-block.yml"
	TestFixturePermissionsNone           = "permissions/no-permissions.yml"
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

	// Common test assertion format strings for reducing duplication.
	TestMsgGotWant           = "got %v, want %v"                // Used in test runners and assertions
	TestErrNoErrorGotNone    = "expected error but got none"    // Used in error validation helpers
	TestMsgFailedReadFile    = "failed to read file %s: %v"     // Used in file assertion helpers
	TestMsgFileContent       = "File content:\n%s"              // Used in file content logging
	TestMsgExpectedNonEmpty  = "expected non-empty result"      // Used for non-empty result assertions
	TestMsgFailedReadOutput  = "Failed to read output file: %v" // Used for output file read errors
	TestMsgExpected1InfoCall = "expected 1 Info call, got %d"   // Used in logger mock tests
	TestMsgExportConfigError = "ExportConfig() error = %v"      // Used in config export tests
)

// Validation test constants.
const (
	TestVersionSemantic = "v1.2.3"
	TestVersionPlain    = "1.2.3"
	TestCaseNameEmpty   = "empty string"
	TestBranchMain      = "main"
	TestGitRefMain      = "refs/heads/main"
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
	TestMsgExpectedNonNilConfig  = "expected non-nil config"
)

// Test commands - moved from appconstants for better separation.
const (
	TestCmdGen      = "gen"
	TestCmdConfig   = "config"
	TestCmdValidate = "validate"
	TestCmdDeps     = "deps"
	TestCmdShow     = "show"
	TestCmdList     = "list"
	TestCmdUpgrade  = "upgrade"
)

// Test file paths and names - moved from appconstants.
const (
	TestTmpDir                     = "/tmp"
	TestTmpActionFile              = "/tmp/action.yml"
	TestPathTempAction             = "/tmp/test-action/action.yml"
	TestErrorScenarioOldDeps       = "error-scenarios/action-with-old-deps.yml"
	TestErrorScenarioInvalidYAML   = "error-scenarios/invalid-yaml-syntax.yml"
	TestErrorScenarioMissingFields = "error-scenarios/missing-required-fields.yml"
)

// TestMinimalAction is the minimal action YAML content for testing.
const TestMinimalAction = "name: Test\ndescription: Test\nruns:\n  using: composite\n  steps: []"

// TestScenarioNoDeps is the common test scenario description for actions with no dependencies.
const TestScenarioNoDeps = "handles action with no dependencies"

// Test messages and error strings - moved from appconstants.
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

// Test scenario names - moved from appconstants.
const (
	TestScenarioColorEnabled  = "with color enabled"
	TestScenarioColorDisabled = "with color disabled"
	TestScenarioQuietEnabled  = "quiet mode enabled"
	TestScenarioQuietDisabled = "quiet mode disabled"
)

// Test URLs and paths - moved from appconstants.
const (
	TestURLHelp           = "https://example.com/help"
	TestURLGitHubAPI      = "https://api.github.com/"
	TestURLGitHub         = "https://github.com/"
	TestURLGitHubUserRepo = "https://github.com/user/repo"
	TestKeyFile           = "file"
	TestKeyPath           = "path"
)

// Test repository and organization values - moved from appconstants.
const (
	TestValue   = "test"
	TestVersion = "v1.0.0"
)

// Test dependency actions - moved from appconstants.
const (
	TestActionCheckoutV3  = "actions/checkout@v3"
	TestActionCheckoutSHA = "692973e3d937129bcbf40652eb9f2f61becf3332"
	TestActionSetupNodeV3 = "actions/setup-node@v3"
	TestActionSetupGoV4   = "actions/setup-go@v4"
)

// Test paths and output - moved from appconstants.
const (
	TestOutputPath = "/tmp/output"
)

// Test HTML content - moved from appconstants.
const (
	TestHTMLNewContent        = "New content"
	TestHTMLClosingTag        = "\n</html>"
	TestMsgFailedToReadOutput = "Failed to read output file: %v"
)

// Test detector messages - moved from appconstants.
const (
	TestMsgFailedToCreateAction = "Failed to create action.yml: %v"
	TestPermRead                = "read"
	TestPermWrite               = "write"
	TestPermContents            = "contents"
)

// Test repository names - moved from appconstants.
const (
	TestRepoTestOrgTestRepo = "test-org/test-repo"
	TestRepoTestRepo        = "test/repo"
)

// Integration test directory and file names - moved from appconstants.
const (
	TestDirDotGitHub       = ".github"
	TestFileGitIgnore      = ".gitignore"
	TestFileGHActionReadme = "gh-action-readme.yml"
	TestBinaryName         = "gh-action-readme"
)

// Integration test CLI flags - moved from appconstants.
const (
	TestFlagOutputFormat = "--output-format"
	TestFlagRecursive    = "--recursive"
	TestFlagTheme        = "--theme"
	TestFlagVerbose      = "--verbose"
)

// Integration test output messages - moved from appconstants.
const (
	TestMsgCurrentConfig     = "Current Configuration"
	TestMsgDependenciesFound = "Dependencies found"
)

// Integration test file patterns - moved from appconstants.
const (
	TestPatternHTML      = "*.html"
	TestPatternREADME    = "README*.md"
	TestPatternREADMEAll = "**/README*.md"
)

// Config test constants - moved from appconstants.
const (
	TestFileGHReadmeYAML = ".ghreadme.yaml"
	TestFileConfigYAML   = "config.yaml"
	TestTokenConfig      = "config-token"
	TestTokenStd         = "ghp_test1234567890abcdefghijklmnopqrstuvwxyz"
	TestTokenEnv         = "env-token"
	TestFileCustomConfig = "custom-config.yml"
)

// Theme constants for testing - reducing string duplication across test files.
const (
	TestThemeDefault      = "default"
	TestThemeGitHub       = "github"
	TestThemeGitLab       = "gitlab"
	TestThemeMinimal      = "minimal"
	TestThemeProfessional = "professional"
	TestThemeASCIIDoc     = "asciidoc"
)

// Template path constants for testing - reducing hardcoded template paths.
const (
	TestTemplateReadme       = "readme.tmpl"
	TestTemplateWithPrefix   = "templates/readme.tmpl"
	TestTemplateGitHub       = "themes/github/readme.tmpl"
	TestTemplateGitLab       = "themes/gitlab/readme.tmpl"
	TestTemplateMinimal      = "themes/minimal/readme.tmpl"
	TestTemplateProfessional = "themes/professional/readme.tmpl"
	TestTemplateASCIIDoc     = "themes/asciidoc/readme.adoc"
)

// Dependency analyzer test constants - moved from appconstants.
const (
	TestVersionV4_1_1 = "v4.1.1"
	TestVersionV4_0_0 = "v4.0.0"
	TestSHAForTesting = "8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e"
)

// File discovery test error messages for reducing string duplication in tests.
const (
	// TestErrDiscoveredFileCountFormat is used when file discovery returns unexpected count.
	TestErrDiscoveredFileCountFormat = "DiscoverActionFiles() returned %d files, want %d"

	// TestErrFileNotFoundInResults is used when expected file is missing from discovery.
	TestErrFileNotFoundInResults = "Expected file %s not found in results"

	// TestErrDiscoveredNestedFilesSkipped is used when nested files should be skipped.
	TestErrDiscoveredNestedFilesSkipped = "DiscoverActionFiles() returned %d files, want 0 (nested dirs should be skipped)"

	// TestErrDiscoveredNonRecursive is used for non-recursive discovery tests.
	TestErrDiscoveredNonRecursive = "DiscoverActionFiles() non-recursive returned %d files, want %d"

	// TestErrMsgParseActionYAML is used when action.yml parsing fails.
	TestErrMsgParseActionYAML = "failed to parse action.yml"

	// TestErrMsgInvalidConfig is used when configuration is invalid.
	TestErrMsgInvalidConfig = "invalid configuration"
)

// Assertion message formats for reducing string duplication in tests.
const (
	// TestMsgShouldIgnoreDirectory is used in shouldIgnoreDirectory tests.
	TestMsgShouldIgnoreDirectory = "shouldIgnoreDirectory(%q, %v) = %v, want %v"

	// TestMsgWalkFuncError is used in walkFunc tests.
	TestMsgWalkFuncError = "walkFunc() with valid directory should return nil, got: %v"

	// TestMsgFileContentMismatch is used when file content doesn't match expectations.
	TestMsgFileContentMismatch = "file content mismatch in %s"
)

// Malformed YAML fixture paths for reducing string duplication in error scenario tests.
const (
	// TestFixtureMalformedBracket has unclosed bracket for testing YAML parse errors.
	TestFixtureMalformedBracket = "error-scenarios/malformed-bracket.yml"

	// TestFixtureMalformedIndentation has invalid indentation for testing YAML parse errors.
	TestFixtureMalformedIndentation = "error-scenarios/malformed-indentation.yml"
)

// Additional assertion message formats for reducing string duplication in tests.
const (
	// TestMsgExpectedError is used when error is expected but not returned.
	TestMsgExpectedError = "expected error, got nil"

	// TestMsgUnexpectedSuccess is used when expecting success but got error.
	TestMsgUnexpectedSuccess = "expected success, got error: %v"

	// TestMsgCountMismatch is used when counts don't match expectations.
	TestMsgCountMismatch = "expected %d items, got %d"
)

// Config-related test constants for reducing string duplication in config tests.
const (
	// TestConfigEmpty is an empty JSON config.
	TestConfigEmpty = "{}"

	// TestConfigMinimal is a minimal JSON config with version.
	TestConfigMinimal = `{"version": "1.0.0"}`
)

// Validation message constants for reducing string duplication in validation tests.
const (
	// TestMsgCannotBeEmpty is a common validation error message.
	TestMsgCannotBeEmpty = "cannot be empty"

	// TestMsgInvalidVariableName is a common validation error for variable names.
	TestMsgInvalidVariableName = "Invalid variable name"
)

// Template helper test constants for reducing string duplication in template tests.
const (
	// Test organization and repository names for template data tests.
	TestOrgName  = "test-org"
	TestRepoName = "test-repo"
	MyOrgName    = "my-org"
	MyRepoName   = "my-repo"
	RepoName     = "repo"

	// Config test organization and repository names for RepoOverrides tests.
	OrgName         = "org"
	ExistingOrgName = "existing"
	NewOrgName      = "new"
	OrgRepo         = "org/repo"
	ExistingRepo    = "existing/repo"
	NewRepo         = "new/repo"

	// Analyzer fixture path for template helper tests.
	AnalyzerFixturePath = "../../testdata/analyzer/"
)

// Config fixture path constants for reducing string duplication.
const (
	// Global configs.
	TestConfigGlobalDefault = "configs/global-config-default.yml"
	//nolint:gosec // G101: False positive - this is a test fixture path, not a credential
	TestConfigGlobalBaseToken    = "configs/global-base-token.yml"
	TestConfigRepoGitHub         = "configs/repo-config-github.yml"
	TestConfigRepoSimple         = "configs/repo-config-simple.yml"
	TestConfigActionProfessional = "configs/action-config-professional.yml"
	TestConfigActionSimple       = "configs/action-config-simple.yml"
	TestConfigRepoVerbose        = "configs/repo-config-verbose.yml"
	TestConfigGitHubVerbose      = "configs/github-verbose-simple.yml"
	TestConfigProfessionalQuiet  = "configs/professional-quiet.yml"
	TestConfigMinimalTheme       = "configs/config-minimal-theme.yml"
	TestConfigMinimalSimple      = "configs/minimal-simple.yml"
	TestConfigProfessionalSimple = "configs/professional-simple.yml"
	TestConfigMinimalDist        = "configs/minimal-dist.yml"

	// Invalid/error configs.
	TestConfigInvalidMalformed  = "configs/invalid-config-malformed.yml"
	TestConfigInvalidIncomplete = "configs/invalid-config-incomplete.yml"
	TestConfigInvalidTheme      = "configs/invalid-config-nonexistent-theme.yml"

	// Template fixtures.
	TestTemplateBroken = "template-fixtures/broken-template.tmpl"
)
