---
name: coverage-check
description: Run test coverage check and report per-package breakdown highlighting packages below the 72% gate and 80% new-code target
---

Run the following in the project root:

1. `make test-coverage-check` — verifies overall coverage meets the 72.0% gate. Report PASS or FAIL.
2. `go test -coverprofile=coverage.out -covermode=atomic ./...` followed by `go tool cover -func=coverage.out`
3. Parse the output and produce a sorted table with two threshold columns:
    - **Gate (72%)**: overall minimum enforced by `make test-coverage-check`
    - **Target (80%)**: goal for all new Go code (`**/*.go`)

    | Package           | Coverage | Gate (72%) | Target (80%)        |
    |-------------------|----------|------------|---------------------|
    | ./internal/wizard | 61.2%    | FAIL       | FAIL (gap: 18.8pp)  |
    | ./internal        | 74.1%    | PASS       | FAIL (gap: 5.9pp)   |

4. Show the overall `total:` line at the bottom.
5. List all packages below 72% (gate failures) with their gap in percentage points.
6. List all packages below 80% (target misses) separately.
7. State clearly: PASS (≥72% overall) or FAIL (<72% overall).
