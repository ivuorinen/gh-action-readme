# Composite Example Action


<div align="center">
  <img src="https://img.shields.io/badge/icon-package-blue" alt="package" />
  <img src="https://img.shields.io/badge/status-stable-brightgreen" alt="Status" />
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="License" />
</div>


## Overview

Test Composite Action for gh-action-readme dependency analysis

This GitHub Action provides a robust solution for your CI/CD pipeline with comprehensive configuration options and detailed output information.

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Input Parameters](#input-parameters)
- [Output Parameters](#output-parameters)
- [Examples](#examples)
- [Dependencies](#-dependencies)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Quick Start

Add the following step to your GitHub Actions workflow:

```yaml
name: CI/CD Pipeline
on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout Repository
        uses: actions/checkout@v4
      
      - name: Composite Example Action
        uses: your-org/ @v1
        with:
          node-version: "20"
          working-directory: "."
```

## Configuration

This action supports various configuration options to customize its behavior according to your needs.


### Input Parameters

| Parameter | Description | Type | Required | Default Value |
|-----------|-------------|------|----------|---------------|
| **`node-version`** | Node.js version to use | `string` | ❌ No | `20` |
| **`working-directory`** | Working directory | `string` | ❌ No | `.` |

#### Parameter Details


##### `node-version`

Node.js version to use

- **Type**: String
- **Required**: No
- **Default**: `20`

```yaml
with:
  node-version: "20"
```


##### `working-directory`

Working directory

- **Type**: String
- **Required**: No
- **Default**: `.`

```yaml
with:
  working-directory: "."
```





### Output Parameters

This action provides the following outputs that can be used in subsequent workflow steps:

| Parameter | Description | Usage |
|-----------|-------------|-------|
| **`build-result`** | Build result status | `\${{ steps. .outputs.build-result }}` |

#### Using Outputs

```yaml
- name: Composite Example Action
  id: action-step
  uses: your-org/ @v1
  
- name: Use Output
  run: |
    echo "build-result: \${{ steps.action-step.outputs.build-result }}"
```


## Examples

### Basic Usage

```yaml
- name: Basic Composite Example Action
  uses: your-org/ @v1
  with:
    node-version: "20"
    working-directory: "."
```

### Advanced Configuration

```yaml
- name: Advanced Composite Example Action
  uses: your-org/ @v1
  with:
    node-version: "20"
    working-directory: "."
  env:
    GITHUB_TOKEN: \${{ secrets.GITHUB_TOKEN }}
```

### Conditional Usage

```yaml
- name: Conditional Composite Example Action
  if: github.event_name == 'push'
  uses: your-org/ @v1
  with:
    node-version: "20"
    working-directory: "."
```


## 📦 Dependencies

This action uses the following dependencies:

| Action | Version | Author | Description |
|--------|---------|--------|-------------|
| [Checkout repository](https://github.com/marketplace/actions/checkout) | v4 | [actions](https://github.com/actions) |  |
| [Setup Node.js](https://github.com/marketplace/actions/setup-node) | v4 | [actions](https://github.com/actions) |  |
| Install dependencies | 🔒 | [ivuorinen](https://github.com/ivuorinen) | Shell script execution |
| Run tests | 🔒 | [ivuorinen](https://github.com/ivuorinen) | Shell script execution |
| [Build project](https://github.com/marketplace/actions/setup-node) | v4 | [actions](https://github.com/actions) |  |

<details>
<summary>📋 Dependency Details</summary>


### Checkout repository @ v4


- 📌 **Floating Version**: Using latest version (consider pinning for security)

- 👤 **Author**: [actions](https://github.com/actions)
- 🏪 **Marketplace**: [View on GitHub Marketplace](https://github.com/marketplace/actions/checkout)
- 📂 **Source**: [View Source](https://github.com/actions/checkout)

- **Configuration**:
  ```yaml
  with:
    fetch-depth: 0
    token: ${{ github.token }}
  ```



### Setup Node.js @ v4


- 📌 **Floating Version**: Using latest version (consider pinning for security)

- 👤 **Author**: [actions](https://github.com/actions)
- 🏪 **Marketplace**: [View on GitHub Marketplace](https://github.com/marketplace/actions/setup-node)
- 📂 **Source**: [View Source](https://github.com/actions/setup-node)

- **Configuration**:
  ```yaml
  with:
    cache: npm
    node-version: ${{ inputs.node-version }}
  ```



### Install dependencies


- 🔒 **Pinned Version**: Locked to specific version for security

- 👤 **Author**: [ivuorinen](https://github.com/ivuorinen)

- 📂 **Source**: [View Source](https://github.com/ivuorinen/gh-action-readme/blob/main/action.yml#L30)



### Run tests


- 🔒 **Pinned Version**: Locked to specific version for security

- 👤 **Author**: [ivuorinen](https://github.com/ivuorinen)

- 📂 **Source**: [View Source](https://github.com/ivuorinen/gh-action-readme/blob/main/action.yml#L40)



### Build project @ v4


- 📌 **Floating Version**: Using latest version (consider pinning for security)

- 👤 **Author**: [actions](https://github.com/actions)
- 🏪 **Marketplace**: [View on GitHub Marketplace](https://github.com/marketplace/actions/setup-node)
- 📂 **Source**: [View Source](https://github.com/actions/setup-node)

- **Configuration**:
  ```yaml
  with:
    node-version: ${{ inputs.node-version }}
  ```







### Same Repository Dependencies

- [Install dependencies](https://github.com/ivuorinen/gh-action-readme/blob/main/action.yml#L30) - Shell script execution

- [Run tests](https://github.com/ivuorinen/gh-action-readme/blob/main/action.yml#L40) - Shell script execution



</details>


## Troubleshooting

### Common Issues

1. **Authentication Errors**: Ensure you have set up the required secrets in your repository settings.
2. **Permission Issues**: Check that your GitHub token has the necessary permissions.
3. **Configuration Errors**: Validate your input parameters against the schema.

### Getting Help

- Check the [action.yml](./action.yml) for the complete specification
- Review the [examples](./examples/) directory for more use cases
- Open an issue if you encounter problems

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

1. Fork this repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

If you find this action helpful, please consider:

- ⭐ Starring this repository
- 🐛 Reporting issues
- 💡 Suggesting improvements
- 🤝 Contributing code

---

<div align="center">
  <sub>📚 Documentation generated with <a href="https://github.com/ivuorinen/gh-action-readme">gh-action-readme</a></sub>
</div>