# Contributing to gh-action-readme

Thank you for your interest in contributing to gh-action-readme! This guide will help you get started.

## 🚀 Quick Start

1. **Fork** the repository on GitHub
2. **Clone** your fork locally
3. **Create** a feature branch from `main`
4. **Make** your changes
5. **Test** your changes
6. **Submit** a pull request

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/gh-action-readme.git
cd gh-action-readme

# Create feature branch
git checkout -b feature/my-awesome-feature

# Install development tools
make devtools

# Make your changes and test
make test
make lint

# Commit and push
git add .
git commit -m "feat: add awesome feature"
git push origin feature/my-awesome-feature
```

## 📋 Development Guidelines

### Prerequisites

- **Go 1.24+** (required)
- **golangci-lint** (for code quality)
- **pre-commit** (for git hooks)

### Setup Development Environment

```bash
# Install development tools
make devtools

# Install pre-commit hooks
make pre-commit-install

# Build and verify
make build
make test
make lint
```

### Code Quality Standards

We maintain **zero tolerance** for quality issues:

- ✅ **Zero linting violations** - All code must pass `golangci-lint`
- ✅ **EditorConfig compliance** - Follow `.editorconfig` rules
- ✅ **Test coverage >80%** - Add tests for all new functionality
- ✅ **Documentation** - Update docs for user-facing changes
- ✅ **Security** - No vulnerabilities in dependencies

### Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Test specific package
go test ./internal/generator

# Run integration tests
go test -tags=integration ./...
```

**Test your changes safely:**

```bash
# Use testdata for safe testing (never modifies main README.md)
gh-action-readme gen testdata/example-action/ --theme github
gh-action-readme gen testdata/composite-action/action.yml --output test.md
```

## 🎯 Contribution Areas

### 🎨 Themes

Create new visual themes for different platforms or use cases:

1. Copy existing theme: `cp -r templates/themes/github templates/themes/my-theme`
2. Edit template: `templates/themes/my-theme/readme.tmpl`
3. Test theme: `gh-action-readme gen --theme my-theme testdata/example-action/`
4. Add to theme list in `internal/config.go`

### 🔌 Features

Add new functionality to the core tool:

- New output formats (PDF, EPUB, etc.)
- Enhanced validation rules
- CLI commands and flags
- Template functions

### 🐛 Bug Fixes

Fix issues and improve reliability:

- Check [GitHub Issues](https://github.com/ivuorinen/gh-action-readme/issues)
- Reproduce the bug with test cases
- Fix with minimal, targeted changes
- Add regression tests

### 📚 Documentation

Improve documentation and examples:

- Update user guides in `docs/`
- Add code examples
- Improve error messages
- Write tutorials

## 📝 Pull Request Process

### Before Submitting

- [ ] Code follows style guidelines (`make lint` passes)
- [ ] Tests added for new features (`make test` passes)
- [ ] Documentation updated for user-facing changes
- [ ] No security vulnerabilities (`make security` passes)
- [ ] Commit messages follow conventional format

### PR Requirements

- **Clear title** describing the change
- **Detailed description** explaining what and why
- **Link related issues** using `Closes #123` syntax
- **Screenshots/examples** for UI changes
- **Breaking change notes** if applicable

### Review Process

1. **Automated checks** must pass (CI/CD)
2. **Code review** by maintainers
3. **Testing** on multiple platforms if needed
4. **Merge** after approval

## 🚨 Critical Guidelines

### README Protection

**NEVER modify `/README.md` directly** - it's the main project documentation.

**For testing generation:**

```bash
# ✅ Correct - target testdata
gh-action-readme gen testdata/example-action/ --output test-output.md

# ❌ Wrong - overwrites main README
gh-action-readme gen  # This modifies /README.md
```

### Security Best Practices

- **Never commit secrets** or tokens
- **Validate all inputs** especially file paths
- **Use secure defaults** in configurations
- **Handle errors properly** don't expose internals

## 🎨 Code Style

### Go Style Guide

We follow standard Go conventions plus:

```go
// Package comments for all packages
package generator

// Function comments for all exported functions
// GenerateReadme creates documentation from action.yml files.
// It processes the input with the specified theme and format.
func GenerateReadme(action *ActionYML, theme, format string) (string, error) {
    // Implementation...
}

// Use meaningful variable names
actionYML := parseActionFile(filename)
outputContent := generateFromTemplate(actionYML, theme)

// Handle all errors explicitly
if err != nil {
    return "", fmt.Errorf("failed to parse action: %w", err)
}
```

### Commit Message Format

Follow [Conventional Commits](https://conventionalcommits.org/):

```bash
# Feature additions
git commit -m "feat: add support for PDF output format"

# Bug fixes
git commit -m "fix: resolve template rendering error for empty inputs"

# Documentation
git commit -m "docs: update configuration guide with new options"

# Breaking changes
git commit -m "feat!: change CLI argument structure for consistency"
```

## 🧪 Testing Guidelines

### Test Structure

```go
func TestGenerateReadme(t *testing.T) {
    tests := []struct {
        name     string
        action   *ActionYML
        theme    string
        format   string
        want     string
        wantErr  bool
    }{
        {
            name: "successful generation with github theme",
            action: &ActionYML{
                Name: "Test Action",
                Description: "Test description",
            },
            theme: "github",
            format: "md",
            want: "# Test Action\n\nTest description\n",
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := GenerateReadme(tt.action, tt.theme, tt.format)
            if (err != nil) != tt.wantErr {
                t.Errorf("GenerateReadme() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("GenerateReadme() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## 🏆 Recognition

Contributors are recognized in:

- **GitHub contributors** list
- **Release notes** for significant contributions
- **Hall of Fame** for major features

## 💬 Community

- **Discussions**: [GitHub Discussions](https://github.com/ivuorinen/gh-action-readme/discussions)
- **Issues**: [GitHub Issues](https://github.com/ivuorinen/gh-action-readme/issues)
- **Security**: [Security Policy](docs/security.md)

## 📖 Additional Resources

- **Development Guide**: [docs/development.md](docs/development.md)
- **Architecture**: [CLAUDE.md](CLAUDE.md)
- **Usage Examples**: [docs/usage.md](docs/usage.md)
- **API Reference**: [docs/api.md](docs/api.md)

---

## Happy Contributing! 🎉

By contributing to gh-action-readme, you're helping make GitHub Actions documentation better for everyone.
