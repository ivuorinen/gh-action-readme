package main

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// Shared assertion format strings (kept local to avoid duplicate string
// literals across this file per the no-constant-duplication rule).
const (
	covDepsErrCount       = "got %d results, want %d"
	covDepsErrUnexpectErr = "unexpected error: %v"
)

// covDepsEOFReader is an InputReader that always reports EOF, exercising the
// applyUpdates branch that treats a closed/empty stdin as the safe "N" default.
type covDepsEOFReader struct{}

func (covDepsEOFReader) ReadLine() (string, error) { return "", io.EOF }

// covDepsNewUpgradeCmd builds a cobra command with the flags depsUpgradeHandler
// reads, mirroring the construction used by the production newDepsCmd wiring.
func covDepsNewUpgradeCmd(use string) *cobra.Command {
	cmd := &cobra.Command{Use: use}
	cmd.Flags().Bool(appconstants.FlagCI, false, "")
	cmd.Flags().Bool(appconstants.InputAll, false, "")
	cmd.Flags().Bool(appconstants.InputDryRun, false, "")

	return cmd
}

// TestCovDepsUpgradeHandlerBody drives depsUpgradeHandler past setup (token +
// discovered action files) so the handler body runs: showUpgradeMode,
// collectAllUpdates, and the "no updates needed" / dry-run / pin paths. Without a
// working GitHub token the outdated check yields no updates, so every path
// returns nil without modifying files.
func TestCovDepsUpgradeHandlerBody(t *testing.T) {
	// No t.Parallel(): these subtests mutate the shared globalConfig.
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()

	tests := []struct {
		name    string
		use     string
		fixture string
		setFlag string // bool flag to enable, or "" for none
	}{
		{
			name:    "dry-run mode default upgrade",
			use:     testutil.TestCmdUpgrade,
			fixture: testutil.TestFixtureCompositeWithDeps,
			setFlag: appconstants.InputDryRun,
		},
		{
			name:    "pin command floating deps",
			use:     appconstants.CommandPin,
			fixture: testutil.TestFixtureCompositeWithDeps,
			setFlag: "",
		},
		{
			name:    "all flag with no dependencies",
			use:     testutil.TestCmdUpgrade,
			fixture: testutil.TestFixtureJavaScriptSimple,
			setFlag: appconstants.InputAll,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// CreateDependencyAnalyzer resolves the git repo root from cwd, so the
			// working dir must be an initialized repository once we chdir into it.
			testutil.InitGitRepo(t, tmpDir)
			testutil.WriteActionFixture(t, tmpDir, tt.fixture)
			t.Chdir(tmpDir)

			globalConfig = internal.DefaultAppConfig()
			globalConfig.Quiet = true
			globalConfig.GitHubToken = testutil.TestToken123

			cmd := covDepsNewUpgradeCmd(tt.use)
			if tt.setFlag != "" {
				if err := cmd.Flags().Set(tt.setFlag, "true"); err != nil {
					t.Fatalf("failed to set flag %q: %v", tt.setFlag, err)
				}
			}

			err := depsUpgradeHandler(cmd, []string{})
			if err != nil {
				t.Errorf(covDepsErrUnexpectErr, err)
			}
		})
	}
}

// TestCovDepsCollectAllUpdatesInvalidFile covers the AnalyzeActionFile error
// branch of collectAllUpdates (invalid YAML -> warning -> continue -> no updates).
func TestCovDepsCollectAllUpdatesInvalidFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	actionFile := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	testutil.WriteTestFile(t, actionFile, testutil.TestInvalidYAMLPrefix)

	output := createOutputManager(true)
	analyzer := &dependencies.Analyzer{} // no GitHub client

	updates := collectAllUpdates(output, analyzer, []string{actionFile})
	if len(updates) != 0 {
		t.Errorf(covDepsErrCount, len(updates), 0)
	}
}

// TestCovDepsCheckAllOutdated exercises checkAllOutdated directly for both the
// AnalyzeActionFile error branch (invalid file) and the normal composite path
// (CheckOutdated returns nothing without a usable GitHub client).
func TestCovDepsCheckAllOutdated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) string
	}{
		{
			name: "invalid yaml is skipped with warning",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				actionFile := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
				testutil.WriteTestFile(t, actionFile, testutil.TestInvalidYAMLPrefix)

				return actionFile
			},
		},
		{
			name: "valid composite yields no outdated without client",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()

				return testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureCompositeWithDeps)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			actionFile := tt.setupFunc(t, tmpDir)

			output := createOutputManager(true)
			analyzer := &dependencies.Analyzer{} // no GitHub client

			outdated := checkAllOutdated(output, []string{actionFile}, analyzer)
			if len(outdated) != 0 {
				t.Errorf(covDepsErrCount, len(outdated), 0)
			}
		})
	}
}

// TestCovDepsApplyUpdatesBranches covers the applyUpdates branches not exercised
// by the existing cancellation/automatic tests: EOF-as-cancel, the interactive
// "yes" path that applies, and the automatic apply success path. Empty update
// lists make ApplyPinnedUpdates a no-op, so the success branches return nil.
func TestCovDepsApplyUpdatesBranches(t *testing.T) {
	t.Parallel()

	output := createOutputManager(true)
	analyzer := dependencies.NewAnalyzer(nil, git.RepoInfo{}, nil)

	t.Run("EOF on prompt cancels without error", func(t *testing.T) {
		t.Parallel()

		updates := []dependencies.PinnedUpdate{
			{OldUses: testutil.TestActionCheckoutV3, NewUses: testutil.TestActionCheckoutV4},
		}
		if err := applyUpdates(output, analyzer, updates, false, covDepsEOFReader{}); err != nil {
			t.Errorf("applyUpdates() EOF should cancel without error, got: %v", err)
		}
	})

	t.Run("interactive yes applies successfully", func(t *testing.T) {
		t.Parallel()

		reader := &TestInputReader{responses: []string{appconstants.InputYes}}
		if err := applyUpdates(output, analyzer, nil, false, reader); err != nil {
			t.Errorf(covDepsErrUnexpectErr, err)
		}
		if reader.index != 1 {
			t.Errorf(testutil.TestErrInputReaderNotUsed, reader.index)
		}
	})

	t.Run("automatic applies successfully", func(t *testing.T) {
		t.Parallel()

		if err := applyUpdates(output, analyzer, nil, true, nil); err != nil {
			t.Errorf(covDepsErrUnexpectErr, err)
		}
	})
}
