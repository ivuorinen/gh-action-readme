---
name: mutation-pkg
description: Run gremlins mutation testing on a specified Go package, following the same pattern as make test-mutation
---

The user provides a package path argument (e.g. `./internal/wizard`).

Run the mutation test using the exact Makefile pattern:

```bash
go run github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0 unleash --timeout-coefficient 10 --workers 1 <package>
```

Scale `--timeout-coefficient` for larger packages:

- Small packages (≤10 files): 10
- Medium packages (≤25 files): 20 (matches `test-mutation-validation`)
- Large packages (>25 files): 30+

Report:

- Mutation score: killed / total (percentage)
- List surviving mutations: file:line, mutation type (e.g. ConditionalsBoundary, IncrementDecrement)
- Status: ✅ ≥80% (project target) or ❌ below target
- For each surviving mutation, suggest the specific test case or assertion that would kill it
