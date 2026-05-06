# Architecture Profile

Generated: 2026-05-05

## Detected Patterns

### Pipe and Filter — High confidence

Evidence:

- Core generation pipeline: `ParseActionYML()` → `BuildTemplateData()` → `RenderReadme()` — discrete, composable stages in `internal/parser.go`, `internal/template.go`
- `Generator.GenerateFromFile()` orchestrates the pipe; each stage returns a typed value consumed by the next
- Dependency analysis forms a parallel sub-pipe: `parse → analyze → generate pinned update`
- Multiple output format filters (md, html, json, asciidoc) at the render stage

### Layered / N-Tier — High confidence

Evidence:

- **CLI layer**: `main.go`, `cmd_gen.go`, `cmd_deps.go`, `cmd_config.go` — Cobra commands, no domain logic
- **Core/domain layer**: `internal/` — generator, parser, template, config, validation
- **Infrastructure layer**: `internal/cache/`, `internal/git/`, `internal/dependencies/` — external system access
- **Constants/shared layer**: `appconstants/` — shared across all layers
- Dependency direction: CLI → internal → subpackages; never reversed
- `main.go` imports only `appconstants`, `internal`, `internal/dependencies`, `cobra`

### Interface Segregation (ISP) — High confidence

Evidence:

- `internal/interfaces.go` defines 8 focused interfaces: `MessageLogger`, `ErrorReporter`, `ErrorFormatter`, `ProgressReporter`, `QuietChecker`, `ProgressManager`, `OutputWriter`, `ErrorManager`
- `CompleteOutput` composite interface explicitly documented as backward-compat escape hatch
- `DependencyCache` interface in `internal/dependencies/analyzer.go:88`
- `InputReader` interface in `cmd_deps.go:24` for stdin abstraction
- `executableTemplate` interface in `internal/template.go:295`
- File comment: "defines focused interfaces following Interface Segregation Principle"

### Dependency Injection — High confidence

Evidence:

- `NewGenerator(config)` vs `NewGeneratorWithDependencies(config, output, progress)` — two-constructor pattern
- `Generator` struct fields are interfaces (`CompleteOutput`, `ProgressManager`), not concrete types
- `NewAnalyzer(client, repoInfo, cache DependencyCache)` — cache injected as interface
- `NewCacheAdapter(c *cache.Cache)` and `NewNoOpCache()` — adapter + null-object for testing
- `NullOutput` and `NullProgressManager` null objects for test isolation
- Unit test auto-detection in `NewGenerator()`: injects `NullOutput` in test binaries automatically
- `setupDepsUpgrade(config *AppConfig)` — accepts config for injection

### Adapter Pattern — Medium confidence

Evidence:

- `internal/dependencies/cache_adapter.go`: `CacheAdapter` wraps `internal/cache.Cache` to satisfy `DependencyCache` interface
- `NoOpCache` implements same interface with no-ops — classic null-object adapter
- `wrapHandlerWithErrorHandling()` in `main.go` — adapts error-returning handlers to Cobra's `func(cmd, args)` signature

### Command Pattern — High confidence

Evidence:

- All Cobra command handlers return `error` (non-standard; Cobra uses `func(cmd, args)`)
- `wrapHandlerWithErrorHandling()` bridges the two signatures
- Each command encapsulates its operation: `genHandler`, `depsListHandler`, `depsSecurityHandler`, `depsOutdatedHandler`, `depsUpgradeHandler`
- Handlers are pure functions, decoupled from the command struct definition

### Repository Pattern — Low confidence

Evidence:

- `internal/cache/cache.go` provides generic `Cache` with `Get`/`Set`/persistence — partial repository shape
- No explicit Repository type names; pattern is present but not named

## Detected Combination

### Custom hybrid: Layered CLI tool + Pipe-and-Filter core + ISP-driven Dependency Injection

This is a CLI application, not a web service or domain-rich system. The combination differs from classic enterprise patterns:

- The "domain" is documentation generation; the pipeline IS the domain logic
- Layers are CLI / Core pipeline / Infrastructure, not Presentation / Business / Data
- DI is used for testability and output abstraction, not for swapping business rules

## Inferred Structural Rules

1. **CLI layer must not contain domain logic** — handlers coordinate only; parsing, rendering, and analysis belong in `internal/`
2. **Pipeline stages must be pure and composable** — each stage takes typed input, returns typed output, no side effects except at the terminal stage (file write)
3. **Interfaces must be segregated** — no parameter should accept `CompleteOutput` when `MessageLogger` suffices; use narrowest interface
4. **Constructors must follow the two-constructor pattern** — `New*(config)` for production, `New*WithDependencies(...)` for injection; never add optional fields directly to structs
5. **Infrastructure packages must not import internal core** — `internal/cache`, `internal/git`, `internal/dependencies` must not import `internal` directly (only `appconstants`)
6. **Command handlers must return `error`** — never call `os.Exit()` in handlers; `wrapHandlerWithErrorHandling` owns exit logic
7. **Null objects required for all output interfaces** — `NullOutput`, `NullProgressManager` must cover all interface methods; tests must not depend on real I/O
8. **Adapters live in the consumer package** — `CacheAdapter` lives in `dependencies/`, not in `cache/`; adapters are the consumer's responsibility

## Ambiguities & Contradictions

- `internal/` is a flat package for most core types (`parser.go`, `template.go`, `generator.go`, `config.go`, `output.go`)
  — this grows coupling as the package expands; no sub-domain separation exists within core
- `cmd_deps.go` defines `InputReader` interface and `StdinReader` struct — these belong in `internal/` by the layering rule, not in the CLI layer
- `Generator` struct exposes `Config *AppConfig` as a public field rather than through an interface — breaks DI rule for config itself
- `internal/analyzer_helpers.go` imports `internal/dependencies` — a core file importing a sub-infrastructure package; acceptable if one-directional but worth watching

## Drift

First run — no prior profile to compare against.
