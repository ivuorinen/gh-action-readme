---
name: coverage-gap-finder
description: Identifies untested functions in packages below the 72% threshold and produces a prioritized list of test additions ordered by coverage gain
---

1. Run `go test -coverprofile=coverage.out -covermode=atomic ./...`
2. Run `go tool cover -func=coverage.out` and parse all output lines.
3. Identify packages where the package-level total is below 72.0%.
4. For each low-coverage package, list every function at 0.0% coverage.

For each 0% function, classify it as:

- **EASY** — pure logic, no I/O, no global state — directly testable
- **INJECTABLE** — takes dependencies as parameters — testable with the existing mock/test pattern
- **HARD** — interactive stdin, os.Exit, or external system side effects — requires significant scaffolding
- **COVERED_INTEGRATION** — called by integration_test.go or testdata/-based tests — not truly untested

For EASY and INJECTABLE functions, estimate the coverage gain (statements covered / total statements in the package) if a test were added.

Output:

- Prioritized list sorted by estimated coverage gain (largest first)
- Each item: function name, file:line, classification, estimated gain in percentage points
- Packages that would cross 72% with one or two test additions highlighted
- Overall projection: what total coverage would be if all EASY+INJECTABLE items were tested
