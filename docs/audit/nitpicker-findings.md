# Nitpicker Findings

Generated: 2026-05-04
Last validated: 2026-05-04

## Summary

- Total: 11 | Open: 2 | Fixed: 8 | Invalid: 1

## Open Findings

### Advisory

#### [N10] `GetEmbeddedTemplate` lacks explicit `..` rejection

Category: security
Area: templates_embed/embed.go:22-31
Problem: After stripping the leading `/` and prepending `templates/`, no explicit check rejects `..` components.
For example, `templates/../../etc/passwd` would be passed to `embeddedTemplates.ReadFile`.
Go's `embed.FS` rejects such paths at runtime (returns error), but the caller gets an opaque error rather than "invalid path."
Evidence: `embed.go:27-31`: prepends `templates/` without checking for `..`.
Impact: No path escape is possible (embed.FS is bounded at compile time), but error messages are confusing. Not exploitable.
Fix: Add `strings.Contains(cleanPath, "..")` guard before `embeddedTemplates.ReadFile`,
returning `filepath.ErrBadPattern` for consistency with `ReadTemplate`.

#### [N11] `testutil` package coverage 38.8%

Category: tests
Area: testutil/
Problem: The `testutil` package — which provides fixture builders and helpers used throughout the test suite —
has only 38.8% statement coverage. Bugs in test helpers go undetected and could produce misleading test results.
Evidence: `go test ./... -coverprofile` output: `coverage: 38.8% of statements` for `testutil`.
Impact: No production risk. Test reliability may be lower than reported.
Fix: Add unit tests for uncovered testutil functions, especially fixture builders and context helpers.

## Fixed

### Pass 1 — 2026-05-04

#### [N01] Production code imports test utility package

Fixed: 2026-05-04
Notes: Added `ActionFilenamePrefix = "action"` to `appconstants/files.go`.
Replaced `testutil.ConfigFieldAction` with `appconstants.ActionFilenamePrefix`
in `internal/apperrors/suggestions.go` and removed the `testutil` import.

#### [N02] `validateGitHubToken` bypasses canonical token retrieval

Fixed: 2026-05-04
Notes: Changed `globalConfig.GitHubToken == ""` to `internal.GetGitHubToken(globalConfig) == ""`
in `validateGitHubToken` (`cmd_deps.go:332`).
Updated test to clear env vars before the "empty token" case to prevent interference
from `GITHUB_TOKEN` in the environment.

#### [N03] `CreateDependencyAnalyzer` skips `GetGitHubToken`

Fixed: 2026-05-04
Notes: Changed `g.Config.GitHubToken` to `GetGitHubToken(g.Config)` at `internal/generator.go:108`,
ensuring env-var-only tokens are used for the GitHub API client.

#### [N04] `regexp.MustCompile` called on every invocation — 4 hot-path functions

Fixed: 2026-05-04
Notes: Promoted all regex patterns to package-level `var` in `internal/validation/strings.go`
(`reGitHubURLFull`, `reGitHubURLSimple`, `reWhitespace`) and `internal/validation/validation.go`
(`reCommitSHA`, `reSemanticVersion`). Functions now reference the package-level vars.

#### [N05] `isUnitTestEnvironment()` uses fragile `os.Args[0]` inspection

Fixed: 2026-05-04
Notes: Changed detection to `strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe")`.
This uses Go's documented test binary naming convention rather than searching for package-name substrings,
making it reliable across all packages and platforms.

#### [N06] `StdinReader.ReadLine()` creates a fresh `bufio.Reader` on every call

Fixed: 2026-05-04
Notes: Added `reader *bufio.Reader` field to `StdinReader` in `cmd_deps.go`.
`ReadLine()` now lazy-initializes the reader on first call and reuses it on subsequent calls,
preserving the internal buffer across reads.

#### [N07] Magic number `1` in `wrapHandlerWithErrorHandling`

Fixed: 2026-05-04
Notes: Changed `os.Exit(1)` to `os.Exit(appconstants.ExitCodeError)` in
`wrapHandlerWithErrorHandling` (`main.go:85`).

#### [N08] Deprecated runtimes `node12`/`node16` accepted without warning

Fixed: 2026-05-04
Notes: Added `NodeRuntimeNode12 = "node12"` and `NodeRuntimeNode16 = "node16"` constants
to `appconstants/constants.go`. Added `isDeprecatedRuntime()` function to `internal/validator.go`.
`ValidateActionYML` now appends to `result.Warnings` and `result.Suggestions` when a deprecated
runtime is detected, without failing validation.
Added `TestValidateActionYML_DeprecatedRuntime` test covering both runtimes.

## Invalid

### Pass 1 — 2026-05-04

#### [N09] Main package coverage below 72% threshold

Notes: The finding referenced 71.3% coverage for the `main` package, but the `make test-coverage-check`
target checks the **total** coverage across all packages.
After all fixes, `go tool cover -func` reports 74.0% total, which satisfies the 72.0% threshold.
Per-package coverage for the `main` package (70.9%) is not the metric enforced by CI.
