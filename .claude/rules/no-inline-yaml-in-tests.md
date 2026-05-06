# No Inline YAML in Tests

Never embed YAML or config data inline in test code.
Always use fixture files from `testdata/yaml-fixtures/` loaded via `testutil.MustReadFixture()`.
Always add fixture path constants to `testutil/test_constants.go` before using them.
