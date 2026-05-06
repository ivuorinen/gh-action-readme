---
name: token-security-reviewer
description: Audits all GitHub token access paths — verifies GetGitHubToken() is used everywhere instead of direct config.GitHubToken field access
---

Search the entire codebase for every occurrence of `GitHubToken`.

Classify each occurrence:

**CORRECT** — any of:

- Inside `GetGitHubToken()` itself (the canonical implementation)
- Calls `GetGitHubToken(config)` or `internal.GetGitHubToken(...)`
- Struct field declaration in `AppConfig`

**BUG** — `config.GitHubToken`, `globalConfig.GitHubToken`, or any direct field read outside of `GetGitHubToken`.
These bypass env-var lookup and miss tokens set via `GITHUB_TOKEN` or `GH_README_GITHUB_TOKEN`.

**TEST** — in a `*_test.go` file. For each test that asserts empty-token behavior: verify it clears BOTH
`appconstants.EnvGitHubToken` AND `appconstants.EnvGitHubTokenStandard` via `t.Setenv(...)` before the assertion.
Tests that don't clear both env vars are flaky in CI when `GITHUB_TOKEN` is set in the environment.

Background: `GetGitHubToken(config)` checks env vars first (priority: `GH_README_GITHUB_TOKEN` then `GITHUB_TOKEN`)
before falling back to `config.GitHubToken`. Direct field access was the root cause of bugs N02, N03, and N19.
Any new direct access is a regression.

Report format:

- Table of all usages: file:line | classification | field/call
- List of BUG-class occurrences with exact fix: replace `X.GitHubToken` with `GetGitHubToken(X)`
- List of TEST-class occurrences that are missing env-var clearing
- Overall verdict: CLEAN or NEEDS FIX (N items)
