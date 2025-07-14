# gh-action-readme

**gh-action-readme** is a modern, platform-agnostic CLI tool written in Go for auto-generating
beautiful, accurate `README.md` and HTML documentation for any GitHub Action defined by
an `action.yml`. It scans one or more actions (including all subfolders), validates them,
and outputs human-friendly, up-to-date documentation with usage examples, badges, and more.

---

> **Note:** By design, only one action per directory is supported.
> If multiple `action.yml` files are found in the same directory,
> only one will be processed and documented.

For upcoming features and ideas, see [TODO.md](TODO.md).

## 🛠️ Setting Up Your Development Environment

To ensure code quality and consistency, follow these steps before contributing:

1. **Install [EditorConfig](https://editorconfig.org/) support in your editor.**
   - This project includes a `.editorconfig` file to enforce indentation,
     line length, and newline rules for Go, Markdown, YAML, and JSON files.
   - Most modern editors support EditorConfig natively or via plugin.
2. **Lint Go code with [golangci-lint](https://golangci-lint.run/):**
   - Install: `brew install golangci-lint` or see [official docs](https://golangci-lint.run/usage/install/)
   - Run before and after making changes:

     ```sh
     golangci-lint run

     ```

   - **Do not change `.golangci.yml` to make linting pass.** Fix code, not the rules.

3. **Lint YAML files with [yamllint](https://yamllint.readthedocs.io/):**
   - Install: `brew install yamllint` or `pip install yamllint`
   - Run:

     ```sh
     yamllint .
     ```

   - The `.yamllint.yml` config enforces 2-space indentation, 100-char lines,
     and trailing newlines for YAML.

4. **Lint Markdown files with [markdownlint](https://github.com/DavidAnson/markdownlint):**
   - Install: `npm install -g markdownlint-cli` or see [markdownlint-cli](https://github.com/DavidAnson/markdownlint-cli)
   - Run:

     ```sh
     markdownlint .
     ```

   - The `.markdownlint.yaml` config enforces 2-space indentation, 100-char lines,
     and trailing newlines for Markdown.

5. **Always write and run tests:**
   - Add tests for all new features and bug fixes.
   - Run tests with:

     ```sh
     go test ./...
     ```

6. **Build the app to check for regressions:**

   ```sh
   go build .
   ```

7. **Update `README.md` and docs as needed.**
   - Keep documentation up to date with code and CLI changes.
8. **Run all linting tools after making changes to files, and fix any reported problems.**
   - You can use the Makefile for a one-step linting workflow (see below).

---

---

## 📝 Makefile Usage

This project includes a `Makefile` to automate linting, testing, and building.
The Makefile is platform-agnostic and will install required tools for your OS and architecture.

**Common commands:**

- `make help` — List all available Makefile commands and variables.
- `make install-tools` — Install all required linting and formatting tools (Go, YAML, Markdown).
- `make lint` — Run all linters (`golangci-lint`, `yamllint`, `markdownlint`) on the codebase.
- `make test` — Run all Go tests.
- `make build` — Build the application.

**Tip:** Always run `make lint` after making changes to any files, and fix all reported problems.

---

## Features

- Parses one or many `action.yml` (recursively)
- Validates required fields, offers autofix with defaults
- Generates `README.md` and/or HTML (with custom header/footer)
- Customizable Go text/template system for all fields
- Usage examples, badges, and summary sections by default
- Lists external action dependencies with links and version info
- Easy to update action schema and templates (uses [SchemaStore.org][s])
- Designed for both interactive and automated/CI workflows
- Configurable GitHub org/user for examples and badges
- Clean, DRY, idiomatic Go codebase
- One action per directory: Only one action per directory is supported.
  This is by design and simplifies documentation and output handling.

## Example usage

Generate a README for all found `action.yml` files (uses org from config):

```shell
gh-action-readme gen --config config.yaml .
```

Override org on CLI:

```shell
gh-action-readme gen --org my-github-org --format md .
```

Override action version on CLI:

```shell
gh-action-readme gen --version v2 .
```

Generate HTML docs with a header/footer:

```shell
gh-action-readme gen --format html .
```

Run in CI pipeline:

```yaml
- name: Generate action documentation
  run: gh-action-readme gen --autofill-missing --org myorg .
```

## CLI Flags Reference

Below is a summary of the most important CLI flags for each subcommand:

| Command         | Flag                 | Description                                                           | Default        |
| --------------- | -------------------- | --------------------------------------------------------------------- | -------------- |
| `gen`           | `--format`           | Output format(s): `md`, `html`, or comma-separated (e.g. `md,html`)   | `md`           |
| `gen`           | `--org`              | GitHub org/user (overrides config)                                    | from config    |
| `gen`           | `--config`           | Path to `config.yaml`                                                 | `config.yaml`  |
| `gen`           | `--output-dir`       | Output directory for docs (defaults to action.yml dir)                | action.yml dir |
| `gen`           | `--version`          | GitHub Action version tag or branch (overrides config)                | from config    |
| `gen`           | `--md-output`        | Output filename for Markdown                                          | `README.md`    |
| `gen`           | `--html-output`      | Output filename for HTML                                              | `README.html`  |
| `gen`           | `[root]`             | Root directory to search for `action.yml` files          | `.`            |
| `validate`      | `--config`           | Path to `config.yaml`                                                 | `config.yaml`  |
| `validate`      | `--autofill-missing` | Autofill missing fields using config defaults (in-memory only)        | `false`        |
| `validate`      | `--fix-missing`      | Autofill and write missing fields back to action.yml                  | `false`        |
| `validate`      | `--schema`           | Path to action.yml schema file (default: from config)                 | from config    |
| `schema`        | _(no flags)_         | Show the current action.yml schema file path                          |                |
| `schema update` | _(no flags)_         | Download and update the latest action.yml schema from SchemaStore.org |                |
| _(all)_         | `--verbose`, `-v`    | Enable verbose logging                                                | `false`        |

For a full list of flags and options, run `gh-action-readme <command> --help`.

## Template System

- Uses Go `text/template` for all documentation rendering.
- **Template context:** All fields from the `action.yml` file (as per the official schema)
  are available as top-level variables in templates.
  Additionally, the following variables are always available:
  - `.Org` — GitHub org/user (from config or CLI)
  - `.Repo` — Repository/folder (relative path for uses)
  - `.Version` — Action version/tag/branch (from config or CLI)
  - `{version}` placeholder in templates is replaced with the actual version.
  - `.LongDescription` contains text between `# docs:start` and `# docs:end` comments.
    Paragraph breaks are preserved and, when rendering HTML, the text is converted from Markdown.
  - `.Dependencies` — slice of actions referenced in composite steps.
    Each element has `Name`, `Version`, `Ref`, `Pinned`, and `Local` fields.
- **Header/Footer:**
  - Templates for header and footer are optional and can be customized per format (Markdown/HTML).
  - If a header or footer template file is missing, it is silently skipped (with a warning in logs).
  - Header and footer content is prepended/appended to the main template output.
  - The main README template path can also be overridden via `config.yaml` using
    the `template`, `header`, and `footer` fields.
- **Custom template functions:**
  - Advanced users can extend template rendering with custom Go template functions by modifying
    the codebase (see developer notes).

### Adding Custom Template Functions

Advanced users can add custom Go template functions for use in templates:

1. **Edit `internal/template.go`:**
   Set `TemplateOptions.Funcs` with a `template.FuncMap` when calling `RenderReadme`.

2. **Register your function:**
   Example:

   ```go
   funcs := template.FuncMap{
       "toUpper": strings.ToUpper,
   }
   opts.Funcs = funcs
   RenderReadme(action, opts)
   ```

   Then use `{{.Name | toUpper}}` in your template.

3. **Rebuild the CLI:**
   Run `go build .` to apply your changes.

---

## Project Structure

- `main.go` — CLI entrypoint (Cobra)
- `cmd/` — CLI subcommands and helpers (one file per command, shared helpers)
- `internal/` — Core logic (parser, validator, template, config, html)
- `templates/` — Format-specific templates for README/HTML/header/footer
- `schemas/` — Official `action.yml` schema (embedded in the binary and auto-updatable from SchemaStore)
- `README.md`, `LICENSE`, `config.yaml`, etc.

## Developer Notes

- All CLI commands are implemented in `cmd/` for clarity and maintainability.
- Shared helpers are DRY and use Go's standard library where possible.
- Linting is enforced with `golangci-lint` and the codebase is fully compliant.
- The schema is embedded and updated from [SchemaStore.org][s] for maximum compatibility.
- Paths like `schemas/action.schema.json` are resolved from the project root.

## License

[MIT](LICENSE)

[s]: https://www.schemastore.org/github-action.json
