# Nitpicker Findings

Generated: 2026-05-04
Last validated: 2026-06-17 (pass 20 — full multi-skill audit)

## Summary

- Total: 117 | Open: 2 | Fixed: 111 | Invalid: 4
- Open breakdown: 2 Advisory only (N111/N112 validation-regex) — both cosmetic (affect validation messaging,
  not generation), with no clean fix (short-SHA ambiguity / would break the semver mutation-test suite). All
  Critical/High/Medium findings are fixed.

## Open Findings

### Advisory

#### [N111] IsCommitSHA false-positives on short all-hex tags/refs

Category: correctness
Area: internal/validation/validation.go (reCommitSHA = ^[a-f0-9]{7,40}$)
Problem: Any 7–40-char lowercase-hex string is classified as a commit SHA, so a tag like `deadbee` or a 7-char
hex branch is mis-detected. The 40-char guard in IsVersionPinned mitigates the pinning path.
Decision: No clean fix — 7-char short SHAs are inherently ambiguous with hex tags. Left as-is (cosmetic;
affects detection messaging only, not generation correctness). Surfaced by the parsing lens.

#### [N112] IsSemanticVersion accepts malformed prerelease/build strings

Category: correctness
Area: internal/validation/validation.go (reSemanticVersion)
Problem: The prerelease group allows empty/dotted-only identifiers, so `v1.2.3-..` and leading-zero `01.02.03`
validate as semver though SemVer 2.0.0 forbids both.
Decision: Cosmetic — only affects validation-quality messaging, not generation. Tightening the regex risks
rejecting valid edge cases and would break the semver mutation-test suite; left as-is. Surfaced by the parsing
lens.

## Fixed

### Pass 20 — 2026-06-17

#### [N116b] Field-boundary test asserted only `err != nil`, not which field was missing

Fixed: 2026-06-17
Notes: TestParseAndValidateAction_FieldBoundary now carries an errContains column asserting the missing-field
error names the specific field (name vs description), so a regression in the required-field boundary cannot pass
by erroring for a different reason. With the traversal case (pass 17) this closes the meaningful part of N116;
the remaining pure invalid-YAML presence checks are acceptable as-is.

### Pass 19 — 2026-06-17

#### [N106] cache.Config.MaxSize was dead — the cache was unbounded

Fixed: 2026-06-17
Notes: Decided to make MaxSize real rather than delete it (Entry.Size feeds the `cache stats` output). Added a
maxSize field to Cache (from Config.MaxSize) and evictToMaxSize(), called from cleanup() after expiry removal:
when total entry size exceeds maxSize it evicts entries oldest-expiry-first until within bound (0 = unbounded).
Added TestCacheEvictsToMaxSize.

### Pass 18 — 2026-06-17

#### [N105] Production caches are never Closed — leaked goroutine and unflushed final saves

Fixed: 2026-06-17
Notes: Added Close() to the DependencyCache interface (CacheAdapter delegates to cache.Close; NoOpCache no-ops)
and Analyzer.Close() (nil-safe). The template gen path (analyzeDependencies) owns its cache and defers
analyzer.Close(); the deps command paths defer closeAnalyzer(analyzer) (nil-safe helper) at all three
createAnalyzer sites and setupDepsUpgrade. Stops the cleanup goroutine and runs the final synchronous flush so
the on-disk cache (N98) actually persists. Added TestAnalyzerClose; verified `go test -race` and e2e gen.

### Pass 17 — 2026-06-17

#### [N116] Security-relevant traversal test asserted only `err != nil`

Fixed: 2026-06-17
Notes: TestValidateFilePath (dependencies/parser_test.go) checked only that rejected paths errored, so a
regression changing WHY a path is rejected (e.g. a stat short-circuit before the traversal check) would slip
past. Added an assertion that error cases' messages contain "traversal", pinning the rejection reason. The
remaining non-security error-presence tables are tracked as N116-remainder (advisory).

### Pass 16 — 2026-06-17

#### [N113] Inline YAML in tests violated the no-inline-yaml-in-tests rule (18 sites)

Fixed: 2026-06-17
Notes: 18 test sites embedded action.yml YAML as Go string literals (main_test.go, generator_test.go,
wizard/detector_test.go, dependencies/updater_test.go, testutil/{helpers,test_suites}_test.go). Created 9 shared
fixtures under testdata/yaml-fixtures/actions/ (composite stubs, field-missing variants, per-action dependency
stubs, a malformed-runs invalid fixture), registered path constants in testutil/test_constants.go, and replaced
every inline literal with testutil.MustReadFixture. Verified 0 remaining inline composite-stub literals.

#### [N114] Duplicate string constants across test files violated no-constant-duplication (8 literals)

Fixed: 2026-06-17
Notes: Replaced repeated literals with existing constants — ".git"→appconstants.DirGit (dir-name uses only;
URL-suffix concatenations left intact), "v1.2.3"→testutil.TestVersionSemantic,
"action.yaml"→appconstants.ActionFileNameYAML, "action.yml"→appconstants.ActionFileNameYML,
"actions/javascript/simple.yml"→testutil.TestFixtureJavaScriptSimple — and added 3 new shared constants
(TestNonexistentYML, TestFixtureCompositeActionAnalyzer, TestTokenGHPExisting) for literals with no existing
constant. Verified each constant's value matched before substituting.

#### [N115] AssertEqual had no negative-path test — a no-op assertion would pass the whole suite

Fixed: 2026-06-17
Notes: testutil.AssertEqual takes *testing.T (unmockable), so its failure path was untested — a regression that
gutted it would silently pass every assertion. Extracted the comparison into a pure equalCheck(expected,
actual) (ok, msg) and added TestEqualCheck covering both equal (ok) and unequal (not-ok + non-empty message)
cases, including the map length/value/type branches. Surfaced by the test-quality lens.

### Pass 15 — 2026-06-17

#### [N107] shieldsBadgeEncode omitted URL percent-encoding — `/ ? # %` corrupt the JSON badge URL

Fixed: 2026-06-17
Notes: shieldsBadgeEncode (json_writer.go) did only shields separator escaping, so a legitimate name/color like
`C/C++ build` produced a URL with a stray `/` that broke the shields endpoint. Added url.PathEscape after the
separator escaping (leaves the produced `-`/`_` untouched). Added percent-encode test cases.

#### [N108] Markdown/AsciiDoc table cells were not escaped — a `|` or newline corrupts the table

Fixed: 2026-06-17
Notes: github/professional/asciidoc themes interpolated action description/default/name/etc directly into table
cells via text/template (no escaping), so an honest description containing `|` (or a crafted `[x](url)`) broke
the table. Added an mdCell template func (escapes `|` → `\|`, collapses CR/LF to `<br>`) and applied it to the
input/output/dependency cells across the github, professional, and asciidoc themes.

#### [N109] Badge URLs interpolated raw `.Name`/`.Branding` — broken/spoofable badges

Fixed: 2026-06-17
Notes: github/asciidoc badge URLs used `.Name | replace " " "%20"` (only spaces) and raw branding icon/color;
the professional theme interpolated branding raw into an `<img src>`. A name like `C/C++ build` or a crafted
value produced malformed URLs or a markdown/attribute breakout. Exposed shieldsBadgeEncode as a `badgeSegment`
template func and applied it to all badge segments in the github, asciidoc, and professional themes. Verified
end-to-end render.

#### [N110] isValidOrgRepo accepted org/repo containing `/` or `@` — uses: path/version injection

Fixed: 2026-06-17
Notes: isValidOrgRepo (template.go) only checked non-empty/non-placeholder, so an org like `evil/x` or repo
`r@v9` (from config or git detection) injected extra path segments / a bogus version into the generated `uses:`
statement. Added isValidGitHubPathSegment (^[A-Za-z0-9._-]+$) — rejects `/`, `@`, whitespace, control chars
while keeping dotted repos valid. Added TestIsValidOrgRepo.

### Pass 14 — 2026-06-17

#### [N98] On-disk cache is silently useless — type assertion never matches after JSON reload

Fixed: 2026-06-17
Notes: cacheVersion stored map[string]string and getCachedVersion asserted the same, but cache.json reload
decodes Entry.Value (any) as map[string]interface{}, so every disk-loaded entry failed the assertion — the
on-disk cache provided zero benefit across CLI invocations. Store/read the version entry as map[string]any
(coercing values to string); cache only the repository Description string instead of the *github.Repository SDK
struct (a string survives the JSON round-trip). Surfaced by the completeness-critic lens.

#### [N99] Dependency tests shared the real per-user cache directory (isolation defect exposed by N98)

Fixed: 2026-06-17
Notes: ~17 dependency tests built cache.NewCache(cache.DefaultConfig()) against the shared XDG cache dir; once
N98 made the disk cache work, entries leaked between tests. Added cache.Config.Path (overrides XDG; empty in
production), and isolatedCacheConfig/newIsolatedCache test helpers (per-test t.TempDir + t.Cleanup(Close)).
Converted all dependency-test cache constructions to the isolated helpers.

#### [N100] ErrorWithSuggestions printed errors to stdout when color was enabled

Fixed: 2026-06-17
Notes: internal/output.go ErrorWithSuggestions used color.Red(...) (writes to color.Output = stdout) in the
color branch while the NoColor branch used stderr — so contextual errors landed on stdout on a normal
terminal. Now both branches write to os.Stderr via color.New(color.FgRed).Fprintf, matching ColoredOutput.Error.

#### [N101] Quoted `uses:` references were silently never pinned

Fixed: 2026-06-17
Notes: applyUpdatesToLines matched the literal `uses: actions/checkout@v4`, so a quoted
`uses: "actions/checkout@v4"` line never matched and the dependency was silently left unpinned while reported
as success. Rewrote matching as a structured per-line parse (applyUpdateToLine) that strips the list marker,
surrounding single/double quotes, trailing inline comment, and whitespace before comparing the reference value
exactly. All existing updater fixtures still pass.

#### [N102] mergeBooleanFields zero-value bug for analyze_dependencies / show_security_info

Fixed: 2026-06-17
Notes: N79 fixed the "explicit false can't override a lower-priority true" defect only for use_default_branch.
The same defect remained for analyze_dependencies and show_security_info (genuine per-config knobs). Added
analyzeDependenciesSet/showSecurityInfoSet presence flags (set via viper.IsSet in both loaders and per
repo-override via the generalized markRepoOverridePresenceFlags), and merged these on presence. Verbose/Quiet
remain intentionally OR-merged. Added fixture + TestLoadConfigFromViperFeatureFlagPresence and updated
TestMergeBooleanFields for the corrected divergent semantics.

#### [N103] determineUpdateType misclassified prerelease/build-metadata versions

Fixed: 2026-06-17
Notes: parseVersionParts string-split on "." so "1.0.0-beta" → ["1","0","0-beta"], making v1.0.0 → v1.0.0-beta
classify as a patch update (offering a prerelease as an upgrade); "+build" metadata likewise. Added
coreVersion() to strip the -prerelease/+build suffix before splitting, so same-core versions compare equal.

#### [N104] Permission-comment parser aborted the whole header scan on the first dedent

Fixed: 2026-06-17
Notes: processPermissionEntry returned "break" on a dedented line and the caller broke the entire scanner loop,
so a second `# permissions:` block (or any later comment) after a dedented prose line was silently dropped. The
caller now ends only the current block (inPermissionsBlock=false) and keeps scanning; a new permissions header
is still caught earlier. Added permissions/two-blocks-with-prose.yml fixture + test.

### Pass 13 — 2026-06-16

#### [N93] `git.parseGitHubURL` truncates repository names that contain a dot (N91 fixed only the sibling parser)

Fixed: 2026-06-16
Notes: internal/git/detector.go used `reGitHubURLNoSuffix = github\.com[:/]([^/]+)/([^/\.]+)`, truncating
dotted repo names at the first dot for any remote URL lacking a `.git` suffix (e.g. the HTTPS web-clone URL
`https://github.com/org/my.repo` yielded repo `my`). N91 fixed the same class of bug in
internal/validation/strings.go's `ParseGitHubURL` but never touched this second, independent parser. Replaced
both regexes with one dot-tolerant `reGitHubURL = github\.com[:/]([^/]+)/([^/]+?)(?:\.git)?(?:/.*)?$` mirroring
`validation.reGitHubURLFull`, simplified `parseGitHubURL` to a single match, and added two dotted-repo cases to
TestParseGitHubURL. Verified all prior edge cases (subgroups, trailing slash, query params, SSH) still pass.

#### [N94] `apperrors.Wrap` aliases the original's `Suggestions` slice despite a "copy to avoid mutating" comment

Fixed: 2026-06-16
Notes: Wrap (internal/apperrors/errors.go) cloned `Details` via maps.Clone but copied `Suggestions` by slice
reference, so a later `WithSuggestions` append on the copy could write into the original's backing array.
Changed to `slices.Clone(ce.Suggestions)` (slices already imported). Latent (no current
Wrap(...).WithSuggestions(...) chain), but the copy contract now holds.

#### [N95] `use_default_branch:false` in a `repo_overrides` entry was silently ignored (incomplete N79 fix)

Fixed: 2026-06-16
Notes: N79 made `use_default_branch:false` overridable only for top-level configs via `useDefaultBranchSet`
(set from `viper.IsSet`). Repo-override structs are produced by viper's nested unmarshal, so their flag was
always false and a per-repo `use_default_branch:false` could never disable the default-true behavior. Added
`markRepoOverrideUseDefaultBranch` (internal/viper_helper.go), called from both loaders, which inspects the raw
viper `repo_overrides` map (robust against repo names containing dots/slashes) and sets the flag per override
when the key is explicitly present. Added `ConfigKeyRepoOverrides` constant, fixture
configs/repo-override-use-default-branch-false.yml, and end-to-end test
TestLoadConfigFromViperRepoOverrideUseDefaultBranch.

#### [N96] `StdinReader` doc comment overstated its buffer-reuse guarantee

Fixed: 2026-06-16
Notes: internal/input.go comment said the buffer is "preserved across calls" without scoping it to a single
instance; a caller constructing a fresh StdinReader per read can drop a buffered-but-unconsumed stdin tail.
Reworded to state the guarantee is per-instance and that multi-read callers must reuse one StdinReader.

### Pass 12 — 2026-06-16

#### [N77] configuration.md documents Performance Settings keys (cache_ttl, concurrent_requests, timeout) that AppConfig never defines

Fixed: 2026-06-16
Notes: Removed the phantom Performance Settings section (cache_ttl/concurrent_requests/timeout) and the
GH_ACTION_README_CACHE_TTL/_TIMEOUT env exports from docs/configuration.md — none exist in AppConfig. (See
doc-findings DA16.)

#### [N78] development.md documents a GenerateReadme() function that does not exist

Fixed: 2026-06-16
Notes: docs/development.md replaced the fictional GenerateReadme() with the real entry points
(internal.RenderReadme and Generator.GenerateFromFile/ProcessBatch). (doc-findings DA17.)

#### [N79] use_default_branch cannot be disabled via config — merge only propagates the true value

Fixed: 2026-06-16
Notes: AppConfig gained an unexported useDefaultBranchSet flag, set from viper.IsSet in both config loaders
(loadConfigFromViper, loadAndUnmarshalConfig). mergeBooleanFields now merges UseDefaultBranch on explicit
presence, so use_default_branch:false overrides the default-true. Added
TestMergeBooleanFieldsUseDefaultBranchPresence; verified end-to-end via an action config.yaml.

#### [N80] Production and tests exercise two divergent LoadConfiguration implementations

Fixed: 2026-06-16
Notes: Deleted the test-only package-level LoadConfiguration and loadAndMergeConfig; repointed config_test.go
to NewConfigurationLoader().LoadConfiguration (the production path), so tests now exercise the shipped
implementation.

#### [N81] promptSensitive echoes GitHub tokens to terminal despite "without echoing" contract

Fixed: 2026-06-16
Notes: promptSensitive now disables terminal echo via golang.org/x/term ReadPassword on a TTY (falls back to
the scanner for non-TTY/piped input). x/term promoted to a direct dependency.

#### [N83] deps security exits with error code 1 when no action files found, inconsistent with deps list/outdated

Fixed: 2026-06-16
Notes: depsSecurityHandler now uses handleNoFilesFoundError + a len==0 guard (returns nil on no files),
consistent with deps list/outdated; updated the integration test that asserted the old error.

#### [N84] deps outdated prints contradictory 'All dependencies are up to date!' after warning that no action files exist

Fixed: 2026-06-16
Notes: depsOutdatedHandler adds a len==0 guard after handleNoFilesFoundError, so it no longer prints 'All
dependencies are up to date!' when no action files exist.

#### [N85] Interactive upgrade prompt errors out (exit 1) on EOF/empty input instead of treating it as the default 'N'

Fixed: 2026-06-16
Notes: applyUpdates treats io.EOF (empty/closed stdin) as the default 'N' and cancels with nil instead of
returning an error.

#### [N86] Unused viper field on ConfigurationLoader (dead state)

Fixed: 2026-06-16
Notes: Removed the unused viper field and its initializers from ConfigurationLoader and dropped the now-unused
spf13/viper import.

#### [N87] github_helper doc comment names a wrong/non-existent env var for the GitHub token

Fixed: 2026-06-16
Notes: Corrected the github_helper.go doc comment GHREADME_GITHUB_TOKEN -> GH_README_GITHUB_TOKEN (matches
appconstants.EnvGitHubToken).

#### [N88] Error() iterates Details map in nondeterministic order, producing unstable error output

Fixed: 2026-06-16
Notes: ContextualError.Error() iterates Details via slices.Sorted(maps.Keys(...)) for deterministic, stable
output.

#### [N89] ValidateActionYMLPath swallows non-NotExist stat errors, treating unreadable paths as valid

Fixed: 2026-06-16
Notes: ValidateActionYMLPath now returns any os.Stat error wrapped with %w (not only IsNotExist), so
permission-denied/unreadable paths are surfaced instead of treated as valid.

#### [N90] depsGraphHandler bypasses the error-returning handler convention and the error-handling wrapper

Fixed: 2026-06-16
Notes: depsGraphHandler (and schemaHandler, arch AA11) now return error and are registered via
wrapHandlerWithErrorHandling, matching the project-wide handler convention.

#### [N91] ParseGitHubURL truncates repository names that contain a dot

Fixed: 2026-06-16
Notes: ParseGitHubURL regexes use ([^/]+?) with explicit optional .git stripping so dotted repo names (e.g.
my.repo) are preserved; updated the mutation test that encoded the old [^/.] truncation.

#### [N92] GitHub token input silently truncated for tokens larger than 64KB line (bufio.Scanner default limit)

Fixed: 2026-06-16
Notes: Raised the wizard's bufio.Scanner buffer to 1MB (the 64KB default could truncate long input); the new
term ReadPassword path for sensitive input also bypasses the scanner entirely.

### Pass 11 — 2026-06-16

#### [N48] Docs advertise yaml/toml gen output formats that are rejected at runtime

Fixed: 2026-06-16
Notes: Removed yaml/toml from the `gen` output-format docs (README.md, docs/usage.md, docs/api.md, CLAUDE.md)
and corrected format counts to 4 (md/html/json/asciidoc). CLI help (cmd_gen.go:38) already listed only the 4
working formats. `config export` yaml/toml docs left intact (that path genuinely supports them).

#### [N49] TOML export swallows all write errors, defeating the atomic-write integrity guarantee

Fixed: 2026-06-16
Notes: Introduced a `stickyWriter` (records first write error) in internal/wizard/exporter.go. writeTOMLConfig
and all section writers now take `*stickyWriter`; the exportTOML callback returns `sw.err`, so a mid-write
failure aborts before os.Rename instead of committing a truncated file.

#### [N50] YAML/JSON header WriteString errors ignored before atomic rename

Fixed: 2026-06-16
Notes: exportYAML now writes the headers and runs the YAML encoder through the same `stickyWriter` and returns
`sw.err` after Encode (goccy/go-yaml discards the underlying write error), surfacing header/body write
failures.

#### [N51] Token redaction misses tokens nested in RepoOverrides

Fixed: 2026-06-16
Notes: configRootHandler deep-copies RepoOverrides and redacts each nested GitHubToken (added
appconstants.RedactedPlaceholder). Added TestConfigRootHandlerRedactsNestedTokens capturing stdout to assert
neither the top-level nor nested token leaks.

#### [N52] security.md falsely claims HTML output is not escaped; html/template auto-escapes all action fields

Fixed: 2026-06-16
Notes: docs/security.md rewritten: HTML output IS auto-escaped by html/template; documented the real (minor)
nuance that operator-controlled header/footer config is written verbatim.

#### [N53] Export path-traversal guard rejects legitimate paths containing '..' as a substring

Fixed: 2026-06-16
Notes: ExportConfig traversal guard now rejects only a `..` path segment (slices.Contains over the split path)
instead of any `..` substring, allowing legitimate names such as config..bak.yaml.

#### [N54] Annotated-tag dereference fallback can emit a tag-object SHA as a commit pin

Fixed: 2026-06-16
Notes: getCommitSHAForTag returns "" when an annotated tag's GetTag dereference fails (instead of the
tag-object SHA), so GeneratePinnedUpdate's empty-SHA guard rejects the update rather than writing a broken
pin.

#### [N55] Unknown-theme error hardcodes a theme list that duplicates the canonical themeTemplates map and will drift

Fixed: 2026-06-16
Notes: Unknown-theme error builds its valid-theme list from appconstants.GetSupportedThemes() (single source)
via a shared validateTheme helper.

#### [N56] Icon badge color segment is not shields-encoded, breaking gray-dark color

Fixed: 2026-06-16
Notes: Icon badge color segment is now shields-encoded (shieldsBadgeEncode(branding.Color)) so colors
containing hyphens (e.g. gray-dark) no longer break shields.io segment parsing.

#### [N57] WaitGroup reuse race between concurrent Close() and Delete()/Set()/cleanup()

Fixed: 2026-06-16
Notes: cache.Close sets a `closed` flag under the mutex before saveWG.Wait; saveToDiskAsync skips when closed
and uses WaitGroup.Go, eliminating the WaitGroup-reuse race with concurrent Set/Delete/cleanup. Verified with
`go test -race ./internal/cache`.

#### [N58] Add unit test for new isValidGitHubOrgName function

Fixed: 2026-06-16
Notes: Added TestConfigValidatorIsValidGitHubOrgName covering valid/invalid org names (underscores,
consecutive/leading/trailing hyphens, dots, 39/40-char bounds).

#### [N59] Misleading 'unknown' fallback for cache directory path

Fixed: 2026-06-16
Notes: cmd_cache handlers use the appconstants.CachePathUnknown sentinel (not ScopeUnknown) and skip os.Stat
for the sentinel, avoiding a bogus relative-path stat.

#### [N60] Annotated-tag dereference-failure fallback branch is untested

Fixed: 2026-06-16
Notes: Added TestGetCommitSHAForTag_AnnotatedTagDerefFailure (omits the /git/tags mock so GetTag 404s)
asserting sha == "".

#### [N61] Duplicate string literal "MyAction" across test files violates project rule

Fixed: 2026-06-16
Notes: Added testutil.TestActionNameMyAction and replaced the four "MyAction" literals in
internal_template_test.go and internal_validator_test.go.

#### [N62] TestRepoSlug does not cover consecutive spaces, masking a doc/behavior mismatch

Fixed: 2026-06-16
Notes: Removed the obsolete TestRepoSlug (function deleted in N70); added a consecutive-spaces case to
validation TestSanitizeActionName documenting the actual (non-collapsing) behavior.

#### [N63] Traversal test depends on ambient cwd and asserts only error-presence

Fixed: 2026-06-16
Notes: TestExportConfigRejectsTraversal now isolates via t.Chdir(t.TempDir()), asserts the error mentions
`..`, and asserts the traversal target file was not created.

#### [N64] Misleading comment misstates detectProjectSettings error contract

Fixed: 2026-06-16
Notes: Corrected the misleading comment in wizard_test.go: detectProjectSettings only errors when the cwd
cannot be determined (actionDir then unset); being outside a git repo is not an error.

#### [N65] Remove duplicate `ContextKey*` constants that shadow existing `TestKey*` constants

Fixed: 2026-06-16
Notes: Removed the duplicate `ContextKey*` constants (shadowed existing `TestKey*`); confirmed 0 usages.

#### [N66] Remove unused `TestStr*`, `Tag*`, and `Generic*` constants (dead code; literals not actually replaced)

Fixed: 2026-06-16
Notes: Removed the unused `TestStr*`/`Tag*`/`Generic*` constants (dead code; the literals were never actually
replaced); confirmed 0 usages.

#### [N67] .gitleaks.toml comment falsely claims docs/audit/ is gitignored; files are committed

Fixed: 2026-06-16
Notes: .gitleaks.toml comment corrected: docs/audit/ is tracked in git (not gitignored); excluded from
scanning with a warning not to store real secrets. Allowlist regex unchanged.

#### [N68] Theme validation is inconsistent across output formats

Fixed: 2026-06-16
Notes: validateTheme is called up-front in generateByFormat, so an invalid --theme now fails identically for
JSON as for templated formats (verified via the built binary).

#### [N69] TOML config export emits map sections in nondeterministic order

Fixed: 2026-06-16
Notes: writeMapSection sorts keys before writing, making TOML [permissions]/[variables] export deterministic.

#### [N70] repoSlug reimplements validation.SanitizeActionName and does not strip URL/repo-invalid characters

Fixed: 2026-06-16
Notes: Deleted repoSlug; callers now use validation.SanitizeActionName (single source of truth).

#### [N71] Tab-indented permission comment parsing change has no test

Fixed: 2026-06-16
Notes: Added testdata/yaml-fixtures/permissions/tab-indented.yml (+
testutil.TestFixturePermissionsTabIndented) and a parser test asserting tab-indented permission comment blocks
parse.

#### [N72] Concurrent cache.json writers race on os.WriteFile (pre-existing, now easier to hit)

Fixed: 2026-06-16
Notes: cache saveToDisk serializes writes with a saveMutex and writes atomically via a temp file + os.Rename,
preventing a torn cache.json from concurrent async saves.

#### [N73] Deprecation suggestion still recommends node20 despite node24 now being valid

Fixed: 2026-06-16
Notes: Deprecated-runtime suggestion now recommends appconstants.NodeRuntimeNode24 (current) instead of
node20.

#### [N74] MockHTTPClient doc comment overclaims full concurrency safety

Fixed: 2026-06-16
Notes: MockHTTPClient doc comment corrected: only the Requests slice is mutex-guarded; same-key responses
share one Body reader and must not be read concurrently.

#### [N75] Exported config files now created with 0600 instead of prior 0644 (behavior change)

Fixed: 2026-06-16
Notes: Documented in writeFileAtomic that exported files are intentionally created 0600 (stricter than the
previous 0644).

#### [N76] Atomic export temp file shares destination dir; partial temp files may linger on crash

Fixed: 2026-06-16
Notes: Documented in writeFileAtomic the temp-file pattern and that a hard kill between create and rename can
leave a harmless stale .tmp (never contains secrets; cleaned on the next export).

### Pass 10 — 2026-06-12

#### [N46] Data race in `testutil.MockHTTPClient.Do` breaks `go test -race ./...`

Fixed: 2026-06-12
Notes: `MockHTTPClient.Do` appended to the shared `Requests` slice without
synchronization (`testutil/testutil.go:28`). Parallel `t.Parallel()` analyzer tests
(e.g. `TestAnalyzerGetLatestVersion`, `analyzer_test.go:325`) share one mock client
through the GitHub client's transport, so `go test -race ./internal/dependencies`
reported a data race on the slice while plain `go test` hid it. Added a `sync.Mutex`
to `MockHTTPClient` guarding the `Requests` append (`Responses` is read-only after
construction). `go test -race ./...` now passes on every package. Surfaced after
fast-forwarding to origin (the parallel analyzer tests arrived with #220).

#### [N47] `getDefaultBranch` has an always-true `len(parts) > 0` guard (dead branch)

Fixed: 2026-06-12
Notes: `internal/git/detector.go:184` guarded `return parts[len(parts)-1]` with
`if len(parts) > 0`, but `strings.Split` always returns at least one element, so the
trailing `return appconstants.GitDefaultBranch` fallback was unreachable dead code.
Removed the guard and return the last element directly with an explanatory comment
(behavior identical; verified by `go test ./internal/git`). Not caught by prior
passes N11–N45.

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

#### [N38] CLAUDE.md Go version "1.25+" contradicts `go.mod` 1.26.3

Fixed: 2026-05-05
Notes: Changed "Go 1.25+" to "Go 1.26+" in the gremlins tool note (`CLAUDE.md:303`).
The module requires go 1.26.3 per `go.mod`.

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
Notes: Changed detection to `strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0],
".test.exe")`.
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

### Pass 13 — 2026-06-16

#### [N97] `validate`/`gen` error on a no-action directory while `deps` subcommands exit 0

Notes: Not a defect — intentional asymmetry. `validate`/`gen` are explicit operations on actions, so a
directory with no action files is a usage error worth surfacing (exit 1). The `deps` subcommands are lenient
discovery and return nil on no files (see fixed N83/N84/N37). The two behaviors are correct for their
respective contracts; no code change.

### Pass 12 — 2026-06-16

#### [N82] ErrorWithSuggestions silently discards suggestions, details, and help URLs — users never see them

Notes: False positive. ContextualError.Error() (internal/apperrors/errors.go) already appends Details,
Suggestions, and HelpURL, so ErrorWithSuggestions printing err.Error() DOES show them. The proposed fix (use
FormatContextualError) would double-render, because formatMainError itself calls err.Error() which already
contains those sections. Code left unchanged.

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
