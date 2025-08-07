# Development Guide

Comprehensive guide for developing and contributing to gh-action-readme.

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

## 🛠️ Development Setup

### Prerequisites

- **Go 1.24+** (required)
- **golangci-lint** (for linting)
- **pre-commit** (for git hooks)
- **Docker** (optional, for containerized testing)

### Quick Start

```bash
# Clone repository
git clone https://github.com/ivuorinen/gh-action-readme.git
cd gh-action-readme

# Install development tools
make devtools

# Install pre-commit hooks
make pre-commit-install

# Build and test
make build
make test
make lint
```

## 🏗️ Architecture

### Core Components

- **`main.go`** - CLI with Cobra framework, enhanced gen command
- **`internal/generator.go`** - Core generation logic with custom output paths
- **`internal/config.go`** - Viper configuration (XDG compliant)
- **`internal/output.go`** - Colored terminal output with progress bars
- **`internal/errors/`** - Contextual error handling with suggestions
- **`internal/wizard/`** - Interactive configuration wizard
- **`internal/progress.go`** - Progress indicators for batch operations

### Template System

- **`templates/readme.tmpl`** - Default template
- **`templates/themes/`** - Theme-specific templates
  - `github/` - GitHub-style with badges
  - `gitlab/` - GitLab CI/CD focused
  - `minimal/` - Clean, concise
  - `professional/` - Comprehensive with ToC
  - `asciidoc/` - AsciiDoc format

### Testing Framework

- **`testutil/`** - Comprehensive testing utilities
- **`testdata/`** - Test fixtures and sample actions
- **Table-driven tests** - Consistent testing patterns
- **Mock implementations** - Isolated unit tests

## 🧪 Testing Strategy

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test ./internal/generator

# Verbose output
go test -v ./...

# Race condition detection
go test -race ./...
```

### Test Types

**Unit Tests** (`*_test.go`):

```bash
go test ./internal        # Core logic tests
go test ./testutil        # Test framework tests
```

**Integration Tests** (`*_integration_test.go`):

```bash
go test -tags=integration ./...
```

**Comprehensive Tests** (`comprehensive_test.go`):

```bash
go test ./internal -run Comprehensive
```

### Testing Best Practices

1. **Use testutil framework** for consistent test patterns
2. **Table-driven tests** for multiple scenarios
3. **Mock external dependencies** (GitHub API, filesystem)
4. **Test error conditions** and edge cases
5. **Verify thread safety** with race detection

### Test Coverage

Current coverage: **51.2%** overall

- **Cache module**: 91.4%
- **Errors module**: 91.5%
- **Git detector**: 78.4%
- **Validation**: 76.2%

Target: **>80%** coverage for all new code

## 🔧 Build System

### Makefile Targets

```bash
# Building
make build               # Build binary
make clean              # Clean build artifacts

# Testing
make test               # Run all tests
make test-coverage      # Run tests with coverage
make lint               # Run all linters

# Development
make devtools           # Install dev tools
make format             # Format code
make pre-commit-install # Install git hooks

# Security
make security           # Run security scans
make vulncheck          # Go vulnerability check

# Dependencies
make deps-check         # Check outdated dependencies
make deps-update        # Interactive dependency updates
```

### Build Configuration

```bash
# Build with version info
go build -ldflags "-X main.version=v1.0.0 -X main.commit=$(git rev-parse HEAD)"

# Cross-platform builds (handled by GoReleaser)
GOOS=linux GOARCH=amd64 go build
GOOS=darwin GOARCH=arm64 go build
GOOS=windows GOARCH=amd64 go build
```

## 📝 Code Style & Quality

### Linting Configuration

- **golangci-lint** with 35+ enabled linters
- **EditorConfig** compliance required
- **Pre-commit hooks** for automated checking
- **Zero linting violations** policy

### Code Quality Standards

- ✅ **Cyclomatic complexity** <10 for all functions
- ✅ **Test coverage** >80% for critical modules
- ✅ **Error handling** for all possible errors
- ✅ **Documentation** for all exported functions
- ✅ **Thread safety** for concurrent operations

### Formatting Rules

```bash
# Auto-format code
make format

# Check formatting
gofmt -d .
goimports -d .
```

## 🔄 Adding New Features

### New Theme

1. Create `templates/themes/THEME_NAME/readme.tmpl`
2. Add to `resolveThemeTemplate()` in `config.go`
3. Update `configThemesHandler()` in `main.go`
4. Add tests and documentation

### New Output Format

1. Add constant to `generator.go`
2. Add case to `GenerateFromFile()` switch
3. Implement `generate[FORMAT]()` method
4. Update CLI help and documentation

### New CLI Command

1. Add command in `main.go` using Cobra
2. Implement handler function
3. Add flags and validation
4. Write tests and update documentation

### New Template Functions

Add to `templateFuncs()` in `internal/template.go`:

```go
"myFunction": func(input string) string {
    // Implementation
    return processed
},
```

## 🚀 Performance Optimization

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Benchmark testing
go test -bench=. -benchmem
```

### Performance Guidelines

- **Use sync.Pool** for frequently allocated objects
- **Implement caching** for expensive operations
- **Minimize memory allocations** in hot paths
- **Concurrent processing** for I/O bound operations
- **Profile before optimizing**

## 🔐 Security Practices

### Security Scanning

```bash
make security           # Run all scans
make vulncheck         # Go vulnerabilities
make trivy             # Filesystem scan
make gitleaks          # Secrets detection
```

### Secure Coding

- **Validate all inputs** especially user-provided data
- **Use secure defaults** for all configurations
- **Sanitize file paths** to prevent directory traversal
- **Handle secrets safely** never log or expose tokens
- **Regular dependency updates** with security patches

### GitHub Token Handling

```go
// ❌ Wrong - logging token
log.Printf("Using token: %s", token)

// ✅ Correct - masking sensitive data
log.Printf("Using token: %s", maskToken(token))

func maskToken(token string) string {
    if len(token) < 8 {
        return "***"
    }
    return token[:4] + "***" + token[len(token)-4:]
}
```

## 📚 Documentation

### Documentation Standards

- **godoc comments** for all exported functions
- **README updates** for user-facing changes
- **CHANGELOG entries** following Keep a Changelog format
- **Architecture Decision Records** for significant changes

### Writing Guidelines

```go
// Package generator provides core documentation generation functionality.
//
// The generator processes GitHub Action YAML files and produces formatted
// documentation in multiple output formats (Markdown, HTML, JSON, AsciiDoc).
package generator

// GenerateReadme creates formatted documentation from an ActionYML struct.
//
// The function applies the specified theme and output format to generate
// comprehensive documentation including input/output tables, usage examples,
// and metadata sections.
//
// Parameters:
//   - action: Parsed action.yml data structure
//   - theme: Template theme name (github, gitlab, minimal, professional, default)
//   - format: Output format (md, html, json, asciidoc)
//
// Returns formatted documentation string and any processing error.
func GenerateReadme(action *ActionYML, theme, format string) (string, error) {
    // Implementation...
}
```

## 🤝 Contributing Guidelines

### Contribution Process

1. **Fork repository** and create feature branch
2. **Make changes** following code style guidelines
3. **Add tests** for new functionality
4. **Update documentation** as needed
5. **Run quality checks** (`make test lint`)
6. **Submit pull request** with clear description

### Pull Request Guidelines

- **Clear title** describing the change
- **Detailed description** of what and why
- **Link related issues** if applicable
- **Include tests** for new features
- **Update documentation** for user-facing changes
- **Ensure CI passes** all quality checks

### Code Review Process

- **Two approvals** required for merging
- **All checks must pass** (tests, linting, security)
- **Documentation review** for user-facing changes
- **Performance impact** assessment for core changes

## 🐛 Debugging

### Debug Mode

```bash
# Enable verbose output
gh-action-readme gen --verbose

# Debug configuration
gh-action-readme config show --debug

# Trace execution
TRACE=1 gh-action-readme gen
```

### Common Issues

**Template Errors:**

```bash
# Validate action.yml syntax
gh-action-readme validate --verbose

# Check template syntax
gh-action-readme gen --dry-run --verbose
```

**GitHub API Issues:**

```bash
# Check rate limits
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/rate_limit

# Test API access
gh-action-readme gen --verbose --github-token $GITHUB_TOKEN
```

**Build Issues:**

```bash
# Clean and rebuild
make clean
make build

# Check dependencies
go mod verify
go mod tidy
```

## 📊 Metrics & Monitoring

### Performance Metrics

- **Generation time** per action
- **GitHub API** request count and timing
- **Memory usage** during processing
- **Cache hit rates** for repeated operations

### Quality Metrics

- **Test coverage** percentage
- **Linting violations** count
- **Security vulnerabilities** detected
- **Documentation coverage** for public APIs

### Tracking Tools

```bash
# Performance benchmarks
go test -bench=. -benchmem ./...

# Memory profiling
go test -memprofile=mem.prof ./...

# Coverage reporting
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

**Status**: ENTERPRISE READY ✅
*Enhanced gen command, thread-safety, comprehensive testing, and enterprise features fully implemented.*
