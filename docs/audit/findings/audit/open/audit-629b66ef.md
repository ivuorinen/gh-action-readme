---
id: audit-629b66ef
auditor: audit
severity: low
category: conventions
area: git:615a82e
status: open
found: 2026-07-15
---

# Removing the deps graph subcommand was typed feat, not feat!/BREAKING CHANGE

## Problem

Commit `615a82e feat(deps): ...` removes the user-invocable `deps graph` subcommand — a CLI-contract removal — under a plain `feat` type, bundling four unrelated concerns.

## Evidence

Body: "Remove the unimplemented `deps graph` stub command and its test helper." `gh-action-readme deps graph` now errors "unknown command". No `!`/`BREAKING CHANGE:` footer.

## Impact

Conventional-commit/semantic-release tooling classifies this as minor, not breaking — mis-versioning a release that removed a command.

## Fix

Use `feat!` with a `BREAKING CHANGE: removed the deps graph command` footer; split unrelated removals into their own refactor/chore commits.
