# Configuration

Configure gh-action-readme with persistent settings, environment variables, and advanced options.

## 📁 Configuration File

Create persistent settings with XDG-compliant configuration:

```bash
gh-action-readme config init
```

### Default Location

- **Linux/macOS**: `~/.config/gh-action-readme/config.yaml`
- **Windows**: `%APPDATA%\gh-action-readme\config.yaml`

### Configuration Format

```yaml
# ~/.config/gh-action-readme/config.yaml
theme: github
output_format: md
output_dir: .
verbose: false
github_token: ""
dependencies_enabled: true
cache_ttl: 3600
```

## 🔧 Configuration Options

### Core Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `theme` | string | `default` | Default theme to use |
| `output_format` | string | `md` | Default output format |
| `output_dir` | string | `.` | Default output directory |
| `verbose` | boolean | `false` | Enable verbose logging |

### GitHub Integration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `github_token` | string | `""` | GitHub personal access token |
| `dependencies_enabled` | boolean | `true` | Enable dependency analysis |
| `rate_limit_delay` | int | `1000` | Delay between API calls (ms) |

### Performance Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `cache_ttl` | int | `3600` | Cache TTL in seconds |
| `concurrent_requests` | int | `3` | Max concurrent GitHub API requests |
| `timeout` | int | `30` | Request timeout in seconds |

## 🌍 Environment Variables

Override configuration with environment variables:

```bash
# Core settings
export GH_ACTION_README_THEME=github
export GH_ACTION_README_OUTPUT_FORMAT=html
export GH_ACTION_README_OUTPUT_DIR=docs
export GH_ACTION_README_VERBOSE=true

# GitHub settings
export GITHUB_TOKEN=your_token_here
export GH_ACTION_README_DEPENDENCIES=true

# Performance settings
export GH_ACTION_README_CACHE_TTL=7200
export GH_ACTION_README_TIMEOUT=60
```

### Environment Variable Priority

1. Command line flags (highest priority)
2. Environment variables
3. Configuration file
4. Built-in defaults (lowest priority)

## 🎛️ Interactive Configuration

Use the interactive wizard for guided setup:

```bash
gh-action-readme config wizard
```

### Wizard Features

- **Auto-detection** of project settings
- **GitHub token** setup with validation
- **Theme preview** with examples
- **Export options** (YAML, JSON, TOML)
- **Real-time validation** with suggestions

### Wizard Example

```bash
$ gh-action-readme config wizard

✨ Welcome to gh-action-readme configuration wizard!

🔍 Detected project settings:
  Repository: ivuorinen/my-action
  Language: JavaScript/TypeScript

📋 Select your preferences:
  Theme: github, gitlab, minimal, professional, default
  >> github

  Output format: md, html, json, asciidoc
  >> md

🔑 GitHub Token (optional, for enhanced features):
  >> ghp_xxxxxxxxxxxx
  ✅ Token validated successfully!

💾 Export configuration as:
  Format: yaml, json, toml
  >> yaml

✅ Configuration saved to ~/.config/gh-action-readme/config.yaml
```

## 🎨 Theme Configuration

### Built-in Themes

```bash
# List available themes
gh-action-readme config themes

# Set default theme
gh-action-readme config set theme github
```

### Custom Themes

Create custom themes by copying existing ones:

```bash
# Copy existing theme
cp -r templates/themes/github templates/themes/custom

# Edit template
vim templates/themes/custom/readme.tmpl

# Use custom theme
gh-action-readme gen --theme custom
```

### Theme Directory Structure

```text
templates/themes/your-theme/
├── readme.tmpl           # Main template
├── partials/            # Optional partial templates
│   ├── header.tmpl
│   ├── inputs.tmpl
│   └── examples.tmpl
└── assets/              # Optional theme assets
    ├── styles.css
    └── logo.png
```

## 🔐 GitHub Token Configuration

### Creating a Token

1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Generate new token with `public_repo` scope
3. Copy token and save securely

### Setting Token

```bash
# Environment variable (recommended)
export GITHUB_TOKEN=your_token_here

# Configuration file
gh-action-readme config set github_token your_token_here

# Command line (least secure)
gh-action-readme gen --github-token your_token_here
```

### Token Benefits

- **Higher rate limits** (5000 requests/hour vs 60)
- **Dependency analysis** with detailed metadata
- **Private repository** access
- **Enhanced error messages** for API issues

## 📊 Cache Configuration

### Cache Settings

```yaml
# Cache configuration
cache_enabled: true
cache_dir: ~/.cache/gh-action-readme
cache_ttl: 3600  # 1 hour in seconds
cache_max_size: 100  # MB
```

### Cache Management

```bash
# Clear cache
gh-action-readme config clear-cache

# Check cache status
gh-action-readme config cache-status

# Set cache TTL
gh-action-readme config set cache_ttl 7200  # 2 hours
```

## 🔧 Advanced Configuration

### Custom Output Templates

```yaml
# Custom output naming patterns
output_patterns:
  md: "${name}-README.md"
  html: "docs/${name}.html"
  json: "api/${name}-metadata.json"
```

### Validation Rules

```yaml
# Custom validation settings
validation:
  require_description: true
  require_examples: false
  max_input_count: 50
  enforce_semver: true
```

### Template Variables

```yaml
# Custom template variables
template_vars:
  organization: "my-org"
  support_email: "support@example.com"
  docs_url: "https://docs.example.com"
```

## 📝 Configuration Commands

### View Configuration

```bash
# Show current config
gh-action-readme config show

# Show specific setting
gh-action-readme config get theme
```

### Update Configuration

```bash
# Set individual values
gh-action-readme config set theme professional
gh-action-readme config set verbose true

# Reset to defaults
gh-action-readme config reset

# Remove config file
gh-action-readme config delete
```

### Export/Import Configuration

```bash
# Export current config
gh-action-readme config export --format json > config.json

# Import configuration
gh-action-readme config import config.json

# Merge configurations
gh-action-readme config merge other-config.yaml
```

## 🔍 Debugging Configuration

### Verbose Mode

```bash
# Enable verbose output
gh-action-readme gen --verbose

# Set in config
gh-action-readme config set verbose true
```

### Configuration Validation

```bash
# Validate current configuration
gh-action-readme config validate

# Test configuration with dry run
gh-action-readme gen --dry-run --verbose
```

### Troubleshooting

```bash
# Show effective configuration (merged from all sources)
gh-action-readme config effective

# Show configuration file locations
gh-action-readme config paths

# Reset corrupted configuration
rm ~/.config/gh-action-readme/config.yaml
gh-action-readme config init
```
