# gh-action-readme

![GitHub](https://img.shields.io/badge/GitHub%20Action-Documentation%20Generator-blue) ![License](https://img.shields.io/badge/license-MIT-green) ![Go](https://img.shields.io/badge/Go-1.22+-00ADD8) ![Status](https://img.shields.io/badge/status-production%20ready-brightgreen)

> **The definitive CLI tool for generating beautiful documentation from GitHub Actions `action.yml` files**

Transform your GitHub Actions into professional documentation with multiple themes, output formats, and enterprise-grade features.

## ✨ Features

🎨 **5 Beautiful Themes** - GitHub, GitLab, Minimal, Professional, Default
📄 **4 Output Formats** - Markdown, HTML, JSON, AsciiDoc
🎯 **Smart Validation** - Context-aware suggestions for fixing action.yml files
🚀 **Modern CLI** - Colored output, progress bars, comprehensive help
⚙️ **Enterprise Ready** - XDG-compliant configuration, recursive processing
🔧 **Developer Friendly** - Template customization, batch operations

## 🚀 Quick Start

### Installation

#### 📦 Binary Releases (Recommended)

Download pre-built binaries for your platform:

```bash
# Linux x86_64
curl -L https://github.com/ivuorinen/gh-action-readme/releases/latest/download/gh-action-readme_Linux_x86_64.tar.gz | tar -xz

# macOS x86_64 (Intel)
curl -L https://github.com/ivuorinen/gh-action-readme/releases/latest/download/gh-action-readme_Darwin_x86_64.tar.gz | tar -xz

# macOS ARM64 (Apple Silicon)
curl -L https://github.com/ivuorinen/gh-action-readme/releases/latest/download/gh-action-readme_Darwin_arm64.tar.gz | tar -xz

# Windows x86_64 (PowerShell)
Invoke-WebRequest -Uri "https://github.com/ivuorinen/gh-action-readme/releases/latest/download/gh-action-readme_Windows_x86_64.zip" -OutFile "gh-action-readme.zip"
Expand-Archive gh-action-readme.zip
```

#### 🍺 Package Managers

```bash
# macOS with Homebrew
brew install ivuorinen/tap/gh-action-readme

# Windows with Scoop
scoop bucket add ivuorinen https://github.com/ivuorinen/scoop-bucket
scoop install gh-action-readme

# Using Go
go install github.com/ivuorinen/gh-action-readme@latest
```

#### 🐳 Docker

```bash
# Run directly with Docker
docker run --rm -v $(pwd):/workspace ghcr.io/ivuorinen/gh-action-readme:latest gen

# Or use as base image
FROM ghcr.io/ivuorinen/gh-action-readme:latest
```

#### 🔨 Build from Source

```bash
git clone https://github.com/ivuorinen/gh-action-readme.git
cd gh-action-readme
go build .
```

### Basic Usage

```bash
# Generate README.md from action.yml
gh-action-readme gen

# Use GitHub theme with badges and collapsible sections
gh-action-readme gen --theme github

# Generate JSON for API integration
gh-action-readme gen --output-format json

# Process all action.yml files recursively
gh-action-readme gen --recursive --theme professional
```

## 📋 Examples

### Input: `action.yml`
```yaml
name: My Action
description: Does something awesome
inputs:
  token:
    description: GitHub token
    required: true
  environment:
    description: Target environment
    default: production
outputs:
  result:
    description: Action result
runs:
  using: node20
  main: index.js
```

### Output: Professional README.md

The tool generates comprehensive documentation including:
- 📊 **Parameter tables** with types, requirements, defaults
- 💡 **Usage examples** with proper YAML formatting
- 🎨 **Badges** for marketplace visibility
- 📚 **Multiple sections** (Overview, Configuration, Examples, Troubleshooting)
- 🔗 **Navigation** with table of contents

## 🎨 Themes

| Theme | Description | Best For |
|-------|-------------|----------|
| **github** | Badges, tables, collapsible sections | GitHub marketplace |
| **gitlab** | GitLab CI/CD focused examples | GitLab repositories |
| **minimal** | Clean, concise documentation | Simple actions |
| **professional** | Comprehensive with troubleshooting | Enterprise use |
| **default** | Original simple template | Basic needs |

## 📄 Output Formats

| Format | Description | Use Case |
|--------|-------------|----------|
| **md** | Markdown (default) | GitHub README files |
| **html** | Styled HTML | Web documentation |
| **json** | Structured data | API integration |
| **asciidoc** | AsciiDoc format | Technical docs |

## 🛠️ Commands

### Generation
```bash
gh-action-readme gen [flags]
  -f, --output-format string   md, html, json, asciidoc (default "md")
  -o, --output-dir string      output directory (default ".")
  -t, --theme string           github, gitlab, minimal, professional
  -r, --recursive              search recursively
```

### Validation
```bash
gh-action-readme validate
# Validates action.yml files with helpful suggestions
```

### Configuration
```bash
gh-action-readme config init     # Create default config
gh-action-readme config show     # Show current settings
gh-action-readme config themes   # List available themes
```

## ⚙️ Configuration

Create persistent settings with XDG-compliant configuration:

```bash
gh-action-readme config init
```

Configuration file (`~/.config/gh-action-readme/config.yaml`):
```yaml
theme: github
output_format: md
output_dir: .
verbose: false
```

**Environment Variables:**
```bash
export GH_ACTION_README_THEME=github
export GH_ACTION_README_VERBOSE=true
```

## 🎯 Advanced Usage

### Batch Processing
```bash
# Process multiple repositories
find . -name "action.yml" -execdir gh-action-readme gen --theme github \;

# Recursive processing with JSON output
gh-action-readme gen --recursive --output-format json --output-dir docs/
```

### Custom Themes
```bash
# Copy and modify existing theme
cp -r templates/themes/github templates/themes/custom
# Edit templates/themes/custom/readme.tmpl
gh-action-readme gen --theme custom
```

### Validation with Suggestions
```bash
gh-action-readme validate --verbose
# ❌ Missing required field: description
# 💡 Add 'description: Brief description of what your action does'
```

## 🏗️ Development

### Prerequisites
- Go 1.22+
- golangci-lint

### Build
```bash
go build .
go test ./internal
golangci-lint run
```

### Code Quality
This project maintains high code quality standards:

- ✅ **0 linting violations** - Clean, maintainable codebase
- ✅ **Comprehensive test coverage** - 80%+ coverage across critical modules
- ✅ **Low cyclomatic complexity** - All functions under 10 complexity
- ✅ **Minimal code duplication** - Shared utilities and helper functions
- ✅ **Proper error handling** - All errors properly acknowledged and handled
- ✅ **Standardized formatting** - `gofmt` and `goimports` applied consistently

**Recent Improvements (2025-07-24)**:
- Extracted common functionality into `internal/helpers/` package
- Simplified template path resolution and git operations
- Refactored complex test functions for better maintainability
- Fixed all linting issues including error handling and unused parameters

### Testing
```bash
# Test generation (safe - uses testdata/)
cd testdata/example-action/
../../gh-action-readme gen --theme github

# Run full test suite
go test ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🤝 Contributing

Contributions welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md).

**Quick Start:**
1. Fork the repository
2. Create a feature branch
3. Make changes (see [CLAUDE.md](CLAUDE.md) for development guide)
4. Add tests
5. Submit pull request

## 📊 Comparison

| Feature | gh-action-readme | action-docs | gh-actions-auto-docs |
|---------|------------------|-------------|----------------------|
| **Themes** | 5 themes | 1 basic | 1 basic |
| **Output Formats** | 4 formats | 1 format | 1 format |
| **Validation** | Smart suggestions | Basic | None |
| **Configuration** | XDG compliant | None | Basic |
| **CLI UX** | Modern + colors | Basic | Basic |
| **Templates** | Customizable | Fixed | Fixed |

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) for CLI framework
- [Viper](https://github.com/spf13/viper) for configuration management
- GitHub Actions community for inspiration

---

<div align="center">
  <sub>Built with ❤️ by <a href="https://github.com/ivuorinen">ivuorinen</a></sub>
</div>

