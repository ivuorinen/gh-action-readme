---
name: arch-boundary-checker
description: Checks for package import direction violations against the architectural rules in docs/audit/arch-findings.md
---

Read `docs/audit/arch-findings.md` for the current open and known findings first.

Audit all Go packages for violations of these structural rules:

1. **testutil isolation** — `testutil` must not be imported by any non-`_test.go` file. Search all `.go` files (excluding `_test.go`) for `"github.com/ivuorinen/gh-action-readme/testutil"`.

2. **Subpackage → parent import** — subpackages of `internal` must not import the `internal` root package.
    Known open exception: `internal/wizard` imports `internal` (AA03 — do not re-flag, just note as KNOWN OPEN).
    Check all other `internal/` subpackages.

3. **L0–L2 upward import** — packages in `appconstants`, `internal/git`, `internal/cache`,
    `internal/apperrors`, `internal/validation`, `internal/dependencies` must not import `internal` root or `main`.

4. **main imports downward only** — `main` package must only import L3 and below. It must not directly import `testutil`.

For each package, parse its import block and check against these rules.

Report:

- **PASS** — no violations
- **VIOLATION** — file:line | importing package | imported package | rule violated | minimal fix
- **KNOWN OPEN** — AA03 (internal/wizard → internal): noted but not flagged as new

Update `docs/audit/arch-findings.md` only if new violations are found.
