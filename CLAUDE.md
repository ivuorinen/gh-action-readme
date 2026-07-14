# CLAUDE.md

**gh-action-readme** — CLI tool for GitHub Actions documentation generation.

Deep references live under `docs/` — this file holds only the agent-critical,
non-obvious guidance. Enforced conventions live in `.claude/rules/`.

## 🛡️ Quality Standards

Strict gates aligned with [SonarCloud "Sonar way"](https://docs.sonarsource.com/sonarqube-cloud/standards/managing-quality-gates/):

| Metric | Threshold | Check |
| ------ | --------- | ----- |
| Code Coverage | ≥ 72% (overall); 80% target | `make test-coverage-check` |
| Duplicated Lines | ≤ 3% (new code) | `make lint` (dupl) |
| Security Rating | A (no issues) | `make security` |
| Reliability / Maintainability | A | `make lint` |
| Cyclomatic Complexity | ≤ 10 per function | `make lint` (gocyclo) |
| Line Length | ≤ 120 characters | `make lint` (lll) |

Threshold is `COVERAGE_THRESHOLD := 72.0` in the `Makefile`. All linters, EditorConfig,
and gitleaks run via pre-commit hooks.

## 🚨 CRITICAL: README Protection

Enforced by [`.claude/rules/readme-protection.md`](.claude/rules/readme-protection.md).
When testing generation, always write to `/tmp/` or `testdata/`:

```bash
gh-action-readme gen testdata/ --output /tmp/test-output.md
```

## 📝 Template Updates

Templates are embedded from `templates_embed/templates/` via a `//go:embed` directive in
`templates_embed/embed.go` (embedded FS by default, filesystem fallback for development).
To modify: edit the file, then `go build .`.

- Default: `templates_embed/templates/readme.tmpl`
- Themes: `templates_embed/templates/themes/{github,gitlab,minimal,professional}/readme.tmpl`
- AsciiDoc (an output format, not a `--theme`): `templates_embed/templates/themes/asciidoc/readme.adoc`

Adding a theme or output format is a multi-step change (constant, map/registry, handler,
CLI help) — see `docs/development.md` and follow the existing theme as the template.

## 🏗️ Architecture

### Command Handler Pattern

All Cobra command handlers **return an error** instead of calling `os.Exit()`, so they are
unit-testable. `wrapHandlerWithErrorHandling()` (in `main.go`) adapts them to Cobra's `Run`:
it initializes `globalConfig` if nil (important for tests), runs the handler, and on error
prints via `ColoredOutput` and exits with `appconstants.ExitCodeError`.

```go
func myHandler(cmd *cobra.Command, args []string) error {
    if err := someOperation(); err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }
    return nil
}

var myCmd = &cobra.Command{Use: "my-command", Run: wrapHandlerWithErrorHandling(myHandler)}
```

### Dependency Injection for Testing

I/O- or global-state-touching functions use the **nil-default parameter pattern**: optional
deps default to the real implementation when nil, and tests pass mocks.

```go
func myFunction(output *ColoredOutput, config *AppConfig, reader InputReader) error {
    if config == nil { config = globalConfig }
    if reader == nil { reader = &StdinReader{} }
    // ...
}
```

Production passes `nil`; tests inject mocks. `InputReader` is declared in
`internal/input.go` (production, enables testing); `TestInputReader` lives in `main_test.go`.

### Template Rendering Pipeline

1. **Parser** (`internal/parser.go`) — parses `action.yml` with `goccy/go-yaml`; extracts
  permissions from header comments via `parsePermissionsFromComments()`; merges comment +
  YAML permissions (YAML wins); returns `*ActionYML`.
2. **Template Data Builder** (`internal/template.go:BuildTemplateData()`) — builds
  `TemplateData`, embeds `*ActionYML`, detects git info, extracts the action subdirectory
  for monorepos, builds the `uses:` statement.
3. **Template Functions** (`internal/template.go:templateFuncs()`) — `gitUsesString`,
  `actionVersion`, `gitOrg`, `gitRepo`, plus `lower`/`upper`/`replace`/`join`.
4. **Renderer** (`internal/template.go:RenderReadme()`) — reads the template via
  `templates_embed.ReadTemplate()` and executes it (formats: md, html, json, asciidoc).

### Key Data Structures

```go
type ActionYML struct {
    Name, Description string
    Inputs      map[string]ActionInput
    Outputs     map[string]ActionOutput
    Runs        map[string]any
    Branding    *Branding
    Permissions PermissionMap // scope→level; also accepts the read-all/write-all scalar
}

type TemplateData struct {
    *ActionYML                 // embedded — fields accessible directly
    Git           git.RepoInfo
    Config        *AppConfig
    UsesStatement string       // pre-built "org/repo/path@version"
    ActionPath, RepoRoot string
    Dependencies  []dependencies.Dependency
}
```

When adding an `ActionYML` field, also update `ActionYMLForJSON` in
`internal/json_writer.go`, add parser tests, and update templates if it should render.

### Monorepo Action Path Resolution

`internal/template.go:extractActionSubdirectory()` computes the action's path relative to the
repo root (empty for root-level actions) so `buildUsesString()` emits a correct `uses:`, e.g.
`/repo/actions/csharp-build/action.yml` (root `/repo`) → `org/repo/actions/csharp-build@main`.

### Permissions Parsing

Merged from two sources — **header comments** (lowest priority) and the **YAML field**
(highest). YAML overrides comment values per key; all unique keys are kept. Comment formats
supported: list (`#   - key: value`), object (`#   key: value`), mixed, and inline comments.
The scalar `permissions: read-all` / `write-all` / `none` shorthand is accepted via
`PermissionMap.UnmarshalYAML` (maps to an `all` scope). See `internal/parser.go`.

## ⚙️ Configuration System

### Hierarchy (highest → lowest priority)

1. **Command-line flags** — override everything.
2. **Environment variables** — config overrides use the `GH_ACTION_README_` prefix (e.g.
  `GH_ACTION_README_THEME`, `GH_ACTION_README_QUIET`), applied via viper `AutomaticEnv()`
  so they override config-file values. The GitHub token is a separate lookup using
  `GH_README_GITHUB_TOKEN` / `GITHUB_TOKEN` (note the different `GH_README_` prefix).
  Matches `docs/configuration.md`.
3. **Action-specific config** — `config.yaml` in the action directory.
4. **Repository config** — `.ghreadme.yaml` in the repo root.
5. **Global config** — `~/.config/gh-action-readme/config.yaml`.
6. **Defaults** — built-in fallbacks.

### Version Resolution for `uses: org/repo@VERSION`

`internal/template.go:getActionVersion()`, in priority order:

1. `Config.Version` — explicit override (e.g. `version: "v2.0.0"`).
2. `Config.UseDefaultBranch` + `Git.DefaultBranch` — detected branch (e.g. `@main`).
3. Fallback — `"v1"`.

## 📊 Package Structure

- `main.go` — CLI entry point (Cobra); `cmd_*.go` — per-command handlers (gen, deps, config,
  cache, validate).
- `internal/generator.go` — generation orchestration; `internal/parser.go` — action.yml
  parsing; `internal/template.go` — template data + rendering; `internal/config.go` — Viper
  config; `internal/json_writer.go` — JSON output; `internal/output.go` — colored output;
  `internal/progress.go` — batch progress bars.
- `internal/git/` — repo detection; `internal/validation/` — validation;
  `internal/wizard/` — interactive config wizard; `internal/dependencies/` — action
  dependency analysis; `internal/apperrors/` — contextual errors; `internal/cache/` —
  analysis cache.
- `appconstants/` — constants; `testutil/` — test utilities; `templates_embed/` — embedded
  templates.

## 🧪 Testing Guidelines

- Unit tests live in `internal/*_test.go` alongside source; fixtures in
  `testdata/yaml-fixtures/` (organized by type: `actions/{composite,docker,javascript,…}`,
  `permissions/`, `configs/`, `validation/`, …).
- **CRITICAL: never inline YAML/config in tests.** Use a fixture file loaded via
  `testutil` helpers, and register its path constant in `testutil/test_constants.go` first
  (see `.claude/rules/no-inline-yaml-in-tests.md`).
- Adding a fixture: create the YAML under the right subdir, add its constant to
  `testutil/test_constants.go`, then reference it (e.g.
  `testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureCompositeBasic)`).

Full command reference (build, test, coverage, mutation/property testing, lint, security,
deps, pre-commit, dependency automation, security tooling) is in `docs/development.md`; the
Makefile is the source of truth for targets.

## 🔐 Security

Enforced by pre-commit (`gitleaks`, `golangci-lint`, `editorconfig-checker`) and CI (CodeQL,
`govulncheck`, Trivy, Nancy). All file reads use validated paths to prevent traversal:

```go
// templates_embed/embed.go:ReadTemplate()
cleanPath := filepath.Clean(templatePath)
if cleanPath != templatePath || strings.Contains(cleanPath, "..") {
    return nil, filepath.ErrBadPattern
}
```
