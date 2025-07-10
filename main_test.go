package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_HelpAndVersion(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"gh-action-readme", "--help"}
	code := run()
	if code != 0 {
		t.Errorf("run() with --help returned nonzero exit code: %d", code)
	}

	os.Args = []string{"gh-action-readme", "version"}
	code = run()
	if code != 0 {
		t.Errorf("run() with version returned nonzero exit code: %d", code)
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"gh-action-readme", "doesnotexist"}
	code := run()
	if code == 0 {
		t.Errorf("run() with unknown command should return nonzero exit code")
	}
}

func TestMain_HelpAndVersion(t *testing.T) {
	bin := buildTestBinary(t)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("failed to remove test binary: %v", err)
		}
	}(bin)

	// Test --help
	out, err := exec.Command(
		bin, "--help",
	).CombinedOutput() // #nosec G204 -- test subprocess, controlled input
	if err != nil {
		t.Fatalf("help failed: %v\n%s", err, string(out))
	}
	if !bytes.Contains(out, []byte("gh-action-readme")) ||
		!bytes.Contains(out, []byte("Usage:")) {
		t.Errorf("help output missing expected content: %s", string(out))
	}

	// Test version
	out, err = exec.Command(
		bin, "version",
	).CombinedOutput() // #nosec G204 -- test subprocess, controlled input
	if err != nil {
		t.Fatalf("version failed: %v\n%s", err, string(out))
	}
	if !bytes.Contains(out, []byte("gh-action-readme version")) {
		t.Errorf("version output missing: %s", string(out))
	}
}

func TestMain_ErrorHandling(t *testing.T) {
	bin := buildTestBinary(t)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("failed to remove test binary: %v", err)
		}
	}(bin)

	// Run with an unknown command
	out, err := exec.Command(
		bin, "doesnotexist",
	).CombinedOutput() // #nosec G204 -- test subprocess, controlled input
	if err == nil {
		t.Errorf("expected error for unknown command")
	}
	if !bytes.Contains(out, []byte("unknown command")) &&
		!bytes.Contains(out, []byte("unknown")) {
		t.Errorf("missing unknown command error: %s", string(out))
	}
}

func buildTestBinary(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "gh-action-readme-test")
	cmd := exec.Command(
		"go", "build", "-o", bin, ".",
	) // #nosec G204 -- test subprocess, controlled input
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf(
			"failed to build test binary: %v\n%s",
			err, string(out),
		)
	}

	return bin
}

// Optional: Test main with no args (should print help)
func TestMain_NoArgs(t *testing.T) {
	bin := buildTestBinary(t)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("failed to remove test binary: %v", err)
		}
	}(bin)

	out, err := exec.Command(
		bin,
	).CombinedOutput() // #nosec G204 -- test subprocess, controlled input
	if err != nil {
		// Some CLI tools exit nonzero for help, that's fine
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() == 0 {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if !strings.Contains(string(out), "gh-action-readme") ||
		!strings.Contains(string(out), "Usage:") {
		t.Errorf("no-args output missing expected content: %s", string(out))
	}
}
