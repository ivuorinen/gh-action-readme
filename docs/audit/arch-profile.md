# Architecture Profile

Generated: 2026-06-16

## Detected Patterns

### Pipe and Filter — High confidence

Evidence:

- Core generation pipeline: `ParseActionYML()` → `BuildTemplateData()` → `RenderReadme()` — discrete,
composable stages in `internal/parser.go`, `internal/template.go`
- `Generator.GenerateFromFile()` orchestrates the pipe; each stage returns a typed value consumed by the next
- Dependency analysis forms a parallel sub-pipe: `parse → analyze → generate pinned update`
- Multiple output format filters (md, html, json, asciidoc) at the render stage

### Layered / N-Tier — High confidence

Evidence:

- **CLI layer**: `main.go`, `cmd_gen.go`, `cmd_deps.go`, `cmd_config.go`, `cmd_cache.go`, `cmd_validate.go` —
Cobra commands, coordination only
- **Core/domain layer**: `internal/` — generator, parser, template, config, validation, output, wizard
- **Infrastructure layer**: `internal/cache/`, `internal/git/`, `internal/dependencies/` — external system access
- **Constants/shared layer**: `appconstants/` — shared across all layers, imports nothing internal
- Dependency direction (confirmed via `go list`): CLI → `internal` → subpackages; never reversed
- `appconstants` is a true leaf: imports no other project package
- `main` (root CLI) imports `appconstants`, `internal`, `internal/apperrors`, `internal/cache`,
`internal/dependencies`, `internal/helpers`, `internal/wizard`

### Interface Segregation (ISP) — High confidence

Evidence:

- `internal/interfaces.go` defines focused interfaces: `MessageLogger`, `ErrorReporter`, `ErrorFormatter`,
`ProgressReporter`, `QuietChecker`, `ProgressManager`, `OutputWriter`
- `CompleteOutput` composite interface explicitly documented as backward-compat escape hatch
- `DependencyCache` interface in `internal/dependencies/analyzer.go`
- `InputReader` interface in `cmd_deps.go` for stdin abstraction
- `executableTemplate` interface in `internal/template.go`
- Wizard constructors accept narrowest interface (`internal.MessageLogger`), not
concrete `*ColoredOutput` — confirmed in `detector.go`, `exporter.go`, `validator.go`, `wizard.go`

### Dependency Injection — High confidence

Evidence:

- `NewGenerator(config)` vs `NewGeneratorWithDependencies(config, output, progress)` — two-constructor pattern
- `Generator` struct fields are interfaces (`CompleteOutput`, `ProgressManager`), not concrete types
- `NewAnalyzer(client, repoInfo, cache DependencyCache)` — cache injected as interface
- `NewCacheAdapter(c *cache.Cache)` and `NewNoOpCache()` — adapter + null-object for testing
- `NullOutput` and `NullProgressManager` null objects for test isolation
- Unit test auto-detection in `NewGenerator()`: injects `NullOutput` in test binaries automatically
- `setupDepsUpgrade(config *AppConfig)` and `applyUpdates(..., InputReader)` — accept dependencies for injection

### Adapter Pattern — Medium confidence

Evidence:

- `internal/dependencies/cache_adapter.go`: `CacheAdapter` wraps `internal/cache.Cache` to satisfy
`DependencyCache` interface
- `NoOpCache` implements same interface with no-ops — classic null-object adapter
- `wrapHandlerWithErrorHandling()` in `main.go` — adapts error-returning handlers to Cobra's `func(cmd, args)` signature

### Command Pattern — High confidence

Evidence:

- Most Cobra command handlers return `error` (non-standard; Cobra uses `func(cmd, args)`)
- `wrapHandlerWithErrorHandling()` bridges the two signatures
- Each command encapsulates its operation: `genHandler`, `depsListHandler`, `depsSecurityHandler`,
`depsOutdatedHandler`, `depsUpgradeHandler`, `configRootHandler`, `cacheClearHandler`, etc.
- Handlers are pure functions, decoupled from the command struct definition

### Repository Pattern — Low confidence

Evidence:

- `internal/cache/cache.go` provides generic `Cache` with `Get`/`Set`/persistence — partial repository shape
- No explicit Repository type names; pattern is present but not named

## Detected Combination

### Custom hybrid: Layered CLI tool + Pipe-and-Filter core + ISP-driven Dependency Injection

This is a CLI application, not a web service or domain-rich system. The combination differs from classic
enterprise patterns:

- The "domain" is documentation generation; the pipeline IS the domain logic
- Layers are CLI / Core pipeline / Infrastructure, not Presentation / Business / Data
- DI is used for testability and output abstraction, not for swapping business rules

## Inferred Structural Rules

1. **CLI layer must not contain domain logic** — handlers coordinate only; parsing, rendering, and analysis
belong in `internal/`
2. **Pipeline stages must be pure and composable** — each stage takes typed input, returns typed output, no
side effects except at the terminal stage (file write)
3. **Interfaces must be segregated** — no parameter should accept `CompleteOutput` when `MessageLogger`
suffices; use narrowest interface
4. **Constructors must follow the two-constructor pattern** — `New*(config)` for production,
`New*WithDependencies(...)` for injection; never add optional fields directly to structs
5. **Infrastructure packages must not import internal core** — `internal/cache`, `internal/git`,
`internal/dependencies` must not import `internal` directly (only `appconstants` and, for `dependencies`, the
sibling infra packages `cache`/`git`)
6. **Command handlers should return `error`** — never call `os.Exit()` in handlers;
`wrapHandlerWithErrorHandling` owns exit logic. Trivial inline commands (`version`, `about`) are exempt
7. **Null objects required for all output interfaces** — `NullOutput`, `NullProgressManager` must cover all
interface methods; tests must not depend on real I/O
8. **Adapters live in the consumer package** — `CacheAdapter` lives in `dependencies/`, not in `cache/`;
adapters are the consumer's responsibility
9. **Production code must not import `testutil`** — `testutil` is test-only; only `_test.go` files (or
test-tagged compilation) may import it. Confirmed: zero production files import it
10. **`GetGitHubToken()` is the canonical token accessor** — read the token via
`internal.GetGitHubToken(config)`, not by reading `config.GitHubToken` directly, except inside the
loader/merge/redaction sites that legitimately set or scrub the field

## Package Responsibilities & Boundaries

- `appconstants/` — leaf package of constants; imported by all, imports none. Boundary: must never import any
`internal*` package.
- `internal/` (flat core) — parsing, templating, generation, config, output, interfaces, wizard orchestration
entrypoints. May import its subpackages and `appconstants`.
- `internal/apperrors/` — contextual error formatting. Imports only `appconstants`.
- `internal/cache/` — generic on-disk cache. Imports only `appconstants`. Must not import `internal`.
- `internal/git/` — git repo detection. Imports only `appconstants`. Must not import `internal`.
- `internal/dependencies/` — action dependency analysis. Imports `appconstants`, `internal/cache`,
`internal/git`. Must not import `internal`.
- `internal/validation/` — action.yml validation. Imports `appconstants`, `internal/git`. Must not import `internal`.
- `internal/helpers/` — small helpers. Imports `internal/git` only. Must not import `internal`.
- `internal/wizard/` — interactive config wizard. Imports `appconstants`, `internal`, `internal/git`,
`internal/helpers`. NOTE: it imports parent `internal` for `AppConfig` and `MessageLogger` —
a documented, accepted coupling (see Ambiguities).
- `templates_embed/` — embedded template FS. Imports only `appconstants`.
- `testutil/` — test-only helpers. Imports only `appconstants`. Must be imported solely from test files.

## Ambiguities & Contradictions

- `internal/` is a flat package for most core types (`parser.go`, `template.go`, `generator.go`, `config.go`,
`output.go`) — this grows coupling as the package expands; no sub-domain separation exists within core
- `cmd_deps.go` defines `InputReader` interface and `StdinReader` struct — these arguably belong in
`internal/` by the layering rule, not in the CLI layer
- `Generator` struct exposes `Config *AppConfig` as a public field rather than through an interface — softens
the DI rule for config itself
- `internal/analyzer_helpers.go` (core) imports `internal/dependencies` (sub-infrastructure) — acceptable
because it is one-directional (dependencies does not import internal), but worth watching
- `internal/wizard` imports parent `internal` because `AppConfig` lives in `internal`. The prior audit (AA03)
accepted this as a partial fix: the coupling surface was minimised to interfaces; a full fix needs `AppConfig`
extracted to a shared `internal/types` package. This remains the intended state, not a regression.

## Drift

Compared to the 2026-05-05 profile:

- Import graph unchanged in direction; no new cyclic or reversed dependencies. `go build ./...` clean; no import cycles.
- No production file imports `testutil` (rule 9 holds — the AA01/AA05 fixes persist).
- `internal/wizard` still imports `internal` (rule documented as accepted coupling, not drift).
- Added explicit rules 9 (no production `testutil`) and 10 (`GetGitHubToken` canonical accessor) and a Package
Responsibilities section to make boundaries auditable.
