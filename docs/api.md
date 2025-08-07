# API Reference

Complete command-line interface reference for gh-action-readme.

## 📋 Command Overview

```bash
gh-action-readme [command] [flags]
```

### Available Commands

- **`gen`** - Generate documentation from action.yml files
- **`validate`** - Validate action.yml files with suggestions
- **`config`** - Configuration management commands
- **`version`** - Show version information
- **`help`** - Help about any command

## 🚀 Generation Command

### Basic Syntax

```bash
gh-action-readme gen [directory_or_file] [flags]
```

### Arguments

- **`[directory_or_file]`** - Optional path to action.yml file or directory containing one
  - If omitted, searches current directory for `action.yml` or `action.yaml`
  - Supports both files and directories
  - Examples: `action.yml`, `./actions/checkout/`, `/path/to/action/`

### Flags

#### Output Options

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output-format` | `-f` | string | `md` | Output format: md, html, json, asciidoc |
| `--output-dir` | `-o` | string | `.` | Output directory for generated files |
| `--output` | | string | | Custom output filename (overrides default naming) |

#### Theme Options

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--theme` | `-t` | string | `default` | Theme: github, gitlab, minimal, professional, default |

#### Processing Options

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--recursive` | `-r` | boolean | `false` | Search directories recursively for action.yml files |
| `--quiet` | `-q` | boolean | `false` | Suppress progress output |
| `--verbose` | `-v` | boolean | `false` | Enable verbose logging |

#### GitHub Integration

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--github-token` | | string | | GitHub personal access token (or use GITHUB_TOKEN env) |
| `--no-dependencies` | | boolean | `false` | Disable dependency analysis |

### Examples

#### Basic Generation

```bash
# Generate with default settings
gh-action-readme gen

# Generate from specific file
gh-action-readme gen action.yml

# Generate from directory
gh-action-readme gen ./actions/checkout/
```

#### Output Formats

```bash
# Markdown (default)
gh-action-readme gen --output-format md

# HTML documentation
gh-action-readme gen --output-format html

# JSON metadata
gh-action-readme gen --output-format json

# AsciiDoc format
gh-action-readme gen --output-format asciidoc
```

#### Custom Output

```bash
# Custom filename
gh-action-readme gen --output custom-readme.md

# Custom directory
gh-action-readme gen --output-dir docs/

# Both custom directory and filename
gh-action-readme gen --output-dir docs/ --output action-guide.html
```

#### Themes

```bash
# GitHub marketplace theme
gh-action-readme gen --theme github

# GitLab CI/CD theme
gh-action-readme gen --theme gitlab

# Clean minimal theme
gh-action-readme gen --theme minimal

# Comprehensive professional theme
gh-action-readme gen --theme professional
```

#### Advanced Options

```bash
# Recursive processing
gh-action-readme gen --recursive --theme github

# With GitHub token for enhanced features
gh-action-readme gen --github-token ghp_xxxx --verbose

# Quiet mode for scripts
gh-action-readme gen --theme github --quiet
```

## ✅ Validation Command

### Basic Syntax

```bash
gh-action-readme validate [file_or_directory] [flags]
```

### Arguments

- **`[file_or_directory]`** - Optional path to validate
  - If omitted, validates current directory
  - Supports both files and directories

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--verbose` | `-v` | boolean | `false` | Show detailed validation messages |
| `--quiet` | `-q` | boolean | `false` | Only show errors, suppress warnings |
| `--recursive` | `-r` | boolean | `false` | Validate recursively |

### Examples

```bash
# Validate current directory
gh-action-readme validate

# Validate specific file
gh-action-readme validate action.yml

# Verbose validation with suggestions
gh-action-readme validate --verbose

# Recursive validation
gh-action-readme validate --recursive ./actions/
```

### Validation Output

```text
✅ action.yml is valid
⚠️  Warning: Missing 'branding' section for marketplace visibility
💡 Consider adding:
  branding:
    icon: 'activity'
    color: 'blue'

❌ Error: Missing required field 'description'
💡 Add: description: "Brief description of what your action does"
```

## ⚙️ Configuration Commands

### Basic Syntax

```bash
gh-action-readme config [subcommand] [flags]
```

### Subcommands

#### `init` - Initialize Configuration

```bash
gh-action-readme config init [flags]
```

**Flags:**

- `--force` - Overwrite existing configuration
- `--global` - Create global configuration (default: user-specific)

#### `show` - Display Configuration

```bash
gh-action-readme config show [key] [flags]
```

**Examples:**

```bash
# Show all configuration
gh-action-readme config show

# Show specific key
gh-action-readme config show theme

# Show with file paths
gh-action-readme config show --paths
```

#### `themes` - List Available Themes

```bash
gh-action-readme config themes
```

**Output:**

```text
Available themes:
  github        GitHub marketplace optimized theme
  gitlab        GitLab CI/CD focused theme
  minimal       Clean, minimal documentation
  professional  Comprehensive enterprise theme
  default       Original simple theme
```

#### `wizard` - Interactive Configuration

```bash
gh-action-readme config wizard [flags]
```

**Flags:**

- `--format` - Export format: yaml (default), json, toml
- `--output` - Output file path
- `--no-github-token` - Skip GitHub token setup

**Example:**

```bash
gh-action-readme config wizard --format json --output config.json
```

#### `set` - Set Configuration Value

```bash
gh-action-readme config set <key> <value>
```

**Examples:**

```bash
gh-action-readme config set theme github
gh-action-readme config set verbose true
gh-action-readme config set output_format html
```

#### `get` - Get Configuration Value

```bash
gh-action-readme config get <key>
```

#### `reset` - Reset Configuration

```bash
gh-action-readme config reset [key]
```

**Examples:**

```bash
# Reset all configuration
gh-action-readme config reset

# Reset specific key
gh-action-readme config reset theme
```

## ℹ️ Information Commands

### Version Command

```bash
gh-action-readme version [flags]
```

**Flags:**

- `--short` - Show version number only
- `--json` - Output in JSON format

**Output:**

```text
gh-action-readme version 1.2.0
Built: 2025-08-07T10:30:00Z
Commit: a1b2c3d
Go: go1.24.4
Platform: linux/amd64
```

### Help Command

```bash
gh-action-readme help [command]
```

**Examples:**

```bash
# General help
gh-action-readme help

# Command-specific help
gh-action-readme help gen
gh-action-readme help config wizard
```

## 🌍 Global Flags

These flags are available for all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--config` | | string | | Custom configuration file path |
| `--help` | `-h` | boolean | `false` | Show help for command |
| `--quiet` | `-q` | boolean | `false` | Suppress non-error output |
| `--verbose` | `-v` | boolean | `false` | Enable verbose logging |

## 📊 Exit Codes

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `3` | File not found |
| `4` | Validation failed |
| `5` | Configuration error |
| `6` | GitHub API error |
| `7` | Template error |

## 🔧 Environment Variables

### Configuration Override

- `GH_ACTION_README_THEME` - Default theme
- `GH_ACTION_README_OUTPUT_FORMAT` - Default output format
- `GH_ACTION_README_OUTPUT_DIR` - Default output directory
- `GH_ACTION_README_VERBOSE` - Enable verbose mode (true/false)
- `GH_ACTION_README_QUIET` - Enable quiet mode (true/false)

### GitHub Integration

- `GITHUB_TOKEN` - GitHub personal access token
- `GH_ACTION_README_NO_DEPENDENCIES` - Disable dependency analysis

### Advanced Options

- `GH_ACTION_README_CONFIG` - Custom configuration file path
- `GH_ACTION_README_CACHE_TTL` - Cache TTL in seconds
- `GH_ACTION_README_TIMEOUT` - Request timeout in seconds

## 🎯 Advanced Usage Patterns

### Batch Processing

```bash
# Process multiple actions with custom themes
for dir in actions/*/; do
  gh-action-readme gen "$dir" --theme github --output "$dir/README.md"
done

# Generate docs for all formats
for format in md html json asciidoc; do
  gh-action-readme gen --output-format "$format" --output "docs/action.$format"
done
```

### CI/CD Integration

```bash
# GitHub Actions workflow
- name: Generate Action Documentation
  run: |
    gh-action-readme gen --theme github --output README.md
    git add README.md
    git commit -m "docs: update action documentation" || exit 0

# GitLab CI
generate_docs:
  script:
    - gh-action-readme gen --theme gitlab --output-format html --output docs/
  artifacts:
    paths:
      - docs/
```

### Conditional Processing

```bash
#!/bin/bash
# Smart theme selection based on repository
if [[ -f ".gitlab-ci.yml" ]]; then
  THEME="gitlab"
elif [[ -f "package.json" ]]; then
  THEME="github"
else
  THEME="minimal"
fi

gh-action-readme gen --theme "$THEME" --verbose
```

### Error Handling

```bash
#!/bin/bash
set -e

# Generate with error handling
if gh-action-readme gen --theme github --quiet; then
  echo "✅ Documentation generated successfully"
else
  exit_code=$?
  echo "❌ Documentation generation failed (exit code: $exit_code)"

  case $exit_code in
    3) echo "💡 Make sure action.yml exists in the current directory" ;;
    4) echo "💡 Run 'gh-action-readme validate' to check for issues" ;;
    6) echo "💡 Check your GitHub token and network connection" ;;
    *) echo "💡 Run with --verbose flag for more details" ;;
  esac

  exit $exit_code
fi
```

## 🔍 Debugging & Troubleshooting

### Debug Output

```bash
# Maximum verbosity
gh-action-readme gen --verbose

# Configuration debugging
gh-action-readme config show --debug

# Validation debugging
gh-action-readme validate --verbose
```

### Common Issues

**Command not found:**

```bash
# Check installation
which gh-action-readme
gh-action-readme version
```

**Permission denied:**

```bash
# Check file permissions
ls -la action.yml
chmod 644 action.yml
```

**GitHub API rate limit:**

```bash
# Use GitHub token
export GITHUB_TOKEN=your_token_here
gh-action-readme gen --verbose
```

**Template errors:**

```bash
# Validate action.yml first
gh-action-readme validate --verbose
```
