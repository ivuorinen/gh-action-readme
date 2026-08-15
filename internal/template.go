package internal

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	texttemplate "text/template"

	"github.com/google/go-github/v74/github"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/internal/validation"
	templatesembed "github.com/ivuorinen/gh-action-readme/templates_embed"
)

// TemplateOptions defines options for rendering templates.
type TemplateOptions struct {
	TemplatePath string
	HeaderPath   string
	FooterPath   string
	Format       string // md or html
	// OutputDir is the directory the rendered document is written to. It is needed
	// to resolve links to repository files (see the repoFile template func), which
	// are relative to the document, not to the action file.
	OutputDir string
}

// TemplateData represents all data available to templates.
type TemplateData struct {
	// Action Data
	*ActionYML

	// Git Repository Information
	Git git.RepoInfo `json:"git"`

	// Configuration
	Config *AppConfig `json:"config"`

	// Computed Values
	UsesStatement string `json:"uses_statement"`

	// Path information for subdirectory extraction
	ActionPath string `json:"action_path,omitempty"`
	RepoRoot   string `json:"repo_root,omitempty"`

	// OutputDir is the directory the rendered document is written to. Links to
	// repository files are resolved relative to it (see the repoFile template func),
	// so a monorepo action documented into its own directory links correctly.
	OutputDir string `json:"output_dir,omitempty"`

	// License is the resolved license identifier, or "" when unknown. Templates must
	// guard on it — rendering a license the action did not declare is a false claim
	// about someone else's terms.
	License string `json:"license,omitempty"`

	// Permissions shadows the embedded ActionYML.Permissions with the resolved set:
	// the action's own declaration, falling back to the config-level `permissions:`
	// block when the action declares none. It is a separate field rather than a
	// mutation of ActionYML so the parsed action stays untouched across formats.
	Permissions PermissionMap `json:"permissions,omitempty"`

	// Dependencies (populated by dependency analysis)
	Dependencies []dependencies.Dependency `json:"dependencies,omitempty"`
}

// templateFuncs returns a map of custom template functions.
func templateFuncs() texttemplate.FuncMap {
	return texttemplate.FuncMap{
		"lower":         strings.ToLower,
		"upper":         strings.ToUpper,
		"replace":       strings.ReplaceAll,
		"join":          strings.Join,
		"gitOrg":        getGitOrg,
		"gitRepo":       getGitRepo,
		"gitUsesString": getGitUsesString,
		"actionVersion": getActionVersion,
		"mdCell":        mdCell,
		"mdCode":        mdCode,
		"adocCell":      adocCell,
		"adocCode":      adocCode,
		"badgeSegment":  shieldsBadgeEncode,
		"hasDefault":    hasDefault,
		"yamlComment":   yamlComment,
		"repoFile":      repoFileLink,
		"runsOn":        runsOnValue,
		"var":           configVariable,
		"add1":          func(n int) int { return n + 1 },
		// A block-scalar description ("description: |") keeps its trailing newline,
		// which lands on top of the template's own blank line and yields a doubled
		// blank (markdownlint MD012). Used where the value is rendered as a plain
		// paragraph rather than through mdCell.
		"trim": strings.TrimSpace,
	}
}

// runsOnValue renders the configured `runs_on` runners as a YAML `runs-on:` value for
// the workflow examples: a bare scalar for one runner, a flow sequence for several
// (both valid GitHub Actions syntax). Falls back to the built-in default runner when
// the config is empty, so examples are never left with a blank runner.
func runsOnValue(data any) string {
	td, ok := data.(*TemplateData)
	if !ok || td.Config == nil || len(td.Config.RunsOn) == 0 {
		return appconstants.RunnerUbuntuLatest
	}

	runners := td.Config.RunsOn
	if len(runners) == 1 {
		return runners[0]
	}

	return "[" + strings.Join(runners, ", ") + "]"
}

// configVariable looks up a custom template variable declared under `variables:` in
// config. Returns "" for an unknown key so a template can use {{with var . "k"}} to
// render a block only when the variable is set.
func configVariable(data any, key string) string {
	td, ok := data.(*TemplateData)
	if !ok || td.Config == nil {
		return ""
	}

	return td.Config.Variables[key]
}

// repoFileLink returns a link target for a repository-root file (LICENSE,
// CONTRIBUTING.md, …), relative to the directory the document is being written to.
// An action at actions/foo documented into actions/foo/README.md gets "../../LICENSE".
//
// It returns "" when the repo root is unknown or the file does not exist, so
// templates can guard with {{with repoFile . "LICENSE"}} and omit the link rather
// than emitting one that 404s. "LICENSE" additionally matches the conventional
// variants (LICENSE.md, LICENSE, COPYING) via findLicenseFile.
func repoFileLink(data any, name string) string {
	td, ok := data.(*TemplateData)
	if !ok || td.RepoRoot == "" || name == "" {
		return ""
	}

	target := resolveRepoFilePath(td.RepoRoot, name)
	if target == "" {
		return ""
	}

	absOut := resolveSymlinkedAbs(documentDir(td))
	absTarget := resolveSymlinkedAbs(target)
	if absOut == "" || absTarget == "" {
		return ""
	}

	// A document written outside the repository (an absolute --output, or an
	// --output-dir pointing elsewhere) cannot reach a repo file by any sensible
	// relative path. Emit nothing rather than a link that climbs out of the tree.
	if !isWithin(resolveSymlinkedAbs(td.RepoRoot), absOut) {
		return ""
	}

	rel, err := filepath.Rel(absOut, absTarget)
	if err != nil {
		return ""
	}

	return filepath.ToSlash(rel)
}

// documentDir returns the directory the rendered document is written to. OutputDir is
// set by the generator from the fully resolved output path; the ActionPath fallback
// covers template data built outside a generation run (e.g. direct BuildTemplateData
// use in tests).
func documentDir(td *TemplateData) string {
	if td.OutputDir != "" {
		return td.OutputDir
	}
	if td.ActionPath == "" {
		return ""
	}

	return filepath.Dir(td.ActionPath)
}

// isWithin reports whether path is root itself or lives underneath it. Both are
// expected to be absolute and symlink-resolved.
func isWithin(root, path string) bool {
	if root == "" || path == "" {
		return false
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	return rel == "." ||
		(rel != appconstants.PathParent &&
			!strings.HasPrefix(rel, appconstants.PathParent+string(filepath.Separator)))
}

// resolveRepoFilePath returns the real path of a repository-root file, or "" when it
// does not exist. The LICENSE family is matched through findLicenseFile so naming
// variants resolve; any other name is matched exactly.
// The name is supplied by a template, which may be a user-provided --template, so it
// is not trusted to stay inside the repository. Reject absolute paths and anything
// that changes under Clean or walks upward before touching the filesystem; the
// function's contract is "a file in this repository", and a link to /etc/passwd is
// not one.
func resolveRepoFilePath(repoRoot, name string) string {
	if filepath.IsAbs(name) {
		return ""
	}
	cleaned := filepath.Clean(name)
	if cleaned != name || cleaned == appconstants.PathParent ||
		slices.Contains(strings.Split(filepath.ToSlash(cleaned), "/"), appconstants.PathParent) {
		return ""
	}

	if strings.EqualFold(cleaned, "LICENSE") {
		path, _ := findLicenseFile(repoRoot)

		return path
	}

	path := filepath.Join(repoRoot, cleaned)
	if _, err := os.Stat(path); err != nil {
		return ""
	}

	return path
}

// yamlComment flattens v onto a single line so it stays inside the `#` comment it
// is interpolated into. An action.yml description is commonly a block scalar
// (`description: |`), whose second and later lines would otherwise land at column 0
// inside a rendered ```yaml fence — un-commented, and no longer valid YAML. Unlike
// mdCell (which emits <br> for a Markdown table cell), a YAML comment has no line
// break to escape to, so runs of whitespace collapse to single spaces.
func yamlComment(v any) string {
	return strings.Join(strings.Fields(fmt.Sprintf("%v", v)), " ")
}

// hasDefault reports whether an action input declared a default value. ActionInput.
// Default is `any`, so an absent default is nil while an explicit falsey default
// (false, 0, "") is a non-nil zero value. Template truthiness (`if .Default`) treats
// those falsey defaults as unset and drops them, so the documentation cells use this
// presence check to render them instead of silently omitting "(default: false)".
func hasDefault(v any) bool {
	return v != nil
}

// mdCode renders v as a Markdown inline code span that survives table cells and
// values containing backticks. Pipes/newlines are escaped like mdCell; a value
// containing backticks is fenced with one more backtick than its longest internal
// run (the GFM rule) and space-padded so the backticks render literally. A plain
// `...`-wrapped value would otherwise be closed early by an embedded backtick.
func mdCode(v any) string {
	s := mdCell(fmt.Sprintf("%v", v))

	longest, run := 0, 0
	for _, r := range s {
		if r == '`' {
			run++
			if run > longest {
				longest = run
			}

			continue
		}
		run = 0
	}

	if longest == 0 {
		return "`" + s + "`"
	}
	fence := strings.Repeat("`", longest+1)

	return fence + " " + s + " " + fence
}

// adocCell escapes v for an AsciiDoc table cell: the `|` separator is escaped and
// newlines become an AsciiDoc hard break (" +"), unlike mdCell's markdown `<br>`,
// which AsciiDoc renders as literal text.
func adocCell(v any) string {
	s := fmt.Sprintf("%v", v)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " +\n")
	s = strings.ReplaceAll(s, "\r", " +\n")

	return s
}

// adocCode renders v as an AsciiDoc inline monospace span (table-cell safe). A
// value containing a backtick uses the unconstrained double-backtick form so the
// inner backticks render literally instead of closing the span early.
func adocCode(v any) string {
	s := adocCell(v)
	if strings.Contains(s, "`") {
		return "``" + s + "``"
	}

	return "`" + s + "`"
}

// mdCell escapes a value for safe interpolation into a Markdown (or AsciiDoc)
// table cell. A literal "|" would otherwise be read as a column separator and an
// embedded newline would terminate the row, so both honest values (a description
// containing a pipe) and crafted action metadata can corrupt the table. Pipes are
// backslash-escaped and CR/LF are collapsed to a <br> break.
func mdCell(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", "<br>")
	s = strings.ReplaceAll(s, "\r", "<br>")

	return s
}

// getFieldWithFallback extracts a field from TemplateData with Git-then-Config fallback logic.
func getFieldWithFallback(data any, gitGetter, configGetter func(*TemplateData) string, defaultValue string) string {
	if td, ok := data.(*TemplateData); ok {
		if gitValue := gitGetter(td); gitValue != "" {
			return gitValue
		}
		// configGetter dereferences td.Config; guard against a manually
		// constructed TemplateData with a nil Config (consistent with getActionVersion).
		if td.Config != nil {
			if configValue := configGetter(td); configValue != "" {
				return configValue
			}
		}
	}

	return defaultValue
}

// getGitOrg returns the Git organization from template data.
func getGitOrg(data any) string {
	return getFieldWithFallback(data,
		func(td *TemplateData) string { return td.Git.Organization },
		func(td *TemplateData) string { return td.Config.Organization },
		appconstants.DefaultOrgPlaceholder)
}

// getGitRepo returns the Git repository name from template data.
func getGitRepo(data any) string {
	return getFieldWithFallback(data,
		func(td *TemplateData) string { return td.Git.Repository },
		func(td *TemplateData) string { return td.Config.Repository },
		appconstants.DefaultRepoPlaceholder)
}

// getGitUsesString returns a complete uses string for the action.
func getGitUsesString(data any) string {
	td, ok := data.(*TemplateData)
	if !ok {
		return appconstants.DefaultUsesPlaceholder
	}

	org := strings.TrimSpace(getGitOrg(data))
	repo := strings.TrimSpace(getGitRepo(data))

	if !isValidOrgRepo(org, repo) {
		return appconstants.DefaultUsesPlaceholder
	}

	version := formatVersion(getActionVersion(data))

	return buildUsesString(td, org, repo, version)
}

// isValidOrgRepo checks if org and repo are valid.
func isValidOrgRepo(org, repo string) bool {
	return org != "" && repo != "" &&
		org != appconstants.DefaultOrgPlaceholder &&
		repo != appconstants.DefaultRepoPlaceholder &&
		isValidGitHubPathSegment(org) &&
		isValidGitHubPathSegment(repo)
}

// reGitHubPathSegment matches a single GitHub org/repo path segment: letters,
// digits, "-", "_", and "." (so dotted repos like "my.repo" stay valid). It
// rejects "/", "@", whitespace, and control characters, which would otherwise
// inject extra path segments or a bogus version into the generated uses:
// statement (org/repo flow from config files and git detection).
var reGitHubPathSegment = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func isValidGitHubPathSegment(s string) bool {
	return reGitHubPathSegment.MatchString(s)
}

// reVersionRef matches a valid git ref/version for a uses: statement: tags,
// branches, and SHAs are made of letters, digits, and "._/+-" (branch names may
// contain "/"). Rejecting whitespace and control characters stops a multiline
// Config.Version from injecting extra YAML lines into the rendered usage example.
var reVersionRef = regexp.MustCompile(`^[A-Za-z0-9._/+-]+$`)

func isValidVersionRef(v string) bool {
	return reVersionRef.MatchString(v)
}

// formatVersion ensures version has proper @ prefix.
func formatVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return appconstants.VersionRefV1
	}
	if !strings.HasPrefix(version, "@") {
		return "@" + version
	}

	return version
}

// buildUsesString constructs the uses string with optional subdirectory path.
func buildUsesString(td *TemplateData, org, repo, version string) string {
	// Use the validation package's FormatUsesStatement for consistency
	if org == "" || repo == "" {
		return appconstants.DefaultUsesPlaceholder
	}

	// For monorepo actions in subdirectories, extract the actual directory path
	subdir := extractActionSubdirectory(td.ActionPath, td.RepoRoot)

	if subdir != "" {
		// Action is in a subdirectory: org/repo/subdir@version
		return validation.FormatUsesStatement(org, repo+"/"+subdir, version)
	}

	// Action is at repo root: org/repo@version
	return validation.FormatUsesStatement(org, repo, version)
}

// resolveSymlinkedAbs returns the absolute, symlink-resolved form of path. It
// falls back to the plain absolute path when symlinks cannot be evaluated (e.g.
// the path does not exist), and returns "" only when even filepath.Abs fails.
func resolveSymlinkedAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	if resolved, rerr := filepath.EvalSymlinks(abs); rerr == nil {
		return resolved
	}

	return abs
}

// extractActionSubdirectory extracts the subdirectory path for an action relative to repo root.
// For monorepo actions (e.g., org/repo/subdir/action.yml), returns "subdir".
// For repo-root actions (e.g., org/repo/action.yml), returns empty string.
// Returns empty string if paths cannot be determined.
func extractActionSubdirectory(actionPath, repoRoot string) string {
	// Validate inputs
	if actionPath == "" || repoRoot == "" {
		return ""
	}

	// Get absolute, symlink-resolved paths for reliable comparison. Resolving
	// symlinks on both sides keeps filepath.Rel from producing a spurious ".."
	// (and silently dropping the subdir) when the action path is reached through
	// a symlink that diverges from the real repo-root tree.
	absActionPath := resolveSymlinkedAbs(actionPath)
	absRepoRoot := resolveSymlinkedAbs(repoRoot)
	if absActionPath == "" || absRepoRoot == "" {
		return ""
	}

	// Get the directory containing action.yml
	actionDir := filepath.Dir(absActionPath)

	// Calculate relative path from repo root to action directory
	relPath, err := filepath.Rel(absRepoRoot, actionDir)
	if err != nil {
		return ""
	}

	// If relative path is "." or empty, action is at repo root
	if relPath == "." || relPath == "" {
		return ""
	}

	// If the relative path escapes the repo root, the action is outside the repo
	// (shouldn't happen). Match an exact ".." component rather than a "*.." prefix
	// so a legitimate subdirectory named "..build" is not silently dropped.
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return ""
	}

	// Return the subdirectory with forward slashes (filepath.Rel yields OS-native
	// separators) so the generated uses: path is valid on Windows too:
	// "actions/csharp-build", never "actions\csharp-build".
	return filepath.ToSlash(relPath)
}

// getActionVersion returns the action version from template data.
// Priority: 1) Config.Version (explicit override), 2) Default branch (if enabled), 3) "v1" (fallback).
func getActionVersion(data any) string {
	td, ok := data.(*TemplateData)
	if !ok || td.Config == nil {
		return appconstants.VersionTagV1
	}

	// Priority 1: Explicit version override. Strip a single leading @ first — a
	// user may paste "@v2" straight from a uses: line — so a valid ref is not
	// silently dropped (@ is not in the allowed set; formatVersion re-adds it).
	// Reject values with whitespace or control characters so a multiline
	// Config.Version cannot inject extra YAML lines into the rendered uses:
	// example (text/template does not escape it), mirroring the org/repo
	// path-segment check. An invalid value falls through.
	if v := strings.TrimPrefix(td.Config.Version, "@"); v != "" && isValidVersionRef(v) {
		return v
	}

	// Priority 2: Use default branch if enabled and available
	if td.Config.UseDefaultBranch && td.Git.DefaultBranch != "" {
		return td.Git.DefaultBranch
	}

	// Priority 3: Fallback
	return appconstants.VersionTagV1
}

// resolvePermissions returns the permission set to document. The action's own
// declaration (YAML plus header comments, already merged by the parser) wins; the
// config-level `permissions:` block is a fallback for actions that declare none,
// which is the only reading under which that config key can affect output.
func resolvePermissions(action *ActionYML, config *AppConfig) PermissionMap {
	if action != nil && len(action.Permissions) > 0 {
		return action.Permissions
	}
	if config == nil || len(config.Permissions) == 0 {
		return nil
	}

	resolved := make(PermissionMap, len(config.Permissions))
	maps.Copy(resolved, config.Permissions)

	return resolved
}

// buildTemplateData constructs comprehensive template data from action and
// configuration. When res is non-nil, dependency analysis reuses the caller-owned
// shared cache+client (read/written once per run); nil gives standalone per-call
// behavior. The nil-res convenience wrapper is test-only and lives in
// deadcode_wrappers_test.go.
func buildTemplateData(
	action *ActionYML,
	config *AppConfig,
	repoRoot, actionPath string,
	res *depResources,
) *TemplateData {
	// Guard against a nil config: this is an exported entry point and the
	// template funcs dereference Config unconditionally.
	if config == nil {
		config = DefaultAppConfig()
	}

	data := &TemplateData{
		ActionYML:  action,
		Config:     config,
		ActionPath: actionPath,
		RepoRoot:   repoRoot,
		// Resolved once here so every format agrees. Shadows the embedded
		// ActionYML.License (depth 0 wins over depth 1), which holds only the
		// action-declared value; this is the full config > yaml > comment > detected
		// chain. "" means unknown and templates must render no license section.
		License:     resolveLicense(action, config, repoRoot),
		Permissions: resolvePermissions(action, config),
	}

	// Populate Git information
	if repoRoot != "" {
		if info, err := git.DetectRepository(repoRoot); err == nil {
			data.Git = *info
		}
	}

	// Override with configuration values if available
	if config.Organization != "" {
		data.Git.Organization = config.Organization
	}
	if config.Repository != "" {
		data.Git.Repository = config.Repository
	}

	// Build uses statement
	data.UsesStatement = getGitUsesString(data)

	// Add dependency analysis if enabled
	if config.AnalyzeDependencies && actionPath != "" {
		data.Dependencies = analyzeDependencies(actionPath, config, data.Git, res)
	}

	return data
}

// depResources holds the GitHub client and dependency cache shared across a
// generation run, so cache.json is read/rewritten once per run instead of once
// per action file. Build with newDepResources; release with Close.
type depResources struct {
	cache  dependencies.DependencyCache
	client *github.Client
}

// newDepResources builds the shared GitHub client and dependency cache once.
// The client is nil when no token is configured (graceful degradation); the
// cache falls back to a no-op cache when creation fails.
func newDepResources(config *AppConfig) *depResources {
	res := &depResources{}

	if token := GetGitHubToken(config); token != "" {
		if client, err := NewGitHubClient(token); err == nil {
			res.client = client.Client
		}
	}

	// *cache.Cache implements DependencyCache directly; on creation failure leave
	// res.cache nil, which the analyzer's nil-guards treat as caching disabled.
	if cacheInstance, err := cache.NewCache(cache.DefaultConfig()); err == nil {
		res.cache = cacheInstance
	}

	return res
}

// Close stops the cache's background goroutine and flushes pending writes once.
func (r *depResources) Close() error {
	if r == nil || r.cache == nil {
		return nil
	}

	return r.cache.Close()
}

// analyzeDependencies performs dependency analysis on the action file. When res
// is non-nil the shared cache+client are reused and left open for the owner to
// close; when nil it builds and owns its own cache+client for this single call.
func analyzeDependencies(
	actionPath string,
	config *AppConfig,
	gitInfo git.RepoInfo,
	res *depResources,
) []dependencies.Dependency {
	if res == nil {
		res = newDepResources(config)
		// This call owns the resources it built; release them on return.
		defer func() { _ = res.Close() }()
	}

	analyzer := dependencies.NewAnalyzer(res.client, gitInfo, res.cache)

	// Analyze dependencies
	deps, err := analyzer.AnalyzeActionFile(actionPath)
	if err != nil {
		// Log error but don't fail - return empty dependencies
		return []dependencies.Dependency{}
	}

	return deps
}

// templateExecuter is a common interface for html/template and text/template.
type templateExecuter interface {
	Execute(io.Writer, any) error
}

// parseReadmeTemplate parses raw template content into an executable template.
// HTML format uses html/template for automatic XSS-safe escaping.
func parseReadmeTemplate(content []byte, format string) (templateExecuter, error) {
	if format == appconstants.OutputFormatHTML {
		funcs := htmltemplate.FuncMap(templateFuncs())

		return htmltemplate.New(appconstants.TemplateNameReadme).Funcs(funcs).Parse(string(content))
	}

	return texttemplate.New(appconstants.TemplateNameReadme).Funcs(templateFuncs()).Parse(string(content))
}

// RenderReadme renders a README using a Go template and the parsed action.yml data.
func RenderReadme(action any, opts TemplateOptions) (string, error) {
	tmplContent, err := templatesembed.ReadTemplate(opts.TemplatePath)
	if err != nil {
		return "", err
	}

	tmpl, err := parseReadmeTemplate(tmplContent, opts.Format)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	// Header/footer partials apply to every text format, not just HTML: the caller
	// (Generator.resolveHeaderFooter) is responsible for only supplying a partial
	// appropriate to the format, so the built-in HTML scaffolding is never injected
	// into Markdown.
	if opts.HeaderPath != "" {
		rendered, e := renderPartial(opts.HeaderPath, opts.Format, action)
		if e != nil {
			return "", fmt.Errorf("header %q: %w", opts.HeaderPath, e)
		}
		buf.WriteString(rendered)
	}

	if err := tmpl.Execute(&buf, action); err != nil {
		return "", err
	}

	if opts.FooterPath != "" {
		rendered, e := renderPartial(opts.FooterPath, opts.Format, action)
		if e != nil {
			return "", fmt.Errorf("footer %q: %w", opts.FooterPath, e)
		}
		buf.WriteString(rendered)
	}

	return buf.String(), nil
}

// renderPartial reads the header/footer template at path and executes it with the
// same template data as the body, so placeholders like {{.Name}} render instead of
// leaking literally into the output (header.tmpl carries a {{.Name}} in its <title>).
func renderPartial(path, format string, data any) (string, error) {
	content, err := templatesembed.ReadTemplate(path)
	if err != nil {
		return "", err
	}

	tmpl, err := parseReadmeTemplate(content, format)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return "", err
	}

	return b.String(), nil
}
