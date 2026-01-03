# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**gh-action-readme** - CLI tool for GitHub Actions documentation generation

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
- Test fixtures: `testdata/example-action/`, `testdata/composite-action/`
- Integration tests: Manual CLI testing with testdata

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
