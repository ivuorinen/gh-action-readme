// Package cmd provides the "gen" CLI subcommand for gh-action-readme.
package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/internal"
)

// GenCmd returns the cobra.Command for the "gen" subcommand.
func GenCmd() *cobra.Command {
	var (
		formatsStr      string
		org             string
		configPath      string
		outputDir       string
		versionOverride string
		mdOutputName    string
		htmlOutputName  string
		relaxedMode     bool
		stopOnErrors    bool
		dryRun          bool
	)
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate README.md and/or HTML for all action.yml files.",
		Long: `Generate documentation for all found action.yml files.

Supports multiple formats (Markdown, HTML) in one run.

Usage examples:

  # Generate Markdown README for all actions (uses org from config)
  gh-action-readme gen --config config.yaml

  # Generate both Markdown and HTML docs
  gh-action-readme gen --format=md,html

  # Override GitHub org and output directory
  gh-action-readme gen --org my-org --output-dir docs/

  # Specify custom output filenames
  gh-action-readme gen --md-output=README.md --html-output=index.html

  # Override action version for badge/examples
  gh-action-readme gen --version v2

  # Use custom header/footer templates
  gh-action-readme gen --format html \
    --header templates/header.html.tmpl \
    --footer templates/footer.html.tmpl

  # Run in CI pipeline (non-interactive)
  gh-action-readme gen --autofill-missing --org myorg

For more, see README.md or run with --help.

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenCommand(
				formatsStr, org, configPath, outputDir, versionOverride,
				mdOutputName, htmlOutputName, relaxedMode, stopOnErrors, dryRun,
			)
		},
	}
	addGenFlags(
		cmd, &formatsStr, &org, &configPath, &outputDir, &versionOverride,
		&mdOutputName, &htmlOutputName, &relaxedMode, &stopOnErrors, &dryRun,
	)

	return cmd
}

// addGenFlags adds all flags for the gen command.
func addGenFlags(
	cmd *cobra.Command,
	formatsStr, org, configPath, outputDir, versionOverride,
	mdOutputName, htmlOutputName *string,
	relaxedMode, stopOnErrors, dryRun *bool,
) {
	cmd.Flags().StringVar(
		formatsStr, "format", "md",
		"Output format(s): md,html or comma-separated (default: md)",
	)
	cmd.Flags().StringVar(
		org, "org", "",
		"GitHub org/user (overrides config)",
	)
	cmd.Flags().StringVar(
		configPath, "config", "config.yaml",
		"Path to config.yaml",
	)
	cmd.Flags().StringVar(
		outputDir, "output-dir", "",
		"Output directory for docs (defaults to action.yml dir)",
	)
	cmd.Flags().StringVar(
		versionOverride, "version", "",
		"GitHub Action version tag or branch (overrides config)",
	)
	cmd.Flags().StringVar(
		mdOutputName, "md-output", "",
		"Output filename for Markdown (default: README.md)",
	)
	cmd.Flags().StringVar(
		htmlOutputName, "html-output", "",
		"Output filename for HTML (default: README.html)",
	)
	cmd.Flags().BoolVar(
		relaxedMode, "relaxed", false,
		"Allow unknown fields in action.yml (relaxed schema validation). Strict mode is default.",
	)
	cmd.Flags().BoolVar(
		stopOnErrors, "stop-on-errors", false,
		"Stop processing on first error (default: process all actions and report all errors)",
	)
	cmd.Flags().BoolVar(
		dryRun, "dry-run", false,
		"Render documentation but do not write any files (for preview/testing)",
	)
}

// runGenCommand orchestrates the gen command logic.
func runGenCommand(
	formatsStr, org, configPath, outputDir, versionOverride,
	mdOutputName, htmlOutputName string,
	relaxedMode, stopOnErrors, dryRun bool,
) error {
	cfg, cfgErr := getConfig(configPath)
	if cfgErr != nil {
		logrus.Errorf("Failed to load config: %v", cfgErr)

		return cfgErr
	}
	orgVal := getOrg(cfg, org)
	ver := getVersion(cfg, versionOverride)
	formats := extractFormats(formatsStr)
	actionFiles := findActionYMLFiles(".")

	results := runGenWorkers(
		actionFiles, formats, cfg, orgVal, ver,
		mdOutputName, htmlOutputName, outputDir, relaxedMode, stopOnErrors, dryRun,
	)

	return reportGenSummary(results, dryRun)
}

// runGenWorkers processes actions concurrently and returns results.
func runGenWorkers(
	actionFiles []string,
	formats []string,
	cfg *internal.Config,
	orgVal, ver string,
	mdOutputName, htmlOutputName, outputDir string,
	relaxedMode, stopOnErrors, dryRun bool,
) []result {
	maxWorkers := 4
	resultsCh := make(chan result, len(actionFiles))
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	abortCh := make(chan struct{})
	var aborted bool

	for _, actionPath := range actionFiles {
		wg.Add(1)
		go func(actionPath string) {
			defer wg.Done()
			runGenWorkerTask(
				actionPath, formats, cfg, orgVal, ver,
				mdOutputName, htmlOutputName, outputDir,
				relaxedMode, stopOnErrors, dryRun,
				resultsCh, sem, abortCh, &aborted,
			)
		}(actionPath)
		if aborted {
			break
		}
	}
	wg.Wait()
	close(resultsCh)

	results := make([]result, 0, len(actionFiles))
	for res := range resultsCh {
		results = append(results, res)
	}

	return results
}

// runGenWorkerTask handles the logic for a single action file in a worker.
func runGenWorkerTask(
	actionPath string,
	formats []string,
	cfg *internal.Config,
	orgVal, ver string,
	mdOutputName, htmlOutputName, outputDir string,
	relaxedMode, stopOnErrors, dryRun bool,
	resultsCh chan<- result,
	sem chan struct{},
	abortCh chan struct{},
	aborted *bool,
) {
	select {
	case sem <- struct{}{}:
	case <-abortCh:
		return
	}
	var errs []string
	action, parseErr := internal.ParseActionYML(actionPath)
	if parseErr != nil {
		logrus.Errorf("Could not parse %s: %v", actionPath, parseErr)
		errs = append(errs, "parse error")
		resultsCh <- result{actionPath, errs}
		if stopOnErrors {
			*aborted = true
			close(abortCh)
		}
		<-sem

		return
	}
	// Strict mode: check for unknown fields unless relaxedMode is set
	if !relaxedMode {
		schemaPath := cfg.Schema
		if schemaPath == "" {
			schemaPath = "schemas/action.schema.json"
		}
		schemaErrs, schemaErr := internal.ValidateActionYMLSchema(
			actionPath,
			schemaPath,
		)
		if schemaErr != nil {
			logrus.Errorf(
				"Schema validation error for %s: %v",
				actionPath,
				schemaErr,
			)
			errs = append(errs, "schema validation error")
			resultsCh <- result{actionPath, errs}
			if stopOnErrors {
				*aborted = true
				close(abortCh)
			}
			<-sem

			return
		}
		if len(schemaErrs) > 0 {
			logrus.Errorf(
				"Schema validation failed for %s: %v",
				actionPath,
				schemaErrs,
			)
			errs = append(errs, "schema validation failed")
			resultsCh <- result{actionPath, errs}
			if stopOnErrors {
				*aborted = true
				close(abortCh)
			}
			<-sem

			return
		}
	}
	if dryRun {
		processGenDryRun(
			actionPath, action, formats, cfg, orgVal, ver,
			mdOutputName, htmlOutputName, errs,
		)
	} else {
		if genErr := generateDocsForAction(
			generateDocsOptions{
				actionPath:     actionPath,
				action:         action,
				formats:        formats,
				cfg:            cfg,
				org:            orgVal,
				ver:            ver,
				mdOutputName:   mdOutputName,
				htmlOutputName: htmlOutputName,
				outputDir:      outputDir,
				dryRun:         dryRun,
			},
		); genErr != nil {
			logrus.Errorf("Error generating docs for %s: %v", actionPath, genErr)
			errs = append(errs, "doc generation error")
			resultsCh <- result{actionPath, errs}
			if stopOnErrors {
				*aborted = true
				close(abortCh)
			}
			<-sem

			return
		}
	}
	resultsCh <- result{actionPath, errs}
	<-sem
}

// processGenDryRun handles dry-run rendering for gen command.
func processGenDryRun(
	actionPath string,
	action *internal.ActionYML,
	formats []string,
	cfg *internal.Config,
	orgVal, ver, mdOutputName, htmlOutputName string,
	errs []string,
) {
	projectRoot, _ := os.Getwd()
	repoRel, _ := filepath.Rel(projectRoot, filepath.Dir(actionPath))
	repoRel = filepath.ToSlash(repoRel)
	for _, format := range formats {
		tmplOpts := internal.TemplateOptions{
			TemplateContent: cfg.Template,
			HeaderBase:      cfg.Header,
			FooterBase:      cfg.Footer,
			Format:          format,
			Org:             orgVal,
			Repo:            repoRel,
			Version:         ver,
		}
		if action == nil {
			logrus.Errorf("Dry-run: Skipping %s because action is nil", actionPath)
			errs = append(errs, "action is nil")

			continue
		}
		out, renderErr := internal.RenderReadme(action, tmplOpts)

		if renderErr != nil {
			logrus.Errorf(
				"Dry-run: Failed to render for %s (%s): %v",
				actionPath,
				format,
				renderErr,
			)
			errs = append(errs, renderErr.Error())

			continue
		}
		outName := resolveOutputFilename(format, mdOutputName, htmlOutputName)
		if outName != "" && ver != "" {
			outName = strings.Replace(outName, "{version}", ver, 1)
		}
		logrus.Infof(
			"[dry-run] Would generate documentation for %s (%s):\n%s",
			actionPath, outName, out,
		)
	}
}

// reportGenSummary prints a summary and returns error if needed.
func reportGenSummary(results []result, dryRun bool) error {
	anyError := false
	actionsProcessed := 0
	actionsWithErrors := 0
	filesWritten := 0
	filesDryRun := 0
	for _, res := range results {
		actionsProcessed++
		if len(res.errs) > 0 {
			actionsWithErrors++
			anyError = true
		} else {
			if dryRun {
				filesDryRun++
			} else {
				filesWritten++
			}
		}
	}
	if dryRun {
		logrus.Infof(
			"Summary: %d actions processed, %d successful, %d errors, "+
				"%d files would be written (dry-run).",
			actionsProcessed,
			actionsProcessed-actionsWithErrors,
			actionsWithErrors,
			filesDryRun,
		)
	} else {
		logrus.Infof(
			"Summary: %d actions processed, %d successful, %d errors, %d files written.",
			actionsProcessed,
			actionsProcessed-actionsWithErrors,
			actionsWithErrors,
			filesWritten,
		)
	}
	if anyError || actionsWithErrors > 0 {
		return errors.New("one or more documentation generation errors occurred")
	}

	return nil
}
