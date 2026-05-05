package dependencies

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testUpdaterRepo  = "repo"
	testUpdaterOwner = "test"
)

// newTestAnalyzer creates an Analyzer with cache for testing.
// Returns the analyzer and a cleanup function.
// Pattern used 7+ times in updater_test.go.
func newTestAnalyzer(t *testing.T) (*Analyzer, func()) {
	t.Helper()

	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
	testutil.AssertNoError(t, err)

	analyzer := &Analyzer{
		Cache: NewCacheAdapter(cacheInstance),
	}

	return analyzer, testutil.CleanupCache(t, cacheInstance)
}

// validatePinnedUpdateSuccess validates that the update succeeded and backup was cleaned up.
func validatePinnedUpdateSuccess(t *testing.T, actionPath string, validateBackup bool, analyzer *Analyzer) {
	t.Helper()

	if validateBackup {
		testutil.AssertBackupNotExists(t, actionPath)
	}

	// Verify file is still valid YAML
	err := analyzer.validateActionFile(actionPath)
	testutil.AssertNoError(t, err)
}

// validatePinnedUpdateRollback validates that the rollback succeeded and file is unchanged.
func validatePinnedUpdateRollback(t *testing.T, actionPath, originalContent string) {
	t.Helper()

	testutil.ValidateRollback(t, actionPath, originalContent)

	// Backup should be removed after rollback
	testutil.AssertBackupNotExists(t, actionPath)
}

// TestApplyPinnedUpdates tests the ApplyPinnedUpdates method.
// Note: These tests identify a bug where the `- ` list marker is not preserved
// when updating YAML. The current implementation replaces entire lines with
// just "uses: " prefix, losing the list marker. Tests are written to document
// current behavior while validating the logic works.
func TestApplyPinnedUpdates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		actionContent  string
		updates        []PinnedUpdate
		wantErr        bool
		validateBackup bool
		checkRollback  bool
	}{
		createSingleUpdateTestCase(singleUpdateParams{
			name:           "list format updates now work correctly (bug fixed)",
			fixturePath:    "dependencies/simple-list-step.yml",
			oldUses:        testutil.TestCheckoutV4OldUses,
			newUses:        testutil.TestCheckoutPinnedV417,
			commitSHA:      testutil.TestActionCheckoutSHA,
			version:        testutil.TestVersionV417,
			updateType:     appconstants.UpdateTypePatch,
			wantErr:        false,
			validateBackup: true,
			checkRollback:  false,
		}),
		createSingleUpdateTestCase(singleUpdateParams{
			name:           "updates work when uses is not in list format",
			fixturePath:    "dependencies/named-step.yml",
			oldUses:        testutil.TestCheckoutV4OldUses,
			newUses:        testutil.TestCheckoutPinnedV417,
			commitSHA:      testutil.TestActionCheckoutSHA,
			version:        testutil.TestVersionV417,
			updateType:     appconstants.UpdateTypePatch,
			wantErr:        false,
			validateBackup: true,
			checkRollback:  false,
		}),
		{
			name:          "multiple updates in non-list format",
			actionContent: testutil.MustReadFixture("dependencies/multiple-steps.yml"),
			updates: []PinnedUpdate{
				{
					FilePath:   "", // Will be set by test
					OldUses:    testutil.TestCheckoutV4OldUses,
					NewUses:    testutil.TestCheckoutPinnedV417,
					CommitSHA:  testutil.TestActionCheckoutSHA,
					Version:    testutil.TestVersionV417,
					UpdateType: appconstants.UpdateTypePatch,
					LineNumber: 0,
				},
				{
					FilePath:   "", // Will be set by test
					OldUses:    testutil.TestActionSetupNodeV3,
					NewUses:    "actions/setup-node@1a4e6d7c9f8e5b2a3c4d5e6f7a8b9c0d1e2f3a4b # v4.0.0",
					CommitSHA:  "1a4e6d7c9f8e5b2a3c4d5e6f7a8b9c0d1e2f3a4b",
					Version:    "v4.0.0",
					UpdateType: appconstants.UpdateTypeMajor,
					LineNumber: 0,
				},
			},
			wantErr:        false,
			validateBackup: true,
			checkRollback:  false,
		},
		createSingleUpdateTestCase(singleUpdateParams{
			name:           "preserves indentation in non-list format",
			fixturePath:    "dependencies/step-with-parameters.yml",
			oldUses:        testutil.TestCheckoutV4OldUses,
			newUses:        testutil.TestCheckoutPinnedV417,
			commitSHA:      testutil.TestActionCheckoutSHA,
			version:        testutil.TestVersionV417,
			updateType:     appconstants.UpdateTypePatch,
			wantErr:        false,
			validateBackup: true,
			checkRollback:  false,
		}),
		createSingleUpdateTestCase(singleUpdateParams{
			name:           "handles already pinned dependencies",
			fixturePath:    "dependencies/already-pinned.yml",
			oldUses:        testutil.TestCheckoutPinnedV417,
			newUses:        testutil.TestCheckoutPinnedV417,
			commitSHA:      testutil.TestActionCheckoutSHA,
			version:        testutil.TestVersionV417,
			updateType:     appconstants.UpdateTypeNone,
			wantErr:        false,
			validateBackup: true,
			checkRollback:  false,
		}),
		{
			name:          "invalid YAML triggers rollback",
			actionContent: testutil.MustReadFixture("dependencies/simple-test-step.yml"),
			updates: []PinnedUpdate{
				{
					FilePath:   "", // Will be set by test
					OldUses:    "actions/checkout@v4",
					NewUses:    "{unclosed",
					CommitSHA:  "",
					Version:    "",
					UpdateType: appconstants.UpdateTypeNone,
					LineNumber: 0,
				},
			},
			wantErr:        true,
			validateBackup: false,
			checkRollback:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary directory and action file
			dir, cleanup := testutil.TempDir(t)
			defer cleanup()

			actionPath := testutil.WriteActionFile(t, dir, tt.actionContent)

			// Store original content for rollback check
			originalContent, _ := os.ReadFile(actionPath) // #nosec G304 -- test file path

			// Set file path in updates
			for i := range tt.updates {
				tt.updates[i].FilePath = actionPath
			}

			// Create analyzer
			analyzer, cleanupAnalyzer := newTestAnalyzer(t)
			defer cleanupAnalyzer()

			// Apply updates
			err := analyzer.ApplyPinnedUpdates(tt.updates)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyPinnedUpdates() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr {
				validatePinnedUpdateSuccess(t, actionPath, tt.validateBackup, analyzer)
			}

			if tt.checkRollback {
				validatePinnedUpdateRollback(t, actionPath, string(originalContent))
			}
		})
	}
}

// validateUpdateFileSuccess validates that the file was updated correctly and backup was cleaned up.
func validateUpdateFileSuccess(t *testing.T, actionPath, expectedYAML string, checkBackup bool) {
	t.Helper()

	testutil.AssertFileContentEquals(t, actionPath, expectedYAML)

	if checkBackup {
		testutil.AssertBackupNotExists(t, actionPath)
	}
}

// validateUpdateFileRollback validates that the rollback succeeded and file is unchanged.
func validateUpdateFileRollback(t *testing.T, actionPath, initialYAML string) {
	t.Helper()

	testutil.AssertFileContentEquals(t, actionPath, initialYAML)
}

// TestUpdateActionFile tests the updateActionFile method directly.
func TestUpdateActionFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		initialYAML   string
		updates       []PinnedUpdate
		expectedYAML  string
		expectError   bool
		checkBackup   bool
		rollbackCheck bool
	}{
		{
			name:        "finds and replaces uses statement in non-list format",
			initialYAML: testutil.MustReadFixture("dependencies/test-checkout-v4.yml"),
			updates: []PinnedUpdate{
				{
					OldUses: testutil.TestCheckoutV4OldUses,
					NewUses: testutil.TestCheckoutPinnedV411,
				},
			},
			expectedYAML: testutil.MustReadFixture("dependencies/test-checkout-pinned.yml"),
			expectError:  false,
			checkBackup:  true,
		},
		{
			name:        "handles different version formats",
			initialYAML: testutil.MustReadFixture("dependencies/test-checkout-v4-1-0.yml"),
			updates: []PinnedUpdate{
				{
					OldUses: "actions/checkout@v4.1.0",
					NewUses: testutil.TestCheckoutPinnedV411,
				},
			},
			expectedYAML: testutil.MustReadFixture("dependencies/test-checkout-pinned.yml"),
			expectError:  false,
			checkBackup:  true,
		},
		{
			name:        "handles multiple references to same action",
			initialYAML: testutil.MustReadFixture("dependencies/test-multiple-checkout.yml"),
			updates: []PinnedUpdate{
				{
					OldUses: testutil.TestCheckoutV4OldUses,
					NewUses: testutil.TestCheckoutPinnedV411,
				},
			},
			expectedYAML: testutil.MustReadFixture("dependencies/test-multiple-checkout-pinned.yml"),
			expectError:  false,
			checkBackup:  true,
		},
		{
			name:        "preserves whitespace and comments",
			initialYAML: testutil.MustReadFixture("dependencies/test-checkout-with-comment.yml"),
			updates: []PinnedUpdate{
				{
					OldUses: testutil.TestCheckoutV4OldUses,
					NewUses: testutil.TestCheckoutPinnedV411,
				},
			},
			expectedYAML: testutil.MustReadFixture("dependencies/test-checkout-with-comment-pinned.yml"),
			expectError:  false,
			checkBackup:  true,
		},
		{
			name:        "invalid YAML triggers rollback",
			initialYAML: testutil.MustReadFixture(testutil.TestFixtureSimpleCheckout),
			updates: []PinnedUpdate{
				{
					OldUses: testutil.TestCheckoutV4OldUses,
					NewUses: "\"unclosed string that breaks YAML parsing", // Unclosed quote breaks YAML
				},
			},
			expectedYAML:  "", // Should rollback to original
			expectError:   true,
			checkBackup:   false,
			rollbackCheck: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory and file
			dir, cleanup := testutil.TempDir(t)
			defer cleanup()

			actionPath := testutil.WriteActionFile(t, dir, tt.initialYAML)

			// Create analyzer
			analyzer, cleanupAnalyzer := newTestAnalyzer(t)
			defer cleanupAnalyzer()

			// Apply update
			err := analyzer.updateActionFile(actionPath, tt.updates)

			// Check error expectation
			if (err != nil) != tt.expectError {
				t.Errorf("updateActionFile() error = %v, expectError %v", err, tt.expectError)

				return
			}

			if !tt.expectError {
				validateUpdateFileSuccess(t, actionPath, tt.expectedYAML, tt.checkBackup)
			}

			if tt.rollbackCheck {
				validateUpdateFileRollback(t, actionPath, tt.initialYAML)
			}
		})
	}
}

// TestValidateActionFile tests the validateActionFile method.
func TestValidateActionFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yamlContent string
		expectValid bool
	}{
		{
			name:        "valid composite action",
			yamlContent: testutil.MustReadFixture("dependencies/simple-list-step.yml"),
			expectValid: true,
		},
		{
			name:        "valid JavaScript action",
			yamlContent: testutil.MustReadFixture("dependencies/valid-javascript-action.yml"),
			expectValid: true,
		},
		{
			name:        "valid Docker action",
			yamlContent: testutil.MustReadFixture("dependencies/valid-docker-action.yml"),
			expectValid: true,
		},
		{
			name:        "missing name field",
			yamlContent: testutil.MustReadFixture("dependencies/missing-name.yml"),
			expectValid: false,
		},
		{
			name:        "missing description field",
			yamlContent: testutil.MustReadFixture("dependencies/missing-description.yml"),
			expectValid: false,
		},
		{
			name:        "missing runs field",
			yamlContent: testutil.MustReadFixture("dependencies/missing-runs.yml"),
			expectValid: false,
		},
		{
			name:        "invalid YAML syntax",
			yamlContent: testutil.MustReadFixture("dependencies/invalid-syntax.yml"),
			expectValid: false,
		},
		{
			name:        "invalid using field",
			yamlContent: testutil.MustReadFixture("dependencies/invalid-using.yml"),
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file
			dir, cleanup := testutil.TempDir(t)
			defer cleanup()

			actionPath := testutil.WriteActionFile(t, dir, tt.yamlContent)

			// Create analyzer
			analyzer, cleanupAnalyzer := newTestAnalyzer(t)
			defer cleanupAnalyzer()

			// Validate
			err := analyzer.validateActionFile(actionPath)

			if tt.expectValid && err != nil {
				t.Errorf("validateActionFile() expected valid but got error: %v", err)
			}

			if !tt.expectValid && err == nil {
				t.Errorf("validateActionFile() expected invalid but got nil error")
			}
		})
	}
}

// TestGetLatestTagEdgeCases tests edge cases for getLatestTag.
func TestGetLatestTagEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockSetup   func() *Analyzer
		owner       string
		repo        string
		expectError bool
	}{
		{
			name: "no tags available",
			mockSetup: func() *Analyzer {
				mockClient := testutil.MockGitHubClient(map[string]string{
					"GET https://api.github.com/repos/" + testUpdaterOwner + "/" + testUpdaterRepo + "/tags": "[]",
				})
				cacheInstance, _ := cache.NewCache(cache.DefaultConfig())

				return &Analyzer{
					GitHubClient: mockClient,
					Cache:        NewCacheAdapter(cacheInstance),
				}
			},
			owner:       testUpdaterOwner,
			repo:        testUpdaterRepo,
			expectError: true,
		},
		{
			name: "GitHub client nil",
			mockSetup: func() *Analyzer {
				return &Analyzer{
					GitHubClient: nil,
					Cache:        nil,
				}
			},
			owner:       testUpdaterOwner,
			repo:        testUpdaterRepo,
			expectError: true,
		},
		{
			name: "malformed tag response",
			mockSetup: func() *Analyzer {
				mockClient := testutil.MockGitHubClient(map[string]string{
					"GET https://api.github.com/repos/" + testUpdaterOwner + "/" + testUpdaterRepo + "/tags": "invalid json",
				})
				cacheInstance, _ := cache.NewCache(cache.DefaultConfig())

				return &Analyzer{
					GitHubClient: mockClient,
					Cache:        NewCacheAdapter(cacheInstance),
				}
			},
			owner:       testUpdaterOwner,
			repo:        testUpdaterRepo,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			analyzer := tt.mockSetup()
			if analyzer.Cache != nil {
				// Clean up cache if it exists
				defer func() {
					if ca, ok := analyzer.Cache.(*CacheAdapter); ok {
						_ = ca.cache.Close()
					}
				}()
			}

			_, _, err := analyzer.getLatestVersion(tt.owner, tt.repo)

			if (err != nil) != tt.expectError {
				t.Errorf("getLatestVersion() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// assertCacheVersionNotFound validates that no version was found in the cache.
func assertCacheVersionNotFound(t *testing.T, version, sha string, found bool) {
	t.Helper()

	if found {
		t.Error("getCachedVersion() should return false")
	}
	if version != "" {
		t.Errorf("version = %q, want empty", version)
	}
	if sha != "" {
		t.Errorf("sha = %q, want empty", sha)
	}
}

// TestCacheVersionEdgeCases tests edge cases for cacheVersion and getCachedVersion.
func TestCacheVersionEdgeCases(t *testing.T) {
	t.Parallel()

	// Parametrized tests for getCachedVersion edge cases
	notFoundCases := []struct {
		name     string
		setupFn  func(*testing.T) (*Analyzer, func())
		cacheKey string
	}{
		{
			name: "nil cache",
			setupFn: func(_ *testing.T) (*Analyzer, func()) {
				return &Analyzer{Cache: nil}, func() {
					// No cleanup needed for nil cache
				}
			},
			cacheKey: testutil.CacheTestKey,
		},
		{
			name: "invalid data type",
			setupFn: func(t *testing.T) (*Analyzer, func()) {
				t.Helper()
				c, err := cache.NewCache(cache.DefaultConfig())
				testutil.AssertNoError(t, err)
				_ = c.Set(testutil.CacheTestKey, "invalid-string")

				return &Analyzer{Cache: NewCacheAdapter(c)}, testutil.CleanupCache(t, c)
			},
			cacheKey: testutil.CacheTestKey,
		},
		{
			name: "empty cache entry",
			setupFn: func(t *testing.T) (*Analyzer, func()) {
				t.Helper()
				c, err := cache.NewCache(cache.DefaultConfig())
				testutil.AssertNoError(t, err)

				return &Analyzer{Cache: NewCacheAdapter(c)}, testutil.CleanupCache(t, c)
			},
			cacheKey: "nonexistent-key",
		},
	}

	for _, tc := range notFoundCases {
		t.Run("getCachedVersion with "+tc.name, func(t *testing.T) {
			t.Parallel()
			analyzer, cleanup := tc.setupFn(t)
			defer cleanup()
			version, sha, found := analyzer.getCachedVersion(tc.cacheKey)
			assertCacheVersionNotFound(t, version, sha, found)
		})
	}

	t.Run("cacheVersion with nil cache", func(t *testing.T) {
		t.Parallel()

		analyzer := &Analyzer{Cache: nil}
		// Should not panic
		analyzer.cacheVersion(testutil.CacheTestKey, "v1.0.0", "abc123")
	})

	t.Run("cacheVersion stores and retrieves correctly", func(t *testing.T) {
		t.Parallel()

		cacheInstance, err := cache.NewCache(cache.DefaultConfig())
		testutil.AssertNoError(t, err)
		defer testutil.CleanupCache(t, cacheInstance)()

		analyzer := &Analyzer{Cache: NewCacheAdapter(cacheInstance)}

		// Cache a version
		analyzer.cacheVersion(testutil.CacheTestKey, "v1.2.3", "def456")

		// Retrieve it
		version, sha, found := analyzer.getCachedVersion(testutil.CacheTestKey)

		if !found {
			t.Error("getCachedVersion() should return true after cacheVersion()")
		}
		if version != "v1.2.3" {
			t.Errorf("getCachedVersion() version = %s, want v1.2.3", version)
		}
		if sha != "def456" {
			t.Errorf("getCachedVersion() sha = %s, want def456", sha)
		}
	})
}

// TestUpdateActionFileBackupAndRollback tests backup creation and rollback functionality.
func TestUpdateActionFileBackupAndRollback(t *testing.T) {
	t.Parallel()

	t.Run("backup created before modification", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		defer cleanup()

		originalContent := testutil.MustReadFixture(testutil.TestFixtureSimpleCheckout)
		actionPath := testutil.WriteActionFile(t, dir, originalContent)

		analyzer, cleanupAnalyzer := newTestAnalyzer(t)
		defer cleanupAnalyzer()

		updates := []PinnedUpdate{
			{
				OldUses: testutil.TestCheckoutV4OldUses,
				NewUses: testutil.TestCheckoutPinnedV411,
			},
		}

		err := analyzer.updateActionFile(actionPath, updates)
		testutil.AssertNoError(t, err)

		// Backup should be removed after successful update
		testutil.AssertBackupNotExists(t, actionPath)
	})

	t.Run("rollback on validation failure", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		defer cleanup()

		originalContent := testutil.MustReadFixture(testutil.TestFixtureSimpleCheckout)
		actionPath := testutil.WriteActionFile(t, dir, originalContent)

		analyzer, cleanupAnalyzer := newTestAnalyzer(t)
		defer cleanupAnalyzer()

		// Create an update that produces unparseable YAML (unclosed flow mapping)
		updates := []PinnedUpdate{
			{
				OldUses: "actions/checkout@v4",
				NewUses: "{unclosed",
			},
		}

		err := analyzer.updateActionFile(actionPath, updates)
		if err == nil {
			t.Error("updateActionFile() should return error for invalid YAML")
		}

		// File should be rolled back to original
		testutil.AssertFileContentEquals(t, actionPath, originalContent)

		// Backup should be removed after rollback
		testutil.AssertBackupNotExists(t, actionPath)
	})

	t.Run("file permission errors", func(t *testing.T) {
		// Skip on Windows as permission handling is different
		if runtime.GOOS == "windows" {
			t.Skip("Skipping permission test on Windows")
		}

		dir, cleanup := testutil.TempDir(t)
		defer cleanup()

		actionPath := filepath.Join(dir, appconstants.ActionFileNameYML)
		testutil.WriteTestFile(t, actionPath, "name: Test\ndescription: Test\nruns:\n  using: composite\n  steps: []")

		// Make file read-only
		err := os.Chmod(actionPath, 0444) // #nosec G302 -- intentionally read-only for test
		testutil.AssertNoError(t, err)

		analyzer, cleanupAnalyzer := newTestAnalyzer(t)
		defer cleanupAnalyzer()

		updates := []PinnedUpdate{
			{
				OldUses: "anything",
				NewUses: "something",
			},
		}

		err = analyzer.updateActionFile(actionPath, updates)
		if err == nil {
			t.Error("updateActionFile() should return error for read-only file")
		}
	})
}

// TestApplyPinnedUpdatesGroupedByFile tests updates to multiple files.
func TestApplyPinnedUpdatesGroupedByFile(t *testing.T) {
	t.Parallel()

	dir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create two action files in non-list format (to avoid YAML bug)
	action1Path := filepath.Join(dir, "action1.yml")
	action2Path := filepath.Join(dir, "action2.yml")

	action1Content := testutil.MustReadFixture("dependencies/action1-checkout.yml")
	action2Content := testutil.MustReadFixture("dependencies/action2-setup-node.yml")

	testutil.WriteTestFile(t, action1Path, action1Content)
	testutil.WriteTestFile(t, action2Path, action2Content)

	analyzer, cleanupAnalyzer := newTestAnalyzer(t)
	defer cleanupAnalyzer()

	// Create updates for both files
	updates := []PinnedUpdate{
		{
			FilePath: action1Path,
			OldUses:  testutil.TestCheckoutV4OldUses,
			NewUses:  testutil.TestCheckoutPinnedV411,
		},
		{
			FilePath: action2Path,
			OldUses:  testutil.TestActionSetupNodeV3,
			NewUses:  "actions/setup-node@def456 # v4.0.0",
		},
	}

	err := analyzer.ApplyPinnedUpdates(updates)
	testutil.AssertNoError(t, err)

	// Verify both files were updated
	content1 := testutil.SafeReadFile(t, action1Path, dir)
	if !strings.Contains(string(content1), testutil.TestCheckoutPinnedV411) {
		t.Errorf("action1.yml was not updated correctly, got:\n%s", string(content1))
	}

	content2 := testutil.SafeReadFile(t, action2Path, dir)
	if !strings.Contains(string(content2), "actions/setup-node@def456 # v4.0.0") {
		t.Errorf("action2.yml was not updated correctly, got:\n%s", string(content2))
	}
}
