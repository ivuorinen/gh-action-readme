// Package cmd contains CLI subcommands for gh-action-readme.
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/schemas"
)

func TestVersionAndAboutCommands(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logrus.SetOutput(logBuf)
	defer logrus.SetOutput(os.Stderr)
	root := newTestRootCmd()
	runCmd(root, "version")
	if !strings.Contains(logBuf.String(), "gh-action-readme version") {
		t.Errorf("version output missing: %s", logBuf.String())
	}
	logBuf.Reset()
	runCmd(root, "about")
	if !strings.Contains(logBuf.String(), "gh-action-readme: Generates README.md and HTML") {
		t.Errorf("about output missing: %s", logBuf.String())
	}
}

// --- BEGIN: Testdata action validation helpers and tests ---

// validateActionYMLWithRepoConfig validates a single action.yml using the repo config and schema.
// Returns error if validation fails, nil if valid.
func validateActionYMLWithRepoConfig(actionPath string) error {
	// Use config and schema from repo root
	repoRoot := ".."
	configPath := filepath.Join(repoRoot, "config.yaml")
	schemaPath := filepath.Join(repoRoot, "schemas", "action.schema.json")

	cfg, err := internal.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.Schema != "" {
		schemaPath = filepath.Join(repoRoot, cfg.Schema)
	}
	// Validate using schema
	schemaErrs, err := internal.ValidateActionYMLSchema(actionPath, schemaPath)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}
	if len(schemaErrs) > 0 {
		return fmt.Errorf("schema errors: %v", schemaErrs)
	}
	// Validate required fields
	action, loadErr := internal.ParseActionYML(actionPath)
	if loadErr != nil {
		return fmt.Errorf("failed to load action.yml: %w", loadErr)
	}
	result := internal.ValidateActionYML(action)
	if len(result.MissingFields) > 0 {
		return fmt.Errorf("missing required fields: %v", result.MissingFields)
	}

	return nil
}

// collectTestdataActionDirs returns all testdata subdirs matching the given prefix.
// This test runs from the cmd/ directory, so testdata is ../testdata.
func collectTestdataActionDirs(t *testing.T, prefix string) []string {
	t.Helper()
	var dirs []string
	entries, err := os.ReadDir("../testdata")
	if err != nil {
		t.Fatalf("failed to read testdata dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			dirs = append(dirs, filepath.Join("../testdata", entry.Name()))
		}
	}

	return dirs
}

// findActionYMLInDir returns the path to action.yml in the given dir, or "" if not found.
func findActionYMLInDir(dir string) string {
	for _, name := range []string{"action.yml", "action.yaml"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func Test_ExampleActions_PassValidation(t *testing.T) {
	t.Parallel()
	exampleDirs := collectTestdataActionDirs(t, "example-")
	for _, dir := range exampleDirs {
		t.Run(
			dir, func(t *testing.T) {
				t.Parallel()
				actionPath := findActionYMLInDir(dir)
				if actionPath == "" {
					t.Fatalf("no action.yml found in %s", dir)
				}
				err := validateActionYMLWithRepoConfig(actionPath)
				if err != nil {
					t.Errorf("expected valid action, got error: %v", err)
				}

				if err != nil {
					t.Errorf("expected valid action, got error: %v", err)
				}
			},
		)
	}
}

// Test dry-run mode: validation should succeed and not write files (simulate by running validation)
func Test_ExampleAction_DryRun(t *testing.T) {
	dir := "../testdata/example-action"
	actionPath := findActionYMLInDir(dir)
	if actionPath == "" {
		t.Fatalf("no action.yml found in %s", dir)
	}
	// Just run validation, since our helper does not write files
	err := validateActionYMLWithRepoConfig(actionPath)
	if err != nil {
		t.Errorf("expected dry-run validation to pass, got error: %v", err)
	}
}

// Test malformed YAML: should return a parse error
func Test_BrokenAction_MalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.yml")
	_ = os.WriteFile(badFile, []byte("::::"), 0o600)
	err := validateActionYMLWithRepoConfig(badFile)
	if err == nil {
		t.Errorf("expected error for malformed YAML, got nil")
	}
	// Accept either a parse error or schema errors for missing required fields/unknown fields
	if err != nil &&
		!strings.Contains(err.Error(), "failed to load action.yml") &&
		!strings.Contains(err.Error(), "schema errors") {
		t.Errorf("unexpected error for malformed YAML: %v", err)
	}
}

func Test_BrokenActions_FailValidation(t *testing.T) {
	t.Parallel()
	brokenDirs := collectTestdataActionDirs(t, "broken-action-")
	for _, dir := range brokenDirs {
		t.Run(
			dir, func(t *testing.T) {
				t.Parallel()
				actionPath := findActionYMLInDir(dir)
				if actionPath == "" {
					t.Fatalf("no action.yml found in %s", dir)
				}
				err := validateActionYMLWithRepoConfig(actionPath)
				if err == nil {
					t.Errorf("expected schema validation to fail, but it passed")
				}
			},
		)
	}
}

// --- END: Testdata action validation helpers and tests ---

// --- BEGIN: Concurrency stress tests ---

// TestConcurrency_ManyValidAndInvalidActions runs validation on many temp dirs in parallel.
//
//nolint:paralleltest,tparallel
func TestConcurrency_ManyValidAndInvalidActions(t *testing.T) {
	t.Parallel()
	var projectRoot string
	var errRepo error
	projectRoot, errRepo = findProjectRoot()
	if errRepo != nil {
		t.Fatalf("failed to find project root: %v", errRepo)
	}

	// Prepare paths for later use
	repoSchema := filepath.Join(projectRoot, "schemas", "action.schema.json")
	testCases := filepath.Join(projectRoot, "testdata")

	// Use a valid and invalid action.yml as templates
	validSrc := filepath.Join(testCases, "example-action", "action.yml")
	invalidSrc := filepath.Join(testCases, "broken-action-extra-field", "action.yml")

	// initialize counters
	var validCount int32
	var invalidCount int32
	const numValid int32 = 30   // How many valid actions to create
	const numInvalid int32 = 20 // How many invalid actions to create

	tmpDir := t.TempDir()
	allDirs := make([]string, 0, 50)

	// Create valid dirs
	for i := range numValid {
		dir := filepath.Join(tmpDir, "valid", fmt.Sprintf("action-%d", i))
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		dst := filepath.Join(dir, "action.yml")
		data, readErr := os.ReadFile(validSrc) // #nosec G304 - Local file
		if readErr != nil {
			t.Fatalf("read validSrc: %v", readErr)
		}
		if writeErr := os.WriteFile(dst, data, 0o600); writeErr != nil {
			t.Fatalf("write valid action: %v", writeErr)
		}
		allDirs = append(allDirs, dir)
	}

	// Create invalid dirs
	for i := range numInvalid {
		dir := filepath.Join(tmpDir, "invalid", fmt.Sprintf("action-%d", i))
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		dst := filepath.Join(dir, "action.yml")
		data, readErr := os.ReadFile(invalidSrc) // #nosec G304 - Local file
		if readErr != nil {
			t.Fatalf("read invalidSrc: %v", readErr)
		}
		if writeErr := os.WriteFile(dst, data, 0o600); writeErr != nil {
			t.Fatalf("write invalid action: %v", writeErr)
		}
		allDirs = append(allDirs, dir)
	}

	// Shuffle allDirs to randomize order
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec G404 -- test shuffle, not crypto
	r.Shuffle(len(allDirs), func(i, j int) { allDirs[i], allDirs[j] = allDirs[j], allDirs[i] })

	// Validate all in parallel

	t.Run(
		"validate-many-actions", func(t *testing.T) {
			for _, dir := range allDirs {
				t.Run(
					dir, func(t *testing.T) {
						t.Parallel()
						validateActionDir(t, dir, repoSchema, &validCount, &invalidCount)
					},
				)
			}
		},
	)

	if got := atomic.LoadInt32(&validCount); got != numValid {
		t.Errorf("expected %d valid actions, got %d", numValid, got)
	}

	if got := atomic.LoadInt32(&invalidCount); got != numInvalid {
		t.Errorf("expected %d invalid actions, got %d", numInvalid, got)
	}
}

// validateActionDir is a helper to validate a single action directory and update counters.
func validateActionDir(t *testing.T, dir, repoSchema string, validCount, invalidCount *int32) {
	t.Helper()
	actionPath := findActionYMLInDir(dir)
	if actionPath == "" {
		t.Fatalf("no action.yml found in %s", dir)
	}

	schemaErrs, err := internal.ValidateActionYMLSchema(actionPath, repoSchema)
	if err != nil {
		if strings.Contains(dir, "/valid/") {
			t.Errorf("expected valid action, got error: %v", err)
		} else {
			atomic.AddInt32(invalidCount, 1)
		}

		return
	}
	if len(schemaErrs) > 0 {
		if strings.Contains(dir, "/valid/") {
			t.Errorf("expected valid action, got schema errors: %v", schemaErrs)
		} else {
			atomic.AddInt32(invalidCount, 1)
		}

		return
	}

	action, loadErr := internal.ParseActionYML(actionPath)
	if loadErr != nil {
		if strings.Contains(dir, "/valid/") {
			t.Errorf("expected valid action, got parse error: %v", loadErr)
		} else {
			atomic.AddInt32(invalidCount, 1)
		}

		return
	}
	result := internal.ValidateActionYML(action)
	if len(result.MissingFields) > 0 {
		if strings.Contains(dir, "/valid/") {
			t.Errorf("expected valid action, got missing fields: %v", result.MissingFields)
		} else {
			atomic.AddInt32(invalidCount, 1)
		}

		return
	}
	if strings.Contains(dir, "/valid/") {
		atomic.AddInt32(validCount, 1)
	} else {
		t.Errorf("expected invalid action to fail validation, but it passed")
	}
}

// --- END: Concurrency stress tests ---

// --- BEGIN: CLI flag tests ---

func TestCLI_MissingConfigFlag(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logrus.SetOutput(logBuf)
	defer logrus.SetOutput(os.Stderr)
	root := newTestRootCmd()
	// Remove config.yaml if present, or use a non-existent config
	cmdOut := runCmd(root, "validate", "--config", "doesnotexist.yaml")
	if !strings.Contains(logBuf.String(), "Failed to load config") &&
		!strings.Contains(cmdOut, "Failed to load config") {
		t.Errorf(
			"expected error for missing config, got: log: %s, cmd: %s",
			logBuf.String(),
			cmdOut,
		)
	}
}

func TestCLI_UnknownFlag(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logrus.SetOutput(logBuf)
	defer logrus.SetOutput(os.Stderr)
	root := newTestRootCmd()
	cmdOut := runCmd(root, "validate", "--notaflag")
	if !strings.Contains(logBuf.String(), "unknown flag") &&
		!strings.Contains(logBuf.String(), "flag provided but not defined") &&
		!strings.Contains(cmdOut, "unknown flag") &&
		!strings.Contains(cmdOut, "flag provided but not defined") {
		t.Errorf(
			"expected error for unknown flag, got: log: %s, cmd: %s",
			logBuf.String(),
			cmdOut,
		)
	}
}

func TestCLI_HelpFlag(t *testing.T) {
	root := newTestRootCmd()
	cmdOut := runCmd(root, "validate", "--help")
	if !strings.Contains(cmdOut, "Usage:") && !strings.Contains(cmdOut, "validate [flags]") {
		t.Errorf("expected help output, got: %s", cmdOut)
	}
}

// --- END: CLI flag tests ---

func TestSchemaCommand(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logrus.SetOutput(logBuf)
	defer logrus.SetOutput(os.Stderr)
	root := newTestRootCmd()
	runCmd(root, "schema")
	if !strings.Contains(logBuf.String(), "Schema: "+schemas.RelPath) {
		t.Errorf("schema output missing: %s", logBuf.String())
	}
}

func TestSchemaUpdateCommand(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logrus.SetOutput(logBuf)
	defer logrus.SetOutput(os.Stderr)
	root := newTestRootCmd()
	runCmd(root, "schema", "update")
	if !strings.Contains(logBuf.String(), "Downloading latest schema from") {
		t.Errorf("schema update output missing: %s", logBuf.String())
	}
}

func TestGenCommand_Minimal(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logrus.SetOutput(logBuf)
	defer logrus.SetOutput(os.Stderr)
	// Set up a minimal action.yml and config.yaml in a temp dir
	tmpDir := t.TempDir()
	actionDir := filepath.Join(tmpDir, "action")
	if err := os.MkdirAll(actionDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	actionPath := filepath.Join(actionDir, "action.yml")
	configPath := filepath.Join(tmpDir, "config.yaml")
	templateDir := filepath.Join(tmpDir, "templates")
	schemaDir := filepath.Join(tmpDir, "schemas")
	if err := os.MkdirAll(templateDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(schemaDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Minimal action.yml
	actionContent := `
name: foo
description: bar
runs:
  using: node20
`
	if err := os.WriteFile(actionPath, []byte(actionContent), 0o600); err != nil {
		t.Fatalf("write action.yml: %v", err)
	}
	// Minimal config.yaml
	configContent := `
defaults:
  name: "Default"
  description: "Default"
  runs: {}
  branding:
    icon: "zap"
    color: "yellow"
  version: main
github_org: "testorg"
template: "` + templateDir + `/readme"
header: "` + templateDir + `/header"
footer: "` + templateDir + `/footer"
schema: "` + schemaDir + `/action.schema.json"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}
	// Minimal templates
	mdTmpl := `# {{.Name}}
{{.Description}}
`
	if err := os.WriteFile(
		filepath.Join(templateDir, "readme.md.tmpl"),
		[]byte(mdTmpl), 0o600,
	); err != nil {
		t.Fatalf("write readme.md.tmpl: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(templateDir, "header.md.tmpl"),
		[]byte(""), 0o600,
	); err != nil {
		t.Fatalf("write header.md.tmpl: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(templateDir, "footer.md.tmpl"),
		[]byte(""), 0o600,
	); err != nil {
		t.Fatalf("write footer.md.tmpl: %v", err)
	}
	// Minimal schema
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["name", "description", "runs"],
  "properties": {
    "name": {"type": "string"},
    "description": {"type": "string"},
    "runs": {"type": "object"}
  }
}`
	if err := os.WriteFile(
		filepath.Join(schemaDir, "action.schema.json"),
		[]byte(schemaContent), 0o600,
	); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	// Run gen command
	oldWd, _ := os.Getwd()
	defer func(dir string) {
		err := os.Chdir(dir)
		if err != nil {
			t.Fatalf("failed to change back to original dir: %v", err)
		}
	}(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	root := newTestRootCmd()
	runCmd(root, "gen", "--config", configPath)
	if !strings.Contains(logBuf.String(), "Generated documentation for") {
		t.Errorf("gen output missing: %s", logBuf.String())
	}
	// Check README.md was generated
	readmePath := filepath.Join(actionDir, "README.md")
	data, err := os.ReadFile(readmePath) // #nosec G304 -- test file, path is controlled
	if err != nil {
		t.Fatalf("README.md not generated: %v", err)
	}
	if !strings.Contains(string(data), "# foo") {
		t.Errorf("README.md missing content: %s", string(data))
	}
}

func TestValidateCommand_Minimal(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logrus.SetOutput(logBuf)
	defer logrus.SetOutput(os.Stderr)
	tmpDir := t.TempDir()
	actionPath := filepath.Join(tmpDir, "action.yml")
	configPath := filepath.Join(tmpDir, "config.yaml")
	schemaDir := filepath.Join(tmpDir, "schemas")
	if err := os.MkdirAll(schemaDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	actionContent := `
name: foo
description: bar
runs:
  using: node20
`
	if err := os.WriteFile(actionPath, []byte(actionContent), 0o600); err != nil {
		t.Fatalf("write action.yml: %v", err)
	}
	configContent := `
defaults:
  name: "Default"
  description: "Default"
  runs: {}
  branding:
    icon: "zap"
    color: "yellow"
  version: main
github_org: "testorg"
template: "templates/readme"
header: "templates/header"
footer: "templates/footer"
schema: "` + schemaDir + `/action.schema.json"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}
	// Minimal schema
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["name", "description", "runs"],
  "properties": {
    "name": {"type": "string"},
    "description": {"type": "string"},
    "runs": {"type": "object"}
  }
}`
	if err := os.WriteFile(
		filepath.Join(schemaDir, "action.schema.json"),
		[]byte(schemaContent), 0o600,
	); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	oldWd, _ := os.Getwd()
	defer func(dir string) {
		err := os.Chdir(dir)
		if err != nil {
			t.Fatalf("failed to change back to original dir: %v", err)
		}
	}(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	root := newTestRootCmd()
	runCmd(root, "validate", "--config", configPath)
	if !strings.Contains(logBuf.String(), "All required fields present") {
		t.Errorf("validate output missing: %s", logBuf.String())
	}
}

func newTestRootCmd() *cobra.Command {
	logrus.SetLevel(logrus.InfoLevel)
	rootCmd := &cobra.Command{Use: "gh-action-readme"}
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	rootCmd.AddCommand(GenCmd())
	rootCmd.AddCommand(ValidateCmd())
	rootCmd.AddCommand(SchemaCmd())
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the version number",
			Run: func(cmd *cobra.Command, args []string) {
				logrus.Infof("gh-action-readme version test")
			},
		},
	)
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "about",
			Short: "About this tool",
			Run: func(cmd *cobra.Command, args []string) {
				logrus.Info(
					"gh-action-readme: Generates README.md and HTML for " +
						"GitHub Actions. MIT License.",
				)
			},
		},
	)

	return rootCmd
}

func TestGenCommand_MissingTemplate(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logrus.SetOutput(logBuf)
	defer logrus.SetOutput(os.Stderr)
	tmpDir := t.TempDir()
	actionDir := filepath.Join(tmpDir, "action")
	if err := os.MkdirAll(actionDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	actionPath := filepath.Join(actionDir, "action.yml")
	configPath := filepath.Join(tmpDir, "config.yaml")
	schemaDir := filepath.Join(tmpDir, "schemas")
	if err := os.MkdirAll(schemaDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Minimal action.yml
	actionContent := `
name: foo
description: bar
runs:
  using: node20
`
	if err := os.WriteFile(actionPath, []byte(actionContent), 0o600); err != nil {
		t.Fatalf("write action.yml: %v", err)
	}
	// Config with missing template
	configContent := `
defaults:
  name: "Default"
  description: "Default"
  runs: {}
  branding:
    icon: "zap"
    color: "yellow"
  version: main
github_org: "testorg"
template: "notemplates/readme"
header: "notemplates/header"
footer: "notemplates/footer"
schema: "schemas/action.schema.json"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}
	// Minimal schema
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["name", "description", "runs"],
  "properties": {
    "name": {"type": "string"},
    "description": {"type": "string"},
    "runs": {"type": "object"}
  }
}`
	if err := os.WriteFile(
		filepath.Join(schemaDir, "action.schema.json"),
		[]byte(schemaContent), 0o600,
	); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	oldWd, _ := os.Getwd()
	defer func(dir string) {
		err := os.Chdir(dir)
		if err != nil {
			t.Fatalf("failed to change back to original dir: %v", err)
		}
	}(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	root := newTestRootCmd()
	cmdOut := runCmd(root, "gen", "--config", configPath)
	if !strings.Contains(logBuf.String(), "not found") && !strings.Contains(cmdOut, "not found") {
		t.Errorf(
			"expected error for unknown flag, got: log: %s, cmd: %s",
			logBuf.String(),
			cmdOut,
		)
	}
}

func TestGenCommand_MissingConfig(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logrus.SetOutput(logBuf)
	defer logrus.SetOutput(os.Stderr)
	root := newTestRootCmd()
	cmdOut := runCmd(root, "gen", "--config", "doesnotexist.yaml")
	t.Logf("logBuf: %q", logBuf.String())
	t.Logf("cmdOut: %q", cmdOut)
	if !strings.Contains(logBuf.String(), "Failed to load config") &&
		!strings.Contains(cmdOut, "Failed to load config") {
		t.Errorf(
			"expected error for missing config, got: log: %s, cmd: %s",
			logBuf.String(),
			cmdOut,
		)
	}
}

func TestGenCommand_InvalidFlag(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logrus.SetOutput(logBuf)
	defer logrus.SetOutput(os.Stderr)
	root := newTestRootCmd()
	cmdOut := runCmd(root, "gen", "--notaflag")
	if !strings.Contains(logBuf.String(), "unknown flag") &&
		!strings.Contains(logBuf.String(), "flag provided but not defined") &&
		!strings.Contains(cmdOut, "unknown flag") &&
		!strings.Contains(cmdOut, "flag provided but not defined") {
		t.Errorf(
			"expected error for missing config, got: log: %s, cmd: %s",
			logBuf.String(),
			cmdOut,
		)
	}
}

func runCmd(root *cobra.Command, args ...string) string {
	// t.Helper() not available here, but this is a helper for tests.
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	_ = root.Execute()

	return buf.String()
}
