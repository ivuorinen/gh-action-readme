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
analyze_dependencies: true
```

## 🔧 Configuration Options

### Core Settings

| Option | Type | Default | Description |
| -------- | ------ | --------- | ------------- |
| `theme` | string | `default` | Default theme to use |
| `output_format` | string | `md` | Default output format |
| `output_dir` | string | `.` | Default output directory |
| `verbose` | boolean | `false` | Enable verbose logging |

### GitHub Integration

| Option | Type | Default | Description |
| -------- | ------ | --------- | ------------- |
| `github_token` | string | `""` | GitHub personal access token |
| `analyze_dependencies` | boolean | `true` | Enable dependency analysis |
| `show_security_info` | boolean | `false` | Show security/permissions info |

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
export GH_ACTION_README_ANALYZE_DEPENDENCIES=true
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
└── readme.tmpl           # Main template (readme.adoc for the asciidoc template)
```

## 🔐 GitHub Token Configuration

### Creating a Token

1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Generate new token with `public_repo` scope
3. Copy token and save securely

### Setting Token

```bash
# Environment variable (recommended)
export GH_README_GITHUB_TOKEN=your_token_here
# or
export GITHUB_TOKEN=your_token_here
```

### Token Benefits

- **Higher rate limits** (5000 requests/hour vs 60)
- **Dependency analysis** with detailed metadata
- **Private repository** access
- **Enhanced error messages** for API issues

## 📊 Cache Configuration

### Cache Settings

Caching is built in and not user-configurable. The cache lives in the
XDG cache directory (`~/.cache/gh-action-readme`) and uses fixed defaults:
a 15-minute TTL for API responses and a 100 MB maximum size.

### Cache Management

```bash
# Clear cache
gh-action-readme cache clear

# Check cache path
gh-action-readme cache path

# Check cache stats
gh-action-readme cache stats
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
```

### Update Configuration

Edit the config file directly (`~/.config/gh-action-readme/config.yaml`) or re-run the wizard:

```bash
gh-action-readme config wizard
```

To reset to defaults, delete the config file and re-initialize:

```bash
rm ~/.config/gh-action-readme/config.yaml
gh-action-readme config init
```

## 🔍 Debugging Configuration

### Verbose Mode

```bash
# Enable verbose output
gh-action-readme gen --verbose
```

### Troubleshooting

```bash
# Show current configuration
gh-action-readme config show

# Reset corrupted configuration
rm ~/.config/gh-action-readme/config.yaml
gh-action-readme config init
```
