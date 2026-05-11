package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// ── containsString ────────────────────────────────────────────────────────────

func TestContainsString(t *testing.T) {
	t.Parallel()

	t.Run("slice contains item", func(t *testing.T) {
		t.Parallel()
		if !containsString([]string{"a", "b", "c"}, "b") {
			t.Error("expected true")
		}
	})

	t.Run("slice missing item", func(t *testing.T) {
		t.Parallel()
		if containsString([]string{"a", "b"}, "z") {
			t.Error("expected false")
		}
	})

	t.Run("string scalar match", func(t *testing.T) {
		t.Parallel()
		if !containsString("hello", "hello") {
			t.Error("expected true")
		}
	})

	t.Run("string scalar mismatch", func(t *testing.T) {
		t.Parallel()
		if containsString("hello", "world") {
			t.Error("expected false")
		}
	})

	t.Run("empty string returns false", func(t *testing.T) {
		t.Parallel()
		if containsString("", "x") {
			t.Error("expected false for empty string")
		}
	})

	t.Run("unsupported type returns false", func(t *testing.T) {
		t.Parallel()
		if containsString(42, "x") {
			t.Error("expected false for unsupported type")
		}
	})
}

// ── getExpectedFilename ───────────────────────────────────────────────────────

func TestGetExpectedFilename(t *testing.T) {
	t.Parallel()

	cases := []struct {
		format string
		want   string
	}{
		{appconstants.OutputFormatMarkdown, appconstants.ReadmeMarkdown},
		{appconstants.OutputFormatHTML, TestPatternHTML},
		{appconstants.OutputFormatJSON, appconstants.ActionDocsJSON},
		{appconstants.OutputFormatASCIIDoc, appconstants.ReadmeASCIIDoc},
		{"unknown-format", appconstants.ReadmeMarkdown},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.format, func(t *testing.T) {
			t.Parallel()
			got := getExpectedFilename(tc.format)
			if got != tc.want {
				t.Errorf("getExpectedFilename(%q) = %q, want %q", tc.format, got, tc.want)
			}
		})
	}
}

// ── DetectGeneratedFiles ──────────────────────────────────────────────────────

// requireDetectedFile writes filename to dir, calls DetectGeneratedFiles, and
// asserts exactly one detected file matching wantFile.
func requireDetectedFile(t *testing.T, dir, format, filename, wantFile string) {
	t.Helper()
	WriteTestFile(t, filepath.Join(dir, filename), "content")
	files := DetectGeneratedFiles(dir, format)
	if len(files) != 1 || files[0] != wantFile {
		t.Errorf("expected [%s], got %v", wantFile, files)
	}
}

func TestDetectGeneratedFiles(t *testing.T) {
	t.Parallel()

	t.Run("nonexistent directory returns empty", func(t *testing.T) {
		t.Parallel()
		files := DetectGeneratedFiles("/nonexistent/path/xyz", appconstants.OutputFormatMarkdown)
		if len(files) != 0 {
			t.Errorf("expected empty slice, got %v", files)
		}
	})

	t.Run("detects README.md for markdown", func(t *testing.T) {
		t.Parallel()
		requireDetectedFile(
			t,
			t.TempDir(),
			appconstants.OutputFormatMarkdown,
			appconstants.ReadmeMarkdown,
			appconstants.ReadmeMarkdown,
		)
	})

	t.Run("detects action-docs.json for json", func(t *testing.T) {
		t.Parallel()
		requireDetectedFile(
			t,
			t.TempDir(),
			appconstants.OutputFormatJSON,
			appconstants.ActionDocsJSON,
			appconstants.ActionDocsJSON,
		)
	})

	t.Run("detects .html file for html format", func(t *testing.T) {
		t.Parallel()
		requireDetectedFile(t, t.TempDir(), appconstants.OutputFormatHTML, "my-action.html", "my-action.html")
	})

	t.Run("detects README.adoc for asciidoc", func(t *testing.T) {
		t.Parallel()
		requireDetectedFile(
			t,
			t.TempDir(),
			appconstants.OutputFormatASCIIDoc,
			appconstants.ReadmeASCIIDoc,
			appconstants.ReadmeASCIIDoc,
		)
	})

	t.Run("skips action.yml file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		WriteTestFile(t, filepath.Join(dir, appconstants.ActionFileNameYML), "name: test")
		requireDetectedFile(
			t,
			dir,
			appconstants.OutputFormatMarkdown,
			appconstants.ReadmeMarkdown,
			appconstants.ReadmeMarkdown,
		)
	})

	t.Run("unknown format defaults to README.md", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		WriteTestFile(t, filepath.Join(dir, appconstants.ReadmeMarkdown), "# Readme")
		files := DetectGeneratedFiles(dir, "unknown")
		if len(files) != 1 {
			t.Errorf("expected 1 file for unknown format, got %v", files)
		}
	})
}

// ── DefaultTestConfig / DefaultMockConfig ─────────────────────────────────────

func TestDefaultTestConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultTestConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.OutputFormat != appconstants.OutputFormatMarkdown {
		t.Errorf("unexpected output format: %q", cfg.OutputFormat)
	}
	if !cfg.Validate {
		t.Error("expected Validate=true by default")
	}
	if cfg.ExtraFlags == nil {
		t.Error("expected non-nil ExtraFlags map")
	}
}

func TestDefaultMockConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultMockConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if !cfg.GitHubClient {
		t.Error("expected GitHubClient=true by default")
	}
	if !cfg.TempDir {
		t.Error("expected TempDir=true by default")
	}
	if !cfg.ColoredOutput {
		t.Error("expected ColoredOutput=true by default")
	}
}

// ── CreateMockSuite ───────────────────────────────────────────────────────────

func TestCreateMockSuite(t *testing.T) {
	t.Parallel()

	t.Run("nil config uses defaults", func(t *testing.T) {
		t.Parallel()
		suite := CreateMockSuite(nil)
		if suite == nil {
			t.Fatal("expected non-nil suite")
		}
		if suite.GitHubClient == nil {
			t.Error("expected GitHub client")
		}
		if suite.ColoredOutput == nil {
			t.Error("expected colored output")
		}
	})

	t.Run("no github client or colored output", func(t *testing.T) {
		t.Parallel()
		suite := CreateMockSuite(&MockConfig{
			GitHubClient:  false,
			ColoredOutput: false,
			Environment:   make(map[string]string),
		})
		if suite.GitHubClient != nil {
			t.Error("expected nil GitHub client")
		}
		if suite.ColoredOutput != nil {
			t.Error("expected nil colored output")
		}
	})

	t.Run("with environment variables", func(t *testing.T) {
		t.Parallel()
		suite := CreateMockSuite(&MockConfig{
			Environment: map[string]string{"KEY": "VALUE"},
		})
		if suite.Environment["KEY"] != "VALUE" {
			t.Error("expected environment variable to be propagated")
		}
	})
}

// ── SetupGitHubMocks ──────────────────────────────────────────────────────────

func TestSetupGitHubMocks(t *testing.T) {
	t.Parallel()

	t.Run("no scenarios returns base responses", func(t *testing.T) {
		t.Parallel()
		responses := SetupGitHubMocks(nil)
		if len(responses) == 0 {
			t.Error("expected non-empty base responses")
		}
	})

	t.Run("rate-limit scenario", func(t *testing.T) {
		t.Parallel()
		responses := SetupGitHubMocks([]string{"rate-limit"})
		if _, ok := responses["GET https://api.github.com/rate_limit"]; !ok {
			t.Error("expected rate_limit response")
		}
	})

	t.Run("not-found scenario", func(t *testing.T) {
		t.Parallel()
		responses := SetupGitHubMocks([]string{"not-found"})
		if _, ok := responses["GET https://api.github.com/repos/nonexistent/repo"]; !ok {
			t.Error("expected not-found response")
		}
	})

	t.Run("latest-release scenario", func(t *testing.T) {
		t.Parallel()
		responses := SetupGitHubMocks([]string{"latest-release"})
		if _, ok := responses["GET https://api.github.com/repos/actions/checkout/releases/latest"]; !ok {
			t.Error("expected latest-release response")
		}
	})
}

// ── CreateGitHubMockSuite ─────────────────────────────────────────────────────

func TestCreateGitHubMockSuite(t *testing.T) {
	t.Parallel()
	suite := CreateGitHubMockSuite(nil)
	if suite == nil {
		t.Fatal("expected non-nil suite")
	}
	if suite.GitHubClient == nil {
		t.Error("expected GitHub client")
	}
}

// ── ValidateActionFixture ─────────────────────────────────────────────────────

func TestValidateActionFixture(t *testing.T) {
	t.Parallel()

	t.Run("valid fixture passes", func(t *testing.T) {
		t.Parallel()
		fixture := &ActionFixture{
			Name:    "test",
			Content: "name: test\ndescription: test",
			IsValid: true,
		}
		ValidateActionFixture(t, fixture)
	})

	t.Run("nil scenario field has no effect", func(t *testing.T) {
		t.Parallel()
		fixture := &ActionFixture{
			Name:     "test",
			Content:  "name: test",
			IsValid:  false,
			Scenario: nil,
		}
		ValidateActionFixture(t, fixture)
	})
}

// ── AssertFixtureValid / AssertFixtureInvalid ─────────────────────────────────

func TestAssertFixtureValid(t *testing.T) {
	t.Parallel()
	AssertFixtureValid(t, TestFixtureCompositeBasic)
}

func TestAssertFixtureInvalid(t *testing.T) {
	t.Parallel()
	AssertFixtureInvalid(t, TestFixtureInvalidMissingDescription)
}

// ── CreateTemporaryAction / CreateTemporaryActionDir ──────────────────────────

func TestCreateTemporaryAction(t *testing.T) {
	t.Parallel()
	path := CreateTemporaryAction(t, TestFixtureCompositeBasic)
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected action file to exist: %v", err)
	}
}

func TestCreateTemporaryActionDir(t *testing.T) {
	t.Parallel()
	dir := CreateTemporaryActionDir(t, TestFixtureCompositeBasic)
	if dir == "" {
		t.Fatal("expected non-empty directory path")
	}
	if _, err := os.Stat(filepath.Join(dir, appconstants.ActionFileNameYML)); err != nil {
		t.Errorf("expected action.yml to exist in dir: %v", err)
	}
}

// ── CreateTestEnvironment ─────────────────────────────────────────────────────

func TestCreateTestEnvironment(t *testing.T) {
	t.Parallel()

	t.Run("nil config creates basic environment", func(t *testing.T) {
		t.Parallel()
		env := CreateTestEnvironment(t, nil)
		if env == nil {
			t.Fatal("expected non-nil environment")
		}
		if env.TempDir == "" {
			t.Error("expected non-empty temp dir")
		}
	})

	t.Run("with action fixture", func(t *testing.T) {
		t.Parallel()
		env := CreateTestEnvironment(t, &EnvironmentConfig{
			ActionFixtures: []string{TestFixtureCompositeBasic},
		})
		if len(env.ActionPaths) != 1 {
			t.Errorf("expected 1 action path, got %d", len(env.ActionPaths))
		}
	})

	t.Run("with config fixture", func(t *testing.T) {
		t.Parallel()
		env := CreateTestEnvironment(t, &EnvironmentConfig{
			ConfigFixture: "default.yml",
		})
		if env.ConfigPath == "" {
			t.Error("expected non-empty config path")
		}
	})

	t.Run("with mocks enabled", func(t *testing.T) {
		t.Parallel()
		env := CreateTestEnvironment(t, &EnvironmentConfig{
			WithMocks: true,
		})
		if env.Mocks == nil {
			t.Error("expected mocks to be configured")
		}
	})
}

// ── RunTestSuite ──────────────────────────────────────────────────────────────

func TestRunTestSuite_Simple(t *testing.T) {
	t.Parallel()
	called := false
	suite := TestSuite{
		Name: "simple-suite",
		Cases: []TestCase{
			{
				Name: "one-case",
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					called = true

					return &TestResult{Success: true, Context: ctx}
				},
			},
		},
	}
	RunTestSuite(t, suite)
	if !called {
		t.Error("expected executor to be called")
	}
}

func TestRunTestSuite_SkippedCase(t *testing.T) {
	t.Parallel()
	suite := TestSuite{
		Name: "suite-with-skip",
		Cases: []TestCase{
			{
				Name:       "skipped-case",
				SkipReason: "intentionally skipped in test",
			},
		},
	}
	RunTestSuite(t, suite)
}

func TestRunTestSuite_WithGlobalSetupAndCleanup(t *testing.T) {
	t.Parallel()
	setupCalled, cleanupCalled := false, false
	suite := TestSuite{
		Name: "suite-global-lifecycle",
		GlobalSetup: func(_ *testing.T) (*TestContext, error) {
			setupCalled = true

			return &TestContext{}, nil
		},
		GlobalCleanup: func(_ *testing.T, _ *TestContext) error {
			cleanupCalled = true

			return nil
		},
		Cases: []TestCase{
			{
				Name: "case-using-global-ctx",
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{Success: true, Context: ctx}
				},
			},
		},
	}
	RunTestSuite(t, suite)
	if !setupCalled {
		t.Error("expected global setup to be called")
	}
	if !cleanupCalled {
		t.Error("expected global cleanup to be called")
	}
}

func TestRunTestSuite_Parallel(t *testing.T) {
	t.Parallel()
	suite := TestSuite{
		Name:     "parallel-suite",
		Parallel: true,
		Cases: []TestCase{
			{
				Name: "parallel-case-1",
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{Success: true, Context: ctx}
				},
			},
			{
				Name: "parallel-case-2",
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{Success: true, Context: ctx}
				},
			},
		},
	}
	RunTestSuite(t, suite)
}

func TestRunTestSuite_CaseSetupAndCleanup(t *testing.T) {
	t.Parallel()
	setupCalled, cleanupCalled := false, false
	suite := TestSuite{
		Name: "suite-case-lifecycle",
		Cases: []TestCase{
			{
				Name: "case-with-lifecycle",
				SetupFunc: func(_ *testing.T, _ *TestContext) error {
					setupCalled = true

					return nil
				},
				CleanupFunc: func(_ *testing.T, _ *TestContext) error {
					cleanupCalled = true

					return nil
				},
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{Success: true, Context: ctx}
				},
			},
		},
	}
	RunTestSuite(t, suite)
	if !setupCalled {
		t.Error("expected case setup to be called")
	}
	if !cleanupCalled {
		t.Error("expected case cleanup to be called")
	}
}

func TestRunTestSuite_WithFixture(t *testing.T) {
	t.Parallel()
	suite := TestSuite{
		Name: "suite-with-fixture",
		Cases: []TestCase{
			{
				Name:    "case-with-fixture",
				Fixture: TestFixtureCompositeBasic,
				Mocks:   &MockConfig{TempDir: true},
			},
		},
	}
	RunTestSuite(t, suite)
}

func TestRunTestSuite_WithTempDirMock(t *testing.T) {
	t.Parallel()
	suite := TestSuite{
		Name: "suite-with-tempdir",
		Cases: []TestCase{
			{
				Name:  "case-with-tempdir",
				Mocks: &MockConfig{TempDir: true},
				Executor: func(t *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					t.Helper()
					if ctx.TempDir == "" {
						t.Error("expected temp dir to be set in context")
					}

					return &TestResult{Success: true, Context: ctx}
				},
			},
		},
	}
	RunTestSuite(t, suite)
}

func TestRunTestSuite_WithExpectations(t *testing.T) {
	t.Parallel()
	suite := TestSuite{
		Name: "suite-with-expectations",
		Cases: []TestCase{
			{
				Name: "success-expectation",
				Expected: &ExpectedResult{
					ShouldSucceed:  true,
					ExpectedOutput: []string{"hello world"},
					ExpectedFiles:  []string{"README.md"},
				},
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{
						Success: true,
						Output:  "hello world",
						Files:   []string{"README.md"},
						Context: ctx,
					}
				},
			},
			{
				Name: "fail-expectation",
				Expected: &ExpectedResult{
					ShouldFail:    true,
					ExpectedError: "something failed",
				},
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{
						Success: false,
						Error:   errors.New("something failed"),
						Context: ctx,
					}
				},
			},
		},
	}
	RunTestSuite(t, suite)
}

func TestRunTestSuite_WithCustomValidation(t *testing.T) {
	t.Parallel()
	customCalled := false
	suite := TestSuite{
		Name: "suite-custom-validation",
		Cases: []TestCase{
			{
				Name: "custom-validator",
				Expected: &ExpectedResult{
					CustomValidation: func(_ *testing.T, _ *TestResult) error {
						customCalled = true

						return nil
					},
				},
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{Success: true, Context: ctx}
				},
			},
		},
	}
	RunTestSuite(t, suite)
	if !customCalled {
		t.Error("expected custom validation to be called")
	}
}

func TestRunTestSuite_WithGlobalConfig(t *testing.T) {
	t.Parallel()
	suite := TestSuite{
		Name: "suite-with-global-config",
		GlobalSetup: func(_ *testing.T) (*TestContext, error) {
			return &TestContext{
				Config: DefaultTestConfig(),
			}, nil
		},
		Cases: []TestCase{
			{
				Name: "inherits-global-config",
				Executor: func(t *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					t.Helper()
					if ctx.Config == nil {
						t.Error("expected config to be inherited from global context")
					}

					return &TestResult{Success: true, Context: ctx}
				},
			},
		},
	}
	RunTestSuite(t, suite)
}

func TestRunTestSuite_WithWildcardFile(t *testing.T) {
	t.Parallel()
	suite := TestSuite{
		Name: "suite-wildcard-file",
		Cases: []TestCase{
			{
				Name: "wildcard-file-match",
				Expected: &ExpectedResult{
					ExpectedFiles: []string{"*.html"},
				},
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{
						Success: true,
						Files:   []string{"my-action.html"},
						Context: ctx,
					}
				},
			},
		},
	}
	RunTestSuite(t, suite)
}

func TestRunTestSuite_WithExitCode(t *testing.T) {
	t.Parallel()
	suite := TestSuite{
		Name: "suite-exit-code",
		Cases: []TestCase{
			{
				Name: "matching-exit-code",
				Expected: &ExpectedResult{
					ExpectedExitCode: 1,
				},
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{
						Success:  false,
						ExitCode: 1,
						Context:  ctx,
					}
				},
			},
		},
	}
	RunTestSuite(t, suite)
}

// ── RunActionTests / RunGeneratorTests / RunValidationTests ───────────────────

func TestRunActionTests(t *testing.T) {
	t.Parallel()
	// Executors run in parallel subtests after the parent returns; verify no panic.
	RunActionTests(t, []ActionTestCase{
		{
			TestCase: TestCase{
				Name: "action-test",
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{Success: true, Context: ctx}
				},
			},
		},
	})
}

func TestRunGeneratorTests(t *testing.T) {
	t.Parallel()
	RunGeneratorTests(t, []GeneratorTestCase{
		{
			TestCase: TestCase{
				Name: "gen-test",
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{Success: true, Context: ctx}
				},
			},
		},
	})
}

func TestRunValidationTests(t *testing.T) {
	t.Parallel()
	RunValidationTests(t, []ValidationTestCase{
		{
			TestCase: TestCase{
				Name: "val-test",
				Executor: func(_ *testing.T, _ TestCase, ctx *TestContext) *TestResult {
					return &TestResult{Success: true, Context: ctx}
				},
			},
		},
	})
}

// ── extractTestCasesGeneric ───────────────────────────────────────────────────

func TestExtractTestCasesGeneric(t *testing.T) {
	t.Parallel()

	t.Run("from ActionTestCase", func(t *testing.T) {
		t.Parallel()
		cases := []ActionTestCase{
			{TestCase: TestCase{Name: "case-a"}},
			{TestCase: TestCase{Name: "case-b"}},
		}
		extracted := extractTestCasesGeneric(cases)
		if len(extracted) != 2 {
			t.Fatalf("expected 2, got %d", len(extracted))
		}
		if extracted[0].Name != "case-a" || extracted[1].Name != "case-b" {
			t.Errorf("unexpected names: %v", extracted)
		}
	})

	t.Run("from GeneratorTestCase", func(t *testing.T) {
		t.Parallel()
		cases := []GeneratorTestCase{
			{TestCase: TestCase{Name: "gen-a"}},
		}
		extracted := extractTestCasesGeneric(cases)
		if len(extracted) != 1 || extracted[0].Name != "gen-a" {
			t.Errorf("unexpected extraction: %v", extracted)
		}
	})

	t.Run("from ValidationTestCase", func(t *testing.T) {
		t.Parallel()
		cases := []ValidationTestCase{
			{TestCase: TestCase{Name: "val-a"}},
		}
		extracted := extractTestCasesGeneric(cases)
		if len(extracted) != 1 || extracted[0].Name != "val-a" {
			t.Errorf("unexpected extraction: %v", extracted)
		}
	})
}

// ── CreateActionTestCases / CreateGeneratorTestCases / CreateValidationTestCases

func TestCreateActionTestCases(t *testing.T) {
	t.Parallel()
	cases := CreateActionTestCases()
	if len(cases) == 0 {
		t.Error("expected non-empty action test cases")
	}
}

func TestCreateGeneratorTestCases(t *testing.T) {
	t.Parallel()
	cases := CreateGeneratorTestCases()
	if len(cases) == 0 {
		t.Error("expected non-empty generator test cases")
	}
}

func TestCreateValidationTestCases(t *testing.T) {
	t.Parallel()
	cases := CreateValidationTestCases()
	if len(cases) == 0 {
		t.Error("expected non-empty validation test cases")
	}
}

// ── TestAllThemes / TestAllFormats / TestValidationScenarios ──────────────────

func TestTestAllThemes(t *testing.T) {
	t.Parallel()
	// Each theme runs as a parallel subtest; assert per-subtest that the theme name is non-empty.
	TestAllThemes(t, func(t *testing.T, theme string) {
		t.Helper()
		if theme == "" {
			t.Error("expected non-empty theme name")
		}
	})
}

func TestTestAllFormats(t *testing.T) {
	t.Parallel()
	TestAllFormats(t, func(t *testing.T, format string) {
		t.Helper()
		if format == "" {
			t.Error("expected non-empty format name")
		}
	})
}

func TestTestValidationScenarios(t *testing.T) {
	t.Parallel()
	// Pass a validator that always returns an error so all invalid-fixture
	// sub-tests pass (TestValidationScenarios expects non-nil errors).
	TestValidationScenarios(t, func(_ *testing.T, fixture string) error {
		return errors.New("validation error for " + fixture)
	})
}
