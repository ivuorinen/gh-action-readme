// Package cmd provides the "validate" CLI subcommand for gh-action-readme.
package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/schemas"
)

// ValidateCmd returns the cobra.Command for the "validate" subcommand.
func ValidateCmd() *cobra.Command {
	var (
		configPath      string
		autofillMissing bool
		fixMissing      bool
		schemaPath      string
		relaxedMode     bool
		stopOnErrors    bool
		dryRun          bool
	)
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate action.yml files and optionally autofill/fix missing fields.",
		Long: `Validate all action.yml files for required fields and schema compliance.
Can autofill/fix missing fields using config defaults.

Usage examples:

  # Validate all actions using config defaults
  gh-action-readme validate --config config.yaml

  # Autofill missing fields (in-memory only)
  gh-action-readme validate --autofill-missing

  # Autofill and write missing fields back to action.yml
  gh-action-readme validate --fix-missing

  # Use a custom schema file
  gh-action-readme validate --schema=schemas/action.schema.json

  # Validate in CI pipeline (non-interactive)
  gh-action-readme validate --config config.yaml

For more, see README.md or run with --help.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidateCommand(
				&configPath, &autofillMissing, &fixMissing, &schemaPath,
				&stopOnErrors, &dryRun,
			)
		},
	}
	addValidateFlags(
		cmd, &configPath, &autofillMissing, &fixMissing, &schemaPath,
		&relaxedMode, &stopOnErrors, &dryRun,
	)

	return cmd
}

// addValidateFlags adds all flags for the validate command.
func addValidateFlags(
	cmd *cobra.Command,
	configPath *string,
	autofillMissing, fixMissing *bool,
	schemaPath *string,
	relaxedMode, stopOnErrors, dryRun *bool,
) {
	cmd.Flags().StringVar(configPath, "config", "config.yaml", "Path to config.yaml")
	cmd.Flags().BoolVar(
		autofillMissing, "autofill-missing", false,
		"Autofill missing fields using config defaults (in-memory only)",
	)
	cmd.Flags().BoolVar(
		fixMissing, "fix-missing", false,
		"Autofill and write missing fields back to action.yml",
	)
	cmd.Flags().StringVar(
		schemaPath, "schema", "",
		"Path to action.yml schema file (default: from config)",
	)
	cmd.Flags().BoolVar(
		relaxedMode, "relaxed", false,
		"Allow unknown fields in action.yml (relaxed mode, default is strict)",
	)
	cmd.Flags().BoolVar(
		stopOnErrors, "stop-on-errors", false,
		"Stop processing further actions on first error (default: process all and report errors)",
	)
	cmd.Flags().BoolVar(
		dryRun, "dry-run", false,
		"Perform validation and autofill checks, but do not write any changes to disk",
	)
}

// runValidateCommand orchestrates the validate command logic.
func runValidateCommand(
	configPath *string,
	autofillMissing, fixMissing *bool,
	schemaPath *string,
	stopOnErrors, dryRun *bool,
) error {
	cfg, err := internal.LoadConfig(*configPath)
	if err != nil {
		logrus.Errorf("Failed to load config: %v", err)

		return err
	}
	if *schemaPath == "" {
		*schemaPath = cfg.Schema
		if *schemaPath == "" {
			*schemaPath = schemas.RelPath
		}
	}
	if !filepath.IsAbs(*schemaPath) {
		if rootDir, rerr := findProjectRoot(); rerr == nil {
			*schemaPath = filepath.Join(rootDir, *schemaPath)
		}
	}
	actionFiles := findActionYMLFiles(".")
	results := runValidateWorkers(
		actionFiles, cfg, *schemaPath, *autofillMissing, *fixMissing,
		*stopOnErrors, *dryRun,
	)

	return reportValidateSummary(results, *dryRun)
}

// runValidateWorkers processes actions concurrently and returns results.
func runValidateWorkers(
	actionFiles []string,
	cfg *internal.Config,
	schemaPath string,
	autofillMissing, fixMissing, stopOnErrors, dryRun bool,
) []result {
	maxWorkers := 4
	resultsCh := make(chan result, len(actionFiles))
	sem := make(chan struct{}, maxWorkers)

	var wg sync.WaitGroup
	var stopFlag sync.Once
	stopChan := make(chan struct{})
	abort := func() { stopFlag.Do(func() { close(stopChan) }) }

	for _, actionPath := range actionFiles {
		wg.Add(1)
		go func(actionPath string) {
			defer wg.Done()
			runValidateWorkerTask(
				actionPath, cfg, schemaPath, autofillMissing, fixMissing,
				stopOnErrors, dryRun, resultsCh, sem, stopChan, abort,
			)
		}(actionPath)
	}
	wg.Wait()
	close(resultsCh)

	results := make([]result, 0, len(actionFiles))
	for res := range resultsCh {
		results = append(results, res)
	}

	return results
}

// runValidateWorkerTask handles the logic for a single action file in a worker.
func runValidateWorkerTask(
	actionPath string,
	cfg *internal.Config,
	schemaPath string,
	autofillMissing, fixMissing, stopOnErrors, dryRun bool,
	resultsCh chan<- result,
	sem chan struct{},
	stopChan <-chan struct{},
	abort func(),
) {
	select {
	case <-stopChan:
		return
	default:
	}
	sem <- struct{}{}
	var errs []string
	action, parseErr := internal.ParseActionYML(actionPath)
	if parseErr != nil {
		logrus.Errorf("Could not parse %s: %v", actionPath, parseErr)
		errs = append(
			errs,
			fmt.Sprintf(
				"Could not parse %s: %v",
				actionPath,
				parseErr,
			),
		)
		resultsCh <- result{actionPath, errs}
		<-sem
		if stopOnErrors {
			abort()
		}

		return
	}
	// Validate required fields
	valResult := internal.ValidateActionYML(action)
	if len(valResult.MissingFields) > 0 {
		msg := fmt.Sprintf(
			"Missing required fields in %s: %v",
			actionPath, valResult.MissingFields,
		)
		logrus.Warnf("%s", msg)
		errs = append(errs, msg)
		if autofillMissing || fixMissing {
			internal.FillMissing(action, cfg.DefaultValues)
			logrus.Infof(
				"Autofilled missing fields in %s using config defaults.",
				actionPath,
			)
			if fixMissing {
				if dryRun {
					logrus.Infof(
						"[dry-run] Would write autofilled YAML to %s",
						actionPath,
					)
				} else {
					// Write autofilled fields back to file
					if writeErr := writeYAMLFile(
						actionPath,
						action,
					); writeErr != nil {
						logrus.Errorf(
							"Failed to write autofilled YAML to %s: %v",
							actionPath, writeErr,
						)
						errs = append(
							errs,
							fmt.Sprintf(
								"Failed to write autofilled YAML to %s: %v",
								actionPath, writeErr,
							),
						)
					}
					// Ensure YAML header and schema comment are present
					if headerErr := fixYAMLHeader(actionPath); headerErr != nil {
						logrus.Warnf(
							"Could not fix YAML header for %s: %v",
							actionPath, headerErr,
						)
					}
					logrus.Infof("Wrote autofilled fields back to %s", actionPath)
				}
			}
		}
	} else {
		logrus.Infof("All required fields present in %s", actionPath)
	}
	// Schema validation
	schemaErrs, schemaErr := internal.ValidateActionYMLSchema(
		actionPath,
		schemaPath,
	)
	switch {
	case schemaErr != nil:
		msg := fmt.Sprintf(
			"Schema validation error for %s: %v",
			actionPath,
			schemaErr,
		)
		logrus.Errorf("%s", msg)
		errs = append(errs, msg)
	case len(schemaErrs) > 0:
		logrus.Warnf("Schema validation failed for %s:", actionPath)
		for _, se := range schemaErrs {
			logrus.Warnf("  %s", se)
			errs = append(
				errs,
				fmt.Sprintf(
					"Schema error in %s: %s",
					actionPath,
					se,
				),
			)
		}
	default:
		logrus.Infof("Schema validation passed for %s", actionPath)
	}

	resultsCh <- result{actionPath, errs}
	<-sem
	if stopOnErrors && len(errs) > 0 {
		abort()
	}
}

// reportValidateSummary prints a summary and returns error if needed.
func reportValidateSummary(results []result, dryRun bool) error {
	anyError := false
	actionsProcessed := 0
	actionsWithErrors := 0
	filesWritten := 0
	filesDryRun := 0
	var errorMessages []string
	for _, res := range results {
		actionsProcessed++
		if len(res.errs) > 0 {
			anyError = true
			actionsWithErrors++
			errorMessages = append(errorMessages, res.errs...)
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
	if anyError {
		logrus.Warn("Validation completed with errors or warnings.")

		return fmt.Errorf(
			"validation completed with errors or warnings:\n%s",
			strings.Join(errorMessages, "\n"),
		)
	}
	logrus.Info("Validation completed successfully.")

	return nil
}
