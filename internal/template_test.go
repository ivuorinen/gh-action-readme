package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// newTemplateData creates a TemplateData with common test values.
// Pass nil for any field to use defaults or zero values.
func newTemplateData(
	actionName string,
	version string,
	useDefaultBranch bool,
	defaultBranch string,
	org string,
	repo string,
	actionPath string,
	repoRoot string,
) *TemplateData {
	var actionYML *ActionYML
	if actionName != "" {
		actionYML = &ActionYML{Name: actionName}
	}

	return &TemplateData{
		ActionYML: actionYML,
		Config: &AppConfig{
			Version:          version,
			UseDefaultBranch: useDefaultBranch,
		},
		Git: git.RepoInfo{
			Organization:  org,
			Repository:    repo,
			DefaultBranch: defaultBranch,
		},
		ActionPath: actionPath,
		RepoRoot:   repoRoot,
	}
}

// TestExtractActionSubdirectory tests the extractActionSubdirectory function.
func TestExtractActionSubdirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		actionPath string
		repoRoot   string
		want       string
	}{
		{
			name:       "subdirectory action",
			actionPath: "/repo/actions/csharp-build/action.yml",
			repoRoot:   "/repo",
			want:       "actions/csharp-build",
		},
		{
			name:       "single level subdirectory",
			actionPath: testutil.TestRepoBuildActionPath,
			repoRoot:   "/repo",
			want:       "build",
		},
		{
			name:       "deeply nested subdirectory",
			actionPath: "/repo/a/b/c/d/action.yml",
			repoRoot:   "/repo",
			want:       "a/b/c/d",
		},
		{
			name:       "root action",
			actionPath: testutil.TestRepoActionPath,
			repoRoot:   "/repo",
			want:       "",
		},
		{
			name:       "empty action path",
			actionPath: "",
			repoRoot:   "/repo",
			want:       "",
		},
		{
			name:       "empty repo root",
			actionPath: testutil.TestRepoActionPath,
			repoRoot:   "",
			want:       "",
		},
		{
			name:       "both empty",
			actionPath: "",
			repoRoot:   "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractActionSubdirectory(tt.actionPath, tt.repoRoot)

			// Normalize paths for cross-platform compatibility
			want := filepath.ToSlash(tt.want)
			got = filepath.ToSlash(got)

			if got != want {
				t.Errorf("extractActionSubdirectory() = %q, want %q", got, want)
			}
		})
	}
}

// TestBuildUsesString tests the buildUsesString function with subdirectory extraction.
func TestBuildUsesString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		td      *TemplateData
		org     string
		repo    string
		version string
		want    string
	}{
		{
			name: "monorepo with subdirectory",
			td: &TemplateData{
				ActionPath: "/repo/actions/csharp-build/action.yml",
				RepoRoot:   "/repo",
			},
			org:     "ivuorinen",
			repo:    "actions",
			version: "@main",
			want:    "ivuorinen/actions/actions/csharp-build@main",
		},
		{
			name: "root action",
			td: &TemplateData{
				ActionPath: testutil.TestRepoActionPath,
				RepoRoot:   "/repo",
			},
			org:     "ivuorinen",
			repo:    "my-action",
			version: "@main",
			want:    "ivuorinen/my-action@main",
		},
		{
			name: "empty org",
			td: &TemplateData{
				ActionPath: testutil.TestRepoBuildActionPath,
				RepoRoot:   "/repo",
			},
			org:     "",
			repo:    "actions",
			version: "@main",
			want:    "your-org/your-action@v1",
		},
		{
			name: "empty repo",
			td: &TemplateData{
				ActionPath: testutil.TestRepoBuildActionPath,
				RepoRoot:   "/repo",
			},
			org:     "ivuorinen",
			repo:    "",
			version: "@main",
			want:    "your-org/your-action@v1",
		},
		{
			name: "missing paths in template data",
			td: &TemplateData{
				ActionPath: "",
				RepoRoot:   "",
			},
			org:     "ivuorinen",
			repo:    "actions",
			version: "@v1",
			want:    "ivuorinen/actions@v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildUsesString(tt.td, tt.org, tt.repo, tt.version)

			// Normalize paths for cross-platform compatibility
			want := filepath.ToSlash(tt.want)
			got = filepath.ToSlash(got)

			if got != want {
				t.Errorf("buildUsesString() = %q, want %q", got, want)
			}
		})
	}
}

// TestGetActionVersion tests the getActionVersion function with priority logic.
func TestGetActionVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data any
		want string
	}{
		{
			name: "config version override",
			data: newTemplateData("", "v2.0.0", true, "main", "", "", "", ""),
			want: "v2.0.0",
		},
		{
			name: "use default branch when enabled",
			data: newTemplateData("", "", true, "main", "", "", "", ""),
			want: "main",
		},
		{
			name: "use default branch master",
			data: newTemplateData("", "", true, "master", "", "", "", ""),
			want: "master",
		},
		{
			name: "fallback to v1 when default branch disabled",
			data: newTemplateData("", "", false, "main", "", "", "", ""),
			want: "v1",
		},
		{
			name: "fallback to v1 when default branch not detected",
			data: newTemplateData("", "", true, "", "", "", "", ""),
			want: "v1",
		},
		{
			name: "fallback to v1 when data is invalid",
			data: "invalid",
			want: "v1",
		},
		{
			name: "fallback to v1 when data is nil",
			data: nil,
			want: "v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getActionVersion(tt.data)
			if got != tt.want {
				t.Errorf("getActionVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestGetGitUsesString tests the complete integration of gitUsesString template function.
func TestGetGitUsesString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data *TemplateData
		want string
	}{
		{
			name: "monorepo action with default branch",
			data: newTemplateData("C# Build", "", true, "main", "ivuorinen", "actions",
				"/repo/csharp-build/action.yml", "/repo"),
			want: "ivuorinen/actions/csharp-build@main",
		},
		{
			name: "monorepo action with explicit version",
			data: newTemplateData("Build Action", "v1.0.0", true, "main", "org", "actions",
				testutil.TestRepoBuildActionPath, "/repo"),
			want: "org/actions/build@v1.0.0",
		},
		{
			name: "root level action with default branch",
			data: newTemplateData("My Action", "", true, "develop", "user", "my-action",
				testutil.TestRepoActionPath, "/repo"),
			want: "user/my-action@develop",
		},
		{
			name: "action with use_default_branch disabled",
			data: newTemplateData(testutil.TestActionName, "", false, "main", "org", "test",
				testutil.TestRepoActionPath, "/repo"),
			want: "org/test@v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getGitUsesString(tt.data)

			// Normalize paths for cross-platform compatibility
			want := filepath.ToSlash(tt.want)
			got = filepath.ToSlash(got)

			if got != want {
				t.Errorf("getGitUsesString() = %q, want %q", got, want)
			}
		})
	}
}

// TestFormatVersion tests the formatVersion function.
func TestFormatVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "empty version",
			version: "",
			want:    "@v1",
		},
		{
			name:    "whitespace only version",
			version: "   ",
			want:    "@v1",
		},
		{
			name:    "version without @",
			version: "v1.2.3",
			want:    testutil.TestVersionV123,
		},
		{
			name:    "version with @",
			version: testutil.TestVersionV123,
			want:    testutil.TestVersionV123,
		},
		{
			name:    "main branch",
			version: "main",
			want:    "@main",
		},
		{
			name:    "version with @ and spaces",
			version: "  @v2.0.0  ",
			want:    "@v2.0.0",
		},
		{
			name:    "sha version",
			version: "abc123",
			want:    "@abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatVersion(tt.version)
			if got != tt.want {
				t.Errorf("formatVersion(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

// TestBuildTemplateData tests the BuildTemplateData function.
func TestBuildTemplateData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		action     *ActionYML
		config     *AppConfig
		repoRoot   string
		actionPath string
		wantOrg    string
		wantRepo   string
	}{
		{
			name: "basic action with config overrides",
			action: &ActionYML{
				Name:        testutil.TestActionName,
				Description: "Test description",
			},
			config: &AppConfig{
				Organization: "testorg",
				Repository:   "testrepo",
			},
			repoRoot:   ".",
			actionPath: "action.yml",
			wantOrg:    "testorg",
			wantRepo:   "testrepo",
		},
		{
			name: "action without config overrides",
			action: &ActionYML{
				Name:        "Another Action",
				Description: "Another description",
			},
			config:     &AppConfig{},
			repoRoot:   ".",
			actionPath: "action.yml",
			wantOrg:    "",
			wantRepo:   "",
		},
		{
			name: "action with dependency analysis enabled",
			action: &ActionYML{
				Name:        "Dep Action",
				Description: "Action with deps",
			},
			config: &AppConfig{
				Organization:        "deporg",
				Repository:          "deprepo",
				AnalyzeDependencies: true,
			},
			repoRoot:   ".",
			actionPath: "../testdata/composite-action/action.yml",
			wantOrg:    "deporg",
			wantRepo:   "deprepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := BuildTemplateData(tt.action, tt.config, tt.repoRoot, tt.actionPath)
			assertTemplateData(t, data, tt.action, tt.config, tt.wantOrg, tt.wantRepo)
		})
	}
}

func assertTemplateData(
	t *testing.T,
	data *TemplateData,
	action *ActionYML,
	config *AppConfig,
	wantOrg, wantRepo string,
) {
	t.Helper()

	if data == nil {
		t.Fatal("BuildTemplateData() returned nil")
	}

	if data.ActionYML != action {
		t.Error("BuildTemplateData() did not preserve ActionYML")
	}

	if data.Config != config {
		t.Error("BuildTemplateData() did not preserve Config")
	}

	if config.Organization != "" && data.Git.Organization != wantOrg {
		t.Errorf("BuildTemplateData() Git.Organization = %q, want %q", data.Git.Organization, wantOrg)
	}

	if config.Repository != "" && data.Git.Repository != wantRepo {
		t.Errorf("BuildTemplateData() Git.Repository = %q, want %q", data.Git.Repository, wantRepo)
	}

	if config.AnalyzeDependencies && data.Dependencies == nil {
		t.Error("BuildTemplateData() expected Dependencies to be set when AnalyzeDependencies is true")
	}
}

// TestAnalyzeDependencies tests the analyzeDependencies function.
func TestAnalyzeDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		actionPath string
		config     *AppConfig
		expectNil  bool
	}{
		{
			name:       "valid composite action without GitHub token",
			actionPath: "../../testdata/analyzer/composite-action.yml",
			config:     &AppConfig{},
			expectNil:  false,
		},
		{
			name:       "nonexistent action file",
			actionPath: "../../testdata/analyzer/nonexistent.yml",
			config:     &AppConfig{},
			expectNil:  false, // Should return empty slice, not nil
		},
		{
			name:       "docker action without token",
			actionPath: "../../testdata/analyzer/docker-action.yml",
			config:     &AppConfig{},
			expectNil:  false,
		},
		{
			name:       "javascript action without token",
			actionPath: "../../testdata/analyzer/javascript-action.yml",
			config:     &AppConfig{},
			expectNil:  false,
		},
		{
			name:       "invalid yaml file",
			actionPath: "../../testdata/analyzer/invalid.yml",
			config:     &AppConfig{},
			expectNil:  false, // Should gracefully handle errors and return empty slice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file with fixture content for valid fixtures
			var actionPath string
			if tt.actionPath != "../../testdata/analyzer/nonexistent.yml" {
				// Extract just the filename from the path
				filename := filepath.Base(tt.actionPath)
				yamlContent := testutil.MustReadAnalyzerFixture(filename)

				tmpDir := t.TempDir()
				actionPath = filepath.Clean(filepath.Join(tmpDir, appconstants.ActionFileNameYML))
				if strings.Contains(actionPath, "..") {
					t.Fatalf("invalid path contains ..: %q", actionPath)
				}
				if err := os.WriteFile(actionPath, []byte(yamlContent), appconstants.FilePermDefault); err != nil {
					t.Fatalf("failed to write temp file: %v", err)
				}
			} else {
				// For nonexistent file test, use a nonexistent path
				actionPath = filepath.Join(t.TempDir(), "nonexistent.yml")
			}

			gitInfo := git.RepoInfo{
				Organization: "testorg",
				Repository:   "testrepo",
			}

			result := analyzeDependencies(actionPath, tt.config, gitInfo)

			if tt.expectNil && result != nil {
				t.Errorf("analyzeDependencies() expected nil, got %v", result)
			}

			if !tt.expectNil && result == nil {
				t.Error("analyzeDependencies() returned nil, expected non-nil slice")
			}
		})
	}
}
