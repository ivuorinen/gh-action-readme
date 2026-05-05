# Documentation Audit Findings

Generated: 2026-05-05
Last validated: 2026-05-05

## Summary

- Total: 15 | Open: 0 | Fixed: 15 | Invalid: 0

## Open Findings

No open findings.

## Fixed

### Pass 3 — 2026-05-05

#### [DA13] `applyUpdates()` and `setupDepsUpgrade()` line numbers stale after linting refactor

Fixed: 2026-05-05
Notes: Updated CLAUDE.md lines 139-140: `applyUpdates()` → `cmd_deps.go:560`; `setupDepsUpgrade()` → `cmd_deps.go:455`. Both functions shifted during the linting refactor commit `5af26be`.

#### [DA14] "New Theme" guide shows switch/case but implementation uses a map

Fixed: 2026-05-05
Notes: Replaced the `case appconstants.ThemeTHEMENAME: templatePath = ...` code snippet in CLAUDE.md with the correct
map-entry form `appconstants.ThemeTHEMENAME: appconstants.TemplatePathTHEMENAME,` and updated the instruction to
reference the `themeTemplates` map at `internal/config.go:189`.

#### [DA15] "New Output Format" guide targets `GenerateFromFile()` but format switch is in `generateByFormat()`

Fixed: 2026-05-05
Notes: Updated CLAUDE.md to reference `internal/generator.go:generateByFormat()` (line ~532) instead of `GenerateFromFile()`.
The `generateByFormat` function contains the `switch g.Config.OutputFormat` block where new format cases must be added.

### Pass 2 — 2026-05-05

#### [DA08] gremlins described as "Go 1.25+ compatible" but project is on Go 1.26.2

Fixed: 2026-05-05
Notes: Changed "Go 1.25+" to "Go 1.26+" in CLAUDE.md line 303.

#### [DA09] CLAUDE.md states current coverage is 72.8% but Makefile says 73.7%

Fixed: 2026-05-05
Notes: Replaced `72.8% overall` with `varies (run make test-coverage to check; target: 80%)` so
the line cannot become stale.

#### [DA10] `TestInputReader` shown without noting it is test-only

Fixed: 2026-05-05
Notes: Added "(defined in `main_test.go` — test-only)" to the code block header.

#### [DA11] `testdata/yaml-fixtures/` fixture structure diagram is incomplete

Fixed: 2026-05-05
Notes: Updated the fixture structure diagram in CLAUDE.md to include all current subdirectories:
`json-fixtures/`, `permissions/`, `scenarios/`, `template-fixtures/`, `validation/` at the top
level; `json-writer/`, `minimal/`, `simple/` under `actions/`.

#### [DA12] `cmd_validate.go` absent from Package Structure

Fixed: 2026-05-05
Notes: Added `cmd_validate.go` — `validate` command implementation to Package Structure in CLAUDE.md.

### Pass 1 — 2026-05-05

#### [DA01] `applyUpdates()` and `setupDepsUpgrade()` attributed to wrong file with wrong line numbers

Fixed: 2026-05-05
Notes: CLAUDE.md lines 139–140 now correctly reference `cmd_deps.go:549` and `cmd_deps.go:444`.

#### [DA02] Package structure lists `internal/errors/` which does not exist

Fixed: 2026-05-05
Notes: CLAUDE.md line 574 now reads `internal/apperrors/`. No remaining references to `internal/errors/`.

#### [DA03] `internal/cache/` package entirely absent from Package Structure

Fixed: 2026-05-05
Notes: CLAUDE.md line 575 now lists `internal/cache/` with description.

#### [DA04] `cmd_cache.go` entirely absent from Package Structure and command list

Fixed: 2026-05-05
Notes: CLAUDE.md line 561 now lists `cmd_cache.go`.

#### [DA05] `internal/analyzer_helpers.go` absent from Package Structure

Fixed: 2026-05-05
Notes: CLAUDE.md line 563 now lists `internal/analyzer_helpers.go`.

#### [DA06] `configThemesHandler` documented as being in `main.go` but is in `cmd_config.go`

Fixed: 2026-05-05
Notes: CLAUDE.md line 503 now reads `cmd_config.go:configThemesHandler()`.

#### [DA07] Contradictory picture from DA06 about `configThemesHandler` location

Fixed: 2026-05-05
Notes: Covered by DA06 fix. No separate change needed.

## Invalid

No invalid findings.
