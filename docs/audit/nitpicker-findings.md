# Nitpicker Findings

Generated: 2026-05-04
Last validated: 2026-05-05

## Summary

- Total: 45 | Open: 0 | Fixed: 42 | Invalid: 3

## Open Findings

## Fixed

### Pass 9 — 2026-05-05

#### [N11] `testutil` package reports 38.8% self-coverage

Fixed: 2026-05-05
Notes: Added `testutil/test_suites_test.go` with 80+ test cases covering RunTestSuite,
RunActionTests, RunGeneratorTests, RunValidationTests, TestAllThemes, TestAllFormats,
TestValidationScenarios, and all helper functions. Coverage improved from 38.8% to 71.6%.

### Pass 8 — 2026-05-05

#### [N45] `applyUpdatesToLines` match fires on `uses:` embedded inside an inline comment of another field

Fixed: 2026-05-05
Notes: Added `if commentIdx := strings.Index(line, "#"); commentIdx >= 0 && commentIdx < idx { continue }`
in `applyUpdatesToLines` (`internal/dependencies/analyzer.go:662`). A line such as
`description: "see uses: actions/checkout@v4"` would pass the leading-`#` guard but contain
a `uses:` match at a position after the first `#`, so the new guard correctly skips it.

### Pass 7 — 2026-05-05

#### [N41] `applyUpdatesToLines` version-prefix collision corrupts longer version refs

Fixed: 2026-05-05
Notes: Added two guards to `applyUpdatesToLines` in `internal/dependencies/analyzer.go:649`:
(1) skip comment lines (`strings.HasPrefix(trimmedLine, "#")`);
(2) require `OldUses` to be a complete token by checking `afterTarget` is empty or starts with
`#` (prevents `actions/checkout@v4` from matching `actions/checkout@v4.1.1`).
Confirmed by both adversarial-reviewer (pass 3) and security-auditor (SEC-024).

#### [N42] CLAUDE.md `applyUpdates()`/`setupDepsUpgrade()` line numbers stale

Fixed: 2026-05-05
Notes: Updated `cmd_deps.go:549` → `cmd_deps.go:560` and `cmd_deps.go:444` → `cmd_deps.go:455`
in CLAUDE.md after linting refactor shifted both functions.

#### [N43] CLAUDE.md "New Theme" guide shows switch/case; implementation uses a map

Fixed: 2026-05-05
Notes: Replaced the `case appconstants.ThemeTHEMENAME:` code snippet with the correct map-entry
form `appconstants.ThemeTHEMENAME: appconstants.TemplatePathTHEMENAME,` referencing the
`themeTemplates` map at `internal/config.go:189`.

#### [N44] CLAUDE.md "New Output Format" guide targets `GenerateFromFile()` but switch is in `generateByFormat()`

Fixed: 2026-05-05
Notes: Updated instruction to reference `internal/generator.go:generateByFormat()` (~line 532)
where the `switch g.Config.OutputFormat` block lives.

### Pass 6 — 2026-05-05

#### [N33] `validateActionType` missing `node24` runtime

Fixed: 2026-05-05
Notes: Added `"node24"` to `validTypes` slice in `validateActionType` at
`internal/dependencies/analyzer.go:241`. Added `TestValidateActionType` covering all valid
runtimes and one invalid runtime to prevent future regressions.

#### [N34] `applyUpdatesToLines` substring match corrupts non-`uses:` fields

Fixed: 2026-05-05
Notes: Changed `strings.Contains(line, update.OldUses)` to
`strings.Contains(line, appconstants.UsesFieldPrefix+update.OldUses)` in
`applyUpdatesToLines` (`internal/dependencies/analyzer.go:650`). The old check matched any
field whose value contained the action reference (e.g. `description: uses actions/checkout@v4`).
Updated two tests that relied on the bug to corrupt non-`uses:` fields: both now use
`OldUses: "actions/checkout@v4"` and `NewUses: "{unclosed"` to trigger an actual YAML parse
failure for rollback testing.

#### [N35] Backup file leaked when `WriteFile` fails in `updateActionFile`

Fixed: 2026-05-05
Notes: Added `_ = os.Remove(backupPath)` before the `return` when `os.WriteFile(cleanPath,...)`
fails in `updateActionFile` (`internal/dependencies/analyzer.go:631`). Previously the `.backup`
file was created and left on disk when the write failed.

#### [N36] `StdinReader.ReadLine()` propagates `io.EOF` on last line without trailing newline

Fixed: 2026-05-05
Notes: Added `if err == io.EOF && trimmed != ""` early return with nil error in
`StdinReader.ReadLine()` (`cmd_deps.go:44`). `bufio.ReadString('\n')` returns `(data, io.EOF)`
simultaneously on the final line when stdin has no trailing newline (common in piped CI usage
such as `echo "y" | gh-action-readme deps upgrade`). Callers treating any non-nil error as
failure would reject valid input.

#### [N37] `depsListHandler` calls `analyzeDependencies` with nil `actionFiles` after swallowed error

Fixed: 2026-05-05
Notes: Added `if len(actionFiles) == 0 { return nil }` guard in `depsListHandler` after
the `handleNoFilesFoundError` check (`cmd_deps.go:121`). When no action files are found,
`handleNoFilesFoundError` swallows the error and returns nil; without the guard, execution
falls through to `analyzeDependencies(output, nil, analyzer)` which prints a misleading
"Dependencies found in action files:" header with no files.

#### [N38] CLAUDE.md Go version "1.25+" contradicts `go.mod` 1.26.2

Fixed: 2026-05-05
Notes: Changed "Go 1.25+" to "Go 1.26+" in the gremlins tool note (`CLAUDE.md:303`).
The module requires go 1.26.2 per `go.mod`.

#### [N39] CLAUDE.md hardcoded coverage percentage becomes stale

Fixed: 2026-05-05
Notes: Replaced `72.8% overall` with `varies (run make test-coverage to check; target: 80%)`
so the line doesn't silently go stale as coverage changes.

#### [N40] CLAUDE.md `TestInputReader` presented as production code without location note

Fixed: 2026-05-05
Notes: Added "(defined in `main_test.go` — test-only)" annotation to the `TestInputReader`
code block header in CLAUDE.md to clarify it is not a production type.

### Pass 5 — 2026-05-05

#### [N24] `shouldIgnoreDirectory` prefix match silently drops `.github/actions/`

Fixed: 2026-05-05
Notes: Removed the dotfile prefix-matching branch from `shouldIgnoreDirectory`. All
directories now use exact match. The default `ignoredDirs` comment already said "keep
.github searchable for .github/actions" — the implementation now matches that intent.
Updated `TestShouldIgnoreDirectory` test cases: "dot prefix pattern match - .github" and
".gitlab" changed from `want: true` to `want: false`; added explicit test asserting
`.github` IS ignored when `DirGitHub` is explicitly in the list.

#### [N25] `walkFunc` aborts entire recursive walk on permission-denied subdirectory

Fixed: 2026-05-05
Notes: Changed `return err` to `return filepath.SkipDir` when `os.IsPermission(err)` in
`walkFunc`. Other errors (e.g. I/O errors on the root) still propagate. Trigger:
`chmod 000 vendor && gh-action-readme gen --recursive .` now yields partial results
instead of a hard error.

#### [N26] `parseGitHubURL` compiles two regexps on every invocation

Fixed: 2026-05-05
Notes: Promoted both GitHub URL patterns to package-level `var` (`reGitHubURLNoSuffix`,
`reGitHubURLWithSuffix`) in `internal/git/detector.go`. Updated `parseGitHubURL` to
iterate over the compiled vars instead of calling `regexp.MustCompile` per call.

#### [N27] Fake GitHub token in test fixture triggers gitleaks

Fixed: 2026-05-05
Notes: Added `# gitleaks:allow` to `testdata/yaml-fixtures/configs/global-config-default.yml`
line containing `ghp_test1234567890abcdefghijklmnopqrstuvwxyz`. The value is obviously
synthetic but matches the GHP token pattern, causing false positives in gitleaks scans.

#### [N28] CLAUDE.md attributes `applyUpdates()` and `setupDepsUpgrade()` to `main.go`

Fixed: 2026-05-05
Notes: Both functions are in `cmd_deps.go`. Updated line references:
`applyUpdates()` → `cmd_deps.go:549`; `setupDepsUpgrade()` → `cmd_deps.go:444`.

#### [N29] CLAUDE.md Package Structure lists non-existent `internal/errors/`

Fixed: 2026-05-05
Notes: Changed `internal/errors/` to `internal/apperrors/` (the actual package path).

#### [N30] CLAUDE.md Package Structure missing `internal/cache/`, `internal/analyzer_helpers.go`, and command files

Fixed: 2026-05-05
Notes: Added `cmd_gen.go`, `cmd_deps.go`, `cmd_config.go`, `cmd_cache.go`,
`internal/analyzer_helpers.go`, and `internal/cache/` to the Package Structure list.

#### [N31] CLAUDE.md "New Theme" step 4 points to `main.go:configThemesHandler()`

Fixed: 2026-05-05
Notes: Changed to `cmd_config.go:configThemesHandler()` (function is at line 126 of that file).

### Pass 4 — 2026-05-05

#### [N10] `GetEmbeddedTemplate` lacks explicit `..` rejection

Fixed: 2026-05-05
Notes: Added `strings.Contains(cleanPath, "..")` guard in `GetEmbeddedTemplate` returning
`filepath.ErrBadPattern` for consistency with `ReadTemplate`. Applied the same guard to
`IsEmbeddedTemplateAvailable` (returns `false`). embed.FS already prevents escape at runtime,
so this is a clarity and defensive-coding fix only.

#### [N23] `CreateAnalyzerOrExit` dead exported function

Fixed: 2026-05-05
Notes: Removed `CreateAnalyzerOrExit` from `internal/analyzer_helpers.go` — it was a one-line
alias for `CreateAnalyzer` with no additional behaviour despite the name implying exit-on-failure.
Only caller was `TestCreateAnalyzerOrExit`; that test was removed. The covered scenario
(successful analyzer creation with valid config) is already tested by `TestCreateAnalyzer`.

### Pass 3 — 2026-05-05

#### [N19] `setupDepsUpgrade` checks `config.GitHubToken` directly, bypassing env-var lookup

Fixed: 2026-05-05
Notes: Replaced `config.GitHubToken == ""` with `internal.GetGitHubToken(config) == ""` in
`setupDepsUpgrade`. Moved the token check before `CreateDependencyAnalyzer()` (fail-fast).
Updated the "no GitHub token" test case to clear both `EnvGitHubToken` and `EnvGitHubTokenStandard`
env vars so the test is reliable in CI environments where `GITHUB_TOKEN` may be set.

#### [N20] `depsUpgradeHandler` has 0% coverage

Fixed: 2026-05-05
Notes: Added `TestDepsUpgradeHandlerErrorPaths` covering the "no GitHub token" error path.
Coverage for `depsUpgradeHandler` moved from 0.0% to 27.3%; `setupDepsUpgrade` moved to 81.2%.
The success/dry-run paths are covered by integration tests that run the compiled binary against
real action files.

#### [N21] `configRootHandler` and `configWizardHandler` have 0% coverage

Fixed: 2026-05-05
Notes: Added `TestConfigRootHandler` — `configRootHandler` coverage moved from 0.0% to 75.0%.
`configWizardHandler` remains at 0.0%: testing it requires mocking stdin (the wizard uses
interactive terminal input); the wizard package's own test suite covers this path indirectly.

#### [N22] `setupDepsUpgrade` silently discards its output parameter

Fixed: 2026-05-05
Notes: Removed `_ *internal.ColoredOutput` from the `setupDepsUpgrade` signature. Updated
call sites in `depsUpgradeHandler` (cmd_deps.go:407) and `TestSetupDepsUpgrade` (main_test.go:2752).

### Pass 2 — 2026-05-05

#### [N12] `main_test.go` redefines 22 constants already in `appconstants` or `testutil`

Fixed: 2026-05-05
Notes: Removed all 22 local constants from the const block. Replaced usages across
`main_test.go` and `integration_test.go` with `appconstants.Command*`, `appconstants.OutputFormat*`,
`appconstants.Theme*`, `testutil.TestFlagOutputFormat`, `testutil.TestFlagTheme`,
`testutil.TestMinimalAction`, and `testutil.TestTmpDir`.

#### [N13] `configuration_loader_test.go` format constants duplicate `appconstants`

Fixed: 2026-05-05
Notes: Removed `testLoaderFormatMD/HTML/JSON` const block. Replaced all usages in
`configuration_loader_test.go` and `configuration_loader_helper_test.go` with
`appconstants.OutputFormatMarkdown`, `appconstants.OutputFormatHTML`, `appconstants.OutputFormatJSON`.

#### [N14] `parser_test.go` permission constants duplicate `testutil.Permission*`

Fixed: 2026-05-05
Notes: Removed `testPermissionRead/Write/Contents/Issues` from `parser_test.go`. Replaced all
usages in `parser_test.go` and `parser_mutation_test.go` with `testutil.PermissionRead`,
`testutil.PermissionWrite`, `testutil.PermissionContents`, `testutil.PermissionIssues`.

#### [N15] `template_test.go` still uses `testPermissionActions` as repo name (4 remaining)

Fixed: 2026-05-05
Notes: Replaced all 4 remaining `testPermissionActions` usages as repo name in `template_test.go`
(lines 260, 282, 381, 395) with `testTplActionsRepo`.

#### [N16] `"testorg"`/`"testrepo"` duplicated across `internal` and `wizard` package tests

Fixed: 2026-05-05
Notes: Changed `testTplTestOrg = "testorg"` → `testutil.WizardOrgTest` and
`testTplTestRepo = "testrepo"` → `testutil.WizardRepoTest` in `template_test.go`.
Changed `testWizardOrg/testWizardRepo` to the same testutil constants in `exporter_test.go`.
Used existing `testutil.WizardOrgTest/WizardRepoTest` (already defined with correct values)
rather than adding new constants.

#### [N17] `config_test.go` defines `testConfigPermWrite = "write"` — duplicates `testutil.PermissionWrite`

Fixed: 2026-05-05
Notes: Changed `testConfigPermWrite = "write"` to `testConfigPermWrite = testutil.PermissionWrite`
in `internal/config_test.go`.

#### [N18] `wizard/validator.go` production permission constants should be in `appconstants`

Fixed: 2026-05-05
Notes: Added `PermissionRead`, `PermissionWrite`, `PermScopeContents`, `PermScopeIssues` to
`appconstants/constants.go`. Updated `wizard/validator.go` local const aliases to reference
`appconstants.*`. Updated `testutil/test_constants.go` `Permission*` constants to reference
`appconstants.*` (adding import), eliminating the dual source of truth.

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
`wrapHandlerWithErrorHandling` (`main.go:85`). Also fixed the second `os.Exit(1)` at
`main.go:160` that was missed in the initial pass.

#### [N08] Deprecated runtimes `node12`/`node16` accepted without warning

Fixed: 2026-05-04
Notes: Added `NodeRuntimeNode12 = "node12"` and `NodeRuntimeNode16 = "node16"` constants
to `appconstants/constants.go`. Added `isDeprecatedRuntime()` function to `internal/validator.go`.
`ValidateActionYML` now appends to `result.Warnings` and `result.Suggestions` when a deprecated
runtime is detected, without failing validation.
Added `TestValidateActionYML_DeprecatedRuntime` test covering both runtimes.

## Invalid

### Pass 5 — 2026-05-05

#### [N32] `ReadTemplate` reads arbitrary absolute paths

Notes: Developer CLI tool — the `--template` flag is intentionally user-controlled for
custom template support. The `#nosec G304` annotation documents this design choice.
Restricting absolute paths to a prefix would break legitimate custom-template workflows
(e.g. `--template /home/user/templates/my.tmpl`). Not exploitable without local machine
access, which implies the "attacker" is the authorized user.

### Pass 1 — 2026-05-04

#### [N09] Main package coverage below 72% threshold

Notes: The finding referenced 71.3% coverage for the `main` package, but the `make test-coverage-check`
target checks the **total** coverage across all packages.
After all fixes, `go tool cover -func` reports 74.0% total, which satisfies the 72.0% threshold.
Per-package coverage for the `main` package (70.9%) is not the metric enforced by CI.
