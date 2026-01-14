package internal

import (
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestAssertTemplateData_Helper tests the assertTemplateData helper function.
func TestAssertTemplateData_Helper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() (*TemplateData, *ActionYML, *AppConfig)
		wantOrg  string
		wantRepo string
	}{
		{
			name: "valid template data",
			setup: func() (*TemplateData, *ActionYML, *AppConfig) {
				action := &ActionYML{
					Name:        "Test Action",
					Description: "A test action",
				}
				config := &AppConfig{
					Organization: testutil.TestOrgName,
					Repository:   testutil.TestRepoName,
				}
				data := &TemplateData{
					ActionYML: action,
					Git: git.RepoInfo{
						Organization: testutil.TestOrgName,
						Repository:   testutil.TestRepoName,
					},
					Config: config,
				}

				return data, action, config
			},
			wantOrg:  testutil.TestOrgName,
			wantRepo: testutil.TestRepoName,
		},
		{
			name: "template data with dependencies",
			setup: func() (*TemplateData, *ActionYML, *AppConfig) {
				action := &ActionYML{
					Name: "Action with deps",
				}
				config := &AppConfig{
					Organization:        testutil.MyOrgName,
					Repository:          testutil.MyRepoName,
					AnalyzeDependencies: true,
				}
				data := &TemplateData{
					ActionYML: action,
					Git: git.RepoInfo{
						Organization: testutil.MyOrgName,
						Repository:   testutil.MyRepoName,
					},
					Config:       config,
					Dependencies: []dependencies.Dependency{}, // Empty slice, not nil
				}

				return data, action, config
			},
			wantOrg:  testutil.MyOrgName,
			wantRepo: testutil.MyRepoName,
		},
		{
			name: "template data with empty organization",
			setup: func() (*TemplateData, *ActionYML, *AppConfig) {
				action := &ActionYML{
					Name: "Test",
				}
				config := &AppConfig{
					Organization: "",
					Repository:   testutil.RepoName,
				}
				data := &TemplateData{
					ActionYML: action,
					Git: git.RepoInfo{
						Organization: "",
						Repository:   testutil.RepoName,
					},
					Config: config,
				}

				return data, action, config
			},
			wantOrg:  "",
			wantRepo: testutil.RepoName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, action, config := tt.setup()

			// Call the helper - it validates the template data
			assertTemplateData(t, data, action, config, tt.wantOrg, tt.wantRepo)
		})
	}
}

// TestPrepareTestActionFile_Helper tests the prepareTestActionFile helper function.
func TestPrepareTestActionFile_Helper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		actionPath string
		wantExists bool
	}{
		{
			name:       "analyzer fixture composite action",
			actionPath: testutil.AnalyzerFixturePath + "composite-action.yml",
			wantExists: true,
		},
		{
			name:       "analyzer fixture docker action",
			actionPath: testutil.AnalyzerFixturePath + "docker-action.yml",
			wantExists: true,
		},
		{
			name:       "analyzer fixture javascript action",
			actionPath: testutil.AnalyzerFixturePath + "javascript-action.yml",
			wantExists: true,
		},
		{
			name:       "nonexistent file path",
			actionPath: testutil.AnalyzerFixturePath + "nonexistent.yml",
			wantExists: true, // Helper creates a path, even if file doesn't exist
		},
		{
			name:       "non-analyzer path",
			actionPath: "some/other/path.yml",
			wantExists: true, // Returns tmpDir path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the helper - it prepares a test action file
			result := prepareTestActionFile(t, tt.actionPath)

			// Verify we got a path
			if result == "" {
				t.Error("prepareTestActionFile returned empty path")
			}

			// Verify it's an absolute path or relative path
			if !filepath.IsAbs(result) && !filepath.IsLocal(result) {
				t.Logf("Note: path may be relative or absolute: %s", result)
			}
		})
	}
}
