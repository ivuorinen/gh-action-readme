# gh-action-readme

![GitHub](https://img.shields.io/badge/GitHub%20Action-Documentation%20Generator-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Go](https://img.shields.io/badge/Go-1.23+-00ADD8)
![Status](https://img.shields.io/badge/status-production%20ready-brightgreen)

[![Security](https://img.shields.io/badge/security-hardened-brightgreen)](SECURITY.md)
[![Go Vulnerability Check](https://github.com/ivuorinen/gh-action-readme/actions/workflows/security.yml/badge.svg)](https://github.com/ivuorinen/gh-action-readme/actions/workflows/security.yml)
[![CodeQL](https://github.com/ivuorinen/gh-action-readme/actions/workflows/codeql.yml/badge.svg)](https://github.com/ivuorinen/gh-action-readme/actions/workflows/codeql.yml)

> **The definitive CLI tool for generating beautiful documentation from GitHub Actions `action.yml` files**

Transform your GitHub Actions into professional documentation with multiple themes, output formats, and enterprise-grade features.

## ✨ Features

🎨 **5 Beautiful Themes** - GitHub, GitLab, Minimal, Professional, Default
📄 **4 Output Formats** - Markdown, HTML, JSON, AsciiDoc
🎯 **Smart Validation** - Context-aware suggestions for fixing action.yml files
🚀 **Modern CLI** - Colored output, progress bars, comprehensive help
⚙️ **Enterprise Ready** - XDG-compliant configuration, recursive processing
🔧 **Developer Friendly** - Template customization, batch operations
📁 **Flexible Targeting** - Directory/file arguments, custom output filenames
🛡️ **Thread Safe** - Race condition protection, concurrent processing ready

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
# Generate README.md from action.yml in current directory
gh-action-readme gen

# Target specific directories or files
gh-action-readme gen testdata/example-action/
gh-action-readme gen testdata/composite-action/action.yml

# Use GitHub theme with custom output filename
gh-action-readme gen --theme github --output custom-readme.md

# Generate JSON for API integration with custom filename
gh-action-readme gen --output-format json --output api-docs.json

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
gh-action-readme gen [directory_or_file] [flags]
  -f, --output-format string   md, html, json, asciidoc (default "md")
  -o, --output-dir string      output directory (default ".")
      --output string          custom output filename
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
gh-action-readme config wizard   # Interactive configuration wizard
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
# Process multiple repositories with custom outputs
find . -name "action.yml" -execdir gh-action-readme gen --theme github --output README-generated.md \;

# Recursive processing with JSON output and custom directory structure
gh-action-readme gen --recursive --output-format json --output-dir docs/

# Target multiple specific actions with different themes
gh-action-readme gen actions/checkout/ --theme github --output docs/checkout.md
gh-action-readme gen actions/setup-node/ --theme professional --output docs/setup-node.md
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

**Recent Improvements (August 6, 2025)**:
- **Enhanced Gen Command**: Added directory/file targeting with `--output` flag for custom filenames
- **Thread Safety**: Implemented RWMutex synchronization for race condition protection
- **GitHub Actions Integration**: Enhanced CI workflow showcasing all new gen command features
- **Code Quality**: Achieved zero linting violations with complete EditorConfig compliance
- **Architecture**: Added contextual error handling, interactive wizard, and progress indicators

### Testing
```bash
# Test generation (safe - uses testdata/)
gh-action-readme gen testdata/example-action/ --theme github --output test-output.md
gh-action-readme gen testdata/composite-action/action.yml --theme professional

# Run full test suite
go test ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🔒 Security

gh-action-readme follows security best practices with comprehensive vulnerability scanning and protection measures:

### Automated Security Scanning
- **govulncheck**: Go-specific vulnerability detection
- **Snyk**: Dependency vulnerability analysis
- **Trivy**: Container and filesystem security scanning
- **gitleaks**: Secrets detection and prevention
- **CodeQL**: Static code analysis for security issues
- **Dependabot**: Automated dependency updates

### Local Security Testing
```bash
# Run all security scans
make security

# Individual security checks
make vulncheck  # Go vulnerability scanning
make snyk       # Dependency analysis
make trivy      # Filesystem scanning
make gitleaks   # Secrets detection
make audit      # Comprehensive security audit
```

### Security Policy
For reporting security vulnerabilities, please see our [Security Policy](SECURITY.md).

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
