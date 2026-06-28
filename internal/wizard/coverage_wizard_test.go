package wizard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	covRunnerUbuntu  = "ubuntu-latest"
	covRunnerWindows = "windows-latest"
)

// TestCovWizardFindActionFilesTraversal covers the two early-return guards in
// findActionFiles: a raw ".." path component, and a path that filepath.Clean
// normalizes (so cleanDir != dir). Both must yield an empty slice.
func TestCovWizardFindActionFilesTraversal(t *testing.T) {
	wizard := testWizard(t, "")

	t.Run("dotdot component", func(t *testing.T) {
		if files := wizard.findActionFiles("../escape"); len(files) != 0 {
			t.Errorf("expected no files for traversal path, got %d", len(files))
		}
	})

	t.Run("non-canonical path", func(t *testing.T) {
		// "foo/./bar" has no ".." component but Clean rewrites it to "foo/bar",
		// so the cleanDir != dir guard fires.
		if files := wizard.findActionFiles("foo/./bar"); len(files) != 0 {
			t.Errorf("expected no files for non-canonical path, got %d", len(files))
		}
	})
}

// TestCovWizardNewConfigExporterNilOutput covers the nil-output default branch of
// NewConfigExporter.
func TestCovWizardNewConfigExporterNilOutput(t *testing.T) {
	t.Parallel()

	exporter := NewConfigExporter(nil)
	if exporter == nil {
		t.Fatal("NewConfigExporter(nil) returned nil")
	}
	if exporter.output == nil {
		t.Error("NewConfigExporter(nil) should install a default output logger")
	}
}

// TestCovWizardSanitizeConfigLegacyDefaults covers the legacy-field branches of
// sanitizeConfig: Template/Header/Footer/Schema equal to the package defaults are
// blanked out in the exported copy.
func TestCovWizardSanitizeConfigLegacyDefaults(t *testing.T) {
	t.Parallel()

	defaults := internal.DefaultAppConfig()
	exporter := NewConfigExporter(internal.NewColoredOutput(true))

	cfg := &internal.AppConfig{
		Organization: testWizardOrg,
		Template:     defaults.Template,
		Header:       defaults.Header,
		Footer:       defaults.Footer,
		Schema:       defaults.Schema,
	}

	sanitized := exporter.sanitizeConfig(cfg, false)
	testutil.AssertEqual(t, "", sanitized.Template)
	testutil.AssertEqual(t, "", sanitized.Header)
	testutil.AssertEqual(t, "", sanitized.Footer)
	testutil.AssertEqual(t, "", sanitized.Schema)
	// Non-default field is preserved.
	testutil.AssertEqual(t, testWizardOrg, sanitized.Organization)
}

// TestCovWizardExportTOMLMultipleRunners covers the multi-element branch of
// writeWorkflowSection (the comma separator between runners).
func TestCovWizardExportTOMLMultipleRunners(t *testing.T) {
	t.Parallel()

	exporter := NewConfigExporter(internal.NewColoredOutput(true))
	cfg := &internal.AppConfig{
		Organization: testWizardOrg,
		Repository:   testWizardRepo,
		Theme:        testWizardThemeDef,
		OutputFormat: testWizardFormatMD,
		RunsOn:       []string{covRunnerUbuntu, covRunnerWindows},
	}

	outputPath := filepath.Join(t.TempDir(), "config.toml")
	if err := exporter.ExportConfig(cfg, FormatTOML, outputPath, false); err != nil {
		t.Fatalf("ExportConfig TOML failed: %v", err)
	}

	data, err := os.ReadFile(outputPath) // #nosec G304 -- test output path
	testutil.AssertNoError(t, err)
	content := string(data)
	if !strings.Contains(content, `runs_on = ["`+covRunnerUbuntu+`", "`+covRunnerWindows+`"]`) {
		t.Errorf("TOML runs_on missing comma-separated runners:\n%s", content)
	}
}

// TestCovWizardExportTOMLMinimal covers the early-return branches of
// writeWorkflowSection and writeMapSection when there are no runners,
// permissions, or variables to write.
func TestCovWizardExportTOMLMinimal(t *testing.T) {
	t.Parallel()

	exporter := NewConfigExporter(internal.NewColoredOutput(true))
	cfg := &internal.AppConfig{
		Organization: testWizardOrg,
		Theme:        testWizardThemeDef,
		OutputFormat: testWizardFormatMD,
	}

	outputPath := filepath.Join(t.TempDir(), "minimal.toml")
	if err := exporter.ExportConfig(cfg, FormatTOML, outputPath, false); err != nil {
		t.Fatalf("ExportConfig TOML failed: %v", err)
	}

	data, err := os.ReadFile(outputPath) // #nosec G304 -- test output path
	testutil.AssertNoError(t, err)
	content := string(data)
	if strings.Contains(content, "runs_on") {
		t.Errorf("minimal config should not emit runs_on:\n%s", content)
	}
	if strings.Contains(content, "[permissions]") || strings.Contains(content, "[variables]") {
		t.Errorf("minimal config should not emit map sections:\n%s", content)
	}
}

// Note: promptSensitive's terminal branch (term.IsTerminal true) is unreachable
// without a real TTY, so it is intentionally left untested here. The stickyWriter
// I/O-error branches and writeFileAtomic's failure paths require an unwritable
// destination (disk-full/permission), which is not portably forced in unit tests.
