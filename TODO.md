# TODO.md

This file tracks planned improvements and changes for the `gh-action-readme` project.
Each task is sorted from simple to expensive, with detailed descriptions for both humans and LLMs.

---

## Planned Actions

### Simple Tasks

1. **Support for switching main templates:**
   - Allow users to specify custom `readme.md.tmpl` or `readme.html.tmpl` files via configuration.
2. **Centralize helper functions:**
   - Move all helper functions to `internal/helpers.go`.
   - Refactor the codebase to use the centralized helpers for cleaner organization.
3. **Validate anchors and references in YAML:**
   - Enhance validation to ensure YAML anchors and references are correctly resolved.
4. **Create validation examples and document them:**
   - Add examples of valid and invalid `action.yml` files to `testdata/`.
   - Document these examples in `README.md` or a dedicated section.
5. **Add support for `config.yml` and validate `config.yaml`:**
   - Validate the `config.yaml` file.
   - Add support for `config.yml` as an alternative.
   - Embed the default `config.yaml` into the binary.
   - Provide a CLI command to generate a baseline configuration for users.
6. **Add global `--ci` flag:**
   - Introduce a `--ci` flag to disable progress bars and other interactive features
     for CI environments.

### Moderate Tasks

1. **Add versioning and fallback for schemas:**
   - Implement a versioning system for `action.schema.json`.
   - Allow users to specify a custom schema file.
   - Validate the schema file before use.
2. **Support all template files in configuration:**
   - Extend configuration to allow specifying paths for all template files
     (e.g., header, footer, main templates).
3. **Add progress bars for long-running operations:**
   - Implement progress bars for tasks like recursive action discovery.
   - Ensure the `--ci` flag disables this feature.

### Expensive Tasks

1. **Document automation and validation examples:**
   - Create detailed examples for CI/CD pipelines.
   - Add documentation for validation scenarios and edge cases.

2. **Optimize recursive action discovery:**
   - Use goroutines to parallelize file discovery and parsing for better performance.

---

## Completed Tasks

1. **Add `make clear` and `make check` targets to the Makefile:**
   - `make clear`: Removes temporary files (e.g., `coverage.out`, `fail.log`).
   - `make check`: Combines linting and testing.
   - Ensure `make lint` runs all tools even if some fail.
   - **Status:** Completed on July 8, 2025.

---

## Notes

- Each task should be implemented incrementally and tested thoroughly.
- Update this file as tasks are completed or new tasks are added.
