package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// newVersionCmd builds the `version` command: the bare version by default, and
// the full ldflags-injected build metadata under --verbose.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   appconstants.CommandVersion,
		Short: "Print the version number",
		Long:  "Print the version number and build information",
		Run: func(cmd *cobra.Command, _ []string) {
			isVerbose, _ := cmd.Flags().GetBool(appconstants.ConfigKeyVerbose)
			if isVerbose {
				fmt.Printf("gh-action-readme version %s\n", version)
				fmt.Printf("  commit: %s\n", commit)
				fmt.Printf("  built at: %s\n", date)
				fmt.Printf("  built by: %s\n", builtBy)
			} else {
				fmt.Println(version)
			}
		},
	}
}

// newAboutCmd builds the `about` command.
func newAboutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "about",
		Short: "About this tool",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("gh-action-readme: Generates README.md and HTML for GitHub Actions. MIT License.")
		},
	}
}
