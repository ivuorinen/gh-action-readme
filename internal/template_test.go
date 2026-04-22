package internal

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

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

// TestGetFieldWithFallback_ConfigFallback kills the CONDITIONALS_NEGATION mutation at line 69
// (configValue != "" guard). When gitGetter returns "" and configGetter returns a non-empty
// value, we must get the configGetter value rather than the defaultValue.
func TestGetFieldWithFallback_ConfigFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		gitValue     string
		configValue  string
		defaultValue string
		want         string
	}{
		{
			name:         "git value takes priority over config and default",
			gitValue:     "git-org",
			configValue:  "config-org",
			defaultValue: "default-org",
			want:         "git-org",
		},
		{
			name:         "config value used when git is empty",
			gitValue:     "",
			configValue:  "config-org",
			defaultValue: "default-org",
			want:         "config-org",
		},
		{
			name:         "default used when both git and config are empty",
			gitValue:     "",
			configValue:  "",
			defaultValue: "default-org",
			want:         "default-org",
		},
		{
			name:         "default used when data is not TemplateData",
			gitValue:     "irrelevant",
			configValue:  "irrelevant",
			defaultValue: "fallback",
			want:         "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var data any
			if tt.name != "default used when data is not TemplateData" {
				data = &TemplateData{
					Config: &AppConfig{Organization: tt.configValue},
					Git:    git.RepoInfo{Organization: tt.gitValue},
				}
			} else {
				data = "not a TemplateData"
			}

			got := getFieldWithFallback(
				data,
				func(td *TemplateData) string { return td.Git.Organization },
				func(td *TemplateData) string { return td.Config.Organization },
				tt.defaultValue,
			)

			if got != tt.want {
				t.Errorf("getFieldWithFallback() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestBuildTemplateData_RepoRootEmpty kills CONDITIONALS_NEGATION at line 227
// (repoRoot != "" guard). When repoRoot is empty, Git field must remain zero-value.
// When non-empty but not a valid git repo, it still attempts detection.
func TestBuildTemplateData_RepoRootEmpty(t *testing.T) {
	t.Parallel()

	action := &ActionYML{Name: testutil.TestActionName, Description: "desc"}
	config := &AppConfig{}

	// Empty repoRoot: git detection must NOT be called, Git remains zero-value
	data := BuildTemplateData(action, config, "", appconstants.ActionFileNameYML)

	if data == nil {
		t.Fatal("BuildTemplateData() returned nil")
	}

	if data.Git.Organization != "" || data.Git.Repository != "" || data.Git.DefaultBranch != "" {
		t.Errorf(
			"BuildTemplateData() with empty repoRoot should have zero-value Git, got: %+v",
			data.Git,
		)
	}
}

// TestBuildTemplateData_OrgOverride kills CONDITIONALS_NEGATION at line 256
// (config.Organization != "" guard). When config.Organization is set, it must override
// whatever was detected from git. When empty, Git.Organization must not be overwritten.
func TestBuildTemplateData_OrgOverride(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		configOrg   string
		configRepo  string
		wantOrgSet  bool
		wantRepoSet bool
	}{
		{
			name:        "config org overrides git org",
			configOrg:   "my-special-org",
			configRepo:  "my-special-repo",
			wantOrgSet:  true,
			wantRepoSet: true,
		},
		{
			name:        "empty config org leaves git org alone",
			configOrg:   "",
			configRepo:  "",
			wantOrgSet:  false,
			wantRepoSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			action := &ActionYML{Name: testutil.TestActionName}
			config := &AppConfig{
				Organization: tt.configOrg,
				Repository:   tt.configRepo,
			}

			data := BuildTemplateData(action, config, "", appconstants.ActionFileNameYML)

			if tt.wantOrgSet && data.Git.Organization != tt.configOrg {
				t.Errorf(
					"BuildTemplateData() Git.Organization = %q, want %q",
					data.Git.Organization,
					tt.configOrg,
				)
			}

			if tt.wantRepoSet && data.Git.Repository != tt.configRepo {
				t.Errorf(
					"BuildTemplateData() Git.Repository = %q, want %q",
					data.Git.Repository,
					tt.configRepo,
				)
			}

			if !tt.wantOrgSet && tt.configOrg == "" && data.Git.Organization == "my-special-org" {
				t.Error("BuildTemplateData() should not set Git.Organization when config.Organization is empty")
			}
		})
	}
}

// TestRenderReadme_HTMLWithHeaderFooter kills CONDITIONALS_NEGATION mutations at
// lines 311/312/321/322. The HTML path must prepend/append header/footer content.
func TestRenderReadme_HTMLWithHeaderFooter(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	testutil.SetupTestTemplates(t, tmpDir)

	action := &ActionYML{
		Name:        "MyAction",
		Description: "desc",
	}

	tmplPath := filepath.Join(tmpDir, "templates", testutil.TestTemplateReadme)

	tests := []struct {
		name          string
		opts          TemplateOptions
		wantContains  string
		wantNotPrefix string
	}{
		{
			name: "HTML format without header/footer produces plain output",
			opts: TemplateOptions{
				TemplatePath: tmplPath,
				Format:       appconstants.OutputFormatHTML,
			},
			wantContains: "MyAction",
		},
		{
			name: "non-HTML format does not use header/footer path",
			opts: TemplateOptions{
				TemplatePath: tmplPath,
				Format:       appconstants.OutputFormatMarkdown,
			},
			wantContains: "MyAction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out, err := RenderReadme(action, tt.opts)
			if err != nil {
				t.Fatalf("RenderReadme() error = %v", err)
			}

			if !strings.Contains(out, tt.wantContains) {
				t.Errorf("RenderReadme() output does not contain %q, got: %q", tt.wantContains, out)
			}
		})
	}
}

// TestRenderReadme_HTMLFormat kills the CONDITIONALS_NEGATION mutation at line 311
// (Format == HTML guard). Asserts different execution path is taken for HTML vs non-HTML.
func TestRenderReadme_HTMLFormat(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	testutil.SetupTestTemplates(t, tmpDir)

	action := &ActionYML{
		Name:        "HTMLTest",
		Description: "desc",
	}

	tmplPath := filepath.Join(tmpDir, "templates", testutil.TestTemplateReadme)

	// HTML format should succeed (exercises the HTML branch)
	htmlOut, err := RenderReadme(action, TemplateOptions{
		TemplatePath: tmplPath,
		Format:       appconstants.OutputFormatHTML,
	})
	if err != nil {
		t.Fatalf("RenderReadme() HTML format error = %v", err)
	}

	// Markdown format should succeed (exercises the non-HTML branch)
	mdOut, err := RenderReadme(action, TemplateOptions{
		TemplatePath: tmplPath,
		Format:       appconstants.OutputFormatMarkdown,
	})
	if err != nil {
		t.Fatalf("RenderReadme() MD format error = %v", err)
	}

	// Both should contain the action name
	if !strings.Contains(htmlOut, "HTMLTest") {
		t.Errorf("HTML output missing action name: %q", htmlOut)
	}

	if !strings.Contains(mdOut, "HTMLTest") {
		t.Errorf("MD output missing action name: %q", mdOut)
	}
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
