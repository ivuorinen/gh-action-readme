# Themes

gh-action-readme includes 5 built-in themes, each optimized for different use cases and visual preferences.

## 🎨 Available Themes

### GitHub Theme

**Best for:** GitHub Marketplace actions, open source projects

```bash
gh-action-readme gen --theme github
```

**Features:**

- GitHub-style badges and shields
- Collapsible sections for better organization
- Table-based parameter documentation
- Action marketplace optimization
- GitHub-specific styling and layout

**Example output:**

- ✅ Professional badges
- 📊 Clean input/output tables
- 🔧 Copy-paste usage examples
- 📁 Collapsible troubleshooting sections

### GitLab Theme

**Best for:** GitLab CI/CD integration, GitLab-hosted projects

```bash
gh-action-readme gen --theme gitlab
```

**Features:**

- GitLab CI/CD pipeline examples
- GitLab-specific badge integration
- Pipeline configuration snippets
- GitLab Pages optimization

### Minimal Theme

**Best for:** Simple actions, lightweight documentation

```bash
gh-action-readme gen --theme minimal
```

**Features:**

- Clean, distraction-free layout
- Essential information only
- Faster loading and parsing
- Mobile-friendly design
- Minimal dependencies

### Professional Theme

**Best for:** Enterprise actions, comprehensive documentation

```bash
gh-action-readme gen --theme professional
```

**Features:**

- Comprehensive table of contents
- Detailed troubleshooting sections
- Advanced usage examples
- Security and compliance notes
- Enterprise-ready formatting

### Default Theme

**Best for:** Basic needs, backward compatibility

```bash
gh-action-readme gen --theme default
```

**Features:**

- Original simple template
- Basic parameter documentation
- Minimal formatting
- Guaranteed compatibility

## 🎯 Theme Comparison

| Feature | GitHub | GitLab | Minimal | Professional | Default |
| --------- | -------- | -------- | --------- | ------------- | --------- |
| **Badges** | ✅ Rich | ✅ GitLab | ❌ None | ✅ Comprehensive | ❌ None |
| **TOC** | ✅ Yes | ✅ Yes | ❌ No | ✅ Advanced | ❌ No |
| **Examples** | ✅ GitHub | ✅ CI/CD | ✅ Basic | ✅ Comprehensive | ✅ Basic |
| **Troubleshooting** | ✅ Collapsible | ✅ Pipeline | ❌ Minimal | ✅ Detailed | ❌ None |
| **File Size** | Medium | Medium | Small | Large | Small |
| **Load Time** | Fast | Fast | Fastest | Slower | Fast |

## 🛠️ Theme Examples

### Input Action

```yaml
name: Deploy to AWS
description: Deploy application to AWS using GitHub Actions
inputs:
  aws-region:
    description: AWS region to deploy to
    required: true
    default: us-east-1
  environment:
    description: Deployment environment
    required: false
    default: production
outputs:
  deployment-url:
    description: URL of the deployed application
runs:
  using: composite
  steps:
    - run: echo "Deploying to AWS..."
```

### GitHub Theme Output

```markdown
# Deploy to AWS

[![GitHub Action](https://img.shields.io/badge/GitHub%20Action-Deploy%20to%20AWS-blue)](https://github.com/marketplace)

> Deploy application to AWS using GitHub Actions

## 🚀 Usage

```yaml
- uses: your-org/deploy-aws@v1
  with:
    aws-region: us-west-2
    environment: staging
```

<details>
<summary>📋 Inputs</summary>

| Input         | Description              | Required | Default       |
| ------------- | ------------------------ | -------- | ------------- |
| `aws-region`  | AWS region to deploy to  | Yes      | `us-east-1`   |
| `environment` | Deployment environment   | No       | `production`  |

</details>
```

### Minimal Theme Output

```markdown
# Deploy to AWS

Deploy application to AWS using GitHub Actions

## Usage

```yaml
- uses: your-org/deploy-aws@v1
  with:
    aws-region: us-west-2
```

## Inputs

- `aws-region` (required): AWS region to deploy to
- `environment` (optional): Deployment environment (default: production)

## Outputs

- `deployment-url`: URL of the deployed application

```text

## 🎨 Custom Themes

### Creating Custom Themes

1. **Copy existing theme:**
```bash
cp -r templates/themes/github templates/themes/my-theme
```

1. **Edit template files:**

```bash
# Main template
vim templates/themes/my-theme/readme.tmpl

# Optional partials (only header.tmpl and footer.tmpl are loaded)
vim templates/themes/my-theme/partials/header.tmpl
vim templates/themes/my-theme/partials/footer.tmpl
```

1. **Use custom theme:**

```bash
gh-action-readme gen --theme my-theme
```

### Theme Structure

```text
templates/themes/my-theme/
├── readme.tmpl           # Main template (required)
└── partials/             # Optional; only these two names are loaded
    ├── header.tmpl       # Rendered before the body
    └── footer.tmpl       # Rendered after the body
```

Only `partials/header.tmpl` and `partials/footer.tmpl` are recognised. There is no
section-level partial mechanism (inputs, outputs, examples) and no theme `assets/`
directory — a theme is one `readme.tmpl` plus, optionally, those two partials.

A theme partial is used unless the `header` / `footer` config key overrides that
part; see [Header and footer partials](configuration.md#header-and-footer-partials)
for the full precedence.

### Template Variables

Available variables in templates:

```go
type ActionData struct {
    Name          string                  // Action name
    Description   string                  // Action description
    Inputs        map[string]ActionInput  // Input parameters
    Outputs       map[string]ActionOutput // Output parameters
    Runs          map[string]interface{}  // Runs configuration
    Branding      *Branding              // Branding info

    // Enhanced data
    Repository    *Repository            // GitHub repo info
    Dependencies  []Dependency           // Analyzed dependencies
    Examples      []Example              // Usage examples
}
```

### Template Functions

Built-in template functions:

```go
// String functions
{{ .Name | title }}              // Title case
{{ .Description | truncate 100 }} // Truncate to 100 chars
{{ .Name | slug }}               // URL-friendly slug

// Formatting functions
{{ .Inputs | toTable }}          // Generate input table
{{ .Dependencies | toList }}      // Generate dependency list
{{ .Examples | toYAML }}         // Format as YAML

// Conditional functions
{{ if hasInputs }}...{{ end }}   // Check if inputs exist
{{ if .Branding }}...{{ end }}   // Check if branding exists
```

### Advanced Template Example

```go-template
{{/* Custom theme header */}}
# {{ .Name }}

{{ if .Branding }}
<p align="center">
  <img src="{{ .Branding.Icon }}" alt="{{ .Name }}" width="64">
</p>
{{ end }}

> {{ .Description }}

## 🚀 Quick Start

```yaml
- uses: {{ .Repository.FullName }}@{{ .Repository.DefaultBranch }}
  with:
    {{- range $key, $input := .Inputs }}
    {{- if $input.Required }}
    {{ $key }}: # {{ $input.Description }}
    {{- end }}
    {{- end }}
```

{{ if .Inputs }}

## 📋 Configuration

{{ template "inputs-table" . }}
{{ end }}

{{ if .Dependencies }}

## 🔗 Dependencies

This action uses the following dependencies:
{{ range .Dependencies }}

- [{{ .Name }}]({{ .SourceURL }}) - {{ .Description }}
{{ end }}
{{ end }}

{{/*Include footer partial if it exists*/}}
{{ template "footer" . }}

```go-template

## 🔧 Theme Configuration

### Set Default Theme
```bash
# Set globally
gh-action-readme config set theme professional

# Use environment variable
export GH_ACTION_README_THEME=github
```

### Theme-Specific Settings

```yaml
# ~/.config/gh-action-readme/config.yaml
theme: github
theme_settings:
  github:
    show_badges: true
    collapse_sections: true
    show_toc: true
  minimal:
    show_examples: true
    compact_tables: true
  professional:
    detailed_examples: true
    show_troubleshooting: true
```

### Dynamic Theme Selection

```bash
# Select theme based on repository type
if [[ -f ".gitlab-ci.yml" ]]; then
  gh-action-readme gen --theme gitlab
elif [[ -f "package.json" ]]; then
  gh-action-readme gen --theme github
else
  gh-action-readme gen --theme minimal
fi
```

## 📱 Responsive Design

All themes support responsive design:

- **Desktop**: Full-width tables and detailed sections
- **Tablet**: Condensed tables with horizontal scrolling
- **Mobile**: Stacked layouts and collapsible sections

## 🎨 Theme Customization Tips

### Colors and Styling

```css
/* Custom CSS for HTML output */
.action-header { color: #0366d6; }
.input-table { border-collapse: collapse; }
.example-code { background: #f6f8fa; }
```

### Badge Customization

```go-template
{{/* Custom badges */}}
![Version](https://img.shields.io/badge/version-{{ .Version }}-blue)
![License](https://img.shields.io/badge/license-{{ .License }}-green)
![Downloads](https://img.shields.io/github/downloads/{{ .Repository.FullName }}/total)
```

### Conditional Content

```go-template
{{/* Show different content based on action type */}}
{{ if eq .Runs.Using "composite" }}
## Composite Action Steps
{{ range .Runs.Steps }}
- {{ .Name }}: {{ .Run }}
{{ end }}
{{ else if eq .Runs.Using "docker" }}
## Docker Configuration
- Image: `{{ .Runs.Image }}`
- Args: `{{ join .Runs.Args " " }}`
{{ end }}
```
