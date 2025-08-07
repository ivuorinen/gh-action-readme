# Usage Guide

Comprehensive guide to using gh-action-readme for generating documentation from GitHub Actions.

## 🚀 Quick Start

Generate documentation from your action.yml:

```bash
# Basic generation
gh-action-readme gen

# With specific theme and output
gh-action-readme gen --theme github --output README.md
```

## 📋 Examples

### Input: `action.yml`

```yaml
name: My Action
description: Does something awesome
inputs:
  token:
    description: GitHub token
    required: true
  environment:
    description: Target environment
    default: production
outputs:
  result:
    description: Action result
runs:
  using: node20
  main: index.js
```

### Output: Professional README.md

The tool generates comprehensive documentation including:

- 📊 **Parameter tables** with types, requirements, defaults
- 💡 **Usage examples** with proper YAML formatting
- 🎨 **Badges** for marketplace visibility
- 📚 **Multiple sections** (Overview, Configuration, Examples, Troubleshooting)
- 🔗 **Navigation** with table of contents

## 🛠️ Commands

### Generation

```bash
gh-action-readme gen [directory_or_file] [flags]
  -f, --output-format string   md, html, json, asciidoc (default "md")
  -o, --output-dir string      output directory (default ".")
      --output string          custom output filename
  -t, --theme string           github, gitlab, minimal, professional
  -r, --recursive              search recursively
```

**Examples:**

```bash
# Generate with GitHub theme
gh-action-readme gen --theme github

# Generate HTML documentation
gh-action-readme gen --output-format html

# Process specific directory
gh-action-readme gen actions/checkout/

# Custom output filename
gh-action-readme gen --output my-action-docs.md

# Recursive processing
gh-action-readme gen --recursive --theme professional
```

### Validation

```bash
gh-action-readme validate

# With verbose output
gh-action-readme validate --verbose
```

**Example Output:**

```text
❌ Missing required field: description
💡 Add 'description: Brief description of what your action does'

✅ All inputs have descriptions
⚠️  Consider adding 'branding' section for marketplace visibility
```

### Configuration

```bash
gh-action-readme config init     # Create default config
gh-action-readme config show     # Show current settings
gh-action-readme config themes   # List available themes
gh-action-readme config wizard   # Interactive configuration
```

## 🎯 Advanced Usage

### Batch Processing

```bash
# Process multiple repositories with custom outputs
find . -name "action.yml" -execdir gh-action-readme gen --theme github --output README-generated.md \;

# Recursive processing with JSON output
gh-action-readme gen --recursive --output-format json --output-dir docs/

# Target multiple specific actions
gh-action-readme gen actions/checkout/ --theme github --output docs/checkout.md
gh-action-readme gen actions/setup-node/ --theme professional --output docs/setup-node.md
```

### Custom Templates

```bash
# Copy and modify existing theme
cp -r templates/themes/github templates/themes/custom
# Edit templates/themes/custom/readme.tmpl
gh-action-readme gen --theme custom
```

### Environment Integration

```bash
# Set default preferences
export GH_ACTION_README_THEME=github
export GH_ACTION_README_VERBOSE=true
export GITHUB_TOKEN=your_token_here

# Use with different output formats
gh-action-readme gen --output-format html --output docs/index.html
gh-action-readme gen --output-format json --output api/action.json
```

## 📄 Output Formats

| Format | Description | Use Case | Extension |
|--------|-------------|----------|-----------|
| **md** | Markdown (default) | GitHub README files | `.md` |
| **html** | Styled HTML | Web documentation | `.html` |
| **json** | Structured data | API integration | `.json` |
| **asciidoc** | AsciiDoc format | Technical docs | `.adoc` |

### Format Examples

```bash
# Markdown (default)
gh-action-readme gen --output-format md

# HTML with custom styling
gh-action-readme gen --output-format html --output docs/action.html

# JSON for API consumption
gh-action-readme gen --output-format json --output api/metadata.json

# AsciiDoc for technical documentation
gh-action-readme gen --output-format asciidoc --output docs/action.adoc
```

## 🎨 Themes

See [themes.md](themes.md) for detailed theme documentation.

| Theme | Best For | Features |
|-------|----------|----------|
| **github** | GitHub marketplace | Badges, collapsible sections |
| **gitlab** | GitLab repositories | CI/CD examples |
| **minimal** | Simple actions | Clean, concise |
| **professional** | Enterprise use | Comprehensive docs |
| **default** | Basic needs | Original template |

## ⚙️ Configuration

See [configuration.md](configuration.md) for detailed configuration options.

## 🔧 Testing Generation

**Safe testing with sample data:**

```bash
# Test with included examples (safe - won't overwrite)
gh-action-readme gen testdata/example-action/ --theme github --output test-output.md
gh-action-readme gen testdata/composite-action/action.yml --theme professional

# Test all themes
for theme in github gitlab minimal professional default; do
  gh-action-readme gen testdata/example-action/ --theme $theme --output "test-${theme}.md"
done

# Test all formats
for format in md html json asciidoc; do
  gh-action-readme gen testdata/example-action/ --output-format $format --output "test.${format}"
done
```

## 🐛 Troubleshooting

### Common Issues

**File not found:**

```bash
# Make sure action.yml exists in current directory or specify path
gh-action-readme gen path/to/action.yml
```

**Permission errors:**

```bash
# Check file permissions
ls -la action.yml
chmod 644 action.yml
```

**GitHub API rate limits:**

```bash
# Set GitHub token for higher rate limits
export GITHUB_TOKEN=your_token_here
gh-action-readme gen --verbose
```

**Template errors:**

```bash
# Validate your action.yml first
gh-action-readme validate --verbose
```

### Getting Help

1. Run with `--verbose` flag for detailed output
2. Check [GitHub Issues](https://github.com/ivuorinen/gh-action-readme/issues)
3. Validate your action.yml syntax
4. Verify file permissions and paths
