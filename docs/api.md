# API Reference

Complete command-line interface reference for gh-action-readme.

## 📋 Command Overview

```bash
gh-action-readme [command] [flags]
```

### Available Commands

- **`gen`** - Generate documentation from action.yml files
- **`validate`** - Validate action.yml files with suggestions
- **`schema`** - Show action.yml JSON schema
- **`config`** - Configuration management commands
- **`cache`** - Cache management commands
- **`deps`** - Dependency analysis commands
- **`version`** - Show version information
- **`about`** - Show about information
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
| ------ | ------- | ------ | --------- | ------------- |
| `--output-format` | `-f` | string | `md` | Output format: md, html, json, asciidoc |
| `--output-dir` | `-o` | string | `.` | Output directory for generated files |
| `--output` | | string | | Custom output filename (overrides default naming) |

#### Theme Options

| Flag      | Short | Type   | Default   | Description                                              |
| --------- | ----- | ------ | --------- | -------------------------------------------------------- |
| `--theme` | `-t`  | string | `default` | Theme: github, gitlab, minimal, professional, default    |

#### Processing Options

| Flag | Short | Type | Default | Description |
| ------ | ------- | ------ | --------- | ------------- |
| `--recursive` | `-r` | boolean | `false` | Search directories recursively for action.yml files |
| `--ignore-dirs` | | string | | Comma-separated list of directory names to ignore |

> Note: `-v/--verbose` and `-q/--quiet` are root persistent flags available on all commands.

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

# With GitHub token for enhanced features (use env var)
export GH_README_GITHUB_TOKEN=ghp_xxxx
gh-action-readme gen --verbose

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

The `validate` command has no own flags. Validation always runs recursively. Use the root persistent flags `-v/--verbose` and `-q/--quiet` as needed.

### Examples

```bash
# Validate current directory
gh-action-readme validate

# Validate specific file
gh-action-readme validate action.yml

# Verbose validation with suggestions
gh-action-readme validate --verbose
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
gh-action-readme config init
```

#### `show` - Display Configuration

```bash
gh-action-readme config show
```

**Examples:**

```bash
# Show all configuration
gh-action-readme config show
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

**Example:**

```bash
gh-action-readme config wizard --format json --output config.json
```

## ℹ️ Information Commands

### Version Command

```bash
gh-action-readme version
```

**Output:**

```text
gh-action-readme version 1.2.0
Built: 2025-08-07T10:30:00Z
Commit: a1b2c3d
Go: go1.26.4
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
| ------ | ------- | ------ | --------- | ------------- |
| `--config` | | string | | Custom configuration file path |
| `--help` | `-h` | boolean | `false` | Show help for command |
| `--quiet` | `-q` | boolean | `false` | Suppress non-error output |
| `--verbose` | `-v` | boolean | `false` | Enable verbose logging |

## 📊 Exit Codes

| Code | Description |
| ------ | ------------- |
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

- `GH_README_GITHUB_TOKEN` - GitHub personal access token
- `GITHUB_TOKEN` - GitHub personal access token (fallback)

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
gh-action-readme config show

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
