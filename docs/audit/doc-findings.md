# Documentation Audit Findings

Generated: 2026-05-05
Last validated: 2026-06-16

## Summary

- Total: 22 | Open: 0 | Fixed: 22 | Invalid: 0

## Open Findings

## Fixed

### Pass 4 — 2026-06-16

#### [DA16] `docs/configuration.md` documents three Performance Settings config keys that do not exist

Fixed: 2026-06-16
Notes: Removed the phantom Performance Settings section and the GH_ACTION_README_CACHE_TTL/_TIMEOUT env
exports from docs/configuration.md (keys absent from AppConfig).

#### [DA17] `docs/development.md` documents a `GenerateReadme()` function that does not exist

Fixed: 2026-06-16
Notes: docs/development.md replaced the fictional GenerateReadme() with the real RenderReadme /
Generator.GenerateFromFile / ProcessBatch entry points.

#### [DA18] `docs/development.md` "New Theme" steps point at the wrong function and wrong file

Fixed: 2026-06-16
Notes: docs/development.md 'New Theme' steps corrected to: add constant in appconstants/themes.go, add entry
to the themeTemplates map in internal/config.go, update configThemesHandler in cmd_config.go.

#### [DA19] `docs/configuration.md` env var `GH_ACTION_README_DEPENDENCIES` does not map to a real key

Fixed: 2026-06-16
Notes: docs/configuration.md env example corrected to GH_ACTION_README_ANALYZE_DEPENDENCIES (the real
analyze_dependencies key).

#### [DA20] CLAUDE.md "New Output Format" line reference for `generateByFormat()` is stale

Fixed: 2026-06-16
Notes: CLAUDE.md 'New Output Format' generateByFormat() line reference made non-line-specific to stop it
drifting (was ~532, now 577).

#### [DA21] `docs/configuration.md` "Theme Directory Structure" shows `partials/` and `assets/` that no theme uses

Fixed: 2026-06-16
Notes: docs/configuration.md Theme Directory Structure reduced to the real readme.tmpl/.adoc layout; the
nonexistent partials/ and assets/ removed.

#### [DA22] CLAUDE.md / docs label AsciiDoc a "theme" though it is only an output format

Fixed: 2026-06-16
Notes: AsciiDoc relabeled as an output format (not a --theme) in CLAUDE.md and docs/development.md.

### Pass 3 — 2026-05-05

#### [DA13] `applyUpdates()` and `setupDepsUpgrade()` line numbers stale after linting refactor

Fixed: 2026-05-05
Notes: Updated CLAUDE.md lines 139-140: `applyUpdates()` → `cmd_deps.go:560`; `setupDepsUpgrade()` →
`cmd_deps.go:455`. Both functions shifted during the linting refactor commit `5af26be`. Re-validated
2026-06-16: `setupDepsUpgrade` is at cmd_deps.go:455 and `applyUpdates` is at cmd_deps.go:562; CLAUDE.md
already reads 562/455 — still accurate.

#### [DA14] "New Theme" guide shows switch/case but implementation uses a map

Fixed: 2026-05-05
Notes: Replaced the `case appconstants.ThemeTHEMENAME: templatePath = ...` code snippet in CLAUDE.md with the
correct
map-entry form `appconstants.ThemeTHEMENAME: appconstants.TemplatePathTHEMENAME,` and updated the instruction
to
reference the `themeTemplates` map at `internal/config.go:189`. Re-validated 2026-06-16: map still at
config.go:189 and CLAUDE.md still correct. (Note: docs/development.md has the same uncorrected drift —
see new finding DA18.)

#### [DA15] "New Output Format" guide targets `GenerateFromFile()` but format switch is in `generateByFormat()`

Fixed: 2026-05-05
Notes: Updated CLAUDE.md to reference `internal/generator.go:generateByFormat()` (line ~532) instead of
`GenerateFromFile()`.
The `generateByFormat` function contains the `switch g.Config.OutputFormat` block where new format cases must
be added.
Re-validated 2026-06-16: function name/file still correct; the approximate line number has since drifted
to 577 — see new finding DA20.

### Pass 2 — 2026-05-05

#### [DA08] gremlins described as "Go 1.25+ compatible" but project is on Go 1.26.3

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
Notes: CLAUDE.md lines 139–140 now correctly reference `cmd_deps.go:560` and `cmd_deps.go:455` (updated again
in DA13 after the linting refactor).

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
Notes: CLAUDE.md line 503 now reads `cmd_config.go:configThemesHandler()`. (Note: docs/development.md
still has the same uncorrected error — see new finding DA18.)

#### [DA07] Contradictory picture from DA06 about `configThemesHandler` location

Fixed: 2026-05-05
Notes: Covered by DA06 fix. No separate change needed.

## Invalid

No invalid findings.
