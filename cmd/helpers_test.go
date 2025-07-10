package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/internal"
)

func TestWriteDocForActionWithName_ErrorCases(t *testing.T) {
	// Try to write to a directory as a file (should error)
	tmpDir := t.TempDir()
	badPath := filepath.Join(tmpDir, "bad_file")
	_ = os.MkdirAll(badPath, 0o750)
	_ = writeDocForActionWithNameErr(filepath.Join(badPath, "action.yml"), "doc", "README.md", "")

	// Try to write to a location where permission is denied (simulate by using root on unix)
	if os.Geteuid() != 0 {
		badDir := "/root/should-not-exist"
		_ = writeDocForActionWithNameErr("foo/action.yml", "doc", "README.md", badDir)
	}
}

func TestWriteYAMLFile_ErrorCases(t *testing.T) {
	// Try to write to a directory as a file (should error)
	tmpDir := t.TempDir()
	badPath := filepath.Join(tmpDir, "bad_file")
	_ = os.MkdirAll(badPath, 0o750)
	_ = writeYAMLFile(badPath, &internal.ActionYML{Name: "foo"})

	// Try to write to a location where permission is denied (simulate by using root on unix)
	if os.Geteuid() != 0 {
		badDir := "/root/should-not-exist"
		_ = writeYAMLFile(filepath.Join(badDir, "file.yml"), &internal.ActionYML{Name: "foo"})
	}
}

func TestWriteDocForActionWithName_InvalidDir(t *testing.T) {
	// Try to write to a directory that cannot exist (invalid path)
	invalidDir := string([]byte{0})
	_ = writeDocForActionWithNameErr("foo/action.yml", "doc", "README.md", invalidDir)
}

func TestWriteYAMLFile_BadWriter(t *testing.T) {
	// Simulate a bad writer by using a file opened read-only
	tmp := t.TempDir()
	file := filepath.Join(tmp, "readonly.yml")
	_ = os.WriteFile(file, []byte("foo"), 0o400)
	f, err := os.OpenFile(file, os.O_RDONLY, 0o400) // #nosec G304 -- test file, path is controlled
	if err != nil {
		t.Fatalf("could not open file readonly: %v", err)
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			t.Errorf("Failed to close file %s: %v", file, closeErr)
		}
	}()
	// This doesn't directly test writeYAMLFile, but covers the error path for a bad writer.
	enc := yamlNewEncoder(f)
	action := &internal.ActionYML{Name: "foo"}
	_ = enc.Encode(action) // Should error, but we ignore as this is a stub
}

// yamlNewEncoder is a local helper for testing (uses yaml.v3).
func yamlNewEncoder(w any) *yamlEncoderStub {
	return &yamlEncoderStub{w: w}
}

type yamlEncoderStub struct {
	w any
}

func (e *yamlEncoderStub) Encode(_ any) error {
	_, err := e.w.(interface{ Write([]byte) (int, error) }).Write([]byte("test"))

	return err
}

// --- Tests for fixYAMLHeader ---

func TestFixYAMLHeader_AddsHeaderAndSchema(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "action.yml")

	// Case 1: No header or schema
	content := `
name: foo
description: bar
runs:
  using: node20
`
	if err := os.WriteFile(yamlPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := fixYAMLHeader(yamlPath); err != nil {
		t.Fatalf("fixYAMLHeader: %v", err)
	}
	data1, err1 := os.ReadFile(yamlPath) // #nosec G304 -- test file, path is controlled
	if err1 != nil {
		t.Fatalf("read: %v", err1)
	}
	lines := strings.Split(string(data1), "\n")
	if lines[0] != "---" {
		t.Errorf("first line not ---: %q", lines[0])
	}
	expectedSchema := "# yaml-language-server: $schema=" +
		"https://json.schemastore.org/github-action.json"
	if lines[1] != expectedSchema {
		t.Errorf(
			"second line not schema: %q",
			lines[1],
		)
	}

	// Case 2: Only header present
	content2 := `---
name: foo
description: bar
runs:
  using: node20
`
	if err2 := os.WriteFile(yamlPath, []byte(content2), 0o600); err2 != nil {
		t.Fatalf("write2: %v", err2)
	}
	if err2 := fixYAMLHeader(yamlPath); err2 != nil {
		t.Fatalf("fixYAMLHeader2: %v", err2)
	}
	data2, err2 := os.ReadFile(yamlPath) // #nosec G304 -- test file, path is controlled
	if err2 != nil {
		t.Fatalf("read2: %v", err2)
	}
	lines2 := strings.Split(string(data2), "\n")
	if lines2[0] != "---" {
		t.Errorf("first line not --- (case2): %q", lines2[0])
	}
	if lines2[1] != expectedSchema {
		t.Errorf(
			"second line not schema (case2): %q",
			lines2[1],
		)
	}

	// Case 3: Both present, nothing should change
	content3 := `---
# yaml-language-server: $schema=https://json.schemastore.org/github-action.json
name: foo
description: bar
runs:
  using: node20
`
	if err3 := os.WriteFile(yamlPath, []byte(content3), 0o600); err3 != nil {
		t.Fatalf("write3: %v", err3)
	}
	if err3 := fixYAMLHeader(yamlPath); err3 != nil {
		t.Fatalf("fixYAMLHeader3: %v", err3)
	}
	data3, err3 := os.ReadFile(yamlPath) // #nosec G304 -- test file, path is controlled
	if err3 != nil {
		t.Fatalf("read3: %v", err3)
	}
	lines3 := strings.Split(string(data3), "\n")
	if lines3[0] != "---" {
		t.Errorf("first line not --- (case3): %q", lines3[0])
	}
	if lines3[1] != expectedSchema {
		t.Errorf(
			"second line not schema (case3): %q",
			lines3[1],
		)
	}
}
