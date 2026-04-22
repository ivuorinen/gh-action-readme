// Package main provides the deps (dependency management) commands.
package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/helpers"
)

// InputReader interface for reading user input (enables testing).
type InputReader interface {
	ReadLine() (string, error)
}

// StdinReader reads from actual stdin.
type StdinReader struct{}

func (r *StdinReader) ReadLine() (string, error) {
	var response string
	_, err := fmt.Scanln(&response)

	return strings.TrimSpace(response), err
}

func newDepsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deps",
		Short: "Dependency management commands",
		Long:  "Analyze and manage GitHub Action dependencies",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all dependencies in action files",
		Run:   wrapHandlerWithErrorHandling(depsListHandler),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "security",
		Short: "Analyze dependency security (pinned vs floating versions)",
		Run:   wrapHandlerWithErrorHandling(depsSecurityHandler),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "outdated",
		Short: "Check for outdated dependencies",
		Run:   wrapHandlerWithErrorHandling(depsOutdatedHandler),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "graph",
		Short: "Generate dependency graph",
		Run:   depsGraphHandler,
	})

	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade dependencies with interactive or CI mode",
		Long:  "Upgrade dependencies to latest versions. Use --ci for automated pinned updates.",
		Run:   wrapHandlerWithErrorHandling(depsUpgradeHandler),
	}
	upgradeCmd.Flags().Bool(appconstants.FlagCI, false, "CI/CD mode: automatically pin all updates to commit SHAs")
	upgradeCmd.Flags().Bool(appconstants.InputAll, false, "Update all outdated dependencies without prompts")
	upgradeCmd.Flags().Bool(appconstants.InputDryRun, false, "Show what would be updated without making changes")
	cmd.AddCommand(upgradeCmd)

	pinCmd := &cobra.Command{
		Use:   appconstants.CommandPin,
		Short: "Pin floating versions to specific commits",
		Long:  "Convert floating versions (like @v4) to pinned commit SHAs with version comments.",
		Run:   wrapHandlerWithErrorHandling(depsUpgradeHandler), // Uses same handler with different flags
	}
	pinCmd.Flags().Bool(appconstants.InputAll, false, "Pin all floating dependencies")
	pinCmd.Flags().Bool(appconstants.InputDryRun, false, "Show what would be pinned without making changes")
	cmd.AddCommand(pinCmd)

	return cmd
}

func depsListHandler(_ *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return wrapError(appconstants.ErrErrorGettingCurrentDir, err)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(
		currentDir,
		true,
		globalConfig.IgnoredDirectories,
		"dependency listing",
	)
	if err := handleNoFilesFoundError(err, output); err != nil {
		return err
	}

	analyzer := createAnalyzer(generator, output)
	totalDeps := analyzeDependencies(output, actionFiles, analyzer)

	if totalDeps > 0 {
		output.Bold("\nTotal dependencies: %d", totalDeps)
	}

	return nil
}

// analyzeDependencies analyzes and displays dependencies.
func analyzeDependencies(
	output *internal.ColoredOutput,
	actionFiles []string,
	analyzer *dependencies.Analyzer,
) int {
	totalDeps := 0
	output.Bold("Dependencies found in action files:")

	// Create progress bar for multiple files
	progressMgr := internal.NewProgressBarManager(output.IsQuiet())

	progressMgr.ProcessWithProgressBar(
		"Analyzing dependencies",
		actionFiles,
		func(actionFile string, bar *progressbar.ProgressBar) {
			if bar == nil {
				output.Info("\n📄 %s", actionFile)
			}
			totalDeps += analyzeActionFileDeps(output, actionFile, analyzer)
		},
	)

	return totalDeps
}

// analyzeActionFileDeps analyzes dependencies in a single action file.
func analyzeActionFileDeps(
	output *internal.ColoredOutput,
	actionFile string,
	analyzer *dependencies.Analyzer,
) int {
	if analyzer == nil {
		output.Printf("  • Cannot analyze (no GitHub token)\n")

		return 0
	}

	deps, err := analyzer.AnalyzeActionFile(actionFile)
	if err != nil {
		output.Warning("  ⚠️  Error analyzing: %v", err)

		return 0
	}

	if len(deps) == 0 {
		output.Printf("  • No dependencies (not a composite action)\n")

		return 0
	}

	for _, dep := range deps {
		if dep.IsPinned {
			output.Success("  🔒 %s @ %s - %s", dep.Name, dep.Version, dep.Description)
		} else {
			output.Warning("  📌 %s @ %s - %s", dep.Name, dep.Version, dep.Description)
		}
	}

	return len(deps)
}

func depsSecurityHandler(_ *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)

	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(
		currentDir,
		true,
		globalConfig.IgnoredDirectories,
		"security analysis",
	)
	if err != nil {
		return fmt.Errorf(appconstants.ErrFailedToDiscoverActionFiles, err)
	}

	analyzer := createAnalyzer(generator, output)
	if analyzer == nil {
		output.Warning(
			"⚠️  Analyzer disabled: GitHub token not configured. " +
				"Use GITHUB_TOKEN or GH_README_GITHUB_TOKEN environment variable.",
		)

		return nil // Analyzer can be nil if token isn't configured, gracefully handle
	}

	pinnedCount, floatingDeps := analyzeSecurityDeps(output, actionFiles, analyzer)
	displaySecuritySummary(output, currentDir, pinnedCount, floatingDeps)

	return nil
}

// analyzeSecurityDeps analyzes dependencies for security issues.
func analyzeSecurityDeps(
	output *internal.ColoredOutput,
	actionFiles []string,
	analyzer *dependencies.Analyzer,
) (int, []struct {
	file string
	dep  dependencies.Dependency
}) {
	pinnedCount := 0
	var floatingDeps []struct {
		file string
		dep  dependencies.Dependency
	}

	output.Bold("Security Analysis of GitHub Action Dependencies:")

	// Create progress bar for multiple files
	progressMgr := internal.NewProgressBarManager(output.IsQuiet())

	progressMgr.ProcessWithProgressBar(
		"Security analysis",
		actionFiles,
		func(actionFile string, _ *progressbar.ProgressBar) {
			deps, err := analyzer.AnalyzeActionFile(actionFile)
			if err != nil {
				return
			}

			for _, dep := range deps {
				if dep.IsPinned {
					pinnedCount++
				} else {
					floatingDeps = append(floatingDeps, struct {
						file string
						dep  dependencies.Dependency
					}{actionFile, dep})
				}
			}
		},
	)

	return pinnedCount, floatingDeps
}

// displaySecuritySummary shows security analysis results.
func displaySecuritySummary(
	output *internal.ColoredOutput,
	currentDir string,
	pinnedCount int,
	floatingDeps []struct {
		file string
		dep  dependencies.Dependency
	},
) {
	output.Success("\n🔒 Pinned versions: %d (Recommended for security)", pinnedCount)
	floatingCount := len(floatingDeps)

	if floatingCount > 0 {
		output.Warning("📌 Floating versions: %d (Consider pinning)", floatingCount)
		displayFloatingDeps(output, currentDir, floatingDeps)
		output.Info("\nRecommendation: Pin dependencies to specific commits or semantic versions for better security.")
	} else if pinnedCount > 0 {
		output.Info("\n✅ All dependencies are properly pinned!")
	}
}

// displayFloatingDeps shows floating dependencies details.
func displayFloatingDeps(
	output *internal.ColoredOutput,
	currentDir string,
	floatingDeps []struct {
		file string
		dep  dependencies.Dependency
	},
) {
	output.Bold("\nFloating dependencies that should be pinned:")
	for _, fd := range floatingDeps {
		relPath, _ := filepath.Rel(currentDir, fd.file)
		output.Warning("  • %s @ %s", fd.dep.Name, fd.dep.Version)
		output.Printf("    in %s\n", relPath)
	}
}

func depsOutdatedHandler(_ *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return wrapError(appconstants.ErrErrorGettingCurrentDir, err)
	}

	generator := internal.NewGenerator(globalConfig)
	actionFiles, err := generator.DiscoverActionFilesWithValidation(
		currentDir,
		true,
		globalConfig.IgnoredDirectories,
		"outdated dependency analysis",
	)
	if err := handleNoFilesFoundError(err, output); err != nil {
		return err
	}

	analyzer := createAnalyzer(generator, output)
	if !validateGitHubToken(output) {
		return nil // Not an error, just no token available
	}

	if analyzer == nil {
		return nil // Analyzer can be nil if token isn't configured, gracefully handle
	}

	allOutdated := checkAllOutdated(output, actionFiles, analyzer)
	displayOutdatedResults(output, allOutdated)

	return nil
}

// validateGitHubToken checks if GitHub token is available.
func validateGitHubToken(output *internal.ColoredOutput) bool {
	if globalConfig.GitHubToken == "" {
		contextualErr := apperrors.New(appconstants.ErrCodeGitHubAuth, "GitHub token not found").
			WithSuggestions(apperrors.GetSuggestions(appconstants.ErrCodeGitHubAuth, map[string]string{})...).
			WithHelpURL(apperrors.GetHelpURL(appconstants.ErrCodeGitHubAuth))

		output.Warning("⚠️  %s", contextualErr.Error())

		return false
	}

	return true
}

// checkAllOutdated checks all action files for outdated dependencies.
func checkAllOutdated(
	output *internal.ColoredOutput,
	actionFiles []string,
	analyzer *dependencies.Analyzer,
) []dependencies.OutdatedDependency {
	output.Bold("Checking for outdated dependencies...")
	var allOutdated []dependencies.OutdatedDependency

	for _, actionFile := range actionFiles {
		deps, err := analyzer.AnalyzeActionFile(actionFile)
		if err != nil {
			output.Warning(appconstants.ErrErrorAnalyzing, actionFile, err)

			continue
		}

		outdated, err := analyzer.CheckOutdated(deps)
		if err != nil {
			output.Warning(appconstants.ErrErrorCheckingOutdated, actionFile, err)

			continue
		}

		allOutdated = append(allOutdated, outdated...)
	}

	return allOutdated
}

// displayOutdatedResults shows outdated dependency results.
func displayOutdatedResults(output *internal.ColoredOutput, allOutdated []dependencies.OutdatedDependency) {
	if len(allOutdated) == 0 {
		output.Success("✅ All dependencies are up to date!")

		return
	}

	output.Warning("Found %d outdated dependencies:", len(allOutdated))
	for _, outdated := range allOutdated {
		output.Printf("  • %s: %s → %s (%s update)",
			outdated.Current.Name,
			outdated.Current.Version,
			outdated.LatestVersion,
			outdated.UpdateType)
		if outdated.IsSecurityUpdate {
			output.Warning("    🔒 Potential security update")
		}
	}

	output.Info("\nRun 'gh-action-readme deps upgrade' to update dependencies")
}

func depsUpgradeHandler(cmd *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return wrapError(appconstants.ErrErrorGettingCurrentDir, err)
	}

	// Setup and validation
	analyzer, actionFiles, err := setupDepsUpgrade(output, currentDir, nil)
	if err != nil {
		// setupDepsUpgrade returns descriptive errors, so just pass them through
		return err
	}

	// Parse flags and show mode
	ciMode, _ := cmd.Flags().GetBool(appconstants.FlagCI)
	allFlag, _ := cmd.Flags().GetBool(appconstants.InputAll)
	dryRun, _ := cmd.Flags().GetBool(appconstants.InputDryRun)
	isPinCmd := cmd.Use == appconstants.CommandPin

	showUpgradeMode(output, ciMode, isPinCmd)

	// Collect all updates
	allUpdates := collectAllUpdates(output, analyzer, actionFiles)
	if len(allUpdates) == 0 {
		output.Success("✅ No updates needed - all dependencies are current and pinned!")

		return nil
	}

	// Show and apply updates
	showPendingUpdates(output, allUpdates, currentDir)
	if !dryRun {
		if err := applyUpdates(output, analyzer, allUpdates, ciMode || allFlag, nil); err != nil {
			return err
		}
	} else {
		output.Info("\n🔍 Dry run complete - no changes made")
	}

	return nil
}

// setupDepsUpgrade handles initial setup and validation for dependency upgrades.
// The config parameter allows injection for testing (pass nil to use globalConfig).
func setupDepsUpgrade(
	_ *internal.ColoredOutput,
	currentDir string,
	config *internal.AppConfig,
) (*dependencies.Analyzer, []string, error) {
	// Default to globalConfig if not provided (backward compatible)
	if config == nil {
		if globalConfig == nil {
			globalConfig = internal.DefaultAppConfig()
		}
		config = globalConfig
	}

	generator := internal.NewGenerator(config)
	actionFiles, err := generator.DiscoverActionFiles(currentDir, true, config.IgnoredDirectories)
	if err != nil {
		return nil, nil, fmt.Errorf("error discovering action files: %w", err)
	}

	if len(actionFiles) == 0 {
		return nil, nil, errors.New(appconstants.ErrNoActionFilesFound)
	}

	analyzer, err := generator.CreateDependencyAnalyzer()
	if err != nil {
		return nil, nil, fmt.Errorf("could not create dependency analyzer: %w", err)
	}

	if config.GitHubToken == "" {
		return nil, nil, errors.New("no GitHub token found, set GITHUB_TOKEN environment variable")
	}

	return analyzer, actionFiles, nil
}

// showUpgradeMode displays the current upgrade mode to the user.
func showUpgradeMode(output *internal.ColoredOutput, ciMode, isPinCmd bool) {
	switch {
	case ciMode:
		output.Bold("🤖 CI/CD Mode: Automated dependency updates with pinned commit SHAs")
	case isPinCmd:
		output.Bold("📌 Pinning floating dependencies to commit SHAs")
	default:
		output.Bold("🔄 Interactive dependency upgrade")
	}
}

// collectAllUpdates gathers all available updates from action files.
func collectAllUpdates(
	output *internal.ColoredOutput,
	analyzer *dependencies.Analyzer,
	actionFiles []string,
) []dependencies.PinnedUpdate {
	var allUpdates []dependencies.PinnedUpdate

	for _, actionFile := range actionFiles {
		deps, err := analyzer.AnalyzeActionFile(actionFile)
		if err != nil {
			output.Warning(appconstants.ErrErrorAnalyzing, actionFile, err)

			continue
		}

		outdated, err := analyzer.CheckOutdated(deps)
		if err != nil {
			output.Warning(appconstants.ErrErrorCheckingOutdated, actionFile, err)

			continue
		}

		for _, outdatedDep := range outdated {
			update, err := analyzer.GeneratePinnedUpdate(
				actionFile,
				outdatedDep.Current,
				outdatedDep.LatestVersion,
				outdatedDep.LatestSHA,
			)
			if err != nil {
				output.Warning("Error generating update for %s: %v", outdatedDep.Current.Name, err)

				continue
			}
			allUpdates = append(allUpdates, *update)
		}
	}

	return allUpdates
}

// showPendingUpdates displays what updates will be applied.
func showPendingUpdates(
	output *internal.ColoredOutput,
	allUpdates []dependencies.PinnedUpdate,
	currentDir string,
) {
	output.Info("Found %d dependencies to update:", len(allUpdates))
	for _, update := range allUpdates {
		relPath, _ := filepath.Rel(currentDir, update.FilePath)
		output.Printf("  • %s (%s update)", update.OldUses, update.UpdateType)
		output.Printf("    → %s", update.NewUses)
		output.Printf("    in %s", relPath)
	}
}

// applyUpdates applies the collected updates either automatically or interactively.
// The reader parameter allows injection of input for testing (pass nil to use stdin).
func applyUpdates(
	output *internal.ColoredOutput,
	analyzer *dependencies.Analyzer,
	allUpdates []dependencies.PinnedUpdate,
	automatic bool,
	reader InputReader,
) error {
	// Default to stdin if not provided
	if reader == nil {
		reader = &StdinReader{}
	}

	if automatic {
		output.Info("\n🚀 Applying updates...")
		if err := analyzer.ApplyPinnedUpdates(allUpdates); err != nil {
			return fmt.Errorf(appconstants.ErrFailedToApplyUpdatesWrapped, err)
		}
		output.Success("✅ Successfully updated %d dependencies with pinned commit SHAs", len(allUpdates))
	} else {
		// Interactive mode
		output.Info("\n❓ This will modify your action.yml files. Continue? (y/N): ")
		response, err := reader.ReadLine()
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		if !slices.Contains([]string{"y", appconstants.InputYes}, strings.ToLower(response)) {
			output.Info("Canceled")

			return nil
		}

		output.Info("🚀 Applying updates...")
		if err := analyzer.ApplyPinnedUpdates(allUpdates); err != nil {
			return fmt.Errorf(appconstants.ErrFailedToApplyUpdatesWrapped, err)
		}
		output.Success("✅ Successfully updated %d dependencies", len(allUpdates))
	}

	return nil
}

func depsGraphHandler(_ *cobra.Command, _ []string) {
	output := createOutputManager(globalConfig.Quiet)
	output.Bold("Dependency Graph:")
	output.Info("Generating visual dependency graph...")
	output.Printf("This feature is not yet implemented\n")
}
