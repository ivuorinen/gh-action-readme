// Package main provides the validate and schema commands.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/helpers"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   appconstants.CommandValidate + " [path]",
		Short: "Validate action.yml files.",
		Args:  cobra.MaximumNArgs(1),
		Run:   wrapHandlerWithErrorHandling(validateHandler),
	}
}

func newSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   appconstants.CommandSchema,
		Short: "Show the action.yml schema info.",
		Run:   wrapHandlerWithErrorHandling(schemaHandler),
	}
}

func validateHandler(_ *cobra.Command, args []string) error {
	target, err := validateTargetDir(args)
	if err != nil {
		return err
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(
		target,
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

// validateTargetDir resolves the directory to validate: the first positional
// argument when given (which must exist), otherwise the current directory. A
// nonexistent path is an error rather than a silent fall-through to the CWD.
func validateTargetDir(args []string) (string, error) {
	if len(args) == 0 {
		currentDir, err := helpers.GetCurrentDir()
		if err != nil {
			return "", fmt.Errorf("unable to determine current directory: %w", err)
		}

		return currentDir, nil
	}

	target := args[0]
	if _, err := os.Stat(target); err != nil {
		return "", fmt.Errorf("cannot validate %q: %w", target, err)
	}

	return target, nil
}

func schemaHandler(_ *cobra.Command, _ []string) error {
	output := internal.NewColoredOutput(globalConfig.Quiet)
	if globalConfig.Schema == "" {
		output.Printf("Schema: (not configured — set 'schema' in .ghreadme.yaml)\n")

		return nil
	}
	output.Printf("Schema: %s (replaceable, editable)\n", globalConfig.Schema)

	return nil
}
