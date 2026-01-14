# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**gh-action-readme** - CLI tool for GitHub Actions documentation generation

## ⚠️ Code Quality Anti-Patterns - DO NOT REPEAT

**CRITICAL:** The following patterns have caused quality issues in the past. These mistakes must not be repeated:

### 🚫 High Cognitive Complexity

#### Never write functions with cognitive complexity > 15

**Bad - Repeated Mistakes:**

- Nested conditionals in test assertions
- Complex error checking logic duplicated across tests
- Deep nesting in validation functions

**Always:**

- Extract complex logic into helper functions
- Create test helper functions for repeated assertion patterns
- Keep functions focused on a single responsibility
- Break down complex conditions into smaller, testable pieces

**Example:**
Instead of 19 lines of nested error checking, create a helper:

```go
// ❌ BAD - High complexity
func TestValidation(t *testing.T) {
    if result.HasErrors {
        found := false
        for _, err := range result.Errors {
            if strings.Contains(err.Message, expected) {
                found = true
                break
            }
        }
        if !found {
            t.Errorf("error not found")
        }
    } else {
        // more nesting...
    }
}

// ✅ GOOD - Use helper
func TestValidation(t *testing.T) {
    assertValidationError(t, result, "field", true, "expected message")
}
```

### 🚫 Duplicate String Literals

#### Never repeat string literals across test files

**Bad - Repeated Mistakes:**

- File paths like `"/tmp/action.yml"` repeated 22 times
- Action references like `"actions/checkout@v3"` duplicated
- Error messages and test scenarios hardcoded everywhere

**Always:**

- Use constants from `appconstants/` for production strings
- Use constants from `testutil/test_constants.go` for test-only strings
- Add new constants when you see duplication (>2 uses)

**Red Flag Patterns:**

- Same string literal in multiple test files
- Same file path repeated in different tests
- Same error message in multiple assertions

### 🚫 Inline YAML and Config Data in Tests

#### Never embed YAML or config data directly in test code

**Bad - Repeated Mistakes:**

- Inline YAML strings with backticks in test functions
- Config data hardcoded in test setup
- Template content embedded in test files

**Always:**

- Create fixture files in `testdata/yaml-fixtures/`
- Use `testutil.MustReadFixture()` to load fixtures
- Add constants to `testutil/test_constants.go` for fixture paths
- Reuse fixtures across multiple tests

**Example:**

```go
// ❌ BAD - Inline YAML
testConfig := `
theme: default
output_format: md
`

// ✅ GOOD - Use fixture
testConfig := string(testutil.MustReadFixture(testutil.TestConfigDefault))
```

**Fixture Organization:**

- `testdata/yaml-fixtures/configs/` - Config files
- `testdata/yaml-fixtures/actions/` - Action files
- `testdata/yaml-fixtures/template-fixtures/` - Template files

### 🚫 Co-Authored-By Lines in Commits

#### Never add Co-Authored-By or similar bylines to commit messages

**Bad - Repeated Mistakes:**

- Adding `Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>` to commits
- Including attribution lines at end of commit messages
- Adding signature or generated-by lines

**Always:**

- Write clean commit messages following conventional commits format
- Omit any co-author, attribution, or signature lines
- Focus commit message on what changed and why

**Example:**

```text
❌ BAD:
refactor: move inline YAML to fixtures

Benefits:
- Improved maintainability
- Better separation

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>

✅ GOOD:
refactor: move inline YAML to fixtures for better test maintainability

- Created 16 new config fixtures
- Replaced 19 inline YAML instances
- All tests passing with no regressions
```

**When user says "no bylines":**

- This means: Remove ALL attribution/co-author lines
- Do NOT argue or explain why they might be useful
- Just comply immediately and recommit without bylines

### ✅ Prevention Mechanisms

**Before writing ANY code:**

1. Check `testutil/test_constants.go` for existing constants
2. Check `testdata/yaml-fixtures/` for existing fixtures
3. Consider if your function will exceed complexity limits
4. Plan helper functions for complex logic upfront

**Before committing:**

1. Run `make lint` - catches complexity and duplication
2. Pre-commit hooks will catch most issues
3. SonarCloud will flag remaining issues in PR

**Remember:** It's easier to write clean code initially than to refactor after quality issues are raised.

## 🛡️ Quality Standards

This project enforces strict quality gates aligned with [SonarCloud "Sonar way"](https://docs.sonarsource.com/sonarqube-cloud/standards/managing-quality-gates/):

| Metric | Threshold | Check Command |
| ------ | --------- | ------------- |
| Code Coverage | ≥ 80% (new code) | `make test-coverage-check` |
| Duplicated Lines | ≤ 3% (new code) | `make lint` (via dupl) |
| Security Rating | A (no issues) | `make security` |
| Reliability Rating | A (no bugs) | `make lint` |
| Maintainability | A (tech debt ≤ 5%) | `make lint` |
| Cyclomatic Complexity | ≤ 10 per function | `make lint` (via gocyclo) |
| Line Length | ≤ 120 characters | `make lint` (via lll) |

**Current Coverage:** 72.8% overall (target: 80%)
**Coverage Threshold:** Set in `Makefile` as `COVERAGE_THRESHOLD := 80.0`

**Pre-commit Quality Checks:**

- All linters run automatically via pre-commit hooks
- EditorConfig compliance enforced
- Security scans (gitleaks) prevent secret commits

## 📝 Template Updates

**Templates are embedded from:** `templates_embed/templates/`

**To modify templates:**

1. Edit template files directly in `templates_embed/templates/`
2. Rebuild the binary: `go build .`
3. Templates are automatically embedded via `//go:embed` directive

**Available template locations:**

- Default: `templates_embed/templates/readme.tmpl`
- GitHub theme: `templates_embed/templates/themes/github/readme.tmpl`
- GitLab theme: `templates_embed/templates/themes/gitlab/readme.tmpl`
- Minimal theme: `templates_embed/templates/themes/minimal/readme.tmpl`
- Professional: `templates_embed/templates/themes/professional/readme.tmpl`
- AsciiDoc theme: `templates_embed/templates/themes/asciidoc/readme.adoc`

**Template embedding:** Handled by `templates_embed/embed.go` using Go's embed directive.
The embedded filesystem is used by default, with fallback to filesystem for development.

## 🚨 CRITICAL: README Protection

**NEVER overwrite `/README.md`** - The root README.md is the main project documentation.

**For testing generation commands:**

```bash
# Safe testing approaches
gh-action-readme gen testdata/example-action/
gh-action-readme gen testdata/composite-action/action.yml
gh-action-readme gen testdata/ --output /tmp/test-output.md
```

## 🏗️ Architecture Overview

### Command Handler Pattern

**All Cobra command handlers return errors** instead of calling `os.Exit()` directly. This enables comprehensive unit testing.

**Pattern:**

```go
// Handler function signature - returns error
func myHandler(cmd *cobra.Command, args []string) error {
    if err := someOperation(); err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }
    return nil
}

// Wrapped in command definition for Cobra compatibility
var myCmd = &cobra.Command{
    Use:   "my-command",
    Short: "Description",
    Run:   wrapHandlerWithErrorHandling(myHandler),
}
```

The `wrapHandlerWithErrorHandling()` wrapper (in `main.go`):

- Initializes `globalConfig` if nil (important for testing)
- Calls the handler and captures the error
- Displays error via `ColoredOutput` and exits with code 1 if error occurs

**Testing handlers:**

```go
func TestMyHandler(t *testing.T) {
    cmd := &cobra.Command{}
    cmd.Flags().String("some-flag", "default", "")

    err := myHandler(cmd, []string{})

    // Can now test error conditions without os.Exit()
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

### Dependency Injection for Testing

Functions that interact with I/O or global state use **nil-default parameter pattern** for testability:

```go
// Production signature with optional injectable dependencies
func myFunction(output *ColoredOutput, config *AppConfig, reader InputReader) error {
    // Default to real implementations if not provided
    if config == nil {
        config = globalConfig
    }
    if reader == nil {
        reader = &StdinReader{}  // Real stdin
    }
    // ... function logic
}

// Production usage (pass nil for defaults)
err := myFunction(output, nil, nil)

// Test usage (inject mocks)
mockConfig := internal.DefaultAppConfig()
mockReader := &TestInputReader{responses: []string{"y"}}
err := myFunction(output, mockConfig, mockReader)
```

**Examples in codebase:**

- `applyUpdates()` - accepts `InputReader` for stdin mocking (main.go:1094)
- `setupDepsUpgrade()` - accepts `*AppConfig` for config injection (main.go:1001)

**Test interfaces:**

```go
// InputReader for mocking user input
type InputReader interface {
    ReadLine() (string, error)
}

type TestInputReader struct {
    responses []string
    index     int
}
```

### Template Rendering Pipeline

1. **Parser** (`internal/parser.go`):

- Parses `action.yml` files using `goccy/go-yaml`
- Extracts permissions from header comments via `parsePermissionsFromComments()`
- Merges comment and YAML permissions (YAML takes precedence)
- Returns `*ActionYML` struct with all parsed data

1. **Template Data Builder** (`internal/template.go`):

- `BuildTemplateData()` creates comprehensive `TemplateData` struct
- Embeds `*ActionYML` (all action fields accessible via `.Name`, `.Inputs`, `.Permissions`, etc.)
- Detects git repository info (org, repo, default branch)
- Extracts action subdirectory for monorepo support
- Builds `uses:` statement with proper path/version

1. **Template Functions** (`internal/template.go:templateFuncs()`):

- `gitUsesString` - Generates complete `org/repo/path@version` string
- `actionVersion` - Determines version (config override → default branch → "v1")
- `gitOrg`, `gitRepo` - Extract git repository information
- Standard Go template functions: `lower`, `upper`, `replace`, `join`

1. **Renderer** (`internal/template.go:RenderReadme()`):

- Reads template from embedded filesystem via `templates_embed.ReadTemplate()`
- Executes template with `TemplateData`
- Supports multiple formats (md, html, json, asciidoc)

### Key Data Structures

**ActionYML** - Parsed action.yml data:

```go
type ActionYML struct {
    Name        string
    Description string
    Inputs      map[string]ActionInput
    Outputs     map[string]ActionOutput
    Runs        map[string]any
    Branding    *Branding
    Permissions map[string]string  // From comments or YAML field
}
```

**TemplateData** - Complete data for rendering:

```go
type TemplateData struct {
    *ActionYML           // Embedded - all fields accessible directly
    Git          git.RepoInfo
    Config       *AppConfig
    UsesStatement string  // Pre-built "org/repo/path@version"
    ActionPath   string   // For subdirectory extraction
    RepoRoot     string
    Dependencies []dependencies.Dependency
}
```

### Monorepo Action Path Resolution

The tool automatically detects and handles monorepo actions:

```text
Input:  /repo/actions/csharp-build/action.yml
Root:   /repo
Output: org/repo/actions/csharp-build@main
```

Implementation in `internal/template.go:extractActionSubdirectory()`:

- Calculates relative path from repo root to action directory
- Returns empty string for root-level actions
- Returns subdirectory path for monorepo actions
- Used by `buildUsesString()` to construct proper `uses:` statements

### Permissions Parsing

Supports three sources (merged with priority):

1. **Header comments** (lowest priority):

```yaml
# permissions:
#   - contents: read
#   - issues: write
```

1. **YAML field** (highest priority):

```yaml
permissions:
  contents: write
```

1. **Merged result**: YAML overrides comment values for duplicate keys, all unique keys included

## 🛠️ Development Commands

### Building and Running

```bash
# Build binary
go build .

# Run without installing
go run . gen testdata/example-action/
go run . validate
go run . config show

# Build and run tests
make build
make test
make test-coverage          # With coverage report
make test-coverage-html     # HTML coverage + open in browser
```

### Testing

```bash
# Run all tests
go test ./...

# Run specific test package
go test ./internal
go test ./internal/wizard

# Run specific test by name
go test ./internal -run TestParsePermissions
go test ./internal -v -run "TestParseActionYML_.*Permissions"

# Run tests with race detection
go test -race ./...

# Test all themes
for theme in default github gitlab minimal professional; do
  ./gh-action-readme gen testdata/example-action/ --theme $theme --output /tmp/test-$theme.md
done
```

### Linting and Quality

```bash
# Run all linters via pre-commit
make lint

# Run golangci-lint directly
golangci-lint run
golangci-lint run --timeout=5m

# Check editor config compliance
make editorconfig

# Auto-fix editorconfig issues
make editorconfig-fix

# Format code
make format
```

### Security Scanning

```bash
# Run all security checks
make security

# Individual security tools
make vulncheck    # Go vulnerability check
make audit        # Nancy dependency audit
make trivy        # Container security scanner
make gitleaks     # Secret detection
```

### Dependencies

```bash
# Check for outdated dependencies
make deps-check

# Update dependencies interactively
make deps-update

# Update all dependencies to latest
make deps-update-all

# Install development tools
make devtools
```

### Pre-commit Hooks

```bash
# Install hooks (run once per clone)
make pre-commit-install

# Update hooks to latest versions
make pre-commit-update

# Run hooks manually
pre-commit run --all-files
```

## ⚙️ Configuration System

### Configuration Hierarchy (highest to lowest priority)

1. **Command-line flags** - Override everything
2. **Action-specific config** - `.ghreadme.yaml` in action directory
3. **Repository config** - `.ghreadme.yaml` in repo root
4. **Global config** - `~/.config/gh-action-readme/config.yaml`
5. **Environment variables** - `GH_README_GITHUB_TOKEN`, `GITHUB_TOKEN`
6. **Defaults** - Built-in fallbacks

### Version Resolution for Usage Examples

Priority order for `uses: org/repo@VERSION`:

1. `Config.Version` - Explicit override (e.g., `version: "v2.0.0"`)
2. `Config.UseDefaultBranch` + `Git.DefaultBranch` - Detected branch (e.g., `@main`)
3. Fallback - `"v1"`

Implemented in `internal/template.go:getActionVersion()`.

### Permissions Parsing

**Comment Format Support:**

- List format: `# permissions:\n#   - key: value`
- Object format: `# permissions:\n#   key: value`
- Mixed format: Both styles in same block
- Inline comments: `# key: value  # explanation`

**Merge Behavior:**

```go
// internal/parser.go:ParseActionYML()
if a.Permissions == nil && commentPermissions != nil {
    a.Permissions = commentPermissions  // Use comments
} else if a.Permissions != nil && commentPermissions != nil {
    // Merge: YAML overrides, add missing from comments
    for key, value := range commentPermissions {
        if _, exists := a.Permissions[key]; !exists {
            a.Permissions[key] = value
        }
    }
}
```

## 🔄 Adding New Features

### New Theme

1. Create template file:

```bash
touch templates_embed/templates/themes/THEME_NAME/readme.tmpl
```

1. Add theme constant to `appconstants/constants.go`:

```go
ThemeTHEMENAME     = "theme-name"
TemplatePathTHEMENAME = "templates/themes/theme-name/readme.tmpl"
```

**Note:** The template path constant still uses `templates/` prefix (not
`templates_embed/templates/`) as this is the logical path used by the code.
The physical file lives in `templates_embed/templates/` but is referenced as
`templates/` in the code.

1. Add to theme resolver in `internal/config.go:resolveThemeTemplate()`:

```go
case appconstants.ThemeTHEMENAME:
    templatePath = appconstants.TemplatePathTHEMENAME
```

1. Update `main.go:configThemesHandler()` to list the new theme

2. Rebuild: `go build .`

### New Output Format

1. Add format constant to `appconstants/constants.go`

2. Add case in `internal/generator.go:GenerateFromFile()`:

```go
case appconstants.OutputFormatNEW:
    return g.generateNEW(action, outputDir, actionPath)
```

1. Implement generator method:

```go
func (g *Generator) generateNEW(action *ActionYML, outputDir, actionPath string) error {
    opts := TemplateOptions{
        TemplatePath: g.resolveTemplatePathForFormat(),
        Format:       "new",
    }
    // ... implementation
}
```

1. Update CLI help text in `main.go`

### New Template Function

Add to `internal/template.go:templateFuncs()`:

```go
func templateFuncs() template.FuncMap {
    return template.FuncMap{
        "myFunc": myFuncImplementation,
        // ... existing functions
    }
}
```

### New Parser Field

When adding fields to `ActionYML`:

1. Update struct in `internal/parser.go`
2. Update `ActionYMLForJSON` in `internal/json_writer.go` (for JSON output)
3. Add field to JSON struct initialization
4. Add tests in `internal/parser_test.go`
5. Update templates if field should be displayed

## 📊 Package Structure

- **`main.go`** - CLI entry point (Cobra commands)
- **`internal/generator.go`** - Core generation orchestration
- **`internal/parser.go`** - Action.yml parsing (including permissions from comments)
- **`internal/template.go`** - Template data building and rendering
- **`internal/config.go`** - Configuration management (Viper)
- **`internal/json_writer.go`** - JSON output format
- **`internal/output.go`** - Colored CLI output
- **`internal/progress.go`** - Progress bars for batch operations
- **`internal/git/`** - Git repository detection
- **`internal/validation/`** - Action.yml validation
- **`internal/wizard/`** - Interactive configuration wizard
- **`internal/dependencies/`** - Dependency analysis for actions
- **`internal/errors/`** - Contextual error handling
- **`appconstants/`** - Application constants
- **`testutil/`** - Testing utilities
- **`templates_embed/`** - Embedded template filesystem

## 🧪 Testing Guidelines

### Test File Locations

- Unit tests: `internal/*_test.go` alongside source files
- Test fixtures: `testdata/yaml-fixtures/` (organized by type)
- Integration tests: Manual CLI testing with testdata

### Test Fixture Organization

**CRITICAL:** Always use fixtures, never inline YAML in tests.

**Fixture Structure:**

```text
testdata/yaml-fixtures/
├── actions/
│   ├── composite/         # Composite actions
│   ├── javascript/        # JavaScript actions
│   ├── docker/           # Docker actions
│   └── invalid/          # Invalid actions for error testing
├── dependencies/         # Actions with specific dependencies
├── configs/             # Configuration files
└── error-scenarios/     # Edge cases and error conditions
```

**Using Fixtures in Tests:**

```go
// Use fixture constants from testutil/test_constants.go
testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureCompositeBasic)

// Available fixture constants:
// - TestFixtureJavaScriptSimple
// - TestFixtureCompositeBasic
// - TestFixtureCompositeWithDeps
// - TestFixtureCompositeMultipleNamedSteps
// - TestFixtureActionWithCheckoutV3/V4
// See testutil/test_constants.go for complete list
```

**Adding New Fixtures:**

1. Create YAML file in appropriate subdirectory: `testdata/yaml-fixtures/actions/composite/my-new-fixture.yml`
2. Add constant to `testutil/test_constants.go`: `TestFixtureMyNewFixture = "actions/composite/my-new-fixture.yml"`
3. Use in tests: `testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureMyNewFixture)`

### Running Specific Tests

```bash
# Parser tests
go test ./internal -v -run TestParse

# Permissions tests
go test ./internal -run ".*Permissions"

# Template tests
go test ./internal -run ".*Template|.*Uses"

# Generator tests
go test ./internal -run "TestGenerator"

# Wizard tests
go test ./internal/wizard -v
```

### Template Testing After Updates

```bash
# 1. Rebuild with updated templates
go build .

# 2. Test all themes
for theme in default github gitlab minimal professional; do
  ./gh-action-readme gen testdata/example-action/ --theme $theme --output /tmp/test-$theme.md
  echo "=== $theme theme ==="
  grep -i "permissions" /tmp/test-$theme.md && echo "✅ Found" || echo "❌ Missing"
done

# 4. Test JSON output
./gh-action-readme gen testdata/example-action/ -f json -o /tmp/test.json
cat /tmp/test.json | python3 -m json.tool | grep -A 3 permissions
```

## 📦 Dependency Management

**Automated Updates:**

- Renovate bot runs weekly (Mondays 4am UTC)
- Auto-merges minor/patch updates
- Major updates require manual review
- Groups `golang.org/x` packages together

**Manual Updates:**

```bash
make deps-check          # Show outdated
make deps-update         # Interactive with go-mod-upgrade
make deps-update-all     # Update all to latest
```

## 🔐 Security

**Pre-commit Hooks:**

- `gitleaks` - Secret detection
- `golangci-lint` - Static analysis including security checks
- `editorconfig-checker` - File format validation

**Security Scanning:**

- CodeQL analysis on push/PR
- Go vulnerability check (govulncheck)
- Trivy container scanning
- Nancy dependency audit

**Path Validation:**
All file reads use validated paths to prevent path traversal:

```go
// templates_embed/embed.go:ReadTemplate()
cleanPath := filepath.Clean(templatePath)
if cleanPath != templatePath || strings.Contains(cleanPath, "..") {
    return nil, filepath.ErrBadPattern
}
```
