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
analyze_dependencies: false # default is false; set true to opt in
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
| `analyze_dependencies` | boolean | `false` | Enable dependency analysis |
| `show_security_info` | boolean | `false` | Add a Security section reporting dependency pinning (needs `analyze_dependencies`) |

### Legacy and Reserved Settings

### Header and footer partials

`header` and `footer` name template files wrapped around the rendered body. They
apply to **every text output format** (Markdown, AsciiDoc, HTML) — JSON is unaffected.

Each part is resolved independently, so overriding one keeps the other:

1. an explicitly configured `header` / `footer`;
2. the selected theme's `partials/header.tmpl` / `partials/footer.tmpl`, if it ships one;
3. for HTML only, the built-in HTML partials (the `<head>`/`<body>` scaffolding).

Setting only `footer` therefore replaces the footer while leaving the theme's header
in place. The built-in HTML partials are never injected into Markdown or AsciiDoc.

### Permissions, runners, security info, and custom variables

| Key | Effect on output |
| --- | --- |
| `permissions` | Rendered in the Permissions section **when the action declares none of its own**. An `action.yml` declaration (YAML or `# permissions:` header comment) always wins. |
| `runs_on` | Supplies the `runs-on:` value in the generated workflow examples. One runner renders as a scalar, several as a YAML flow sequence (`[ubuntu-latest, macos-latest]`). |
| `show_security_info` | When true *and* dependency analysis is on, adds a Security section reporting how many dependencies are pinned and listing the floating ones. |
| `variables` | Custom values reachable from any template as `{{var . "key"}}`. Unknown keys return `""`, so `{{with var . "key"}}…{{end}}` renders a block only when the variable is set. |

```yaml
permissions:
  contents: read
runs_on: [ubuntu-latest, macos-latest]
show_security_info: true
variables:
  support_url: https://example.com/support
```

### Dependency analysis in JSON output

With `analyze_dependencies: true`, `-f json` emits a top-level `dependencies` array
mirroring the Dependencies section the Markdown themes render. The key is omitted
entirely when analysis is disabled, so consumers written against earlier output are
unaffected.

```json
{
  "action": { "...": "..." },
  "dependencies": [
    { "name": "actions/checkout", "uses": "actions/checkout@v4", "version": "v4", "...": "..." }
  ]
}
```

### License

The license shown in generated docs is resolved highest-priority-first:

1. the `license` config key;
2. a top-level `license:` key in `action.yml`;
3. a `# license: <id>` header comment in `action.yml`;
4. detection from the repository's `LICENSE` / `LICENCE` / `COPYING` file, where an
    `SPDX-License-Identifier:` tag wins over title matching.

If none resolve, **no license section or badge is rendered**. The tool documents
actions it does not own, so it never asserts a license it cannot verify.

```yaml
# action.yml
# license: Apache-2.0
name: My Action
```

In JSON output the resolved value appears as `action.license` (omitted when unknown),
alongside the human-facing licence badge.

Links to repository files (`LICENSE` and friends) are emitted relative to the
generated **document**, so a monorepo action documented into `actions/foo/` links to
`../../LICENSE`, and `--output nested/README.md` links to `../../../LICENSE`. The link
is omitted when the file does not exist, and when the document is written outside the
repository (an absolute `--output`), since no relative path could reach it.

Some configuration keys are accepted for backward compatibility or planned
features but do **not** affect output in the current release:

| Key | Status |
| --- | --- |
| `template` | Legacy custom-template path. Ignored while a `theme` is set, and `default` is set out of the box. Set `theme: ""` to use the `template` path instead. |
| `schema` | Informational only. The `schema` command and key print/record the schema path, but the tool does not validate `action.yml` against a JSON schema — validation is structural. |

> Note: the `permissions:` block *inside* an `action.yml` file always wins. The config
> key of the same name is a fallback, used only when the action declares none.

## 🌍 Environment Variables

Override configuration with environment variables:

```bash
# Core settings
export GH_ACTION_README_THEME=github
export GH_ACTION_README_OUTPUT_FORMAT=html
export GH_ACTION_README_OUTPUT_DIR=docs
export GH_ACTION_README_VERBOSE=true

# GitHub settings (GH_README_GITHUB_TOKEN takes precedence over GITHUB_TOKEN;
# see "Setting Token" below)
export GH_README_GITHUB_TOKEN=your_token_here
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
