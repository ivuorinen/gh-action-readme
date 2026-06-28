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

func configRootHandler(_ *cobra.Command, _ []string) error {
	output := internal.NewColoredOutput(globalConfig.Quiet)
	path, err := internal.GetConfigPath()
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToGetConfigPath, err)
	}
	output.Info("Configuration file location: %s", path)
	if globalConfig.Verbose {
		output.Info("Current config: %+v", redactConfigForLog(globalConfig))
	}

	return nil
}

// redactConfigForLog returns a copy of config with every GitHub token (the
// top-level field and each nested repo override) replaced by a placeholder, so
// the result is safe to dump via %+v without leaking secrets into terminals, CI
// logs, or screen shares. Returns nil unchanged. Shared by `config show` and
// `gen --verbose`; any future config dump must route through here too.
func redactConfigForLog(config *internal.AppConfig) *internal.AppConfig {
	if config == nil {
		return nil
	}
	redacted := *config
	if redacted.GitHubToken != "" {
		redacted.GitHubToken = appconstants.RedactedPlaceholder
	}
	// RepoOverrides is a map of AppConfig, each with its own GitHubToken. The
	// shallow copy above aliases the same map, and %+v recurses into it, so a
	// token configured under a repo override would print in plaintext. Deep-copy
	// the map and redact each nested token.
	if len(redacted.RepoOverrides) > 0 {
		overrides := make(map[string]internal.AppConfig, len(redacted.RepoOverrides))
		for name, override := range redacted.RepoOverrides {
			if override.GitHubToken != "" {
				override.GitHubToken = appconstants.RedactedPlaceholder
			}
			overrides[name] = override
		}
		redacted.RepoOverrides = overrides
	}

	return &redacted
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   appconstants.CommandConfig,
		Short: "Configuration management commands",
		Run:   wrapHandlerWithErrorHandling(configRootHandler),
	}

	// Add subcommands
	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize default configuration file",
		Run:   wrapHandlerWithErrorHandling(configInitHandler),
	})

	initCmd := &cobra.Command{
		Use:   appconstants.CommandWizard,
		Short: "Interactive configuration wizard",
		Long:  "Launch an interactive wizard to set up your configuration step by step",
		Run:   wrapHandlerWithErrorHandling(configWizardHandler),
	}
	initCmd.Flags().String(appconstants.FlagFormat, "yaml", "Export format: yaml, json, toml")
	initCmd.Flags().String(appconstants.FlagOutput, "", "Output path (default: XDG config directory)")
	cmd.AddCommand(initCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   appconstants.CommandShow,
		Short: "Show current configuration",
		Run:   wrapHandlerWithErrorHandling(configShowHandler),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   appconstants.CommandThemes,
		Short: "List available themes",
		Run:   wrapHandlerWithErrorHandling(configThemesHandler),
	})

	return cmd
}

func configInitHandler(_ *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)

	// Check if config already exists
	configPath, err := internal.GetConfigPath()
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToGetConfigPath, err)
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

func configShowHandler(_ *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)

	output.Bold("Current Configuration:")
	output.Printf("Theme: %s\n", globalConfig.Theme)
	output.Printf("Output Format: %s\n", globalConfig.OutputFormat)
	output.Printf("Output Directory: %s\n", globalConfig.OutputDir)
	output.Printf("Template: %s\n", globalConfig.Template)
	output.Printf("Schema: %s\n", globalConfig.Schema)
	output.Printf("Verbose: %t\n", globalConfig.Verbose)
	output.Printf("Quiet: %t\n", globalConfig.Quiet)

	return nil
}

func configThemesHandler(_ *cobra.Command, _ []string) error {
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

	return nil
}

// newConfigWizard builds the interactive wizard. It is a var so tests can inject
// a wizard reading scripted input (via wizard.NewConfigWizardWithInput) and drive
// configWizardHandler end-to-end without a real terminal.
var newConfigWizard = wizard.NewConfigWizard

func configWizardHandler(cmd *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)

	// Create and run the wizard
	configWizard := newConfigWizard(output)
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

	// Preserve repo_overrides from the existing config: the wizard writes the live
	// config but never prompts for overrides, so without this a second run would
	// wipe them. allowLiveOverwrite=true lets the wizard populate the live config
	// (the guard only blocks share-style exports).
	if config.RepoOverrides == nil && globalConfig != nil {
		config.RepoOverrides = globalConfig.RepoOverrides
	}

	exportFormat := resolveExportFormat(format)

	if err := exporter.ExportConfig(config, exportFormat, outputPath, true); err != nil {
		return fmt.Errorf("failed to export configuration: %w", err)
	}

	output.Info("\nConfiguration wizard completed successfully!")
	output.Info("You can now use 'gh-action-readme gen' to generate documentation.")

	return nil
}
