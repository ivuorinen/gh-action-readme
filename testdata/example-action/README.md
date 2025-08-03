# Example Action

![check](https://img.shields.io/badge/icon-check-green) ![GitHub](https://img.shields.io/badge/GitHub%20Action- -blue) ![License](https://img.shields.io/badge/license-MIT-green)

> Test Action for gh-action-readme

## 🚀 Quick Start

```yaml
name: My Workflow
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Example Action
        uses: ivuorinen/gh-action-readme/example-action@v1
        with:
          input1: "foo"
          input2: "value"
```


## 📥 Inputs

| Parameter | Description | Required | Default |
|-----------|-------------|----------|---------|
| `input1` | First input | ✅ | `foo` |
| `input2` | Second input | ❌ | - |



## 📤 Outputs

| Parameter | Description |
|-----------|-------------|
| `result` | Result output |


## 💡 Examples

<details>
<summary>Basic Usage</summary>

```yaml
- name: Example Action
  uses: ivuorinen/gh-action-readme/example-action@v1
  with:
    input1: "foo"
    input2: "example-value"
```
</details>

<details>
<summary>Advanced Configuration</summary>

```yaml
- name: Example Action with custom settings
  uses: ivuorinen/gh-action-readme/example-action@v1
  with:
    input1: "foo"
    input2: "custom-value"
```
</details>



## 🔧 Development

See the [action.yml](./action.yml) for the complete action specification.

## 📄 License

This action is distributed under the MIT License. See [LICENSE](LICENSE) for more information.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

<div align="center">
  <sub>🚀 Generated with <a href="https://github.com/ivuorinen/gh-action-readme">gh-action-readme</a></sub>
</div>
