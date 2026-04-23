// Package main provides the validate and schema commands.
package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/helpers"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate action.yml files.",
		Run:   wrapHandlerWithErrorHandling(validateHandler),
	}
}

func newSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "schema",
		Short: "Show the action.yml schema info.",
		Run:   schemaHandler,
	}
}

func validateHandler(_ *cobra.Command, _ []string) error {
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return fmt.Errorf("unable to determine current directory: %w", err)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(
		currentDir,
		true,
		globalConfig.IgnoredDirectories,
		"validation",
	) // Recursive for validation
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToDiscoverActionFiles, err)
	}

	// Validate the discovered files
	if err := generator.ValidateFiles(actionFiles); err != nil {
		return err
	}

	generator.Output.Success("\nAll validations passed successfully!")

	return nil
}

func schemaHandler(_ *cobra.Command, _ []string) {
	output := internal.NewColoredOutput(globalConfig.Quiet)
	output.Printf("Schema: %s (replaceable, editable)\n", globalConfig.Schema)
}
