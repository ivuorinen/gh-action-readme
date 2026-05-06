---
name: fixture-add
description: Scaffold a new YAML test fixture and register its path constant in testutil/test_constants.go
---

The user provides: fixture type and content description.

Steps:

1. Choose the correct subdirectory under `testdata/yaml-fixtures/`:
    - GitHub Action → `actions/javascript/`, `actions/composite/`, or `actions/docker/`
    - Invalid/error case → `actions/invalid/` or `error-scenarios/`
    - Config file → `configs/`
    - Dependency-heavy action → `dependencies/`

2. Create the YAML fixture file following the naming convention of existing fixtures in that directory.
    Inspect one existing fixture file first to match the YAML structure and formatting.

3. Add a constant to `testutil/test_constants.go` following the `TestFixture<Name>` pattern,
    pointing to the path relative to `testdata/yaml-fixtures/` (e.g. `"actions/composite/my-fixture.yml"`).

4. Run `go build ./testutil/...` to verify it compiles.

5. Show the user the usage pattern:

    ```go
    testutil.WriteActionFixture(t, tmpDir, testutil.TestFixture<Name>)
    // or for config fixtures:
    testutil.MustReadFixture(testutil.TestFixture<Name>)
    ```
