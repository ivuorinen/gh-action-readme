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
gh-action-readme validate [path] [flags]
```

### Arguments

- **`[path]`** - Optional file or directory to validate
  - Defaults to the current working directory when omitted
  - A nonexistent path is an error (the command does not silently fall back to the current directory)
  - Validation always runs recursively from the resolved path

### Flags

The `validate` command has no own flags. Validation always runs recursively. Use the root persistent flags `-v/--verbose` and `-q/--quiet` as needed.

### Examples

```bash
# Validate the current directory tree
gh-action-readme validate

# Validate a specific directory
gh-action-readme validate ./actions/checkout/

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

## 📦 Dependency Commands

```bash
gh-action-readme deps [subcommand] [flags]
```

Analyze and manage the GitHub Actions referenced (`uses:`) by composite action files,
discovered recursively from the current working directory. Most subcommands require a
GitHub token (see [GitHub Integration](#github-integration)).

### Subcommands

#### `list` - List all dependencies

```bash
gh-action-readme deps list
```

Lists every dependency found in discovered action files, marking pinned versus floating versions.

#### `security` - Analyze dependency security

```bash
gh-action-readme deps security
```

Reports how many dependencies are pinned (to commit SHAs) versus floating, and lists the floating ones that should be pinned.

#### `outdated` - Check for outdated dependencies

```bash
gh-action-readme deps outdated
```

Checks each dependency against the latest available release.

#### `upgrade` - Upgrade dependencies

```bash
gh-action-readme deps upgrade [flags]
```

**⚠️ Destructive:** applies changes by rewriting your `action.yml`/`action.yaml` files.
Use `--dry-run` to preview first.

| Flag | Type | Default | Description |
| ------ | ------ | --------- | ------------- |
| `--ci` | boolean | `false` | CI/CD mode: automatically pin all updates to commit SHAs |
| `--all` | boolean | `false` | Update all outdated dependencies without prompts |
| `--dry-run` | boolean | `false` | Show what would be updated without making changes |

#### `pin` - Pin floating versions

```bash
gh-action-readme deps pin [flags]
```

**⚠️ Destructive:** converts floating versions (like `@v4`) to pinned commit SHAs with
version comments, rewriting your action files. Use `--dry-run` to preview first.

| Flag | Type | Default | Description |
| ------ | ------ | --------- | ------------- |
| `--all` | boolean | `false` | Pin all floating dependencies |
| `--dry-run` | boolean | `false` | Show what would be pinned without making changes |

## 🗃️ Cache Commands

```bash
gh-action-readme cache [subcommand]
```

Manage the XDG-compliant dependency analysis cache.

### Subcommands

#### `clear` - Clear the cache

```bash
gh-action-readme cache clear
```

#### `stats` - Show cache statistics

```bash
gh-action-readme cache stats
```

Prints cache location, total entries, expired entries, and total size.

#### `path` - Show the cache directory path

```bash
gh-action-readme cache path
```

> Note: `cache stats` and `cache path` intentionally print their primary output (the
> statistics and the path) even under `--quiet`, so they remain usable in scripts.

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
| `--quiet` | `-q` | boolean | `false` | Suppress non-error output (but `cache stats` and `cache path` still emit their primary output — see note below) |
| `--verbose` | `-v` | boolean | `false` | Enable verbose logging |

> Note: `--quiet` suppresses decorative/informational messages, but the `cache stats` and
> `cache path` subcommands deliberately print their primary output (statistics and path) even
> under `--quiet` so they stay usable in scripts.

## 📊 Exit Codes

| Code | Description |
| ------ | ------------- |
| `0` | Success |
| `1` | Any error (the CLI does not differentiate error categories by exit code) |

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
