# No Duplicate String Constants

Never repeat string literals across test files (threshold: > 2 uses).
Use constants from `appconstants/` for production strings.
Use constants from `testutil/test_constants.go` for test-only strings.
Check both packages for existing constants before adding new ones.
