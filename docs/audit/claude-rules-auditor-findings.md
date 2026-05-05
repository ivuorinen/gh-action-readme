# Claude Rules Audit Findings

Generated: 2026-05-05
Last validated: 2026-05-05

## Summary

- Rules files audited: 11 (6 project `.claude/rules/` + 5 user `~/.claude/rules/`)
- CLAUDE.md files audited: 3 (project, user-global, home)
- Validation errors: 0 | Misplaced rules: 0 | Redundant rules: 0 | Suggestions: 1

## Open Findings

### Advisory

#### [RA19] Project CLAUDE.md still 871 lines — RA02 partial fix did not reduce line count

Category: suggestion
Area: `CLAUDE.md`
Problem: RA02 was marked Fixed in Pass 1 with note "Severity reduced to Advisory after rule migration." However, migrating
rules to `.claude/rules/` files while leaving source text in CLAUDE.md did not reduce the line count at all — the file
remains at 871 lines, still above the 400-line High threshold. Removing the redundant sections from RA17 would reduce it
by approximately 155 lines (to ~716 lines), which still exceeds the 400-line threshold.
Evidence: `wc -l CLAUDE.md` → 871 lines.
Impact: Large CLAUDE.md files reduce rule adherence and inflate context size on every session.
Fix: After applying RA17, continue extracting file-type-specific rules (Go testing guidelines, template sections) into
path-scoped `.claude/rules/` files, or split the architecture and development command documentation into a subdirectory
`CLAUDE.md`.

## Fixed

### Pass 2 — 2026-05-05

#### [RA17] Project CLAUDE.md retains 6 behavioral mandate sections now covered by `.claude/rules/`

Fixed: 2026-05-05
Notes: Removed the entire "Code Quality Anti-Patterns - DO NOT REPEAT" block (no-high-complexity, no-constant-duplication,
no-inline-yaml-in-tests, no-commit-bylines, commit-message-limits, and prevention mechanisms sections). Removed the
"NEVER overwrite /README.md" mandate line from the README Protection section; kept the section header and bash examples
as development context.

#### [RA18] `~/.claude/CLAUDE.md` retains all mandates now covered by `~/.claude/rules/`

Fixed: 2026-05-05
Notes: Removed the entire "Code Quality" section (7 mandates). Removed 7 behavioral mandate lines from "General
Guidelines" (git commit, full paths, which, no-verify, test errors, rg/fd/date, linting rights). Retained "Command
Shortcuts" section and the "Hmm..." workflow instruction, which are CONTEXT not mandates.

### Pass 1 — 2026-05-05

#### [RA01] Project `.claude/rules/` absent; 6+ atomic behavioral rules in project CLAUDE.md

Fixed: 2026-05-05
Notes: Created `.claude/rules/` directory in the project root and migrated all six behavioral mandates to focused rule
files (RA07–RA12). The Critical finding is resolved now that `.claude/rules/` exists with dedicated rule files.

#### [RA02] Project CLAUDE.md exceeds 400-line threshold (871 lines)

Fixed: 2026-05-05 (partial)
Notes: Behavioral rules migrated out of CLAUDE.md sections into `.claude/rules/` files. Full reduction of CLAUDE.md line
count requires further refactoring of architectural documentation into subdirectory files; left as ongoing work. Severity
reduced to Advisory after rule migration. See RA19 for current status.

#### [RA03] `~/.claude/CLAUDE.md` contains 13 atomic behavioral rules with no rule files

Fixed: 2026-05-05
Notes: Created four rule files in `~/.claude/rules/` covering all 13 mandates:

- `git-conduct.md` — Never commit on own, never use --no-verify
- `tool-preferences.md` — rg/fd preferences, full paths, which for tool location, date before writing dates
- `code-quality-gates.md` — EditorConfig, linting, autofixers, no linting config simplification
- `test-quality.md` — Always fix test errors and warnings

#### [RA04] `~/CLAUDE.md` context-mode routing rules as behavioral mandates — Advisory

Fixed: 2026-05-05
Notes: Determined to be plugin-managed content. No action taken. Finding remains Advisory-only; not a defect.

#### [RA05] `communication-style.md` — no defects

Fixed: 2026-05-05
Notes: File validated as correct. No action needed.

#### [RA06] Nitpicker Advisory findings N10, N11 — no rule candidates

Fixed: 2026-05-05
Notes: Both are Advisory severity; no rule candidate extraction triggered per extraction rules.

#### [RA07] Created `no-commit-bylines.md`

Fixed: 2026-05-05
Notes: Created `.claude/rules/no-commit-bylines.md`. Never add Co-Authored-By or attribution lines to commits.

#### [RA08] Created `readme-protection.md`

Fixed: 2026-05-05
Notes: Created `.claude/rules/readme-protection.md`. Never overwrite `/README.md`; use `/tmp/` or `testdata/` for test output.

#### [RA09] Created `no-inline-yaml-in-tests.md`

Fixed: 2026-05-05
Notes: Created `.claude/rules/no-inline-yaml-in-tests.md`. Always use `testdata/yaml-fixtures/` and `testutil.MustReadFixture()`.

#### [RA10] Created `no-constant-duplication.md`

Fixed: 2026-05-05
Notes: Created `.claude/rules/no-constant-duplication.md`. Use `appconstants/` and `testutil/test_constants.go`; never repeat literals > 2 uses.

#### [RA11] Created `no-high-complexity.md`

Fixed: 2026-05-05
Notes: Created `.claude/rules/no-high-complexity.md`. Cyclomatic complexity cap at 15; extract helpers.

#### [RA12] Created `commit-message-limits.md`

Fixed: 2026-05-05
Notes: Created `.claude/rules/commit-message-limits.md`. Subject lines under 100 characters.

#### [RA13] Created `~/.claude/rules/git-conduct.md`

Fixed: 2026-05-05
Notes: Created user-level rule file covering git commit and --no-verify mandates.

#### [RA14] Created `~/.claude/rules/tool-preferences.md`

Fixed: 2026-05-05
Notes: Created user-level rule file covering rg, fd, full paths, which, and date-check mandates.

#### [RA15] Created `~/.claude/rules/code-quality-gates.md`

Fixed: 2026-05-05
Notes: Created user-level rule file covering EditorConfig, linting, autofixers, and no-simplification mandates.

#### [RA16] Created `~/.claude/rules/test-quality.md`

Fixed: 2026-05-05
Notes: Created user-level rule file covering test error and warning fix mandate.

## Invalid

(none)
