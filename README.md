# gh-action-readme

![GitHub](https://img.shields.io/badge/GitHub%20Action-Documentation%20Generator-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Go](https://img.shields.io/badge/Go-1.24+-00ADD8)
![Status](https://img.shields.io/badge/status-production%20ready-brightgreen)

[![Security](https://img.shields.io/badge/security-hardened-brightgreen)](docs/security.md)
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

```bash
# macOS with Homebrew
brew install ivuorinen/tap/gh-action-readme

# Using Go
go install github.com/ivuorinen/gh-action-readme@latest

# Download binary from releases
curl -L https://github.com/ivuorinen/gh-action-readme/releases/latest/download/gh-action-readme_Linux_x86_64.tar.gz | tar -xz
```

📖 **[Complete Installation Guide →](docs/installation.md)**

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

Choose from 5 built-in themes: `github`, `gitlab`, `minimal`, `professional`, `default`

📖 **[Theme Gallery & Examples →](docs/themes.md)**

## 📄 Output Formats

Supports 4 formats: `md`, `html`, `json`, `asciidoc`

## 🛠️ Commands

```bash
# Generation
gh-action-readme gen [directory_or_file] [flags]

# Validation with suggestions
gh-action-readme validate

# Interactive configuration
gh-action-readme config wizard
```

📖 **[Complete CLI Reference →](docs/api.md)**

## ⚙️ Configuration

```bash
# Interactive setup wizard
gh-action-readme config wizard

# XDG-compliant config file
gh-action-readme config init
```

📖 **[Configuration Guide →](docs/configuration.md)**

## 🎯 Advanced Usage

```bash
# Batch processing with custom themes
gh-action-readme gen --recursive --theme github --output-dir docs/

# Custom themes
cp -r templates/themes/github templates/themes/custom
gh-action-readme gen --theme custom
```

📖 **[Complete Usage Guide →](docs/usage.md)**

## 🏗️ Development

```bash
# Build and test
go build .
go test ./...
golangci-lint run
```

Maintains enterprise-grade code quality with 0 linting violations and 80%+ test coverage.

📖 **[Development Guide →](docs/development.md)**

## 🔒 Security

Comprehensive security scanning with govulncheck, Trivy, gitleaks, and CodeQL.

```bash
make security  # Run all security scans
```

📖 **[Security Policy →](docs/security.md)**

## 🤝 Contributing

Contributions welcome! Fork, create feature branch, add tests, submit PR.

📖 **[Contributing Guide →](CONTRIBUTING.md)**

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
