# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- GoReleaser configuration for automated releases
- Multi-platform binary builds (Linux, macOS, Windows)
- Docker images with multi-architecture support
- Homebrew formula for easy installation on macOS
- Scoop bucket for Windows package management
- Binary signing with cosign
- SBOM (Software Bill of Materials) generation
- Enhanced version command with build information

### Changed

- Updated GitHub Actions workflow for automated releases
- Improved release process with GoReleaser

### Infrastructure

- Added Dockerfile for containerized deployments
- Set up automated Docker image publishing to GitHub Container Registry
- Added support for ARM64 and AMD64 architectures

## [0.1.0] - Initial Release

### Added

- Core CLI framework with Cobra
- Documentation generation from action.yml files
- Multiple output formats (Markdown, HTML, JSON, AsciiDoc)
- Five beautiful themes (default, github, gitlab, minimal, professional)
- Smart validation with helpful error messages
- XDG-compliant configuration system
- Recursive file processing
- Colored terminal output with progress indicators
- Advanced dependency analysis system
- GitHub API integration with rate limiting
- Security analysis (pinned vs floating versions)
- Dependency upgrade automation
- CI/CD mode for automated updates
- Comprehensive test suite (80%+ coverage)
- Zero linting violations

### Features

- **CLI Commands**: gen, validate, schema, version, about, config, deps, cache
- **Configuration**: Multi-level hierarchy with hidden config files
- **Dependency Management**: Outdated detection, security analysis, version pinning
- **Template System**: Customizable themes with rich dependency information
- **GitHub Integration**: API client with caching and rate limiting
- **Cache Management**: XDG-compliant caching with TTL support

### Quality

- Comprehensive code quality improvements
- Extracted helper functions for code deduplication
- Reduced cyclomatic complexity in all functions
- Proper error handling throughout codebase
- Standardized formatting with gofmt and goimports

[Unreleased]: https://github.com/ivuorinen/gh-action-readme/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ivuorinen/gh-action-readme/releases/tag/v0.1.0
