package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/cmd"
)

var (
	version  = "0.1.0"
	verbose  bool
	logLevel string
)

func main() {
	os.Exit(run())
}

func run() int {
	rootCmd := &cobra.Command{
		Use:   "gh-action-readme",
		Short: "Auto-generate beautiful README and HTML documentation for GitHub Actions.",
		Long: `gh-action-readme is a CLI tool for parsing one or many action.yml files and
generating informative, modern, and customizable documentation.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// --log-level takes precedence, then --verbose, then default to info
			switch {
			case logLevel != "":
				level, err := logrus.ParseLevel(logLevel)
				if err != nil {
					logrus.Warnf("Invalid log level '%s', defaulting to info", logLevel)
					logrus.SetLevel(logrus.InfoLevel)
				} else {
					logrus.SetLevel(level)
				}
			case verbose:
				logrus.SetLevel(logrus.DebugLevel)
			default:
				logrus.SetLevel(logrus.InfoLevel)
			}
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVar(
		&logLevel,
		"log-level",
		"",
		"Set log level: debug, info, warn, error (default: info)",
	)

	rootCmd.AddCommand(cmd.GenCmd())
	rootCmd.AddCommand(cmd.ValidateCmd())
	rootCmd.AddCommand(cmd.SchemaCmd())
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the version number",
			Run: func(cmd *cobra.Command, args []string) {
				logrus.Infof("gh-action-readme version %s", version)
			},
		},
	)
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "about",
			Short: "About this tool",
			Run: func(cmd *cobra.Command, args []string) {
				logrus.Info(
					"gh-action-readme: Generates README.md and " +
						"HTML for GitHub Actions. MIT License.",
				)
			},
		},
	)

	if err := rootCmd.Execute(); err != nil {
		logrus.Errorf("Error: %v", err)

		return 1
	}

	return 0
}
