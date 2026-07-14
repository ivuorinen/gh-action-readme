# No testutil in Production Code

Never import `testutil` from a non-test file.
Only `_test.go` files may import `testutil`.
Put test helpers shared across tests in `*_test.go` files (or a test-tagged
file), never in production `.go` files.
