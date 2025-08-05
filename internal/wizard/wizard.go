// Package wizard provides an interactive configuration wizard for gh-action-readme.
package wizard

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/internal/helpers"
)

// ConfigWizard handles interactive configuration setup.
type ConfigWizard struct {
	output    *internal.ColoredOutput
	scanner   *bufio.Scanner
	config    *internal.AppConfig
	repoInfo  *git.RepoInfo
	actionDir string
}

// NewConfigWizard creates a new configuration wizard instance.
func NewConfigWizard(output *internal.ColoredOutput) *ConfigWizard {
	return &ConfigWizard{
		output:  output,
		scanner: bufio.NewScanner(os.Stdin),
		config:  internal.DefaultAppConfig(),
	}
}

// Run executes the interactive configuration wizard.
func (w *ConfigWizard) Run() (*internal.AppConfig, error) {
	w.output.Bold("🧙 Welcome to gh-action-readme Configuration Wizard!")
	w.output.Info("This wizard will help you set up your configuration step by step.\n")

	// Step 1: Auto-detect project settings
	if err := w.detectProjectSettings(); err != nil {
		w.output.Warning("Could not auto-detect project settings: %v", err)
	}

	// Step 2: Configure basic settings
	w.configureBasicSettings()

	// Step 3: Configure template and output settings
	w.configureTemplateSettings()

	// Step 4: Configure features
	w.configureFeatures()

	// Step 5: Configure GitHub integration
	w.configureGitHubIntegration()

	// Step 6: Summary and confirmation
	if err := w.showSummaryAndConfirm(); err != nil {
		return nil, fmt.Errorf("configuration canceled: %w", err)
	}

	w.output.Success("\n✅ Configuration completed successfully!")
	return w.config, nil
}

// detectProjectSettings auto-detects project settings from the current environment.
func (w *ConfigWizard) detectProjectSettings() error {
	w.output.Bold("🔍 Step 1: Auto-detecting project settings...")

	// Detect current directory
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	w.actionDir = currentDir

	// Detect git repository
	repoRoot := helpers.FindGitRepoRoot(currentDir)
	if repoRoot != "" {
		repoInfo, err := git.DetectRepository(repoRoot)
		if err == nil {
			w.repoInfo = repoInfo
			w.config.Organization = repoInfo.Organization
			w.config.Repository = repoInfo.Repository
			w.output.Success("  📁 Repository: %s/%s", w.config.Organization, w.config.Repository)
		}
	}

	// Check for existing action files
	actionFiles := w.findActionFiles(currentDir)
	if len(actionFiles) > 0 {
		w.output.Success("  🎯 Found %d action file(s)", len(actionFiles))
	}

	return nil
}

// configureBasicSettings handles basic configuration prompts.
func (w *ConfigWizard) configureBasicSettings() {
	w.output.Bold("\n⚙️  Step 2: Basic Settings")

	// Organization
	w.config.Organization = w.promptWithDefault("Organization/Owner", w.config.Organization)

	// Repository
	w.config.Repository = w.promptWithDefault("Repository Name", w.config.Repository)

	// Version (optional)
	version := w.promptWithDefault("Version (optional)", "")
	if version != "" {
		w.config.Version = version
	}
}

// configureTemplateSettings handles template and output configuration.
func (w *ConfigWizard) configureTemplateSettings() {
	w.output.Bold("\n🎨 Step 3: Template & Output Settings")

	w.configureThemeSelection()
	w.configureOutputFormat()
	w.configureOutputDirectory()
}

// configureThemeSelection handles theme selection.
func (w *ConfigWizard) configureThemeSelection() {
	w.output.Info("Available themes:")
	themes := w.getAvailableThemes()

	w.displayThemeOptions(themes)

	themeChoice := w.promptWithDefault("Choose theme (1-5)", "1")
	if choice, err := strconv.Atoi(themeChoice); err == nil && choice >= 1 && choice <= len(themes) {
		w.config.Theme = themes[choice-1].name
	}
}

// configureOutputFormat handles output format selection.
func (w *ConfigWizard) configureOutputFormat() {
	w.output.Info("\nAvailable output formats:")
	formats := []string{"md", "html", "json", "asciidoc"}

	w.displayFormatOptions(formats)

	formatChoice := w.promptWithDefault("Choose output format (1-4)", "1")
	if choice, err := strconv.Atoi(formatChoice); err == nil && choice >= 1 && choice <= len(formats) {
		w.config.OutputFormat = formats[choice-1]
	}
}

// configureOutputDirectory handles output directory configuration.
func (w *ConfigWizard) configureOutputDirectory() {
	w.config.OutputDir = w.promptWithDefault("Output directory", w.config.OutputDir)
}

// getAvailableThemes returns the list of available themes.
func (w *ConfigWizard) getAvailableThemes() []struct {
	name string
	desc string
} {
	return []struct {
		name string
		desc string
	}{
		{"default", "Original simple template"},
		{"github", "GitHub-style with badges and collapsible sections"},
		{"gitlab", "GitLab-focused with CI/CD examples"},
		{"minimal", "Clean and concise documentation"},
		{"professional", "Comprehensive with troubleshooting and ToC"},
	}
}

// displayThemeOptions displays the theme options with current selection.
func (w *ConfigWizard) displayThemeOptions(themes []struct {
	name string
	desc string
}) {
	for i, theme := range themes {
		marker := " "
		if theme.name == w.config.Theme {
			marker = "►"
		}
		w.output.Printf("  %s %d. %s - %s", marker, i+1, theme.name, theme.desc)
	}
}

// displayFormatOptions displays the output format options with current selection.
func (w *ConfigWizard) displayFormatOptions(formats []string) {
	for i, format := range formats {
		marker := " "
		if format == w.config.OutputFormat {
			marker = "►"
		}
		w.output.Printf("  %s %d. %s", marker, i+1, format)
	}
}

// configureFeatures handles feature configuration.
func (w *ConfigWizard) configureFeatures() {
	w.output.Bold("\n🚀 Step 4: Features")

	// Dependency analysis
	w.output.Info("Dependency analysis provides detailed information about GitHub Action dependencies.")
	analyzeDeps := w.promptYesNo("Enable dependency analysis?", w.config.AnalyzeDependencies)
	w.config.AnalyzeDependencies = analyzeDeps

	// Security information
	w.output.Info("Security information shows pinned vs floating versions and security recommendations.")
	showSecurity := w.promptYesNo("Show security information?", w.config.ShowSecurityInfo)
	w.config.ShowSecurityInfo = showSecurity
}

// configureGitHubIntegration handles GitHub API configuration.
func (w *ConfigWizard) configureGitHubIntegration() {
	w.output.Bold("\n🐙 Step 5: GitHub Integration")

	// Check for existing token
	existingToken := internal.GetGitHubToken(w.config)
	if existingToken != "" {
		w.output.Success("GitHub token already configured ✓")
		return
	}

	w.output.Info("GitHub integration requires a personal access token for:")
	w.output.Printf("  • Enhanced dependency analysis")
	w.output.Printf("  • Latest version checking")
	w.output.Printf("  • Repository information")
	w.output.Printf("  • Rate limit improvements")

	setupToken := w.promptYesNo("Set up GitHub token now?", false)
	if !setupToken {
		w.output.Info("You can set up the token later using environment variables:")
		w.output.Printf("  export GITHUB_TOKEN=your_personal_access_token")
		return
	}

	w.output.Info("\nTo create a personal access token:")
	w.output.Printf("  1. Visit: https://github.com/settings/tokens")
	w.output.Printf("  2. Click 'Generate new token (classic)'")
	w.output.Printf("  3. Select scopes: 'repo' (for private repos) or 'public_repo' (for public only)")
	w.output.Printf("  4. Copy the generated token")

	token := w.promptSensitive("Enter your GitHub token (or press Enter to skip)")
	if token != "" {
		// Validate token format (basic check)
		if strings.HasPrefix(token, "ghp_") || strings.HasPrefix(token, "github_pat_") {
			w.config.GitHubToken = token
			w.output.Success("GitHub token configured ✓")
		} else {
			w.output.Warning("Token format looks unusual. You can update it later if needed.")
			w.config.GitHubToken = token
		}
	}
}

// showSummaryAndConfirm displays configuration summary and asks for confirmation.
func (w *ConfigWizard) showSummaryAndConfirm() error {
	w.output.Bold("\n📋 Step 6: Configuration Summary")

	w.output.Info("Your configuration:")
	w.output.Printf("  Repository: %s/%s", w.config.Organization, w.config.Repository)
	if w.config.Version != "" {
		w.output.Printf("  Version: %s", w.config.Version)
	}
	w.output.Printf("  Theme: %s", w.config.Theme)
	w.output.Printf("  Output Format: %s", w.config.OutputFormat)
	w.output.Printf("  Output Directory: %s", w.config.OutputDir)
	w.output.Printf("  Dependency Analysis: %t", w.config.AnalyzeDependencies)
	w.output.Printf("  Security Information: %t", w.config.ShowSecurityInfo)

	tokenStatus := "Not configured"
	if w.config.GitHubToken != "" {
		tokenStatus = "Configured ✓" // #nosec G101 -- status message, not actual token
	} else if internal.GetGitHubToken(w.config) != "" {
		tokenStatus = "Configured via environment ✓" // #nosec G101 -- status message, not actual token
	}
	w.output.Printf("  GitHub Token: %s", tokenStatus)

	return w.confirmConfiguration()
}

// confirmConfiguration asks user to confirm the configuration.
func (w *ConfigWizard) confirmConfiguration() error {
	w.output.Info("")
	confirmed := w.promptYesNo("Save this configuration?", true)
	if !confirmed {
		return fmt.Errorf("configuration canceled by user")
	}
	return nil
}

// promptWithDefault prompts for input with a default value.
func (w *ConfigWizard) promptWithDefault(prompt, defaultValue string) string {
	if defaultValue != "" {
		w.output.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		w.output.Printf("%s: ", prompt)
	}

	if w.scanner.Scan() {
		input := strings.TrimSpace(w.scanner.Text())
		if input == "" {
			return defaultValue
		}
		return input
	}

	return defaultValue
}

// promptSensitive prompts for sensitive input (like tokens) without echoing.
func (w *ConfigWizard) promptSensitive(prompt string) string {
	w.output.Printf("%s: ", prompt)
	if w.scanner.Scan() {
		return strings.TrimSpace(w.scanner.Text())
	}
	return ""
}

// promptYesNo prompts for a yes/no answer.
func (w *ConfigWizard) promptYesNo(prompt string, defaultValue bool) bool {
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}

	w.output.Printf("%s [%s]: ", prompt, defaultStr)

	if w.scanner.Scan() {
		input := strings.ToLower(strings.TrimSpace(w.scanner.Text()))
		switch input {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		case "":
			return defaultValue
		default:
			w.output.Warning("Please answer 'y' or 'n'. Using default.")
			return defaultValue
		}
	}

	return defaultValue
}

// findActionFiles discovers action files in the given directory.
func (w *ConfigWizard) findActionFiles(dir string) []string {
	var actionFiles []string

	// Check for action.yml and action.yaml
	for _, filename := range []string{"action.yml", "action.yaml"} {
		actionPath := filepath.Join(dir, filename)
		if _, err := os.Stat(actionPath); err == nil {
			actionFiles = append(actionFiles, actionPath)
		}
	}

	return actionFiles
}
