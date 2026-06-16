# Architecture Audit Findings

Generated: 2026-05-05
Last validated: 2026-06-16

## Summary

- Total: 11 | Open: 0 | Fixed: 10 | Invalid: 1
- Pass 3 re-validation (2026-06-16): all 8 prior Fixed findings re-confirmed still resolved; the 1 prior
Invalid finding (AA04) remains Invalid. 2 new findings filed (AA11, AA12), both Medium pattern/placement
inconsistencies.
- No Critical or High findings. `go build ./...` is clean — no import cycles, no reversed dependencies. Zero
production files import `testutil`. `GetGitHubToken()` is used consistently as the canonical token accessor.

## Open Findings

## Fixed

### Pass 3 — 2026-06-16

#### [AA11] Two Cobra handlers bypass the error-returning + `wrapHandlerWithErrorHandling` convention

Fixed: 2026-06-16
Notes: depsGraphHandler (cmd_deps.go) and schemaHandler (cmd_validate.go) now return error and are registered
via wrapHandlerWithErrorHandling, matching the other 13 handlers. (= nitpicker N90.)

#### [AA12] `InputReader` / `StdinReader` abstraction lives in the CLI layer instead of `internal/`

Fixed: 2026-06-16
Notes: InputReader and StdinReader moved from cmd_deps.go (CLI layer) to internal/input.go; cmd_deps.go now
references internal.InputReader / internal.StdinReader. CLAUDE.md updated to match.

Re-validation of prior findings. No code changes made in this pass; each prior Fixed finding was re-verified
against the current tree (`go list` import graph, targeted `grep`, `go build ./...`) and confirmed still
resolved.

- [AA01] Re-confirmed: `grep` for `gh-action-readme/testutil` across all non-`_test.go`, non-`testutil/` files
returns zero results.
- [AA02] Re-confirmed: `go list` shows `internal/helpers` imports only `internal/git`; no parent `internal`
import.
- [AA03] Re-confirmed (documented partial fix, intended state): `internal/wizard` still imports `internal` for
`AppConfig` / `MessageLogger` / `MessagingOutput` / `GetGitHubToken` / `DiscoverActionFilesNonRecursive`
(`detector.go`, `exporter.go`, `validator.go`, `wizard.go`). Coupling surface remains interface-only; full fix
(extract `AppConfig` to `internal/types`) deferred. Not a regression.
- [AA05] Re-confirmed: no non-`_test.go` file imports `testutil`.
- [AA06] Re-confirmed: no concrete `*ColoredOutput` field reintroduced.
- [AA07] Re-confirmed: interface-typed output parameters intact.
- [AA08] Re-confirmed: wizard constructors accept `internal.MessageLogger` / `internal.MessagingOutput`
(`detector.go:32`, `exporter.go:38`, `validator.go:60`, `wizard.go:29`).
- [AA09] Re-confirmed: `cmd_deps.go` functions accept narrow interfaces (e.g. `validateGitHubToken(output
internal.MessageLogger)` at `cmd_deps.go:343`); no `*internal.ColoredOutput` parameters.

### Pass 2 — 2026-05-05

#### [AA03] Four `internal/wizard` production files import parent package `internal`

Fixed: 2026-05-05
Notes: Partial fix — the `internal` import remains because `internal/wizard` needs `AppConfig` (defined in
`internal`).
The coupling surface was reduced by replacing all `*internal.ColoredOutput` fields/parameters in wizard files
with
`internal.MessageLogger` or `internal.MessagingOutput` interfaces (narrowest applicable). A full fix would
require moving
`AppConfig` to a shared `internal/types` package; deferred as it would be a larger refactor with no immediate
correctness
benefit.

#### [AA05] Three test-helper files import `testutil` without `_test.go` suffix

Fixed: 2026-05-05
Notes: Renamed all three files:

- `main_test_helper.go` → `main_test_helper_test.go`
- `templates_embed/embed_test_helpers.go` → `templates_embed/embed_test_helpers_test.go`
- `internal/wizard/detector_test_helper.go` → `internal/wizard/detector_test_helper_test.go`
All packages compile and all tests pass after rename.

#### [AA06] `internal/errorhandler.go` holds a concrete `*ColoredOutput` field instead of an interface

Fixed: 2026-05-05
Notes: Changed `ErrorHandler.output` from `*ColoredOutput` to `ErrorReporter`. Updated `NewErrorHandler`
signature to
accept `ErrorReporter`. `ErrorHandler` only ever calls `output.ErrorWithSuggestions(err)` — `ErrorReporter` is
the
narrowest interface covering that method.

#### [AA07] `internal/analyzer_helpers.go` accepts concrete `*ColoredOutput` instead of interface

Fixed: 2026-05-05
Notes: Changed `CreateAnalyzer` parameter from `*ColoredOutput` to `MessageLogger`. The function only calls
`output.Warning(...)`.

#### [AA08] `internal/wizard` constructors accept concrete `*ColoredOutput` instead of interface

Fixed: 2026-05-05
Notes: Updated all four wizard files:

- `wizard.go`: `*internal.ColoredOutput` → `internal.MessageLogger` (uses Info/Warning/Success/Bold/Printf)
- `exporter.go`: `*internal.ColoredOutput` → `internal.MessageLogger` (uses Success only)
- `detector.go`: `*internal.ColoredOutput` → `internal.MessageLogger` (uses Warning/Success)
- `validator.go`: `*internal.ColoredOutput` → `internal.MessagingOutput` (uses both Error and MessageLogger
methods; `MessagingOutput` added to `internal/interfaces.go`)
Also removed the redundant local aliases (`permScopeContents`, `permScopeIssues`, `permissionRead`,
`permissionWrite`) from
`validator.go` and replaced all references in wizard test files with `appconstants` constants directly.

#### [AA09] `cmd_deps.go` internal functions accept `*internal.ColoredOutput` (12 occurrences)

Fixed: 2026-05-05
Notes: Replaced all `*internal.ColoredOutput` parameters in `cmd_deps.go` with narrower interfaces:

- `analyzeDependencies`: `OutputWriter` (calls `IsQuiet()` + MessageLogger methods)
- `analyzeSecurityDeps`: `OutputWriter` (calls `IsQuiet()` + MessageLogger methods)
- All 10 remaining functions: `MessageLogger` (use only Info/Success/Warning/Bold/Printf)

### Pass 1 — 2026-05-05

#### [AA01] `internal/dependencies/updater_test_helper.go` imports `testutil` in production-compiled file

Fixed: 2026-05-05
Notes: Renamed `updater_test_helper.go` → `updater_test_helper_test.go`. File now has `_test.go` suffix and is
excluded
from the production binary. Package declaration kept as `package dependencies` (white-box test access).

#### [AA02] `internal/helpers` imports parent package `internal`

Fixed: 2026-05-05
Notes: Deleted `internal/helpers/analyzer.go` (which imported `internal` for
`CreateAnalyzer`/`CreateAnalyzerOrExit`).
Moved those functions into `internal/analyzer_helpers.go` (same package as `internal`, no import cycle).
Removed dead
function `SetupGeneratorContext` from `internal/helpers/common.go` (zero callers confirmed).
`internal/helpers` now
imports only `internal/git` — no parent package import.

## Invalid

### Pass 2 — 2026-06-16

No findings invalidated in this pass.

### Pass 1 — 2026-05-05

#### [AA04] `internal/testoutput.go` ships test doubles in production-compiled code

Invalid: 2026-05-05 (re-confirmed 2026-06-16)
Notes: Finding was wrong. `internal/generator.go` calls `NewNullOutput()` and `NewNullProgressManager()`
inside
production `NewGenerator()` via `isUnitTestEnvironment()`. These null implementations are required at compile
time by
production code, not exclusively by tests. Renaming to `_test.go` causes undefined symbol errors. The
functions are
legitimately production code.

#### [AA10 — NOT FILED] `main.go:84,159` and `internal/errorhandler.go:28` call `os.Exit` — not a violation

Notes: arch-profile Rule 6 says "command handlers must return `error` — never call `os.Exit()` in handlers;
`wrapHandlerWithErrorHandling` owns exit logic." `main.go:84` IS inside `wrapHandlerWithErrorHandling` — the
designated
exit point. `main.go:159` is in `main()` after `rootCmd.Execute()` — the correct terminal location.
`internal/errorhandler.go:28` is `ErrorHandler.HandleError()`, a dedicated exit-on-fatal-error function that
deliberately
owns the exit path. None are handler functions. No violation; finding not filed. Re-confirmed 2026-06-16:
these are still
the only three `os.Exit` sites in non-test code.
