# CLAUDE.md - Development Guide

**gh-action-readme** - CLI tool for GitHub Actions documentation generation

## 🚨 CRITICAL: README Protection

**NEVER overwrite `/README.md`** - The root README.md is the main project documentation.

**For testing generation commands:**
```bash
cd testdata/
../gh-action-readme gen [options]
```

## 🏗️ Architecture

**Core Components:**
- `main.go` - CLI with Cobra framework
- `internal/generator.go` - Core generation logic
- `internal/config.go` - Viper configuration (XDG compliant)
- `internal/output.go` - Colored terminal output
- `internal/json_writer.go` - JSON format support

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
gh-action-readme gen [flags]          # Generate documentation
gh-action-readme validate            # Validate action.yml files
gh-action-readme config {init|show|themes}  # Configuration management
gh-action-readme version             # Show version
gh-action-readme about               # About tool
```

**Key Flags:**
- `--theme` - Select template theme
- `--output-format` - Choose format (md, html, json, asciidoc)
- `--recursive` - Process directories recursively
- `--verbose` - Detailed output
- `--quiet` - Suppress output

## 🔧 Development Workflow

**Build:** `go build .`
**Test:** `go test ./internal`
**Lint:** `golangci-lint run`

**Testing Generation (SAFE):**
```bash
cd testdata/example-action/
../../gh-action-readme gen --theme github
```

## 📊 Feature Matrix

| Feature | Status | Files |
|---------|--------|-------|
| CLI Framework | ✅ | `main.go` |
| File Discovery | ✅ | `generator.go:174` |
| Template Themes | ✅ | `templates/themes/` |
| Output Formats | ✅ | `generator.go:67-78` |
| Validation | ✅ | `internal_validator.go` |
| Configuration | ✅ | `config.go` |
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
# Core functionality
cd testdata/ && ../gh-action-readme gen

# All themes
for theme in github gitlab minimal professional; do
  cd testdata/ && ../gh-action-readme gen --theme $theme
done

# All formats
for format in md html json asciidoc; do
  cd testdata/ && ../gh-action-readme gen --output-format $format
done
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
- Binary-relative template paths
- Efficient file discovery
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

---

**Status: PRODUCTION READY ✅**
*All core features implemented and tested.*

