// Package cmd provides CLI subcommands and helper functions for gh-action-readme.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/ivuorinen/gh-action-readme/internal"
)

// splitAndTrim splits a string by sep, trims whitespace from each part, and omits empty results.
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}

	return out
}

// getConfig loads the config from the given path, or uses the default if empty.
func getConfig(configPath string) (*internal.Config, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}

	return internal.LoadConfig(configPath)
}

// getOrg returns the org from CLI if set, otherwise from config.
func getOrg(cfg *internal.Config, org string) string {
	if org == "" {
		return cfg.GitHubOrg
	}

	return org
}

// getVersion returns the version from CLI if set, otherwise from config, or "main" as fallback.
func getVersion(cfg *internal.Config, versionOverride string) string {
	ver := versionOverride
	if ver == "" {
		ver = cfg.DefaultValues.Version
		if ver == "" {
			ver = "main"
		}
	}

	return ver
}

// extractFormats parses the formats string and returns a slice of valid output formats.
func extractFormats(formatsStr string) []string {
	var formats []string
	for _, f := range splitAndTrim(formatsStr, ",") {
		switch f {
		case "md", "markdown":
			formats = append(formats, "md")
		case "html":
			formats = append(formats, "html")
		}
	}
	if len(formats) == 0 {
		formats = append(formats, "md")
	}

	return formats
}

type generateDocsOptions struct {
	actionPath     string
	action         *internal.ActionYML
	formats        []string
	cfg            *internal.Config
	org            string
	ver            string
	mdOutputName   string
	htmlOutputName string
	outputDir      string
	dryRun         bool
}

// generateDocsForAction renders and writes documentation for a given action in all requested
// formats. If dryRun is true, it only logs what would be written and does not write files.
// Returns an error if any rendering or writing fails.
func generateDocsForAction(opts generateDocsOptions) error {
	projectRoot, _ := os.Getwd()
	repoRel, _ := filepath.Rel(projectRoot, filepath.Dir(opts.actionPath))
	repoRel = filepath.ToSlash(repoRel)
	var errs []string
	for _, format := range opts.formats {
		tmplOpts := internal.TemplateOptions{
			TemplateContent: opts.cfg.Template,
			HeaderBase:      opts.cfg.Header,
			FooterBase:      opts.cfg.Footer,
			Format:          format,
			Org:             opts.org,
			Repo:            repoRel,
			Version:         opts.ver,
		}
		out, renderErr := internal.RenderReadme(opts.action, tmplOpts)
		if renderErr != nil {
			logrus.Errorf(
				"Failed to render for %s (%s): %v",
				opts.actionPath, format, renderErr,
			)
			errs = append(errs, renderErr.Error())

			continue
		}
		// Determine output filename per format
		outName := resolveOutputFilename(format, opts.mdOutputName, opts.htmlOutputName)
		// Replace {version} in output name if present
		if outName != "" && opts.ver != "" {
			outName = strings.Replace(outName, "{version}", opts.ver, 1)
		}
		if opts.dryRun {
			logrus.Infof(
				"[dry-run] Would write documentation for %s (%s):\n%s",
				opts.actionPath,
				outName,
				out,
			)

			continue
		}
		if err := writeDocForActionWithNameErr(
			opts.actionPath,
			out,
			outName,
			opts.outputDir,
		); err != nil {
			logrus.Errorf(
				"Failed to write documentation for %s (%s): %v",
				opts.actionPath, outName, err,
			)
			errs = append(errs, err.Error())

			continue
		}
		logrus.Infof("Generated documentation for %s (%s)", opts.actionPath, outName)
	}
	if len(errs) > 0 {
		return fmt.Errorf(
			"documentation generation errors: %v",
			strings.Join(errs, "; "),
		)
	}

	return nil
}

// resolveOutputFilename returns the output filename for the given format and optional overrides.
func resolveOutputFilename(format, mdOutputName, htmlOutputName string) string {
	switch format {
	case "md":
		if mdOutputName != "" {
			return mdOutputName
		}

		return "README.md"
	case "html":
		if htmlOutputName != "" {
			return htmlOutputName
		}

		return "README.html"
	default:
		return ""
	}
}

// fileExists returns true if the file or directory exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// downloadURL retrieves the content at the given URL and returns it as a byte slice.
func downloadURL(url string) ([]byte, error) {
	resp, err := httpGet(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := resp.Body.Close()
		if cerr != nil {
			logrus.Errorf("Failed to close response body: %v", cerr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}

// httpGet wraps http.Get to allow testing/mocking if needed.
// #nosec G107 -- Usage is intentional and safe for CLI tool (user-controlled URLs are not
// accepted from untrusted input).
func httpGet(url string) (*http.Response, error) {
	return http.Get(url)
}

// findProjectRoot searches upwards for a marker file (.git or go.mod)
// and returns the directory path.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		gitPath := filepath.Join(dir, ".git")
		goModPath := filepath.Join(dir, "go.mod")
		if fileExists(gitPath) || fileExists(goModPath) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", errors.New("project root marker (.git or go.mod) not found")
}

// findActionYMLFiles recursively finds all action.yml or action.yaml files under root.
func findActionYMLFiles(root string) []string {
	var results []string
	_ = filepath.Walk(
		root, func(path string, info os.FileInfo, err error) error {
			if err == nil &&
				!info.IsDir() &&
				(filepath.Base(path) == "action.yml" || filepath.Base(path) == "action.yaml") {
				results = append(results, path)
			}

			return nil
		},
	)

	return results
}

// writeDocForActionWithNameErr writes the generated documentation to the output file.
// Returns an error if writing fails.
func writeDocForActionWithNameErr(actionPath, doc, outName, outputDir string) error {
	dir := resolveOutputDir(actionPath, outputDir)
	cleanDir := filepath.Clean(dir)
	if err := os.MkdirAll(cleanDir, 0o750); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", cleanDir, err)
	}
	outPath := filepath.Join(cleanDir, outName)
	cleanOutPath := filepath.Clean(outPath)
	if err := os.WriteFile(cleanOutPath, []byte(doc), 0o600); err != nil {
		return fmt.Errorf("failed to write output file %s: %w", cleanOutPath, err)
	}

	return nil
}

// resolveOutputDir returns the output directory, preferring outputDir if set.
func resolveOutputDir(actionPath, outputDir string) string {
	if outputDir != "" {
		return outputDir
	}

	return filepath.Dir(actionPath)
}

// writeYAMLFile writes the ActionYML struct to the given path as YAML.
// Returns an error if writing fails.
func writeYAMLFile(path string, action *internal.ActionYML) error {
	cleanPath := filepath.Clean(path)
	f, err := os.Create(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", cleanPath, err)
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			logrus.Error(fmt.Sprintf("Failed to close file %s: %v", cleanPath, closeErr))
		}
	}()
	enc := yaml.NewEncoder(f)
	if encodeErr := enc.Encode(action); encodeErr != nil {
		return fmt.Errorf("failed to write YAML to %s: %w", cleanPath, encodeErr)
	}

	return nil
}

// fixYAMLHeader ensures the first two lines of the YAML file are:
// 1. ---
// 2. # yaml-language-server: $schema=https://json.schemastore.org/github-action.json
// If not, it inserts them and rewrites the file.
func fixYAMLHeader(path string) error {
	const (
		yamlStart    = "---"
		schemaHeader = "# yaml-language-server: $schema=" +
			"https://json.schemastore.org/github-action.json"
	)
	cleanPath := filepath.Clean(path)
	data, readErr := os.ReadFile(cleanPath)
	if readErr != nil {
		return readErr
	}
	lines := strings.Split(string(data), "\n")
	changed := false

	// Ensure first line is '---'
	if len(lines) == 0 || lines[0] != yamlStart {
		lines = append([]string{yamlStart}, lines...)
		changed = true
	}

	// Ensure second line is schema header
	if len(lines) < 2 || lines[1] != schemaHeader {
		if len(lines) == 1 {
			lines = append(lines, schemaHeader)
		} else {
			lines = append(lines[:1], append([]string{schemaHeader}, lines[1:]...)...)
		}
		changed = true
	}

	if changed {
		// Remove trailing empty lines
		for len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		// Add back a single trailing newline
		out := strings.Join(lines, "\n") + "\n"
		if writeErr := os.WriteFile(cleanPath, []byte(out), 0o600); writeErr != nil {
			return writeErr
		}
	}

	return nil
}
