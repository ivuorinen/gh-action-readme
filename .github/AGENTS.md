# AGENTS.md

> **Purpose:**
> This document provides a comprehensive, human- and LLM-friendly overview of the `gh-action-readme`
> project. It covers the architecture, features, directory structure, CLI commands, configuration,
> extensibility, and key APIs for both developers and automation agents.

---

## Overview

**gh-action-readme** is a modern, cross-platform Go CLI tool for auto-generating `README.md`
and HTML documentation for GitHub Actions based on `action.yml` files. It supports schema
validation, customizable templates, CI/CD workflows, and is designed to be DRY, idiomatic,
and easy to extend or automate.

---

## Features

- **Recursive Action Discovery:**
  Finds and parses all `action.yml` files in a directory tree (including subfolders).

- **Schema-based Parsing:**
  Uses the official [action.yml schema](https://www.schemastore.org/github-action.json)
  for robust parsing and validation. The schema is auto-updatable via `gh-action-readme schema update`.

- **README & HTML Generation:**
  Generates beautiful, comprehensive `README.md` and/or HTML documentation for each action.
  Includes usage examples, badges, metadata, and org/user info from config or CLI.
  Uses Go text/template for Markdown and HTML, with customizable header/footer.

- **Validation & Autofix:**
  Validates that all required fields are present in `action.yml`.
  Can autofill/fix missing fields using config defaults.

- **Configurable Defaults:**
  All `action.yml` fields’ default values are configurable via `config.yaml` or CLI flags.

- **Modern, DRY, and Simple Codebase:**
  Shared helpers are deduplicated and use the Go standard library where possible.
  Minimal custom code for string splitting, trimming, and file handling.

- **Helpful CLI/UX:**
  Prints actionable help when required fields are missing or invalid.
  Includes `--version` and `--about` flags.
  MIT licensed, platform-agnostic.

- **CI/CD Ready:**
  Can run non-interactively for pipeline automation (e.g., auto-update README.md).

- **Extensible and Maintainable:**
  Modular design for easy field addition, template swapping, or schema changes.
  All code is linted and formatted with `golangci-lint` and `gofumpt`.

---

## Agent & Contributor Instructions

To ensure code quality, maintainability, and a smooth development process,
**all agents and contributors must follow these rules**:

> **Note:** By design, only one action per directory is supported. If multiple `action.yml` files
> are found in the same directory, only one will be processed and documented.
> This is intentional and simplifies output handling and documentation generation.

- **Run `golangci-lint` before and after making any code changes.**
  Fix all reported problems. Do not proceed with commits or PRs if any linter
  errors or warnings remain.

- **Always write tests for new functionality.**
  Ensure that all new features, bug fixes, and changes are covered by tests.
  Keep test coverage as high as possible.

- **Try to build the application after making changes.**
  This verifies that no regressions or build errors have been introduced.

- **Never change linter configurations just to make validation or linting pass.**
  The linter configuration is strict by design. Fix the code, not the rules.
  Linter configurations should be updated only when necessary.
  Configuration files are `.golangci.yml`, `.markdownlint.yml` and `.yamllint.yml`.

- **Update `README.md` and other documentation as needed.**
  Keep user-facing docs up to date with any changes to features, flags, or usage.

- **Always write clean, easy to understand code and documentation.**
  Favor clarity and maintainability over cleverness or shortcuts.

- **Run all linting tools individually after making changes to files,
  and fix any reported problems or warnings.**
  - Do **not** use `make lint` for this purpose. Instead, run each linter one by one
    so problems are detected and easier to fix:
    - `golangci-lint run`
    - `yamllint .`
    - `markdownlint . --config .markdownlint.yaml`
  - **You must fix both errors and warnings reported by any linter.**

---

### 🛠️ Setting Up Your Development Environment

To make it easy to write compliant code and keep the project healthy,
set up your environment as follows:

1. **Install [EditorConfig](https://editorconfig.org/) support in your editor.**
   - This project includes a `.editorconfig` file to enforce indentation, line length,
     and newline rules for Go, Markdown, YAML, and JSON files.
   - Most modern editors support EditorConfig natively or via plugin.

2. **Lint Go code with [golangci-lint](https://golangci-lint.run/):**
   - Install: `brew install golangci-lint` or see [official docs](https://golangci-lint.run/usage/install/)
   - Run before and after making changes:
     ```bash
     golangci-lint run
     ```
   - **Do not change `.golangci.yml` to make linting pass.** Fix code, not the rules.

3. **Lint YAML files with [yamllint](https://yamllint.readthedocs.io/):**
   - Install: `brew install yamllint` or `pip install yamllint`
   - Run:
     ```bash
     yamllint .
     ```
   - The `.yamllint.yml` config enforces 2-space indentation, 100-char lines, and trailing newlines for YAML.

4. **Lint Markdown files with [markdownlint](https://github.com/DavidAnson/markdownlint):**
   - Install: `npm install -g markdownlint-cli` or see [markdownlint-cli](https://github.com/DavidAnson/markdownlint-cli)
   - Run:
     ```bash
     markdownlint .
     ```
   - The `.markdownlint.yaml` config enforces 2-space indentation, 100-char lines, and trailing newlines for Markdown.

5. **Always write and run tests:**
   - Add tests for all new features and bug fixes.
   - Run tests with:
     ```bash
     go test ./...
     ```

6. **Build the app to check for regressions:**

   ```bash
   go build .
   ```

7. **Update `README.md` and docs as needed.**
   - Keep documentation up to date with code and CLI changes.

---

### 📝 Makefile Usage

This project includes a `Makefile` to automate linting, testing, and building.
The Makefile is platform-agnostic and will install required tools for your OS and architecture.

**Common commands:**

- `make help` — List all available Makefile commands and variables.
- `make install-tools` — Install all required linting and formatting tools (Go, YAML, Markdown).
- `make lint` — Run all linters (`golangci-lint`, `yamllint`, `markdownlint`) on the codebase.
- `make test` — Run all Go tests.
- `make build` — Build the application.

**Tip:** Always run each linter individually after making changes to any files, and fix all
reported problems and warnings. This makes it easier to see and address issues specific to
each linter.

---

## Directory Structure

```
.
├── cmd/                # CLI subcommands (gen, validate, schema) and helpers
│   ├── gen.go
│   ├── validate.go
│   ├── schema.go
│   └── helpers.go
├── internal/           # Core logic: parsing, validation, templating, config
│   ├── config.go
│   ├── defaults.go
│   ├── html.go
│   ├── parser.go
│   ├── template.go
│   └── validator.go
├── templates/          # Go text/template files for Markdown and HTML output
├── schemas/            # JSON schema for action.yml (embedded and auto-updatable)
├── testdata/           # Example actions and test cases
├── main.go             # CLI entrypoint (wires up commands)
├── config.yaml         # Default config for templates, org, etc.
├── Makefile            # Build/test/lint targets
├── README.md           # Project documentation
└── .golangci.yml       # Linter configuration
```

---

## CLI Commands

- **`gen`**
  Generate documentation for all found `action.yml` files.
  _Flags:_ `--format`, `--org`, `--config`, `--output-dir`, `--version`,
  `--md-output`, `--html-output`

- **`validate`**
  Validate all `action.yml` files for required fields and schema compliance.
  _Flags:_ `--config`, `--autofill-missing`, `--fix-missing`, `--schema`

- **`schema`**
  Show or update the action.yml schema file.
  _Subcommand:_ `update` (downloads latest schema from SchemaStore)

- **`version`**
  Print the tool version.

- **`about`**
  Print a short description.

---

## Configuration

- **`config.yaml`**
  - Custom paths for templates and schema file
  - Default values for all `action.yml` fields
  - Output formatting options
  - GitHub org/user for code samples and badges

---

## Template System

- Uses Go `text/template` for all documentation rendering.
- **Template context:** All fields from the `action.yml` file (as per the official schema)
  are available as top-level variables in templates.
  Additionally, the following variables are always available:
  - `.Org` — GitHub org/user (from config or CLI)
  - `.Repo` — Repository/folder (relative path for uses)
  - `.Version` — Action version/tag/branch (from config or CLI)
- **Header/Footer:**
  - Templates for header and footer are optional and can be customized per format (Markdown/HTML).
  - If a header or footer template file is missing, it is silently skipped (with a warning in logs).
  - Header and footer content is prepended/appended to the main template output.
- **Custom template functions:**
  - Advanced users can extend template rendering with custom Go template functions by
    modifying the codebase (see developer notes).

---

## Project Root Path Handling

When writing or updating files such as `schemas/action.schema.json`,
**always resolve the path from the project root**, not the current working directory.
The project root should be detected by looking for a marker file such as `.git` or `go.mod`.
This ensures that schema and other important files are placed in the correct location
regardless of where the CLI is invoked.

**Example:**  
If you are updating the schema, search upwards from the current directory for `.git` or `go.mod`,
and use that directory as the base for `schemas/action.schema.json`.

---

## Extending gh-action-readme

### Adding Custom Template Functions

Advanced users can add custom Go template functions for use in templates:

1. **Edit `internal/template.go`:**
   Set `TemplateOptions.Funcs` with a `template.FuncMap` when calling `RenderReadme`.

2. **Register your function:**
   Example:

   ```go
   functions := template.FuncMap{
       "toUpper": strings.ToUpper,
   }
   opts.Funcs = functions
   RenderReadme(action, opts)
   ```

   Then use `{{.Name | toUpper}}` in your template.

3. **Rebuild the CLI:**
   Run `go build .` to apply your changes.

---

## Key APIs and Data Structures

- **Config Loading**
  - `internal.Config` — Loads from `config.yaml`
  - `internal.LoadConfig(path string) (*Config, error)`

- **Action Parsing**
  - `internal.ActionYML` — Struct representing the parsed `action.yml`
  - `internal.ParseActionYML(path string) (*ActionYML, error)`

- **Validation**
  - `internal.ValidateActionYML(action *ActionYML) ValidationResult` — Checks for required fields
  - `internal.ValidateActionYMLSchema(actionYMLPath, schemaPath string) ([]string, error)`
    — Validates against JSON schema

- **Documentation Generation**
  - `internal.TemplateOptions` — Holds template paths, org, repo, version, format
  - `internal.RenderReadme(action any, opts TemplateOptions) (string, error)`
    — Renders documentation using Go templates

- **Helpers**
  - File discovery, string manipulation, and YAML writing are in `cmd/helpers.go`

---

## Usage Scenarios

### Basic README Generation

```sh
gh-action-readme gen --config config.yaml
```

### Override org

```sh
gh-action-readme gen --org my-github-org
```

### Validate and Autofix action.yml

```sh
gh-action-readme validate --fix-missing
```

### Generate HTML with header/footer

```sh
gh-action-readme gen --format html --header myheader.html.tmpl --footer myfooter.html.tmpl
```

### CI Pipeline Example

```yaml
- name: Auto-generate Action README
  run: gh-action-readme gen --autofill-missing --org myorg
```

### Update Schema

```sh
gh-action-readme schema update
```

---

## Extensibility

- Modular design for easy field addition, template swapping, or schema changes
- All code is linted and formatted with `golangci-lint` and `gofumpt`
- Schema can be updated from SchemaStore for future compatibility

---

## License

MIT License

---

_For LLMs and automation agents: All CLI commands are implemented in `cmd/` as `cobra.Command`
factories. Core logic is in `internal/`. The schema file is auto-updatable. The codebase is DRY,
idiomatic, and ready for automation or extension._

_For humans: See `README.md` for user-facing documentation and usage examples._

---

## Notice for Agents

A `TODO.md` file has been added to the project root. It contains a detailed plan for upcoming changes and improvements. Please refer to it for guidance on tasks and priorities.
