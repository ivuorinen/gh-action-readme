// Package internal provides core logic for gh-action-readme, including template rendering.
package internal

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
)

// TemplateOptions holds options for rendering documentation templates.
type TemplateOptions struct {
	TemplateContent  string           // Path base for main template (e.g. templates/readme)
	HeaderBase       string           // Path base for header template (e.g. templates/header)
	FooterBase       string           // Path base for footer template (e.g. templates/footer)
	Format           string           // Output format ("md" or "html")
	Org              string           // GitHub org/user name
	Repo             string           // Repository/folder (relative path for uses)
	Version          string           // Action version/tag/branch
	HTMLTemplatePath string           // Optional HTML template path override
	Funcs            template.FuncMap // Optional custom template functions
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
	// Determine template paths with overrides.
	var mainPath string
	if opts.Format == "html" && opts.HTMLTemplatePath != "" {
		mainPath = opts.HTMLTemplatePath
	}
	if mainPath == "" {
		base := opts.TemplateContent
		if base == "" {
			base = "templates/readme"
		}
		mainPath = resolveTemplatePath(base, opts.Format)
	}
	headerPath := resolveTemplatePath(opts.HeaderBase, opts.Format)
	footerPath := resolveTemplatePath(opts.FooterBase, opts.Format)

	headerPath = filepath.Clean(headerPath)
	footerPath = filepath.Clean(footerPath)
	mainPath = filepath.Clean(mainPath)

	header, _ := readOptionalTemplate(headerPath)
	footer, _ := readOptionalTemplate(footerPath)

	mainTmpl, err := os.ReadFile(mainPath) // #nosec G304
	if err != nil {
		return "", fmt.Errorf("main template %q not found: %v", mainPath, err)
	}
	tmpl := template.New(filepath.Base(mainPath))
	if opts.Funcs != nil {
		tmpl = tmpl.Funcs(opts.Funcs)
	}
	tmpl, parseErr := tmpl.Parse(string(mainTmpl))
	if parseErr != nil {
		return "", fmt.Errorf("parse error in main template %q: %v", mainPath, parseErr)
	}

	ctx := buildTemplateContext(action, opts)

	if opts.Format == "html" {
		if desc, ok := ctx["LongDescription"].(string); ok && desc != "" {
			var buf bytes.Buffer
			if convErr := goldmark.Convert([]byte(desc), &buf); convErr == nil {
				ctx["LongDescription"] = buf.String()
			}
		}
	}

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

	result := buf.String()
	if opts.Version != "" {
		result = strings.ReplaceAll(result, "{version}", opts.Version)
	}

	return result, nil
}

// readTemplateFile reads a template file from disk. The path should be validated by the caller.
func readTemplateFile(path string) ([]byte, error) {
	// #nosec G304 -- Path is controlled by config and validated by the caller.
	data, err := os.ReadFile(path)

	return data, err
}

// readOptionalTemplate reads a template file and returns nil if it does not exist.
// A warning is logged if the file is missing or failed to read.
func readOptionalTemplate(path string) ([]byte, error) {
	data, err := readTemplateFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("Template %q not found, skipping", path)

			return nil, nil
		}
		logrus.Warnf("Failed to load template %q: %v", path, err)

		return nil, err
	}

	return data, nil
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
