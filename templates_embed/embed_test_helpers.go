package templatesembed

import "testing"

// assertTemplateLoaded validates template loading results.
// This helper reduces cognitive complexity in embed tests by centralizing
// the template loading validation logic that was repeated across test functions.
func assertTemplateLoaded(t *testing.T, content []byte, err error, expectError bool, minContentLength int) {
	t.Helper()

	if expectError {
		if err == nil {
			t.Error("expected error but got none")
		}

		return
	}

	// Success case
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(content) < minContentLength {
		t.Errorf("content too short: got %d bytes, want at least %d", len(content), minContentLength)
	}
}
