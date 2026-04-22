// Package main provides the config management commands.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/wizard"
)

var exportFormats = map[string]wizard.ExportFormat{
	appconstants.OutputFormatJSON: wizard.FormatJSON,
	appconstants.OutputFormatTOML: wizard.FormatTOML,
}

// resolveExportFormat converts a format string to wizard.ExportFormat.
func resolveExportFormat(format string) wizard.ExportFormat {
	if f, ok := exportFormats[format]; ok {
		return f
	}

	return wizard.FormatYAML
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
		Run: func(_ *cobra.Command, _ []string) {
			output := internal.NewColoredOutput(globalConfig.Quiet)
			path, err := internal.GetConfigPath()
			if err != nil {
				output.Error("Error getting config path: %v", err)

				return
			}
			output.Info("Configuration file location: %s", path)
			if globalConfig.Verbose {
				output.Info("Current config: %+v", globalConfig)
			}
		},
	}

	// Add subcommands
	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize default configuration file",
		Run:   wrapHandlerWithErrorHandling(configInitHandler),
	})

	initCmd := &cobra.Command{
		Use:   "wizard",
		Short: "Interactive configuration wizard",
		Long:  "Launch an interactive wizard to set up your configuration step by step",
		Run:   wrapHandlerWithErrorHandling(configWizardHandler),
	}
	initCmd.Flags().String(appconstants.FlagFormat, "yaml", "Export format: yaml, json, toml")
	initCmd.Flags().String(appconstants.FlagOutput, "", "Output path (default: XDG config directory)")
	cmd.AddCommand(initCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run:   configShowHandler,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "themes",
		Short: "List available themes",
		Run:   configThemesHandler,
	})

	return cmd
}

func configInitHandler(_ *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)

	// Check if config already exists
	configPath, err := internal.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		output.Warning("Configuration file already exists at: %s", configPath)
		output.Info("Use 'gh-action-readme config show' to view current configuration")

		return nil
	}

	// Create default config
	if err := internal.WriteDefaultConfig(); err != nil {
		return fmt.Errorf("failed to write default configuration: %w", err)
	}

	output.Success("Created default configuration at: %s", configPath)
	output.Info("Edit this file to customize your settings")

	return nil
}

func configShowHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)

	output.Bold("Current Configuration:")
	output.Printf("Theme: %s\n", globalConfig.Theme)
	output.Printf("Output Format: %s\n", globalConfig.OutputFormat)
	output.Printf("Output Directory: %s\n", globalConfig.OutputDir)
	output.Printf("Template: %s\n", globalConfig.Template)
	output.Printf("Schema: %s\n", globalConfig.Schema)
	output.Printf("Verbose: %t\n", globalConfig.Verbose)
	output.Printf("Quiet: %t\n", globalConfig.Quiet)
}

func configThemesHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)

	output.Bold("Available Themes:")
	themes := []struct {
		name string
		desc string
	}{
		{appconstants.ThemeDefault, "Original simple template"},
		{appconstants.ThemeGitHub, "GitHub-style with badges and collapsible sections"},
		{appconstants.ThemeGitLab, "GitLab-focused with CI/CD examples"},
		{appconstants.ThemeMinimal, "Clean and concise documentation"},
		{appconstants.ThemeProfessional, "Comprehensive with troubleshooting and ToC"},
	}

	for _, theme := range themes {
		if theme.name == globalConfig.Theme {
			output.Success("• %s - %s (current)", theme.name, theme.desc)
		} else {
			output.Printf("• %s - %s\n", theme.name, theme.desc)
		}
	}

	output.Info("\nUse --theme flag or set 'theme' in config file to change theme")
}

func configWizardHandler(cmd *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)

	// Create and run the wizard
	configWizard := wizard.NewConfigWizard(output)
	config, err := configWizard.Run()
	if err != nil {
		return fmt.Errorf("wizard failed: %w", err)
	}

	// Get export format and output path
	format, _ := cmd.Flags().GetString(appconstants.FlagFormat)
	outputPath, _ := cmd.Flags().GetString(appconstants.FlagOutput)

	// Create exporter and export configuration
	exporter := wizard.NewConfigExporter(output)

	// Use default output path if not specified
	if outputPath == "" {
		exportFormat := resolveExportFormat(format)
		defaultPath, err := exporter.GetDefaultOutputPath(exportFormat)
		if err != nil {
			return fmt.Errorf("failed to get default output path: %w", err)
		}
		outputPath = defaultPath
	}

	// Export the configuration
	exportFormat := resolveExportFormat(format)

	if err := exporter.ExportConfig(config, exportFormat, outputPath); err != nil {
		return fmt.Errorf("failed to export configuration: %w", err)
	}

	output.Info("\nConfiguration wizard completed successfully!")
	output.Info("You can now use 'gh-action-readme gen' to generate documentation.")

	return nil
}
