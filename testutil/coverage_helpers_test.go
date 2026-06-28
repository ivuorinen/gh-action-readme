package testutil

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// Local constants to avoid duplicate string literals across these tests.
const (
	covActionContent = "name: Coverage Action\ndescription: cov\n"
	covFileName      = "cov.yml"
	covWarningText   = "deprecated input detected"
	covErrText       = "boom"
)

// covFakeCache implements the interface required by CleanupCache.
type covFakeCache struct {
	closed bool
}

func (f *covFakeCache) Close() error {
	f.closed = true

	return nil
}

// skipIfNoGit skips the test when the git binary is unavailable.
func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath(appconstants.GitCommand); err != nil {
		t.Skip("git binary not available")
	}
}

func TestCovWriteFileInDir(t *testing.T) {
	dir := t.TempDir()
	path := WriteFileInDir(t, dir, covFileName, covActionContent)
	if path != filepath.Join(dir, covFileName) {
		t.Fatalf("unexpected path: %q", path)
	}
	AssertFileExists(t, path)
}

func TestCovWriteActionFixture(t *testing.T) {
	dir := t.TempDir()
	path := WriteActionFixture(t, dir, TestFixtureJavaScriptSimple)
	if filepath.Base(path) != appconstants.ActionFileNameYML {
		t.Fatalf("expected action.yml, got %q", path)
	}
	AssertFileExists(t, path)
}

func TestCovWriteActionFixtureAs(t *testing.T) {
	dir := t.TempDir()
	path := WriteActionFixtureAs(t, dir, covFileName, TestFixtureJavaScriptSimple)
	if filepath.Base(path) != covFileName {
		t.Fatalf("expected %q, got %q", covFileName, path)
	}
	AssertFileExists(t, path)
}

func TestCovCreateActionInTempDir(t *testing.T) {
	tmpDir, actionPath := CreateActionInTempDir(t, covActionContent)
	if tmpDir == "" {
		t.Fatal("expected non-empty tmpDir")
	}
	AssertFileExists(t, actionPath)
}

func TestCovCreateNestedAction(t *testing.T) {
	base := t.TempDir()
	dirPath, actionPath := CreateNestedAction(t, base, filepath.Join("actions", "build"), covActionContent)
	if dirPath != filepath.Join(base, "actions", "build") {
		t.Fatalf("unexpected dirPath: %q", dirPath)
	}
	AssertFileExists(t, actionPath)
}

func TestCovCreateTestSubdir(t *testing.T) {
	base := t.TempDir()
	sub := CreateTestSubdir(t, base, "a", "b")
	if sub != filepath.Join(base, "a", "b") {
		t.Fatalf("unexpected subdir: %q", sub)
	}
	if info, err := os.Stat(sub); err != nil || !info.IsDir() {
		t.Fatalf("expected directory at %q: %v", sub, err)
	}
}

func TestCovCreateActionSubdir(t *testing.T) {
	base := t.TempDir()
	path := CreateActionSubdir(t, base, "nested", TestFixtureCompositeBasic)
	AssertFileExists(t, path)
}

func TestCovAssertSliceContainsAll(t *testing.T) {
	// Success path: every expected substring is present.
	AssertSliceContainsAll(t, []string{ValidationHelloWorld, "foo bar"}, []string{"world", "foo"})
}

func TestCovValidateTestPath(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sub", "file.txt")
	got := ValidateTestPath(t, path, root)
	if got != path {
		t.Fatalf("expected %q, got %q", path, got)
	}
}

func TestCovSafeReadFile(t *testing.T) {
	dir := t.TempDir()
	path := WriteFileInDir(t, dir, covFileName, covActionContent)
	got := SafeReadFile(t, path, dir)
	if string(got) != covActionContent {
		t.Fatalf("unexpected content: %q", string(got))
	}
}

func TestCovCreateTempActionFile(t *testing.T) {
	path := CreateTempActionFile(t, covActionContent)
	AssertFileExists(t, path)
	got := SafeReadFile(t, path, filepath.Dir(path))
	if string(got) != covActionContent {
		t.Fatalf("unexpected content: %q", string(got))
	}
}

func TestCovWriteConfigFile(t *testing.T) {
	base := t.TempDir()
	path := WriteConfigFile(t, base, "theme: default\n")
	AssertFileExists(t, path)
}

func TestCovInitGitRepo(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	InitGitRepo(t, dir)
	if info, err := os.Stat(filepath.Join(dir, appconstants.DirGit)); err != nil || !info.IsDir() {
		t.Fatalf("expected .git directory: %v", err)
	}
}

func TestCovCreateGitRepoWithRemote(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	configPath := CreateGitRepoWithRemote(t, dir, TestURLGitHubMyorgMyrepo)
	AssertFileExists(t, configPath)
	got := SafeReadFile(t, configPath, filepath.Join(dir, appconstants.DirGit))
	if string(got) == "" {
		t.Fatal("expected non-empty git config")
	}
}

func TestCovSetupXDGEnv(t *testing.T) {
	xdg := t.TempDir()
	home := t.TempDir()
	SetupXDGEnv(t, xdg, home)
	if os.Getenv(EnvVarXDGConfigHome) != xdg {
		t.Fatalf("XDG_CONFIG_HOME not set: %q", os.Getenv(EnvVarXDGConfigHome))
	}
	if os.Getenv(EnvVarHOME) != home {
		t.Fatalf("HOME not set: %q", os.Getenv(EnvVarHOME))
	}
}

func TestCovSetupTokenEnv(t *testing.T) {
	SetupTokenEnv(t, "tool-tok", appconstants.TokenFallback)
	if os.Getenv(appconstants.EnvGitHubToken) != "tool-tok" {
		t.Fatalf("tool token not set: %q", os.Getenv(appconstants.EnvGitHubToken))
	}
	if os.Getenv(appconstants.EnvGitHubTokenStandard) != appconstants.TokenFallback {
		t.Fatalf("standard token not set: %q", os.Getenv(appconstants.EnvGitHubTokenStandard))
	}
}

func TestCovCleanupCache(t *testing.T) {
	cache := &covFakeCache{}
	cleanup := CleanupCache(t, cache)
	cleanup()
	if !cache.closed {
		t.Fatal("expected cache to be closed")
	}
}

func TestCovGetGitHubTokenHierarchyTests(t *testing.T) {
	cases := GetGitHubTokenHierarchyTests()
	if len(cases) != 3 {
		t.Fatalf("expected 3 cases, got %d", len(cases))
	}
	for _, tc := range cases {
		if tc.Name == "" || tc.SetupFunc == nil {
			t.Fatalf("invalid case: %+v", tc)
		}
	}
}

func TestCovAssertMessageCounts(t *testing.T) {
	out := &CapturedOutput{
		InfoMessages:    []string{"i1", "i2"},
		ErrorMessages:   []string{"e1"},
		WarningMessages: []string{"w1"},
		BoldMessages:    []string{"b1"},
	}
	AssertMessageCounts(t, "cov", out, 2, 1, 1, 1)
}

func TestCovContainsWarning(t *testing.T) {
	out := &CapturedOutput{WarningMessages: []string{covWarningText}}
	if !out.ContainsWarning("deprecated") {
		t.Fatal("expected ContainsWarning to find substring")
	}
	if out.ContainsWarning("absent-needle") {
		t.Fatal("expected ContainsWarning to return false for missing needle")
	}
}

func TestCovGitHelpers(t *testing.T) {
	base := t.TempDir()
	gitDir := SetupGitDirectory(t, base)
	if filepath.Base(gitDir) != appconstants.DirGit {
		t.Fatalf("unexpected git dir: %q", gitDir)
	}
	cfg := CreateGitConfigWithRemote(t, gitDir, TestURLGitHubMyorgMyrepo, "main")
	AssertFileExists(t, cfg)

	wgPath := WriteGitConfigFile(t, t.TempDir(), "[core]\n")
	AssertFileExists(t, wgPath)
}

func TestCovMockOutputBasics(t *testing.T) {
	c := &CapturedOutput{}

	c.Fprintf(os.Stdout, "hello %s", "world")
	if len(c.PrintfMessages) != 1 || c.PrintfMessages[0] != ValidationHelloWorld {
		t.Fatalf("unexpected Fprintf record: %v", c.PrintfMessages)
	}

	c.QuietMode = true
	if !c.IsQuiet() {
		t.Fatal("expected IsQuiet true")
	}
}

func TestCovMockOutputErrors(t *testing.T) {
	c := &CapturedOutput{}
	err := errors.New(covErrText)

	c.ErrorWithSuggestions(err)
	c.ErrorWithSuggestions(nil) // nil should not be recorded
	if len(c.ErrorWithSuggestionsCalls) != 1 {
		t.Fatalf("expected 1 suggestion call, got %d", len(c.ErrorWithSuggestionsCalls))
	}

	c.ErrorWithContext(appconstants.ErrCodeInvalidYAML, "ctx-msg", nil)
	if len(c.ErrorWithContextCalls) != 1 || c.ErrorWithContextCalls[0] != "ctx-msg" {
		t.Fatalf("unexpected ErrorWithContext: %v", c.ErrorWithContextCalls)
	}

	c.ErrorWithSimpleFix("problem", "fix-it")
	if len(c.ErrorWithSimpleFixCalls) != 1 || c.ErrorWithSimpleFixCalls[0] != "problem: fix-it" {
		t.Fatalf("unexpected ErrorWithSimpleFix: %v", c.ErrorWithSimpleFixCalls)
	}

	if got := c.FormatContextualError(err); got != covErrText {
		t.Fatalf("expected %q, got %q", covErrText, got)
	}
	if got := c.FormatContextualError(nil); got != "" {
		t.Fatalf("expected empty string for nil error, got %q", got)
	}
}

func TestCovErrorFormatterMock(t *testing.T) {
	m := &ErrorFormatterMock{}
	if got := m.FormatContextualError(errors.New(covErrText)); got != covErrText {
		t.Fatalf("expected %q, got %q", covErrText, got)
	}
	if got := m.FormatContextualError(nil); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
	if len(m.FormatContextualErrorCalls) != 1 {
		t.Fatalf("expected 1 recorded call, got %d", len(m.FormatContextualErrorCalls))
	}
}
