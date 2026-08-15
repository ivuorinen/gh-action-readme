package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// Local constants keep repeated literals out of the test body (no-constant-duplication rule).
const (
	covGenNonYamlFile    = "notes.txt"
	covGenNotYamlContent = "just some plain text, not yaml"
)

// TestCovGenResolveAndValidateTargetPath exercises resolveAndValidateTargetPath's
// branches: a real directory, a real file, a nonexistent path, and the
// empty-args path that falls back to the current working directory.
func TestCovGenResolveAndValidateTargetPath(t *testing.T) {
	// Not parallel: a subtest uses t.Chdir, which forbids a parallel ancestor.

	t.Run("existing directory resolves and reports IsDir", covGenTargetPathDirectory)
	t.Run("existing file resolves and reports not IsDir", covGenTargetPathFile)
	t.Run("nonexistent path returns does-not-exist error", covGenTargetPathMissing)
	t.Run("empty args falls back to current directory", covGenTargetPathEmptyArgs)
}

func covGenTargetPathDirectory(t *testing.T) {
	t.Parallel()
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	absPath, info, err := resolveAndValidateTargetPath([]string{tmpDir})
	if err != nil {
		t.Fatalf("resolveAndValidateTargetPath() unexpected error: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected info.IsDir() to be true for a directory, got false")
	}
	if !filepath.IsAbs(absPath) {
		t.Errorf("expected absolute path, got %q", absPath)
	}
}

func covGenTargetPathFile(t *testing.T) {
	t.Parallel()
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	filePath := filepath.Join(tmpDir, covGenNonYamlFile)
	testutil.WriteTestFile(t, filePath, covGenNotYamlContent)

	_, info, err := resolveAndValidateTargetPath([]string{filePath})
	if err != nil {
		t.Fatalf("resolveAndValidateTargetPath() unexpected error: %v", err)
	}
	if info.IsDir() {
		t.Errorf("expected info.IsDir() to be false for a regular file, got true")
	}
}

func covGenTargetPathMissing(t *testing.T) {
	t.Parallel()
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	_, _, err := resolveAndValidateTargetPath([]string{filepath.Join(tmpDir, "missing-target")})
	if err == nil {
		t.Fatal("expected an error for a nonexistent path")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' error, got: %v", err)
	}
}

// covGenTargetPathEmptyArgs is not parallel: it uses t.Chdir to control the
// working directory.
func covGenTargetPathEmptyArgs(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	t.Chdir(tmpDir)

	_, info, err := resolveAndValidateTargetPath([]string{})
	if err != nil {
		t.Fatalf("resolveAndValidateTargetPath() unexpected error: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected current-directory fallback to be a directory")
	}
}

// TestCovGenHandlerRejectsNonYamlFile covers genHandler's file branch where a
// real, existing file with a non-YAML extension is rejected upfront.
func TestCovGenHandlerRejectsNonYamlFile(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	filePath := filepath.Join(tmpDir, covGenNonYamlFile)
	testutil.WriteTestFile(t, filePath, covGenNotYamlContent)

	globalConfig = internal.DefaultAppConfig()

	err := genHandler(newGenCmd(), []string{filePath})
	if err == nil {
		t.Fatal("expected an error for a non-YAML file target")
	}
	if !strings.Contains(err.Error(), "must be a YAML file") {
		t.Errorf("expected a 'must be a YAML file' error, got: %v", err)
	}
}

// TestCovGenLoadGenConfigValidationError covers loadGenConfig's branch where the
// loaded configuration fails validation (a repo config with an invalid theme).
func TestCovGenLoadGenConfigValidationError(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	testutil.WriteTestFile(
		t,
		filepath.Join(tmpDir, testutil.TestFileGHReadmeYAML),
		testutil.MustReadFixture(testutil.TestConfigInvalidTheme),
	)

	config, err := loadGenConfig(tmpDir, tmpDir)
	if err == nil {
		t.Fatalf("expected a validation error, got config: %+v", config)
	}
	if !strings.Contains(err.Error(), "configuration validation error") {
		t.Errorf("expected a 'configuration validation error', got: %v", err)
	}
}

// TestCovGenApplyCommandFlagsOutputFlags covers applyCommandFlags' output-dir and
// output (filename) branches, which the existing TestApplyCommandFlags does not.
func TestCovGenApplyCommandFlagsOutputFlags(t *testing.T) {
	t.Parallel()

	const (
		wantDir      = "docs/out"
		wantFilename = "custom-readme.md"
	)

	config := internal.DefaultAppConfig()
	cmd := newGenCmd()
	if err := cmd.Flags().Set(appconstants.FlagOutputDir, wantDir); err != nil {
		t.Fatalf("failed to set %s: %v", appconstants.FlagOutputDir, err)
	}
	if err := cmd.Flags().Set(appconstants.FlagOutput, wantFilename); err != nil {
		t.Fatalf("failed to set %s: %v", appconstants.FlagOutput, err)
	}

	applyCommandFlags(cmd, config)

	if config.OutputDir != wantDir {
		t.Errorf("OutputDir = %q, want %q", config.OutputDir, wantDir)
	}
	if config.OutputFilename != wantFilename {
		t.Errorf("OutputFilename = %q, want %q", config.OutputFilename, wantFilename)
	}
}

// TestCovGenCloseAnalyzer covers closeAnalyzer for both the nil and non-nil
// analyzer branches.
func TestCovGenCloseAnalyzer(t *testing.T) {
	t.Parallel()

	t.Run("nil analyzer is a no-op", func(t *testing.T) {
		t.Parallel()
		closed := false
		closeAnalyzer(nil)
		closed = true
		if !closed {
			t.Error("closeAnalyzer(nil) should return normally")
		}
	})

	t.Run("non-nil analyzer is closed", func(t *testing.T) {
		t.Parallel()
		analyzer := dependencies.NewAnalyzer(nil, git.RepoInfo{}, nil)
		closed := false
		closeAnalyzer(analyzer)
		closed = true
		if !closed {
			t.Error("closeAnalyzer(analyzer) should return normally")
		}
	})
}

// TestCovGenWrapHandlerSuccessRunsHandler verifies that wrapHandlerWithErrorHandling
// actually invokes the wrapped handler on the success path (handler returns nil),
// which does not call os.Exit.
func TestCovGenWrapHandlerSuccessRunsHandler(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()

	globalConfig = internal.DefaultAppConfig()

	ran := false
	wrapped := wrapHandlerWithErrorHandling(func(_ *cobra.Command, _ []string) error {
		ran = true

		return nil
	})

	wrapped(&cobra.Command{}, []string{})

	if !ran {
		t.Error("wrapHandlerWithErrorHandling did not invoke the wrapped handler")
	}
}

// TestCovGenHandleNoFilesFoundError covers handleNoFilesFoundError's three
// branches: nil input, the no-files warning case (returns nil), and the
// passthrough case for unrelated errors.
func TestCovGenHandleNoFilesFoundError(t *testing.T) {
	t.Parallel()

	output := createOutputManager(true)

	t.Run("nil error returns nil", func(t *testing.T) {
		t.Parallel()
		if err := handleNoFilesFoundError(nil, output); err != nil {
			t.Errorf("expected nil, got: %v", err)
		}
	})

	t.Run("no-files error is swallowed with warning", func(t *testing.T) {
		t.Parallel()
		input := errors.New("discovery: " + appconstants.ErrNoActionFilesFound)
		if err := handleNoFilesFoundError(input, output); err != nil {
			t.Errorf("expected no-files error to be swallowed, got: %v", err)
		}
	})

	t.Run("unrelated error passes through", func(t *testing.T) {
		t.Parallel()
		input := errors.New("some other failure")
		err := handleNoFilesFoundError(input, output)
		if !errors.Is(err, input) {
			t.Errorf("expected the original error to pass through, got: %v", err)
		}
	})
}
