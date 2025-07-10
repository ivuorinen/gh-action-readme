// Package internal provides core logic for gh-action-readme, including template rendering.
package internal

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"text/template"

	"github.com/sirupsen/logrus"
)

// TemplateOptions holds options for rendering documentation templates.
type TemplateOptions struct {
	TemplateBase string          // Path base for main template (e.g. templates/readme)
	HeaderBase   string          // Path base for header template (e.g. templates/header)
	FooterBase   string          // Path base for footer template (e.g. templates/footer)
	Format       string          // Output format ("md" or "html")
	Org          string          // GitHub org/user name
	Repo         string          // Repository/folder (relative path for uses)
	Version      string          // Action version/tag/branch
	Config       *TemplateConfig // Configuration with template paths
}

// TemplateConfig holds custom template paths for rendering.
type TemplateConfig struct {
	MainTemplatePath string // Custom path for the main template
	HTMLTemplatePath string // Custom path for the HTML template
}

// Template context documentation:
// The entire ActionYML struct is passed as the template context.
// Available fields in templates:
//   .Name         - Action name (string)
//   .Description  - Action description (string)
//   .Inputs       - map[string]ActionInput
//   .Outputs      - map[string]ActionOutput
//   .Runs         - map[string]any
//   .Branding     - *Branding
//   .Org          - GitHub org/user (from TemplateOptions)
//   .Repo         - repo/folder (from TemplateOptions)
//   .Version      - Action version/tag/branch (from TemplateOptions)
// You can access all ActionYML fields directly, and Org/Repo/Version as top-level fields.

// resolveTemplatePath returns the template file path for a given base and format.
func resolveTemplatePath(base, format string) string {
	return fmt.Sprintf("%s.%s.tmpl", base, format)
}

// RenderReadme renders the documentation for a GitHub Action using the provided template options.
//
// The template context includes all fields from the ActionYML struct, as well as Org, Repo, and
// Version from TemplateOptions as top-level fields.
//
// Returns the rendered documentation as a string, or an error.
func RenderReadme(action any, opts TemplateOptions) (string, error) {
	var templatePath string
	switch {
	case opts.Format == "html" && opts.Config.HTMLTemplatePath != "":
		templatePath = opts.Config.HTMLTemplatePath
	case opts.Format == "md" && opts.Config.MainTemplatePath != "":
		templatePath = opts.Config.MainTemplatePath
	case opts.Format == "html":
		templatePath = "templates/readme.html.tmpl"
	case opts.Format == "md":
		templatePath = "templates/readme.md.tmpl"
	default:
		return "", fmt.Errorf("unsupported format: %s", opts.Format)
	}

	// Load and parse the template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Render the template
	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, action)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return rendered.String(), nil
}

// RenderReadmeWithFuncs renders the documentation for a GitHub Action using the provided template
// options and custom template functions.
// funcs: map[string]interface{} of custom template functions to register (maybe nil).
func RenderReadmeWithFuncs(
	action any,
	opts TemplateOptions,
	funcs template.FuncMap,
) (string, error) {
	mainPath := resolveTemplatePath(opts.TemplateBase, opts.Format)
	headerPath := resolveTemplatePath(opts.HeaderBase, opts.Format)
	footerPath := resolveTemplatePath(opts.FooterBase, opts.Format)

	headerPath = filepath.Clean(headerPath)
	footerPath = filepath.Clean(footerPath)
	mainPath = filepath.Clean(mainPath)

	header, hErr := readTemplateFile(headerPath)
	footer, fErr := readTemplateFile(footerPath)
	if hErr != nil && !os.IsNotExist(hErr) {
		logrus.Warnf("Failed to read header file: %s", headerPath)
	}
	if os.IsNotExist(hErr) {
		logrus.Warnf("Failed to read header file: %s", headerPath)
		header = nil
	}
	if fErr != nil && !os.IsNotExist(fErr) {
		logrus.Warnf("Failed to load footer template %q: %v\n", footerPath, fErr)
	}
	if os.IsNotExist(fErr) {
		logrus.Warnf("Footer template %q not found, skipping footer.", footerPath)
		footer = nil
	}

	mainTmpl, err := os.ReadFile(mainPath) // #nosec G304
	if err != nil {
		return "", fmt.Errorf("main template %q not found: %v", mainPath, err)
	}
	tmpl := template.New(filepath.Base(mainPath))
	if funcs != nil {
		tmpl = tmpl.Funcs(funcs)
	}
	tmpl, parseErr := tmpl.Parse(string(mainTmpl))
	if parseErr != nil {
		return "", fmt.Errorf("parse error in main template %q: %v", mainPath, parseErr)
	}

	ctx := buildTemplateContext(action, opts)

	buf := &bytes.Buffer{}
	if len(header) > 0 {
		buf.Write(header)
	}

	if execErr := tmpl.Execute(buf, ctx); execErr != nil {
		return "", fmt.Errorf("template execution failed: %v", execErr)
	}

	if len(footer) > 0 {
		buf.Write(footer)
	}

	return buf.String(), nil
}

// readTemplateFile reads a template file from disk. The path should be validated by the caller.
func readTemplateFile(path string) ([]byte, error) {
	// #nosec G304 -- Path is controlled by config and validated by the caller.
	data, err := os.ReadFile(path)

	return data, err
}

func buildTemplateContext(action any, opts TemplateOptions) map[string]any {
	ctx := make(map[string]any)
	v := reflect.ValueOf(action)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			ctx[t.Field(i).Name] = v.Field(i).Interface()
		}
	}

	ctx["Org"] = opts.Org
	ctx["Repo"] = opts.Repo
	ctx["Version"] = opts.Version

	return ctx
}
