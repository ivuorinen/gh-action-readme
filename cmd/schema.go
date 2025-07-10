// Package cmd provides the "schema" CLI subcommand for gh-action-readme.
package cmd

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// SchemaCmd returns the cobra.Command for the "schema" subcommand.
func SchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Show the action.yml schema info or update to latest.",
		Long: `Show the current action.yml schema file path, or update it from the official
GitHub Actions documentation.

Examples:
  gh-action-readme schema
  gh-action-readme schema update
`,
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "update",
			Short: "Update the action.yml schema from the official GitHub Actions documentation.",
			RunE: func(cmd *cobra.Command, args []string) error {
				const schemaURL = "https://www.schemastore.org/github-action.json"
				const schemaRelPath = "schemas/action.schema.json"
				logrus.Infof("Downloading latest schema from %s ...", schemaURL)
				resp, err := downloadURL(schemaURL)
				if err != nil {
					logrus.Errorf("Failed to download schema: %v", err)

					return err
				}
				projectRoot, err := findProjectRoot()
				if err != nil {
					logrus.Errorf("Failed to find project root: %v", err)

					return err
				}
				schemasDir := filepath.Join(projectRoot, "schemas")
				mkdirErr := os.MkdirAll(schemasDir, 0o750)
				if mkdirErr != nil {
					logrus.Errorf("Failed to create schemas directory: %v", mkdirErr)

					return mkdirErr
				}
				schemaPath := filepath.Join(projectRoot, schemaRelPath)
				writeErr := os.WriteFile(schemaPath, resp, 0o600)
				if writeErr != nil {
					logrus.Errorf("Failed to write schema file: %v", writeErr)

					return writeErr
				}
				logrus.Infof("Schema updated successfully at %s", schemaPath)

				return nil
			},
		},
	)
	cmd.Run = func(cmd *cobra.Command, args []string) {
		logrus.Info("Schema: schemas/action.schema.json (replaceable, editable)")
		logrus.Info(
			"To update the schema from the official source, run: gh-action-readme schema update",
		)
	}

	return cmd
}
