# CLAUDE.md - Development Guide

**gh-action-readme** - CLI tool for GitHub Actions documentation generation

## 🚨 CRITICAL: README Protection

**NEVER overwrite `/README.md`** - The root README.md is the main project documentation.

**For testing generation commands:**
```bash
# New enhanced targeting (recommended)
gh-action-readme gen testdata/example-action/
gh-action-readme gen testdata/composite-action/action.yml

# Traditional method (still supported)
cd testdata/
../gh-action-readme gen [options]
```

## 🏗️ Architecture

**Core Components:**
- `main.go` - CLI with Cobra framework, enhanced gen command
- `internal/generator.go` - Core generation logic with custom output paths
- `internal/config.go` - Viper configuration (XDG compliant)
- `internal/output.go` - Colored terminal output with progress bars
- `internal/json_writer.go` - JSON format support
- `internal/errors/` - Contextual error handling with suggestions
- `internal/wizard/` - Interactive configuration wizard
- `internal/progress.go` - Progress indicators for batch operations

**Templates:**
- `templates/readme.tmpl` - Default template
- `templates/themes/` - Theme-specific templates
  - `github/` - GitHub-style with badges
  - `gitlab/` - GitLab CI/CD focused
  - `minimal/` - Clean, concise
  - `professional/` - Comprehensive with ToC
  - `asciidoc/` - AsciiDoc format

## 🛠️ Commands & Usage

**Available Commands:**
```bash
gh-action-readme gen [directory_or_file] [flags]  # Generate documentation
gh-action-readme validate            # Validate action.yml files
gh-action-readme config {init|show|themes|wizard}  # Configuration management
gh-action-readme version             # Show version
gh-action-readme about               # About tool
```

**Key Flags:**
- `--theme` - Select template theme
- `--output-format` - Choose format (md, html, json, asciidoc)
- `--output` - Custom output filename
- `--recursive` - Process directories recursively
- `--verbose` - Detailed output
- `--quiet` - Suppress output

## 🔧 Development Workflow

**Build:** `go build .`
**Test:** `go test ./internal`
**Lint:** `golangci-lint run`
**Dependencies:** `make deps-check` / `make deps-update`

**Testing Generation (SAFE):**
```bash
# Enhanced targeting (recommended)
gh-action-readme gen testdata/example-action/ --theme github --output test-output.md
gh-action-readme gen testdata/composite-action/action.yml --theme professional

# Traditional method (still works)
cd testdata/example-action/
../../gh-action-readme gen --theme github
```

## 📊 Feature Matrix

| Feature | Status | Files |
|---------|--------|-------|
| CLI Framework | ✅ | `main.go` |
| Enhanced Gen Command | ✅ | `main.go:168-180` |
| File Discovery | ✅ | `generator.go:304-324` |
| Template Themes | ✅ | `templates/themes/` |
| Output Formats | ✅ | `generator.go:168-182` |
| Custom Output Paths | ✅ | `generator.go:157-166` |
| Validation | ✅ | `internal/validation/` |
| Configuration | ✅ | `config.go`, `configuration_loader.go` |
| Interactive Wizard | ✅ | `internal/wizard/` |
| Progress Indicators | ✅ | `progress.go` |
| Contextual Errors | ✅ | `internal/errors/` |
| Colored Output | ✅ | `output.go` |

## 🎨 Themes

**Available Themes:**
1. **default** - Original simple template
2. **github** - Badges, tables, collapsible sections
3. **gitlab** - GitLab CI/CD examples
4. **minimal** - Clean, concise documentation
5. **professional** - Comprehensive with troubleshooting

## 📄 Output Formats

**Supported Formats:**
- **md** - Markdown (default)
- **html** - HTML with styling
- **json** - Structured data for APIs
- **asciidoc** - Technical documentation format

## 🧪 Testing Strategy

**Unit Tests:** `internal/*_test.go` (26.2% coverage)
**Integration:** Manual CLI testing
**Templates:** Test with `testdata/example-action/`

**Test Commands:**
```bash
# Core functionality (enhanced)
gh-action-readme gen testdata/example-action/
gh-action-readme gen testdata/composite-action/action.yml

# All themes with custom outputs
for theme in github gitlab minimal professional; do
  gh-action-readme gen testdata/example-action/ --theme $theme --output "test-${theme}.md"
done

# All formats with custom outputs
for format in md html json asciidoc; do
  gh-action-readme gen testdata/example-action/ --output-format $format --output "test.${format}"
done

# Recursive processing
gh-action-readme gen testdata/ --recursive --theme professional
```

## 🚀 Production Features

**Configuration:**
- XDG Base Directory compliant
- Environment variable support
- Theme persistence
- Multiple search paths

**Error Handling:**
- Colored error messages
- Actionable suggestions
- Context-aware validation
- Graceful fallbacks

**Performance:**
- Progress bars for batch operations
- Thread-safe fixture caching with RWMutex
- Binary-relative template paths
- Efficient file discovery
- Custom output path resolution
- Race condition protection
- Minimal dependencies

## 🔄 Adding New Features

**New Theme:**
1. Create `templates/themes/THEME_NAME/readme.tmpl`
2. Add to `resolveThemeTemplate()` in `config.go:67`
3. Update `configThemesHandler()` in `main.go:284`

**New Output Format:**
1. Add constant to `generator.go:14`
2. Add case to `GenerateFromFile()` switch `generator.go:67`
3. Implement `generate[FORMAT]()` method
4. Update CLI help in `main.go:84`

**New Template Functions:**
Add to `templateFuncs()` in `internal_template.go:19`

## 📦 Dependency Management

**Check for updates:**
```bash
make deps-check          # Show outdated dependencies
```

**Update dependencies:**
```bash
make deps-update         # Interactive updates with go-mod-upgrade
make deps-update-all     # Update all to latest versions
```

**Automated updates:**
- Renovate bot runs weekly on Mondays at 4am UTC
- Creates PRs for minor/patch updates (auto-merge enabled)
- Major updates disabled (require manual review)
- Groups golang.org/x packages together
- Runs `go mod tidy` after updates

---

**Status: ENTERPRISE READY ✅**
*Enhanced gen command, thread-safety, comprehensive testing, and enterprise features fully implemented.*

**Latest Updates (August 6, 2025):**
- ✅ Enhanced gen command with directory/file targeting
- ✅ Custom output filename support (`--output` flag)
- ✅ Thread-safe fixture management with race condition protection
- ✅ GitHub Actions workflow integration with new capabilities
- ✅ Complete linting and code quality compliance
- ✅ Zero known race conditions or threading issues
- ✅ Dependency management automation with Renovate and go-mod-upgrade
