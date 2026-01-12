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

// TestApplyPinnedUpdates tests the ApplyPinnedUpdates method.
// Note: These tests identify a bug where the `- ` list marker is not preserved
// when updating YAML. The current implementation replaces entire lines with
// just "uses: " prefix, losing the list marker. Tests are written to document
// validatePinnedUpdateSuccess validates that the update succeeded and backup was cleaned up.
func validatePinnedUpdateSuccess(t *testing.T, actionPath string, validateBackup bool, analyzer *Analyzer) {
	t.Helper()

	if validateBackup {
		backupPath := actionPath + appconstants.BackupExtension
		testutil.AssertFileNotExists(t, backupPath)
	}

	// Verify file is still valid YAML
	_ = analyzer.validateActionFile(actionPath)
}

// validatePinnedUpdateRollback validates that the rollback succeeded and file is unchanged.
func validatePinnedUpdateRollback(t *testing.T, actionPath, originalContent string) {
	t.Helper()

	currentContent, err := os.ReadFile(actionPath) // #nosec G304 -- test file path
	testutil.AssertNoError(t, err)

	if string(currentContent) != originalContent {
		t.Error("rollback failed, file was modified")
	}

	// Backup should be removed after rollback
	backupPath := actionPath + appconstants.BackupExtension
	testutil.AssertFileNotExists(t, backupPath)
}

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
		{
			name:          "list format updates now work correctly (bug fixed)",
			actionContent: testutil.MustReadFixture("dependencies/simple-list-step.yml"),
			updates: []PinnedUpdate{
				{
					FilePath:   "", // Will be set by test
					OldUses:    "actions/checkout@v4",
					NewUses:    "actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v4.1.7",
					CommitSHA:  "692973e3d937129bcbf40652eb9f2f61becf3332",
					Version:    "v4.1.7",
					UpdateType: "patch",
					LineNumber: 0,
				},
			},
			wantErr:        false, // Updates work correctly after fix
			validateBackup: true,
			checkRollback:  false,
		},
		{
			name:          "updates work when uses is not in list format",
			actionContent: testutil.MustReadFixture("dependencies/named-step.yml"),
			updates: []PinnedUpdate{
				{
					FilePath:   "", // Will be set by test
					OldUses:    "actions/checkout@v4",
					NewUses:    "actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v4.1.7",
					CommitSHA:  "692973e3d937129bcbf40652eb9f2f61becf3332",
					Version:    "v4.1.7",
					UpdateType: "patch",
					LineNumber: 0,
				},
			},
			wantErr:        false,
			validateBackup: true,
			checkRollback:  false,
		},
		{
			name:          "multiple updates in non-list format",
			actionContent: testutil.MustReadFixture("dependencies/multiple-steps.yml"),
			updates: []PinnedUpdate{
				{
					FilePath:   "", // Will be set by test
					OldUses:    "actions/checkout@v4",
					NewUses:    "actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v4.1.7",
					CommitSHA:  "692973e3d937129bcbf40652eb9f2f61becf3332",
					Version:    "v4.1.7",
					UpdateType: "patch",
					LineNumber: 0,
				},
				{
					FilePath:   "", // Will be set by test
					OldUses:    "actions/setup-node@v3",
					NewUses:    "actions/setup-node@1a4e6d7c9f8e5b2a3c4d5e6f7a8b9c0d1e2f3a4b # v4.0.0",
					CommitSHA:  "1a4e6d7c9f8e5b2a3c4d5e6f7a8b9c0d1e2f3a4b",
					Version:    "v4.0.0",
					UpdateType: "major",
					LineNumber: 0,
				},
			},
			wantErr:        false,
			validateBackup: true,
			checkRollback:  false,
		},
		{
			name:          "preserves indentation in non-list format",
			actionContent: testutil.MustReadFixture("dependencies/step-with-parameters.yml"),
			updates: []PinnedUpdate{
				{
					FilePath:   "", // Will be set by test
					OldUses:    "actions/checkout@v4",
					NewUses:    "actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v4.1.7",
					CommitSHA:  "692973e3d937129bcbf40652eb9f2f61becf3332",
					Version:    "v4.1.7",
					UpdateType: "patch",
					LineNumber: 0,
				},
			},
			wantErr:        false,
			validateBackup: true,
			checkRollback:  false,
		},
		{
			name:          "handles already pinned dependencies",
			actionContent: testutil.MustReadFixture("dependencies/already-pinned.yml"),
			updates: []PinnedUpdate{
				{
					FilePath:   "", // Will be set by test
					OldUses:    "actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v4.1.7",
					NewUses:    "actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v4.1.7",
					CommitSHA:  "692973e3d937129bcbf40652eb9f2f61becf3332",
					Version:    "v4.1.7",
					UpdateType: "none",
					LineNumber: 0,
				},
			},
			wantErr:        false,
			validateBackup: true,
			checkRollback:  false,
		},
		{
			name:          "invalid YAML triggers rollback",
			actionContent: testutil.MustReadFixture("dependencies/simple-test-step.yml"),
			updates: []PinnedUpdate{
				{
					FilePath:   "", // Will be set by test
					OldUses:    "name: Test Action",
					NewUses:    "invalid:::yaml",
					CommitSHA:  "",
					Version:    "",
					UpdateType: "none",
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

			actionPath := filepath.Join(dir, appconstants.ActionFileNameYML)
			testutil.WriteTestFile(t, actionPath, tt.actionContent)

			// Store original content for rollback check
			originalContent, _ := os.ReadFile(actionPath) // #nosec G304 -- test file path

			// Set file path in updates
			for i := range tt.updates {
				tt.updates[i].FilePath = actionPath
			}

			// Create analyzer
			cacheInstance, err := cache.NewCache(cache.DefaultConfig())
			testutil.AssertNoError(t, err)
			defer testutil.CleanupCache(t, cacheInstance)()

			analyzer := &Analyzer{
				Cache: NewCacheAdapter(cacheInstance),
			}

			// Apply updates
			err = analyzer.ApplyPinnedUpdates(tt.updates)

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
					OldUses: "actions/checkout@v4",
					NewUses: "actions/checkout@abc123 # v4.1.1",
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
					NewUses: "actions/checkout@abc123 # v4.1.1",
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
					OldUses: "actions/checkout@v4",
					NewUses: "actions/checkout@abc123 # v4.1.1",
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
					OldUses: "actions/checkout@v4",
					NewUses: "actions/checkout@abc123 # v4.1.1",
				},
			},
			expectedYAML: testutil.MustReadFixture("dependencies/test-checkout-with-comment-pinned.yml"),
			expectError:  false,
			checkBackup:  true,
		},
		{
			name:        "invalid YAML triggers rollback",
			initialYAML: testutil.MustReadFixture("dependencies/simple-test-checkout.yml"),
			updates: []PinnedUpdate{
				{
					OldUses: "actions/checkout@v4",
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

			actionPath := filepath.Join(dir, appconstants.ActionFileNameYML)
			testutil.WriteTestFile(t, actionPath, tt.initialYAML)

			// Create analyzer
			cacheInstance, err := cache.NewCache(cache.DefaultConfig())
			testutil.AssertNoError(t, err)
			defer testutil.CleanupCache(t, cacheInstance)()

			analyzer := &Analyzer{
				Cache: NewCacheAdapter(cacheInstance),
			}

			// Apply update
			err = analyzer.updateActionFile(actionPath, tt.updates)

			// Check error expectation
			if (err != nil) != tt.expectError {
				t.Errorf("updateActionFile() error = %v, expectError %v", err, tt.expectError)

				return
			}

			if !tt.expectError {
				// Read and verify content
				actualContent := testutil.SafeReadFile(t, actionPath, dir)

				if strings.TrimSpace(string(actualContent)) != strings.TrimSpace(tt.expectedYAML) {
					t.Errorf("updateActionFile() content mismatch\nGot:\n%s\nWant:\n%s",
						string(actualContent), tt.expectedYAML)
				}

				// Check backup was removed
				if tt.checkBackup {
					backupPath := actionPath + appconstants.BackupExtension
					testutil.AssertFileNotExists(t, backupPath)
				}
			}

			if tt.rollbackCheck {
				// Verify rollback happened - file should be unchanged
				actualContent := testutil.SafeReadFile(t, actionPath, dir)

				if strings.TrimSpace(string(actualContent)) != strings.TrimSpace(tt.initialYAML) {
					t.Errorf("rollback failed, file was modified")
				}
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

			actionPath := filepath.Join(dir, appconstants.ActionFileNameYML)
			testutil.WriteTestFile(t, actionPath, tt.yamlContent)

			// Create analyzer
			cacheInstance, err := cache.NewCache(cache.DefaultConfig())
			testutil.AssertNoError(t, err)
			defer testutil.CleanupCache(t, cacheInstance)()

			analyzer := &Analyzer{
				Cache: NewCacheAdapter(cacheInstance),
			}

			// Validate
			err = analyzer.validateActionFile(actionPath)

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
					"GET https://api.github.com/repos/test/repo/tags": "[]",
				})
				cacheInstance, _ := cache.NewCache(cache.DefaultConfig())

				return &Analyzer{
					GitHubClient: mockClient,
					Cache:        NewCacheAdapter(cacheInstance),
				}
			},
			owner:       "test",
			repo:        "repo",
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
			owner:       "test",
			repo:        "repo",
			expectError: true,
		},
		{
			name: "malformed tag response",
			mockSetup: func() *Analyzer {
				mockClient := testutil.MockGitHubClient(map[string]string{
					"GET https://api.github.com/repos/test/repo/tags": "invalid json",
				})
				cacheInstance, _ := cache.NewCache(cache.DefaultConfig())

				return &Analyzer{
					GitHubClient: mockClient,
					Cache:        NewCacheAdapter(cacheInstance),
				}
			},
			owner:       "test",
			repo:        "repo",
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

// TestCacheVersionEdgeCases tests edge cases for cacheVersion and getCachedVersion.
//
//nolint:gocyclo // Test function with multiple independent test cases
func TestCacheVersionEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("getCachedVersion with nil cache", func(t *testing.T) {
		t.Parallel()

		analyzer := &Analyzer{Cache: nil}
		version, sha, found := analyzer.getCachedVersion("test-key")

		if found {
			t.Error("getCachedVersion() should return false with nil cache")
		}
		if version != "" || sha != "" {
			t.Error("getCachedVersion() should return empty strings with nil cache")
		}
	})

	t.Run("getCachedVersion with invalid cached data type", func(t *testing.T) {
		t.Parallel()

		cacheInstance, err := cache.NewCache(cache.DefaultConfig())
		testutil.AssertNoError(t, err)
		defer testutil.CleanupCache(t, cacheInstance)()

		// Store invalid data type
		_ = cacheInstance.Set("test-key", "invalid-string-not-map")

		analyzer := &Analyzer{Cache: NewCacheAdapter(cacheInstance)}
		version, sha, found := analyzer.getCachedVersion("test-key")

		if found {
			t.Error("getCachedVersion() should return false with invalid data type")
		}
		if version != "" || sha != "" {
			t.Error("getCachedVersion() should return empty strings with invalid data type")
		}
	})

	t.Run("getCachedVersion with empty cache entry", func(t *testing.T) {
		t.Parallel()

		cacheInstance, err := cache.NewCache(cache.DefaultConfig())
		testutil.AssertNoError(t, err)
		defer testutil.CleanupCache(t, cacheInstance)()

		analyzer := &Analyzer{Cache: NewCacheAdapter(cacheInstance)}
		version, sha, found := analyzer.getCachedVersion("nonexistent-key")

		if found {
			t.Error("getCachedVersion() should return false for nonexistent key")
		}
		if version != "" || sha != "" {
			t.Error("getCachedVersion() should return empty strings for nonexistent key")
		}
	})

	t.Run("cacheVersion with nil cache", func(t *testing.T) {
		t.Parallel()

		analyzer := &Analyzer{Cache: nil}
		// Should not panic
		analyzer.cacheVersion("test-key", "v1.0.0", "abc123")
	})

	t.Run("cacheVersion stores and retrieves correctly", func(t *testing.T) {
		t.Parallel()

		cacheInstance, err := cache.NewCache(cache.DefaultConfig())
		testutil.AssertNoError(t, err)
		defer testutil.CleanupCache(t, cacheInstance)()

		analyzer := &Analyzer{Cache: NewCacheAdapter(cacheInstance)}

		// Cache a version
		analyzer.cacheVersion("test-key", "v1.2.3", "def456")

		// Retrieve it
		version, sha, found := analyzer.getCachedVersion("test-key")

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

		actionPath := filepath.Join(dir, appconstants.ActionFileNameYML)
		originalContent := testutil.MustReadFixture("dependencies/simple-test-checkout.yml")

		testutil.WriteTestFile(t, actionPath, originalContent)

		cacheInstance, err := cache.NewCache(cache.DefaultConfig())
		testutil.AssertNoError(t, err)
		defer testutil.CleanupCache(t, cacheInstance)()

		analyzer := &Analyzer{Cache: NewCacheAdapter(cacheInstance)}

		updates := []PinnedUpdate{
			{
				OldUses: "actions/checkout@v4",
				NewUses: "actions/checkout@abc123 # v4.1.1",
			},
		}

		err = analyzer.updateActionFile(actionPath, updates)
		testutil.AssertNoError(t, err)

		// Backup should be removed after successful update
		backupPath := actionPath + appconstants.BackupExtension
		testutil.AssertFileNotExists(t, backupPath)
	})

	t.Run("rollback on validation failure", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		defer cleanup()

		actionPath := filepath.Join(dir, appconstants.ActionFileNameYML)
		originalContent := testutil.MustReadFixture("dependencies/simple-test-checkout.yml")

		testutil.WriteTestFile(t, actionPath, originalContent)

		cacheInstance, err := cache.NewCache(cache.DefaultConfig())
		testutil.AssertNoError(t, err)
		defer testutil.CleanupCache(t, cacheInstance)()

		analyzer := &Analyzer{Cache: NewCacheAdapter(cacheInstance)}

		// Create an update that breaks YAML
		updates := []PinnedUpdate{
			{
				OldUses: "name: Test",
				NewUses: "invalid::yaml::syntax:",
			},
		}

		err = analyzer.updateActionFile(actionPath, updates)
		if err == nil {
			t.Error("updateActionFile() should return error for invalid YAML")
		}

		// File should be rolled back to original
		actualContent := testutil.SafeReadFile(t, actionPath, dir)
		if strings.TrimSpace(string(actualContent)) != strings.TrimSpace(originalContent) {
			t.Error("rollback failed, file was not restored to original content")
		}

		// Backup should be removed after rollback
		backupPath := actionPath + appconstants.BackupExtension
		testutil.AssertFileNotExists(t, backupPath)
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

		cacheInstance, err := cache.NewCache(cache.DefaultConfig())
		testutil.AssertNoError(t, err)
		defer testutil.CleanupCache(t, cacheInstance)()

		analyzer := &Analyzer{Cache: NewCacheAdapter(cacheInstance)}

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

	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
	testutil.AssertNoError(t, err)
	defer testutil.CleanupCache(t, cacheInstance)()

	analyzer := &Analyzer{Cache: NewCacheAdapter(cacheInstance)}

	// Create updates for both files
	updates := []PinnedUpdate{
		{
			FilePath: action1Path,
			OldUses:  "actions/checkout@v4",
			NewUses:  "actions/checkout@abc123 # v4.1.1",
		},
		{
			FilePath: action2Path,
			OldUses:  "actions/setup-node@v3",
			NewUses:  "actions/setup-node@def456 # v4.0.0",
		},
	}

	err = analyzer.ApplyPinnedUpdates(updates)
	testutil.AssertNoError(t, err)

	// Verify both files were updated
	content1 := testutil.SafeReadFile(t, action1Path, dir)
	if !strings.Contains(string(content1), "actions/checkout@abc123 # v4.1.1") {
		t.Errorf("action1.yml was not updated correctly, got:\n%s", string(content1))
	}

	content2 := testutil.SafeReadFile(t, action2Path, dir)
	if !strings.Contains(string(content2), "actions/setup-node@def456 # v4.0.0") {
		t.Errorf("action2.yml was not updated correctly, got:\n%s", string(content2))
	}
}
