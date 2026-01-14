package internal

import (
	"testing"

	"github.com/google/go-github/v74/github"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// assertGitHubClient validates GitHub client creation results.
// This helper reduces cognitive complexity in config tests by centralizing
// the client validation logic that was repeated across test cases.
//
//nolint:unused // Prepared for future use in config tests
func assertGitHubClient(t *testing.T, client *github.Client, err error, expectError bool) {
	t.Helper()

	if expectError {
		if err == nil {
			t.Error(testutil.TestErrNoErrorGotNone)
		}
		if client != nil {
			t.Error("expected nil client on error")
		}

		return
	}

	// Success case
	if err != nil {
		t.Errorf(testutil.TestErrUnexpected, err)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
}
