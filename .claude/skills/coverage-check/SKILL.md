---
name: coverage-check
description: Run test coverage check and report per-package breakdown highlighting packages below the 72% threshold
---

Run the following in the project root:

1. `make test-coverage-check` — verifies overall coverage meets the 72.0% threshold. Report PASS or FAIL.
2. `go test -coverprofile=coverage.out -covermode=atomic ./...` followed by `go tool cover -func=coverage.out`
3. Parse the output and produce a sorted table:

    | Package           | Coverage | Status             |
    |-------------------|----------|--------------------|
    | ./internal/wizard | 61.2%    | FAIL (gap: 10.8pp) |
    | ./internal        | 74.1%    | PASS               |

4. Show the overall `total:` line at the bottom.
5. List all packages below 72% with their gap in percentage points.
6. State clearly: PASS (≥72%) or FAIL (<72%).
