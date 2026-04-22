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

// TestGetFieldWithFallback_GitValueReturned kills the CONDITIONALS_NEGATION mutation at
// template.go:69 (gitValue != "" → == "") by verifying that a non-empty gitValue is
// returned directly, not the configValue or defaultValue.
func TestGetFieldWithFallback_GitValueReturned(t *testing.T) {
	td := &TemplateData{
		ActionYML: &ActionYML{},
		Git: git.RepoInfo{
			Organization: "git-org",
		},
		Config: &AppConfig{
			Organization: "config-org",
		},
	}

	got := getFieldWithFallback(
		td,
		func(d *TemplateData) string { return d.Git.Organization },
		func(d *TemplateData) string { return d.Config.Organization },
		"default-org",
	)

	if got != "git-org" {
		t.Errorf("getFieldWithFallback() = %q, want %q", got, "git-org")
	}
}

// TestGetFieldWithFallback_ConfigFallback verifies that configValue is returned when
// gitValue is empty but configValue is set.
func TestGetFieldWithFallback_ConfigFallback(t *testing.T) {
	td := &TemplateData{
		ActionYML: &ActionYML{},
		Git:       git.RepoInfo{},
		Config: &AppConfig{
			Organization: "config-org",
		},
	}

	got := getFieldWithFallback(
		td,
		func(d *TemplateData) string { return d.Git.Organization },
		func(d *TemplateData) string { return d.Config.Organization },
		"default-org",
	)

	if got != "config-org" {
		t.Errorf("getFieldWithFallback() = %q, want %q", got, "config-org")
	}
}

// TestGetFieldWithFallback_DefaultValue verifies that the default is returned when both
// git and config values are empty.
func TestGetFieldWithFallback_DefaultValue(t *testing.T) {
	td := &TemplateData{
		ActionYML: &ActionYML{},
		Git:       git.RepoInfo{},
		Config:    &AppConfig{},
	}

	got := getFieldWithFallback(
		td,
		func(d *TemplateData) string { return d.Git.Organization },
		func(d *TemplateData) string { return d.Config.Organization },
		"default-org",
	)

	if got != "default-org" {
		t.Errorf("getFieldWithFallback() = %q, want %q", got, "default-org")
	}
}

// TestGetFieldWithFallback_NonTemplateData verifies the type-guard branch: passing a
// non-*TemplateData value always returns the default value.
func TestGetFieldWithFallback_NonTemplateData(t *testing.T) {
	got := getFieldWithFallback(
		"not-a-TemplateData",
		func(_ *TemplateData) string { return "git-org" },
		func(_ *TemplateData) string { return "config-org" },
		"default-org",
	)

	if got != "default-org" {
		t.Errorf("getFieldWithFallback() = %q, want %q", got, "default-org")
	}
}

// templateDataParams holds parameters for creating test TemplateData.
type templateDataParams struct {
	actionName       string
	version          string
	useDefaultBranch bool
	defaultBranch    string
	org              string
	repo             string
	actionPath       string
	repoRoot         string
}

// newTemplateData creates a TemplateData with the provided templateDataParams.
// Zero values are preserved as-is; this helper does not apply defaults.
// Callers must set defaults themselves or use a separate defaulting helper.
func newTemplateData(params templateDataParams) *TemplateData {
	var actionYML *ActionYML
	if params.actionName != "" {
		actionYML = &ActionYML{Name: params.actionName}
	}

	return &TemplateData{
		ActionYML: actionYML,
		Config: &AppConfig{
			Version:          params.version,
			UseDefaultBranch: params.useDefaultBranch,
		},
		Git: git.RepoInfo{
			Organization:  params.org,
			Repository:    params.repo,
			DefaultBranch: params.defaultBranch,
		},
		ActionPath: params.actionPath,
		RepoRoot:   params.repoRoot,
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
			name:       testutil.TestCaseNameSubdirAction,
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
			name:       testutil.TestCaseNameRootAction,
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
			name: testutil.TestCaseNameRootAction,
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
			data: newTemplateData(templateDataParams{version: "v2.0.0", useDefaultBranch: true, defaultBranch: "main"}),
			want: "v2.0.0",
		},
		{
			name: "use default branch when enabled",
			data: newTemplateData(templateDataParams{useDefaultBranch: true, defaultBranch: "main"}),
			want: "main",
		},
		{
			name: "use default branch master",
			data: newTemplateData(templateDataParams{useDefaultBranch: true, defaultBranch: "master"}),
			want: "master",
		},
		{
			name: "fallback to v1 when default branch disabled",
			data: newTemplateData(templateDataParams{useDefaultBranch: false, defaultBranch: "main"}),
			want: "v1",
		},
		{
			name: "fallback to v1 when default branch not detected",
			data: newTemplateData(templateDataParams{useDefaultBranch: true}),
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
			data: newTemplateData(templateDataParams{
				actionName:       "C# Build",
				useDefaultBranch: true,
				defaultBranch:    "main",
				org:              "ivuorinen",
				repo:             "actions",
				actionPath:       "/repo/csharp-build/action.yml",
				repoRoot:         "/repo",
			}),
			want: "ivuorinen/actions/csharp-build@main",
		},
		{
			name: "monorepo action with explicit version",
			data: newTemplateData(templateDataParams{
				actionName:       "Build Action",
				version:          "v1.0.0",
				useDefaultBranch: true,
				defaultBranch:    "main",
				org:              "org",
				repo:             "actions",
				actionPath:       testutil.TestRepoBuildActionPath,
				repoRoot:         "/repo",
			}),
			want: "org/actions/build@v1.0.0",
		},
		{
			name: "root level action with default branch",
			data: newTemplateData(templateDataParams{
				actionName:       testutil.TestMyAction,
				useDefaultBranch: true,
				defaultBranch:    "develop",
				org:              "user",
				repo:             "my-action",
				actionPath:       testutil.TestRepoActionPath,
				repoRoot:         "/repo",
			}),
			want: "user/my-action@develop",
		},
		{
			name: "action with use_default_branch disabled",
			data: newTemplateData(templateDataParams{
				actionName:       testutil.TestActionName,
				useDefaultBranch: false,
				defaultBranch:    "main",
				org:              "org",
				repo:             "test",
				actionPath:       testutil.TestRepoActionPath,
				repoRoot:         "/repo",
			}),
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
			want:    testutil.TestVersionWithAt,
		},
		{
			name:    "version with @",
			version: testutil.TestVersionWithAt,
			want:    testutil.TestVersionWithAt,
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
			actionPath: appconstants.ActionFileNameYML,
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
			actionPath: appconstants.ActionFileNameYML,
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

// TestBuildTemplateData_RealGitRepo tests BuildTemplateData with a real git repo so that
// mutations to the repoRoot != "" guard (line 227) and the err == nil guard (line 228)
// produce observably different Git fields.
func TestBuildTemplateData_RealGitRepo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testutil.CreateGitRepoWithRemote(t, tmpDir, "https://github.com/testorg/testrepo.git")

	action := &ActionYML{Name: "Test", Description: "test"}
	config := &AppConfig{}

	data := BuildTemplateData(action, config, tmpDir, filepath.Join(tmpDir, appconstants.ActionFileNameYML))

	if data.Git.Organization != "testorg" {
		t.Errorf("expected Git.Organization = %q, got %q", "testorg", data.Git.Organization)
	}

	if data.Git.Repository != "testrepo" {
		t.Errorf("expected Git.Repository = %q, got %q", "testrepo", data.Git.Repository)
	}
}

// TestRenderReadme_HTMLHeaderFooter tests that HTML rendering includes header and footer content
// from files on the filesystem (kills mutations on lines 305 and 309 of template.go).
func TestRenderReadme_HTMLHeaderFooter(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testutil.SetupTestTemplates(t, tmpDir)

	// Write header and footer as absolute-path files so ReadTemplate uses os.ReadFile
	headerContent := "<html-header>"
	footerContent := "<html-footer>"
	headerPath := filepath.Join(tmpDir, "header.html")
	footerPath := filepath.Join(tmpDir, "footer.html")

	if err := os.WriteFile(headerPath, []byte(headerContent), 0o600); err != nil {
		t.Fatalf("write header: %v", err)
	}

	if err := os.WriteFile(footerPath, []byte(footerContent), 0o600); err != nil {
		t.Fatalf("write footer: %v", err)
	}

	tmplPath := filepath.Join(tmpDir, "templates", testutil.TestTemplateReadme)

	action := &ActionYML{
		Name:        "HTMLAction",
		Description: "html test",
	}

	t.Run("with header and footer", func(t *testing.T) {
		t.Parallel()

		opts := TemplateOptions{
			TemplatePath: tmplPath,
			HeaderPath:   headerPath,
			FooterPath:   footerPath,
			Format:       appconstants.OutputFormatHTML,
		}

		out, err := RenderReadme(action, opts)
		if err != nil {
			t.Fatalf("RenderReadme failed: %v", err)
		}

		if !strings.Contains(out, headerContent) {
			t.Errorf("output missing header content %q; got: %q", headerContent, out)
		}

		if !strings.Contains(out, footerContent) {
			t.Errorf("output missing footer content %q; got: %q", footerContent, out)
		}
	})

	t.Run("without header and footer", func(t *testing.T) {
		t.Parallel()

		opts := TemplateOptions{
			TemplatePath: tmplPath,
			Format:       appconstants.OutputFormatHTML,
		}

		out, err := RenderReadme(action, opts)
		if err != nil {
			t.Fatalf("RenderReadme failed: %v", err)
		}

		if strings.Contains(out, headerContent) {
			t.Errorf("output should not contain header content %q when HeaderPath is empty", headerContent)
		}

		if strings.Contains(out, footerContent) {
			t.Errorf("output should not contain footer content %q when FooterPath is empty", footerContent)
		}
	})
}

// TestAnalyzeDependencies tests the analyzeDependencies function.
// prepareTestActionFile prepares a test action file for analyzeDependencies tests.
func prepareTestActionFile(t *testing.T, actionPath string) string {
	t.Helper()

	if strings.HasPrefix(actionPath, "../../testdata/analyzer/") &&
		actionPath != "../../testdata/analyzer/nonexistent.yml" {
		filename := filepath.Base(actionPath)
		yamlContent := testutil.MustReadAnalyzerFixture(filename)

		tmpDir := t.TempDir()
		tmpPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
		tmpPath = testutil.ValidateTestPath(t, tmpPath, tmpDir)
		testutil.WriteTestFile(t, tmpPath, yamlContent)

		return tmpPath
	}

	// For nonexistent file test
	return filepath.Join(t.TempDir(), "nonexistent.yml")
}

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
		{
			name:       testutil.TestCaseNamePathTraversalAttempt,
			actionPath: "../../etc/passwd",
			config:     &AppConfig{},
			expectNil:  false, // Returns empty slice for invalid paths
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actionPath := prepareTestActionFile(t, tt.actionPath)

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
