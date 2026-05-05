# Adversarial Review — gh-action-readme

Date: 2026-05-05
Reviewer: adversarial-reviewer skill

---

## Summary

| Severity  | Count |
|-----------|-------|
| CRITICAL  | 1     |
| HIGH      | 3     |
| MEDIUM    | 3     |
| LOW       | 2     |
| **Total** | **9** |

---

## CRITICAL

**BUG: ReadTemplate reads arbitrary files when given any absolute path**
File: `templates_embed/embed.go:62-69`
Category: Security — Path Traversal
Severity: CRITICAL

`ReadTemplate` accepts an absolute path and reads it with only `filepath.Clean` normalization — no confinement to any
allowed directory. Any caller that passes user-controlled input (e.g. `--template /etc/passwd`) receives the file contents,
which are then rendered via Go templates. Template execution can expose file contents in output or error messages.

Trigger: `gh-action-readme gen --template /etc/passwd testdata/example-action/`

Fix: Reject all absolute paths that do not reside under a known safe root (e.g. the binary's directory or an explicit allowed-templates prefix). Example guard:

```go
if filepath.IsAbs(templatePath) {
    if !strings.HasPrefix(cleanPath, allowedTemplateRoot) {
        return nil, filepath.ErrBadPattern
    }
}
```

---

## HIGH

**BUG: `regexp.MustCompile` called inside hot function `parseGitHubURL` in `internal/git/detector.go`**
File: `internal/git/detector.go:209`
Category: Resource Management / Logic Error
Severity: HIGH

`parseGitHubURL` compiles two regexps on every invocation inside the loop. This is not a data-loss bug but it panics if
the pattern is ever invalid (no recovery path) and degrades performance on any codepath that calls this function repeatedly
(batch generation, recursive walk with many repos). More critically: if `appconstants.RegexGitSHA` is mutated at runtime
(unlikely but possible via `go test -v` parallel tests that swap constants), the panic is unrecoverable.

All other regexps in the codebase (`validation/strings.go`, `dependencies/analyzer.go`) are correctly compiled as package-level `var`. This one is the outlier.

Trigger: Call `parseGitHubURL` from parallel goroutines — race on `regexp.MustCompile` internal state during compilation.

Fix: Move both patterns to package-level `var` as done everywhere else in the codebase.

---

**BUG: `shouldIgnoreDirectory` prefix logic skips `.github` and `.githooks` when only `.git` is ignored**
File: `internal/parser.go:204-221`
Category: Logic Error
Severity: HIGH

The comment says `".git" matches ".git", ".github", etc.` — it treats this as intentional, but it is wrong behavior. If a
repository stores its own GitHub Actions in `.github/actions/`, those actions are silently skipped during recursive discovery.
The user gets no warning and no output for those actions.

Trigger: Repository with `.github/actions/my-action/action.yml`. Run `gh-action-readme gen --recursive .`. The action is never processed and no error is emitted.

Fix: Use exact match for all entries including hidden directories. If `.github` itself should be ignored, add it
explicitly to the ignored list. The current fuzzy prefix match is a superset of what is intended.

```go
// Replace the HasPrefix branch with exact match:
if dirName == ignored {
    return true
}
```

---

**BUG: `walkFunc` propagates any filesystem error and aborts the entire walk**
File: `internal/parser.go:230-233`
Category: Error Handling
Severity: HIGH

```go
func (w *actionFileWalker) walkFunc(path string, info os.FileInfo, err error) error {
    if err != nil {
        return err   // <-- aborts entire walk
    }
```

On a real repository with any unreadable directory (permission-denied on a vendor or private subdir), the entire recursive
scan fails with that error. The caller wraps it as a hard error. Users running `--recursive` on a monorepo with a restricted
directory lose all results, not just results from that subtree.

The `findActionFilesRecursive` in `internal/wizard/detector.go` already does it correctly — it returns `filepath.SkipDir` on error. `parser.go`'s walker does not.

Trigger: `chmod 000 somedir && gh-action-readme gen --recursive .` — returns error, generates nothing.

Fix:

```go
if err != nil {
    return filepath.SkipDir
}
```

---

## MEDIUM

**BUG: TOCTOU race in `loadActionConfigInternal` and `findFirstExistingConfig`**
File: `internal/config.go:396-401` and `internal/config.go:380-381`
Category: State & Concurrency — TOCTOU
Severity: MEDIUM

Both functions `os.Stat` a path then pass it to `loadConfigFromViper`. Between `Stat` and `ReadConfig` the file can be
deleted or replaced. In `loadActionConfigInternal`, if the file disappears after `Stat` returns non-`IsNotExist`, Viper
returns an opaque read error with no useful context. This is a latent correctness issue in CI environments with cleanup
scripts running in parallel.

Fix: Remove the `Stat` guard; call `loadConfigFromViper` directly and check `os.IsNotExist` on its error.

---

**BUG: `getLatestTag` trusts `tags[0]` is the most recent tag — it is not guaranteed**
File: `internal/dependencies/analyzer.go:527-534`
Category: Logic Error
Severity: MEDIUM

`ListTags` returns tags in the order GitHub's API delivers them, which is reverse-creation-order for lightweight tags but
chronological for annotated tags. The code assumes `tags[0]` is always the latest — this is only true for lightweight tags.
For a repository mixing annotated and lightweight tags (common in the wild), the "latest" tag returned may be stale.

Trigger: Repository with annotated tags for releases and lightweight tags for CI snapshots. `tags[0]` will be a CI snapshot, not the latest release.

Fix: Sort by semantic version after retrieval, or prefer `getLatestRelease` (already tried first) and only fall back to tags as a last resort without claiming `[0]` is newest.

---

**BUG: `parsePermissionsFromComments` error silently discarded — corrupts merge result**
File: `internal/parser.go:48-52`
Category: Error Handling
Severity: MEDIUM

```go
commentPermissions, err := parsePermissionsFromComments(path)
if err != nil {
    // Don't fail if comment parsing fails, just log and continue
    commentPermissions = nil
}
```

The comment says "just log" but there is no log call. The error is discarded. If the file is opened successfully by
`parsePermissionsFromComments` but fails mid-scan (e.g. truncated file, I/O error during read), the caller silently merges
`nil` permissions. Users who rely on comment-format permissions see them stripped with no indication why. The file is
opened twice — once for comments, once for YAML — so partial failure of one silently changes output.

Fix: At minimum `fmt.Fprintf(os.Stderr, ...)` the error. Better: return it as a warning through the output interface.

---

## LOW

**BUG: `setupOutputAndErrorHandling` dereferences `globalConfig` before nil check**
File: `main.go:62`
Category: Logic Error — Nil Dereference
Severity: LOW

`setupOutputAndErrorHandling` calls `createOutputManager(globalConfig.Quiet)` at line 62. `globalConfig` is initialized
in `initConfig`, which is registered as a `PersistentPreRunE`. If any command bypasses that hook (e.g. a command registered
without inheriting the root's `PersistentPreRunE`, or a direct call in test without running Cobra), `globalConfig` is
`nil` at line 62 and the process panics.

The nil guard exists in `wrapHandlerWithErrorHandling` (line 77) but not in `setupOutputAndErrorHandling`. These are called in different orders depending on the command path.

Trigger: Add a Cobra subcommand without `PersistentPreRunE` inheritance, call it from tests.

Fix: Add `if globalConfig == nil { globalConfig = internal.DefaultAppConfig() }` at the top of `setupOutputAndErrorHandling`.

---

**BUG: TOML export write errors silently ignored throughout `writeBehaviorSection` and siblings**
File: `internal/wizard/exporter.go` (multiple `writeXxxSection` methods)
Category: Error Handling
Severity: LOW

All `fmt.Fprintf(file, ...)` and `file.WriteString(...)` calls in the TOML exporter are written as `_, _ = ...`. If the
underlying file write fails (disk full, quota exceeded), the function returns `nil` indicating success. The exported TOML
file is then silently truncated.

Trigger: Export config to a filesystem at quota. `exportTOML` returns `nil`, prints success message, file is partially written.

Fix: Capture write errors and return them. At minimum, accumulate the first error and return it after the last write.
