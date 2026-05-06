---
name: fixture-auditor
description: Finds all inline YAML and config data in test files that violates the no-inline-yaml-in-tests rule and produces a migration plan
---

Search all `*_test.go` files for inline YAML or configuration data. Flag any of:

- Multi-line backtick strings containing YAML keys (`name:`, `runs:`, `inputs:`, `outputs:`, `theme:`, `output_format:`, `on:`)
- `strings.NewReader(` with YAML content as a literal argument
- `[]byte(` with YAML content as a literal argument
- `yaml.Unmarshal([]byte(` with inline content

For each violation:

1. Show file:line and a snippet of the inline content (first 5 lines)
2. Suggest the correct fixture path under `testdata/yaml-fixtures/` based on content type
3. Suggest the constant name following the `TestFixture<Name>` pattern
4. Show the exact replacement code using `testutil.WriteActionFixture()` or `testutil.MustReadFixture()`

Group results by package. List packages in order of violation count (worst first).

Summary table:

| Package | Violations | Estimated effort     |
|---------|------------|----------------------|
| ...     | N          | N fixtures to create |

Total violation count across all packages.
