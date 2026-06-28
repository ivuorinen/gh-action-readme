package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testTplTestOrg     = testutil.WizardOrgTest
	testTplTestRepo    = testutil.WizardRepoTest
	testTplGitOrg      = "git-org"
	testTplConfigOrg   = "config-org"
	testTplRepoRoot    = "/repo"
	testTplOrgName     = "ivuorinen"
	testTplRefMain     = "@main"
	testTplBranchMain  = "main"
	testTplActionsRepo = "actions"
	testTplOrgPlain    = "org"
	testTplRepoPlain   = "repo"
)

// TestGetFieldWithFallback_GitValueReturned kills the CONDITIONALS_NEGATION mutation at
// template.go:69 (gitValue != "" → == "") by verifying that a non-empty gitValue is
// returned directly, not the configValue or defaultValue.
func TestGetFieldWithFallback_GitValueReturned(t *testing.T) {
	td := &TemplateData{
		ActionYML: &ActionYML{},
		Git: git.RepoInfo{
			Organization: testTplGitOrg,
		},
		Config: &AppConfig{
			Organization: testTplConfigOrg,
		},
	}

	got := getFieldWithFallback(
		td,
		func(d *TemplateData) string { return d.Git.Organization },
		func(d *TemplateData) string { return d.Config.Organization },
		"default-org",
	)

	if got != testTplGitOrg {
		t.Errorf("getFieldWithFallback() = %q, want %q", got, testTplGitOrg)
	}
}

// TestGetFieldWithFallback_ConfigFallback verifies that configValue is returned when
// gitValue is empty but configValue is set.
func TestGetFieldWithFallback_ConfigFallback(t *testing.T) {
	td := &TemplateData{
		ActionYML: &ActionYML{},
		Git:       git.RepoInfo{},
		Config: &AppConfig{
			Organization: testTplConfigOrg,
		},
	}

	got := getFieldWithFallback(
		td,
		func(d *TemplateData) string { return d.Git.Organization },
		func(d *TemplateData) string { return d.Config.Organization },
		"default-org",
	)

	if got != testTplConfigOrg {
		t.Errorf("getFieldWithFallback() = %q, want %q", got, testTplConfigOrg)
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
		func(_ *TemplateData) string { return testTplGitOrg },
		func(_ *TemplateData) string { return testTplConfigOrg },
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
			repoRoot:   testTplRepoRoot,
			want:       "actions/csharp-build",
		},
		{
			name:       "single level subdirectory",
			actionPath: testutil.TestRepoBuildActionPath,
			repoRoot:   testTplRepoRoot,
			want:       "build",
		},
		{
			name:       "deeply nested subdirectory",
			actionPath: "/repo/a/b/c/d/action.yml",
			repoRoot:   testTplRepoRoot,
			want:       "a/b/c/d",
		},
		{
			name:       testutil.TestCaseNameRootAction,
			actionPath: testutil.TestRepoActionPath,
			repoRoot:   testTplRepoRoot,
			want:       "",
		},
		{
			// N133: a real subdir whose name merely starts with ".." must not be
			// dropped by the parent-traversal guard.
			name:       "subdirectory name starting with dots",
			actionPath: "/repo/..build/action.yml",
			repoRoot:   testTplRepoRoot,
			want:       "..build",
		},
		{
			name:       "empty action path",
			actionPath: "",
			repoRoot:   testTplRepoRoot,
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

// TestMdCell verifies markdown table-cell escaping of pipes and newlines.
func TestMdCell(t *testing.T) {
	t.Parallel()

	noop := "ordinary text"
	cases := map[string]string{
		noop:         noop, // no special chars → unchanged
		"a|b":        "a\\|b",
		"a | b | c":  "a \\| b \\| c",
		"line1\nl2":  "line1<br>l2",
		"crlf\r\nx":  "crlf<br>x",
		"trailing\r": "trailing<br>",
	}
	for in, want := range cases {
		if got := mdCell(in); got != want {
			t.Errorf("mdCell(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestIsValidOrgRepo verifies org/repo validation rejects path-injection
// characters while preserving legitimate dotted names.
func TestIsValidOrgRepo(t *testing.T) {
	t.Parallel()

	cases := []struct {
		org, repo string
		want      bool
	}{
		{"actions", "checkout", true},
		{"my-org", "my.repo", true}, // dotted repo stays valid
		{"org_1", "repo-2", true},
		{"", testTplRepoPlain, false},
		{testTplOrgPlain, "", false},
		{"evil/x", testTplRepoPlain, false},  // embedded slash injects a path segment
		{testTplOrgPlain, "repo@v9", false},  // embedded @ injects a version
		{testTplOrgPlain, "re po", false},    // whitespace
		{testTplOrgPlain, "repo/sub", false}, // slash in repo
		{appconstants.DefaultOrgPlaceholder, testTplRepoPlain, false},
	}
	for _, c := range cases {
		if got := isValidOrgRepo(c.org, c.repo); got != c.want {
			t.Errorf("isValidOrgRepo(%q, %q) = %v, want %v", c.org, c.repo, got, c.want)
		}
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
				RepoRoot:   testTplRepoRoot,
			},
			org:     testTplOrgName,
			repo:    testTplActionsRepo,
			version: testTplRefMain,
			want:    "ivuorinen/actions/actions/csharp-build@main",
		},
		{
			name: testutil.TestCaseNameRootAction,
			td: &TemplateData{
				ActionPath: testutil.TestRepoActionPath,
				RepoRoot:   testTplRepoRoot,
			},
			org:     testTplOrgName,
			repo:    "my-action",
			version: testTplRefMain,
			want:    "ivuorinen/my-action@main",
		},
		{
			name: "empty org",
			td: &TemplateData{
				ActionPath: testutil.TestRepoBuildActionPath,
				RepoRoot:   testTplRepoRoot,
			},
			org:     "",
			repo:    testTplActionsRepo,
			version: testTplRefMain,
			want:    "your-org/your-action@v1",
		},
		{
			name: "empty repo",
			td: &TemplateData{
				ActionPath: testutil.TestRepoBuildActionPath,
				RepoRoot:   testTplRepoRoot,
			},
			org:     testTplOrgName,
			repo:    "",
			version: testTplRefMain,
			want:    "your-org/your-action@v1",
		},
		{
			name: "missing paths in template data",
			td: &TemplateData{
				ActionPath: "",
				RepoRoot:   "",
			},
			org:     testTplOrgName,
			repo:    testTplActionsRepo,
			version: appconstants.VersionRefV1,
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
			data: newTemplateData(
				templateDataParams{version: "v2.0.0", useDefaultBranch: true, defaultBranch: testTplBranchMain},
			),
			want: "v2.0.0",
		},
		{
			name: "use default branch when enabled",
			data: newTemplateData(templateDataParams{useDefaultBranch: true, defaultBranch: testTplBranchMain}),
			want: testTplBranchMain,
		},
		{
			// N125: a version with whitespace/newlines is rejected (would inject
			// extra lines into the rendered uses: example) and falls through.
			name: "invalid version with newline falls through to default branch",
			data: newTemplateData(
				templateDataParams{
					version:          "v1\n      run: x",
					useDefaultBranch: true,
					defaultBranch:    testTplBranchMain,
				},
			),
			want: testTplBranchMain,
		},
		{
			name: "use default branch master",
			data: newTemplateData(templateDataParams{useDefaultBranch: true, defaultBranch: "master"}),
			want: "master",
		},
		{
			name: "fallback to v1 when default branch disabled",
			data: newTemplateData(templateDataParams{useDefaultBranch: false, defaultBranch: testTplBranchMain}),
			want: appconstants.VersionTagV1,
		},
		{
			name: "fallback to v1 when default branch not detected",
			data: newTemplateData(templateDataParams{useDefaultBranch: true}),
			want: appconstants.VersionTagV1,
		},
		{
			name: "fallback to v1 when data is invalid",
			data: "invalid",
			want: appconstants.VersionTagV1,
		},
		{
			name: "fallback to v1 when data is nil",
			data: nil,
			want: appconstants.VersionTagV1,
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
				defaultBranch:    testTplBranchMain,
				org:              testTplOrgName,
				repo:             testTplActionsRepo,
				actionPath:       "/repo/csharp-build/action.yml",
				repoRoot:         testTplRepoRoot,
			}),
			want: "ivuorinen/actions/csharp-build@main",
		},
		{
			name: "monorepo action with explicit version",
			data: newTemplateData(templateDataParams{
				actionName:       "Build Action",
				version:          "v1.0.0",
				useDefaultBranch: true,
				defaultBranch:    testTplBranchMain,
				org:              "org",
				repo:             testTplActionsRepo,
				actionPath:       testutil.TestRepoBuildActionPath,
				repoRoot:         testTplRepoRoot,
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
				repoRoot:         testTplRepoRoot,
			}),
			want: "user/my-action@develop",
		},
		{
			name: "action with use_default_branch disabled",
			data: newTemplateData(templateDataParams{
				actionName:       testutil.TestActionName,
				useDefaultBranch: false,
				defaultBranch:    testTplBranchMain,
				org:              "org",
				repo:             "test",
				actionPath:       testutil.TestRepoActionPath,
				repoRoot:         testTplRepoRoot,
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
			want:    appconstants.VersionRefV1,
		},
		{
			name:    "whitespace only version",
			version: "   ",
			want:    appconstants.VersionRefV1,
		},
		{
			name:    "version without @",
			version: testutil.TestVersionSemantic,
			want:    testutil.TestVersionWithAt,
		},
		{
			name:    "version with @",
			version: testutil.TestVersionWithAt,
			want:    testutil.TestVersionWithAt,
		},
		{
			name:    "main branch",
			version: testTplBranchMain,
			want:    testTplRefMain,
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
				Organization: testTplTestOrg,
				Repository:   testTplTestRepo,
			},
			repoRoot:   ".",
			actionPath: appconstants.ActionFileNameYML,
			wantOrg:    testTplTestOrg,
			wantRepo:   testTplTestRepo,
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
	testutil.CreateGitRepoWithRemote(
		t,
		tmpDir,
		fmt.Sprintf("https://github.com/%s/%s.git", testTplTestOrg, testTplTestRepo),
	)

	action := &ActionYML{Name: "Test", Description: "test"}
	config := &AppConfig{}

	data := BuildTemplateData(action, config, tmpDir, filepath.Join(tmpDir, appconstants.ActionFileNameYML))

	if data.Git.Organization != testTplTestOrg {
		t.Errorf("expected Git.Organization = %q, got %q", testTplTestOrg, data.Git.Organization)
	}

	if data.Git.Repository != testTplTestRepo {
		t.Errorf("expected Git.Repository = %q, got %q", testTplTestRepo, data.Git.Repository)
	}
}

// TestRenderReadme_HTMLHeaderFooter tests that HTML rendering includes header and footer content
// from files on the filesystem (kills mutations on lines 305 and 309 of template.go).
// TestRenderReadmeDefaultTemplateNoBranding verifies N146: an action without a
// branding block still renders the full README, not just the title (the body was
// erroneously wrapped in {{if .Branding}}).
func TestRenderReadmeDefaultTemplateNoBranding(t *testing.T) {
	t.Parallel()

	action := &ActionYML{
		Name:        "NoBrand",
		Description: "the description",
		Inputs:      map[string]ActionInput{"token": {Description: "a token"}},
	}
	td := BuildTemplateData(action, DefaultAppConfig(), "", "")

	out, err := RenderReadme(td, TemplateOptions{
		TemplatePath: resolveTemplatePath(appconstants.TemplatePathDefault),
		Format:       appconstants.OutputFormatMarkdown,
	})
	if err != nil {
		t.Fatalf("RenderReadme: %v", err)
	}

	for _, want := range []string{"## Usage", "## Inputs", "the description", "token"} {
		if !strings.Contains(out, want) {
			t.Errorf("default README missing %q for a branding-less action:\n%s", want, out)
		}
	}
}

// TestMdCode verifies N144: a value with a backtick is fenced so the Markdown
// inline code span is not closed early, and table specials stay escaped.
func TestMdCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
		want string
	}{
		{"no specials", "v1.0", "`v1.0`"},
		{"pipe escaped", "a|b", "`a\\|b`"},
		{"backtick fenced", "use `json`", "`` use `json` ``"},
		{"non-string", 5, "`5`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := mdCode(tt.in); got != tt.want {
				t.Errorf("mdCode(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestAdocCode verifies N144/N145: AsciiDoc inline code uses the unconstrained
// double-backtick form when the value contains a backtick, and pipes are escaped.
func TestAdocCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
		want string
	}{
		{"simple", "v1", "`v1`"},
		{"backtick unconstrained", "a`b", "``a`b``"},
		{"pipe escaped", "x|y", "`x\\|y`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := adocCode(tt.in); got != tt.want {
				t.Errorf("adocCode(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestAdocCellNewline verifies N145: AsciiDoc cells turn newlines into an AsciiDoc
// hard break (" +"), not the markdown "<br>" that AsciiDoc renders literally.
func TestAdocCellNewline(t *testing.T) {
	t.Parallel()

	if got := adocCell("a\nb"); got != "a +\nb" {
		t.Errorf("adocCell newline = %q, want %q", got, "a +\nb")
	}
	if got := adocCell("x|y"); got != "x\\|y" {
		t.Errorf("adocCell pipe = %q, want %q", got, "x\\|y")
	}
}

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

	t.Run("missing header path surfaces an error", func(t *testing.T) {
		t.Parallel()

		// N126: a mistyped --header path must produce a diagnostic, not output
		// with the fragment silently missing.
		opts := TemplateOptions{
			TemplatePath: tmplPath,
			HeaderPath:   filepath.Join(tmpDir, "does-not-exist.html"),
			Format:       appconstants.OutputFormatHTML,
		}

		if _, err := RenderReadme(action, opts); err == nil {
			t.Error("expected an error for a missing HTML header path, got nil")
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
	return filepath.Join(t.TempDir(), testutil.TestNonexistentYML)
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
				Organization: testTplTestOrg,
				Repository:   testTplTestRepo,
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

// TestHasDefault verifies the template presence check distinguishes an absent
// default (nil) from an explicit falsey default (false, 0, ""), which plain
// template truthiness cannot — so documentation cells render "(default: false)"
// instead of silently dropping it.
func TestHasDefault(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		val  any
		want bool
	}{
		{"nil is absent", nil, false},
		{"false is present", false, true},
		{"zero int is present", 0, true},
		{"empty string is present", "", true},
		{"non-empty value is present", "x", true},
	}
	for _, tc := range cases {
		if got := hasDefault(tc.val); got != tc.want {
			t.Errorf("%s: hasDefault(%#v) = %v, want %v", tc.name, tc.val, got, tc.want)
		}
	}
}
