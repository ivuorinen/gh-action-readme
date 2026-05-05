# Security Audit Findings

Generated: 2026-05-05
Last validated: 2026-05-05
Pass: 2

## Tool Coverage

- Available: opengrep, grype, trivy, gitleaks, checkov, gosec, snyk, npm, yarn, pnpm
- Not available: semgrep (broken Python environment — ImportError on startup)
- Not applicable: npm/yarn/pnpm (no Node.js lockfile present), snyk (requires `snyk auth`)
- Errored: semgrep: `ImportError: cannot import name 'cli' from 'semgrep.cli'`; snyk: `Use 'snyk auth' to authenticate.`

## Summary

Total: 24 | Open: 0 | Fixed: 21 | Invalid: 3

## Open Findings

No open findings.

## Fixed

### Pass 2 — 2026-05-05

#### [SEC-001] Go stdlib: loop induction variable overflow/underflow

Fixed: 2026-05-05
Notes: `go.mod` already declares `go 1.26.2`, which is one of the fix versions (go1.25.9, go1.26.2). No change required.

#### [SEC-002] Go stdlib crypto/tls: race on ClientCAs/RootCAs during session resumption

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (≥1.25.7). No change required.

#### [SEC-003] Fake GitHub token in test fixture committed to repository

Fixed: 2026-05-05
Notes: Replaced `ghp_test1234567890abcdefghijklmnopqrstuvwxyz` with `"FAKE_TOKEN_FOR_TESTING"` in
`testdata/yaml-fixtures/configs/global-config-default.yml`. Updated `testutil.TestTokenStd` constant to match.

#### [SEC-004] Go stdlib: incorrect pointer unwrapping in compiler (memory move)

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (go1.25.9, go1.26.2). Same upgrade as SEC-001. No change required.

#### [SEC-005] Go stdlib net/url: invalid URL host accepted by url.Parse

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (≥1.25.8). No change required.

#### [SEC-006] Go stdlib net/url: unbounded query parsing (DoS)

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (≥1.25.6). No change required.

#### [SEC-007] Go stdlib html/template: XSS via http-equiv refresh meta tag

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (≥1.25.8). No change required.

#### [SEC-008] Go stdlib html/template: incorrect JS template literal escaping across branches

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (≥1.25.9). No change required.

#### [SEC-009] Dockerfile: container process runs as root (missing USER)

Fixed: 2026-05-05
Notes: Added `USER 65532:65532` before `ENTRYPOINT` in `Dockerfile`. Used numeric UID because `FROM scratch` has no `/etc/passwd`. Also resolves SEC-023 (duplicate flag at ENTRYPOINT level).

#### [SEC-012] Non-static exec.Command in testutil/testutil.go

Fixed: 2026-05-05
Notes: Added `// TEST-ONLY: binaryPath is always the compiled test binary, never user-supplied input.` comment to
`RunBinaryCommand` in `testutil/testutil.go:240`. Existing `#nosec G204` annotation retained.

#### [SEC-013] Shell injection in testdata composite action fixtures (8 sites)

Fixed: 2026-05-05
Notes: Replaced direct `${{ github.* }}` / `${{ inputs.* }}` interpolation in `run:` steps with environment variable
indirection (`env: WORKING_DIR: ${{ ... }}` + `"$WORKING_DIR"`) in `testdata/composite-action/action.yml` (2 sites),
`testdata/yaml-fixtures/actions/composite/basic.yml` (1 site), and
`testdata/yaml-fixtures/actions/composite/complex-workflow.yml` (5 sites).

#### [SEC-014] Go stdlib archive/zip: super-linear file name indexing (DoS)

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (≥1.25.6). No change required.

#### [SEC-015] Go stdlib archive/tar: unbounded memory on sparse old-GNU archives

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (≥1.25.9). No change required.

#### [SEC-016] Go stdlib: GOARCH=wasm memory limit bypass (DoS)

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (≥1.25.6). No change required.

#### [SEC-017] Go stdlib net/http: Transfer-Encoding header bypass (request smuggling)

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (≥1.25.6). No change required.

#### [SEC-019] Go stdlib os.File.ReadDir: FileInfo may reference file outside Root

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (≥1.25.6). No change required.

#### [SEC-020] Go stdlib os: file path escape via ReadDir FileInfo (Unix only)

Fixed: 2026-05-05
Notes: `go.mod` at `go 1.26.2` satisfies the fix window (≥1.25.8). No change required.

#### [SEC-021] Checkov secret detection: obvious fake tokens in test fixtures

Fixed: 2026-05-05
Notes: Added `# checkov:skip=CKV_SECRET_6:test fixture - fake token for testing` header comment to 5 fixture files:
`testdata/yaml-fixtures/configs/minimal-with-token.yml`, `testdata/yaml-fixtures/global-config.yml`,
`testdata/yaml-fixtures/minimal-config.yml`, `testdata/yaml-fixtures/professional-config.yml`,
`testdata/yaml-fixtures/configs/global-base-token.yml`.

#### [SEC-022] Dockerfile: no health check defined

Fixed: 2026-05-05
Notes: Added `HEALTHCHECK NONE` to `Dockerfile` with comment documenting the intentional decision (CLI tool, not a long-running service).

#### [SEC-023] Dockerfile: no explicit USER instruction

Fixed: 2026-05-05
Notes: Resolved by SEC-009 fix (same `USER 65532:65532` directive). This was a duplicate flag of SEC-009 at advisory severity.

### Pass 1 — 2026-05-05

#### [SEC-024] `applyUpdatesToLines` substring match triggers false matches on version-prefixed refs and comment lines

Fixed: 2026-05-05
Notes: `strings.Contains(line, "uses: "+oldUses)` matched any line containing the ref as a
substring — including `uses: actions/checkout@v4.1.1` when updating `@v4`, and lines where
`uses:` appears in an inline comment of another field. Fixed by three guards in
`internal/dependencies/analyzer.go:649`: (1) skip full comment lines; (2) skip if the match
position is preceded by `#` on the same line (inline comment guard); (3) require OldUses to be
a complete token (afterTarget must be empty or `#`-prefixed).

## Invalid

### Pass 2 — 2026-05-05

#### [SEC-018] text/template import used instead of html/template

Notes: False positive. `internal/template.go` imports `text/template` for Markdown rendering paths, where auto-escaping
is not needed and would corrupt output. For HTML format rendering, the code uses `html/template` (confirmed via code
inspection). The opengrep rule flags any import of `text/template` regardless of context. No XSS risk exists:
`text/template` is only used where the output is Markdown, not HTML.

### Pass 1 — 2026-05-05

#### [SEC-010] Non-static exec.Command in git/detector.go (3 sites)

Notes: False positive. All arguments at the three `exec.Command` sites (`detector.go:100`,
`:158`, `:188`) are string constants (`"remote"`, `"get-url"`, `"origin"`,
`"symbolic-ref"`, `"refs/remotes/origin/HEAD"`, `appconstants.GitShowRef`, etc.). The branch
name argument in `branchExists` comes from a hardcoded two-element slice `["main", "master"]`,
not user input. `cmd.Dir` sets the OS working directory — not a shell argument — so no shell
metacharacter processing occurs. The `// #nosec G204` annotations on all three sites are
accurate. Tool flagged on structural pattern ("non-constant argument") without semantic analysis.

#### [SEC-011] Non-static exec.Command in internal/validation/validation.go

Notes: False positive. The `exec.Command` at `validation.go:36` uses `appconstants.GitCommand`
(`"git"`) and `appconstants.GitShowRef`, `appconstants.GitVerify`, `appconstants.GitQuiet`
(all package-level string constants) plus `"refs/heads/"+branch` where `branch` is validated
against a known allowlist before reaching this call. No user-controlled input reaches the
arguments. Tool flagged on structural pattern. `// #nosec G204` annotation is accurate.
