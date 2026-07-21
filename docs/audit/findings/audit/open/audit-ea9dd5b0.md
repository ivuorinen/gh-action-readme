---
id: audit-ea9dd5b0
auditor: audit
severity: low
category: docs
area: README.md:211
status: open
found: 2026-07-15
---

# README claims 80%+ overall test coverage; real gate is 72% and actual ~73.7%

## Problem

README states an unscoped "80%+ test coverage" claim contradicting the Makefile threshold and CLAUDE.md.

## Evidence

`README.md:211`: "...0 linting violations and 80%+ test coverage."
(also `:44` ">= 80% (new code)"). `Makefile:13-15`: "Current overall
coverage: 73.7%" / `COVERAGE_THRESHOLD := 72.0`; CLAUDE.md says
">= 72% (overall); 80% target".

## Impact

Documented quality claim is false for overall coverage; misleads contributors.

## Fix

Change line 211 to "72%+ overall coverage (80% target on new code)".
