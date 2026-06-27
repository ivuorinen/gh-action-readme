// Package wizard provides configuration export functionality.
package wizard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
)

// ExportFormat represents the supported export formats.
type ExportFormat string

const (
	// FormatYAML exports configuration as YAML.
	FormatYAML ExportFormat = appconstants.OutputFormatYAML
	// FormatJSON exports configuration as JSON.
	FormatJSON ExportFormat = appconstants.OutputFormatJSON
	// FormatTOML exports configuration as TOML.
	FormatTOML ExportFormat = appconstants.OutputFormatTOML
)

// ConfigExporter handles exporting configuration to various formats.
type ConfigExporter struct {
	output internal.MessageLogger
}

// NewConfigExporter creates a new configuration exporter.
func NewConfigExporter(output internal.MessageLogger) *ConfigExporter {
	if output == nil {
		output = internal.NewColoredOutput(false)
	}

	return &ConfigExporter{
		output: output,
	}
}

// ExportConfig exports the configuration to the specified format and path.
func (e *ConfigExporter) ExportConfig(config *internal.AppConfig, format ExportFormat, outputPath string) error {
	// Reject path traversal before creating any directory or writing the file:
	// outputPath comes from the --output flag and must not escape via a ".."
	// path segment. Check segments rather than a substring so legitimate names
	// that merely contain ".." (e.g. "config..bak.yaml") are still allowed.
	// Absolute paths without a ".." segment remain allowed (the user's choice).
	if slices.Contains(strings.Split(filepath.ToSlash(outputPath), "/"), "..") {
		return fmt.Errorf("refusing to export to a path containing a '..' segment: %q", outputPath)
	}

	// Refuse to overwrite the live config file. Export sanitizes the config
	// (strips the token and drops repo_overrides), so writing over the active
	// config — which is also GetDefaultOutputPath's default target — would
	// silently discard those fields. Require an explicit, different --output.
	if cfgPath, cfgErr := internal.GetConfigPath(); cfgErr == nil && samePath(cfgPath, outputPath) {
		if _, statErr := os.Stat(outputPath); statErr == nil {
			return fmt.Errorf(
				"refusing to overwrite the live config %q (export drops tokens and repo_overrides); "+
					"pass --output to choose a different file", outputPath)
		}
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	// #nosec G301 -- output directory permissions
	if err := os.MkdirAll(outputDir, appconstants.FilePermDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	switch format {
	case FormatYAML:
		return e.exportYAML(config, outputPath)
	case FormatJSON:
		return e.exportJSON(config, outputPath)
	case FormatTOML:
		return e.exportTOML(config, outputPath)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// samePath reports whether a and b resolve to the same filesystem path,
// comparing cleaned absolute forms so "./config.yaml" and an absolute config
// path are recognized as identical.
func samePath(a, b string) bool {
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}

	return absA == absB
}

// GetSupportedFormats returns the list of supported export formats.
func (e *ConfigExporter) GetSupportedFormats() []ExportFormat {
	return []ExportFormat{FormatYAML, FormatJSON, FormatTOML}
}

// GetDefaultOutputPath returns the default output path for a given format.
func (e *ConfigExporter) GetDefaultOutputPath(format ExportFormat) (string, error) {
	configPath, err := internal.GetConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	dir := filepath.Dir(configPath)

	switch format {
	case FormatYAML:
		return filepath.Join(dir, appconstants.ConfigYAML), nil
	case FormatJSON:
		return filepath.Join(dir, "config.json"), nil
	case FormatTOML:
		return filepath.Join(dir, "config.toml"), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// stickyWriter wraps an io.Writer and remembers the first write error so a long
// run of writes can be checked once at the end instead of after every call. Once
// an error has occurred, later writes are skipped. It lets the TOML/YAML export
// helpers report I/O failures (disk-full, quota) that they would otherwise
// discard — which is essential for writeFileAtomic's integrity guarantee: if the
// write callback returns nil despite a failed write, a truncated temp file is
// renamed over the user's valid config.
type stickyWriter struct {
	w   io.Writer
	err error
}

// Write implements io.Writer, recording the first error.
func (s *stickyWriter) Write(p []byte) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	var n int
	n, s.err = s.w.Write(p)

	return n, s.err
}

// WriteString writes a string, recording the first error.
func (s *stickyWriter) WriteString(str string) {
	if s.err != nil {
		return
	}
	_, s.err = io.WriteString(s.w, str)
}

// Printf writes a formatted string, recording the first error.
func (s *stickyWriter) Printf(format string, args ...any) {
	if s.err != nil {
		return
	}
	_, s.err = fmt.Fprintf(s.w, format, args...)
}

// writeFileAtomic writes content via a temp file in the destination directory and
// renames it into place. A failed or partial write therefore never truncates or
// corrupts an existing target file (e.g. the user's current config on disk-full),
// PROVIDED the write callback returns a non-nil error when a write fails.
//
// The temp file is created with 0600 (os.CreateTemp's mode), so exported config
// files are owner-only — intentionally stricter than the previous 0644. On a
// successful run the temp file is renamed into place; on error it is removed. A
// hard kill (SIGKILL) between creation and rename can leave a stale
// .ghreadme-export-*.tmp in the destination directory; this is harmless (it never
// contains secrets — sanitizeConfig strips them) and is overwritten/cleaned on
// the next successful export.
func writeFileAtomic(outputPath string, write func(*os.File) error) (retErr error) {
	tmp, err := os.CreateTemp(filepath.Dir(outputPath), ".ghreadme-export-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		if retErr != nil {
			_ = os.Remove(tmpName) // best-effort cleanup of the abandoned temp file
		}
	}()

	if err := write(tmp); err != nil {
		_ = tmp.Close()

		return err
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to flush temp file: %w", err)
	}
	if err := os.Rename(tmpName, outputPath); err != nil {
		return fmt.Errorf("failed to finalize export: %w", err)
	}

	return nil
}

// exportYAML exports configuration as YAML.
func (e *ConfigExporter) exportYAML(config *internal.AppConfig, outputPath string) error {
	// Create a clean config without sensitive data for export
	exportConfig := e.sanitizeConfig(config)

	if err := writeFileAtomic(outputPath, func(file *os.File) error {
		// Route header and encoder through a sticky writer so a failed write
		// (disk-full, quota) is reported instead of silently producing a
		// truncated file that the atomic rename would commit. goccy/go-yaml's
		// Encode discards the underlying write error, so checking sw.err after
		// Encode is what actually surfaces an I/O failure here.
		sw := &stickyWriter{w: file}
		sw.WriteString(appconstants.MsgConfigHeader)
		sw.WriteString(appconstants.MsgConfigWizardHeader)
		if sw.err != nil {
			return fmt.Errorf("failed to write config header: %w", sw.err)
		}

		encoder := yaml.NewEncoder(sw, yaml.Indent(2))
		if err := encoder.Encode(exportConfig); err != nil {
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
		if sw.err != nil {
			return fmt.Errorf("failed to write YAML: %w", sw.err)
		}

		return nil
	}); err != nil {
		return err
	}

	e.output.Success(appconstants.MsgConfigurationExportedTo, outputPath)

	return nil
}

// exportJSON exports configuration as JSON.
func (e *ConfigExporter) exportJSON(config *internal.AppConfig, outputPath string) error {
	// Create a clean config without sensitive data for export
	exportConfig := e.sanitizeConfig(config)

	if err := writeFileAtomic(outputPath, func(file *os.File) error {
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(exportConfig); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	e.output.Success(appconstants.MsgConfigurationExportedTo, outputPath)

	return nil
}

// exportTOML exports configuration as TOML.
func (e *ConfigExporter) exportTOML(config *internal.AppConfig, outputPath string) error {
	// For now, we'll use a basic TOML export since the TOML library adds dependencies
	// In a full implementation, you would use "github.com/BurntSushi/toml"
	exportConfig := e.sanitizeConfig(config)

	if err := writeFileAtomic(outputPath, func(file *os.File) error {
		// Route all writes through a sticky writer so an I/O failure (disk-full,
		// quota) is reported instead of silently producing a truncated file that
		// the atomic rename would commit over the user's valid config.
		sw := &stickyWriter{w: file}

		// Write TOML header
		sw.WriteString(appconstants.MsgConfigHeader)
		sw.WriteString(appconstants.MsgConfigWizardHeader)

		// Basic TOML export (simplified version)
		e.writeTOMLConfig(sw, exportConfig)

		if sw.err != nil {
			return fmt.Errorf("failed to write TOML: %w", sw.err)
		}

		return nil
	}); err != nil {
		return err
	}

	e.output.Success(appconstants.MsgConfigurationExportedTo, outputPath)

	return nil
}

// sanitizeConfig removes sensitive information from config for export.
func (e *ConfigExporter) sanitizeConfig(config *internal.AppConfig) *internal.AppConfig {
	// Create a copy of the config
	sanitized := *config

	// Remove sensitive information
	sanitized.GitHubToken = ""    // Never export tokens
	sanitized.RepoOverrides = nil // Don't export repo overrides

	// Remove legacy fields if they match defaults
	defaults := internal.DefaultAppConfig()
	if sanitized.Template == defaults.Template {
		sanitized.Template = ""
	}
	if sanitized.Header == defaults.Header {
		sanitized.Header = ""
	}
	if sanitized.Footer == defaults.Footer {
		sanitized.Footer = ""
	}
	if sanitized.Schema == defaults.Schema {
		sanitized.Schema = ""
	}

	return &sanitized
}

// writeTOMLConfig writes a basic TOML configuration. All writes go through the
// sticky writer, so the caller checks sw.err once to detect any I/O failure.
func (e *ConfigExporter) writeTOMLConfig(sw *stickyWriter, config *internal.AppConfig) {
	e.writeRepositorySection(sw, config)
	e.writeTemplateSection(sw, config)
	e.writeFeaturesSection(sw, config)
	e.writeBehaviorSection(sw, config)
	e.writeWorkflowSection(sw, config)
	e.writePermissionsSection(sw, config)
	e.writeVariablesSection(sw, config)
}

// writeRepositorySection writes the repository information section.
func (e *ConfigExporter) writeRepositorySection(sw *stickyWriter, config *internal.AppConfig) {
	sw.Printf("# Repository Information\n")
	if config.Organization != "" {
		sw.Printf("organization = %q\n", config.Organization)
	}
	if config.Repository != "" {
		sw.Printf("repository = %q\n", config.Repository)
	}
	if config.Version != "" {
		sw.Printf("version = %q\n", config.Version)
	}
}

// writeTemplateSection writes the template settings section.
func (e *ConfigExporter) writeTemplateSection(sw *stickyWriter, config *internal.AppConfig) {
	sw.Printf("\n# Template Settings\n")
	sw.Printf("theme = %q\n", config.Theme)
	sw.Printf("output_format = %q\n", config.OutputFormat)
	sw.Printf("output_dir = %q\n", config.OutputDir)
}

// writeFeaturesSection writes the features section.
func (e *ConfigExporter) writeFeaturesSection(sw *stickyWriter, config *internal.AppConfig) {
	sw.Printf("\n# Features\n")
	sw.Printf("analyze_dependencies = %t\n", config.AnalyzeDependencies)
	sw.Printf("show_security_info = %t\n", config.ShowSecurityInfo)
}

// writeBehaviorSection writes the behavior section.
func (e *ConfigExporter) writeBehaviorSection(sw *stickyWriter, config *internal.AppConfig) {
	sw.Printf("\n# Behavior\n")
	sw.Printf("verbose = %t\n", config.Verbose)
	sw.Printf("quiet = %t\n", config.Quiet)
}

// writeWorkflowSection writes the workflow requirements section.
func (e *ConfigExporter) writeWorkflowSection(sw *stickyWriter, config *internal.AppConfig) {
	if len(config.RunsOn) == 0 {
		return
	}

	sw.Printf("\n# Workflow Requirements\n")
	sw.Printf("runs_on = [")
	for i, runner := range config.RunsOn {
		if i > 0 {
			sw.Printf(", ")
		}
		sw.Printf("%q", runner)
	}
	sw.Printf("]\n")
}

// writePermissionsSection writes the permissions section.
func (e *ConfigExporter) writePermissionsSection(sw *stickyWriter, config *internal.AppConfig) {
	e.writeMapSection(sw, "[permissions]", config.Permissions)
}

// writeVariablesSection writes the variables section.
func (e *ConfigExporter) writeVariablesSection(sw *stickyWriter, config *internal.AppConfig) {
	e.writeMapSection(sw, "[variables]", config.Variables)
}

// writeMapSection writes a TOML section with key-value pairs from a map. Keys are
// sorted so the export is deterministic (Go map iteration order is randomized,
// which would otherwise produce spurious diffs between identical exports).
func (e *ConfigExporter) writeMapSection(sw *stickyWriter, sectionName string, data map[string]string) {
	if len(data) == 0 {
		return
	}

	sw.Printf("\n%s\n", sectionName)
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		sw.Printf(appconstants.FormatEnvVar, key, data[key])
	}
}
