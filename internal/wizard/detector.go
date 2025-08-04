// Package wizard provides project setting detection functionality.
package wizard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/internal/helpers"
)

const (
	// Language constants to avoid repetition.
	langJavaScriptTypeScript = "JavaScript/TypeScript"
	langGo                   = "Go"
)

// ProjectDetector handles auto-detection of project settings.
type ProjectDetector struct {
	output     *internal.ColoredOutput
	currentDir string
	repoRoot   string
}

// NewProjectDetector creates a new project detector.
func NewProjectDetector(output *internal.ColoredOutput) (*ProjectDetector, error) {
	currentDir, err := helpers.GetCurrentDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return &ProjectDetector{
		output:     output,
		currentDir: currentDir,
		repoRoot:   helpers.FindGitRepoRoot(currentDir),
	}, nil
}

// DetectedSettings contains auto-detected project settings.
type DetectedSettings struct {
	Organization         string
	Repository           string
	Version              string
	ActionFiles          []string
	IsGitHubAction       bool
	HasDockerfile        bool
	HasCompositeAction   bool
	SuggestedTheme       string
	SuggestedRunsOn      []string
	SuggestedPermissions map[string]string
	ProjectType          string
	Language             string
	Framework            string
}

// DetectProjectSettings auto-detects project settings from the current environment.
func (d *ProjectDetector) DetectProjectSettings() (*DetectedSettings, error) {
	settings := &DetectedSettings{
		SuggestedPermissions: make(map[string]string),
		SuggestedRunsOn:      []string{"ubuntu-latest"},
	}

	// Detect repository information
	if err := d.detectRepositoryInfo(settings); err != nil {
		d.output.Warning("Could not detect repository info: %v", err)
	}

	// Detect action files and project type
	if err := d.detectActionFiles(settings); err != nil {
		d.output.Warning("Could not detect action files: %v", err)
	}

	// Detect project characteristics
	if err := d.detectProjectCharacteristics(settings); err != nil {
		d.output.Warning("Could not detect project characteristics: %v", err)
	}

	// Suggest configuration based on detection
	d.suggestConfiguration(settings)

	return settings, nil
}

// detectRepositoryInfo detects repository information from git.
func (d *ProjectDetector) detectRepositoryInfo(settings *DetectedSettings) error {
	if d.repoRoot == "" {
		return fmt.Errorf("not in a git repository")
	}

	repoInfo, err := git.DetectRepository(d.repoRoot)
	if err != nil {
		return fmt.Errorf("failed to detect repository: %w", err)
	}

	settings.Organization = repoInfo.Organization
	settings.Repository = repoInfo.Repository

	// Try to detect version from various sources
	settings.Version = d.detectVersion()

	d.output.Success("Detected repository: %s/%s", settings.Organization, settings.Repository)
	return nil
}

// detectActionFiles finds and analyzes action files.
func (d *ProjectDetector) detectActionFiles(settings *DetectedSettings) error {
	// Look for action files in current directory and subdirectories
	actionFiles, err := d.findActionFiles(d.currentDir, true)
	if err != nil {
		return fmt.Errorf("failed to find action files: %w", err)
	}

	settings.ActionFiles = actionFiles
	settings.IsGitHubAction = len(actionFiles) > 0

	if len(actionFiles) > 0 {
		d.output.Success("Found %d action file(s)", len(actionFiles))

		// Analyze action files to determine project characteristics
		for _, actionFile := range actionFiles {
			if err := d.analyzeActionFile(actionFile, settings); err != nil {
				d.output.Warning("Could not analyze %s: %v", actionFile, err)
			}
		}
	}

	return nil
}

// detectProjectCharacteristics detects project type, language, and framework.
func (d *ProjectDetector) detectProjectCharacteristics(settings *DetectedSettings) error {
	// Check for common files and patterns
	characteristics := d.analyzeProjectFiles()

	settings.ProjectType = characteristics["type"]
	settings.Language = characteristics["language"]
	settings.Framework = characteristics["framework"]

	// Check for Dockerfile
	dockerfilePath := filepath.Join(d.currentDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		settings.HasDockerfile = true
		d.output.Success("Detected Dockerfile")
	}

	return nil
}

// detectVersion attempts to detect project version from various sources.
func (d *ProjectDetector) detectVersion() string {
	// Check package.json
	if version := d.detectVersionFromPackageJSON(); version != "" {
		return version
	}

	// Check git tags
	if version := d.detectVersionFromGitTags(); version != "" {
		return version
	}

	// Check version files
	if version := d.detectVersionFromFiles(); version != "" {
		return version
	}

	return ""
}

// detectVersionFromPackageJSON detects version from package.json.
func (d *ProjectDetector) detectVersionFromPackageJSON() string {
	packageJSONPath := filepath.Join(d.currentDir, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return ""
	}

	var packageJSON struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(data, &packageJSON); err != nil {
		return ""
	}

	return packageJSON.Version
}

// detectVersionFromGitTags detects version from git tags.
func (d *ProjectDetector) detectVersionFromGitTags() string {
	if d.repoRoot == "" {
		return ""
	}

	// This is a simplified version - in a full implementation,
	// you would use git commands to get the latest tag
	return ""
}

// detectVersionFromFiles detects version from common version files.
func (d *ProjectDetector) detectVersionFromFiles() string {
	versionFiles := []string{"VERSION", "version.txt", ".version"}

	for _, filename := range versionFiles {
		versionPath := filepath.Join(d.currentDir, filename)
		if data, err := os.ReadFile(versionPath); err == nil {
			version := strings.TrimSpace(string(data))
			if version != "" {
				return version
			}
		}
	}

	return ""
}

// findActionFiles discovers action files recursively.
func (d *ProjectDetector) findActionFiles(dir string, recursive bool) ([]string, error) {
	if recursive {
		return d.findActionFilesRecursive(dir)
	}
	return d.findActionFilesInDirectory(dir)
}

// findActionFilesRecursive discovers action files recursively using filepath.Walk.
func (d *ProjectDetector) findActionFilesRecursive(dir string) ([]string, error) {
	var actionFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir // Skip errors by skipping this directory
		}

		if info.IsDir() {
			return d.handleDirectory(info)
		}

		if d.isActionFile(info.Name()) {
			actionFiles = append(actionFiles, path)
		}

		return nil
	})

	return actionFiles, err
}

// handleDirectory decides whether to skip a directory during recursive search.
func (d *ProjectDetector) handleDirectory(info os.FileInfo) error {
	name := info.Name()
	if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
		return filepath.SkipDir
	}
	return nil
}

// findActionFilesInDirectory finds action files only in the specified directory.
func (d *ProjectDetector) findActionFilesInDirectory(dir string) ([]string, error) {
	var actionFiles []string

	for _, filename := range []string{"action.yml", "action.yaml"} {
		actionPath := filepath.Join(dir, filename)
		if _, err := os.Stat(actionPath); err == nil {
			actionFiles = append(actionFiles, actionPath)
		}
	}

	return actionFiles, nil
}

// isActionFile checks if a filename is an action file.
func (d *ProjectDetector) isActionFile(filename string) bool {
	return filename == "action.yml" || filename == "action.yaml"
}

// analyzeActionFile analyzes an action file to extract characteristics.
func (d *ProjectDetector) analyzeActionFile(actionFile string, settings *DetectedSettings) error {
	action, err := d.parseActionFile(actionFile)
	if err != nil {
		return err
	}

	d.analyzeRunsSection(action, settings)
	d.analyzePermissionsSection(action, settings)

	return nil
}

// parseActionFile reads and parses an action YAML file.
func (d *ProjectDetector) parseActionFile(actionFile string) (map[string]any, error) {
	data, err := os.ReadFile(actionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read action file: %w", err)
	}

	var action map[string]any
	if err := yaml.Unmarshal(data, &action); err != nil {
		return nil, fmt.Errorf("failed to parse action YAML: %w", err)
	}

	return action, nil
}

// analyzeRunsSection analyzes the runs section of an action file.
func (d *ProjectDetector) analyzeRunsSection(action map[string]any, settings *DetectedSettings) {
	runs, ok := action["runs"].(map[string]any)
	if !ok {
		return
	}

	// Check if it's a composite action
	if using, ok := runs["using"].(string); ok && using == "composite" {
		settings.HasCompositeAction = true
	}

	// Analyze runs-on requirements if present
	d.processRunsOnField(runs, settings)
}

// processRunsOnField processes the runs-on field from the runs section.
func (d *ProjectDetector) processRunsOnField(runs map[string]any, settings *DetectedSettings) {
	runsOn, ok := runs["runs-on"]
	if !ok {
		return
	}

	if runsOnStr, ok := runsOn.(string); ok {
		settings.SuggestedRunsOn = []string{runsOnStr}
	} else if runsOnSlice, ok := runsOn.([]any); ok {
		for _, runner := range runsOnSlice {
			if runnerStr, ok := runner.(string); ok {
				settings.SuggestedRunsOn = append(settings.SuggestedRunsOn, runnerStr)
			}
		}
	}
}

// analyzePermissionsSection analyzes the permissions section of an action file.
func (d *ProjectDetector) analyzePermissionsSection(action map[string]any, settings *DetectedSettings) {
	permissions, ok := action["permissions"].(map[string]any)
	if !ok {
		return
	}

	for key, value := range permissions {
		if valueStr, ok := value.(string); ok {
			settings.SuggestedPermissions[key] = valueStr
		}
	}
}

// analyzeProjectFiles analyzes project files to determine characteristics.
func (d *ProjectDetector) analyzeProjectFiles() map[string]string {
	characteristics := make(map[string]string)

	files, err := os.ReadDir(d.currentDir)
	if err != nil {
		return characteristics
	}

	for _, file := range files {
		d.detectLanguageFromFile(file.Name(), characteristics)
		d.detectFrameworkFromFile(file.Name(), characteristics)
	}

	d.setDefaultProjectType(characteristics)
	return characteristics
}

// detectLanguageFromFile detects programming language from filename.
func (d *ProjectDetector) detectLanguageFromFile(filename string, characteristics map[string]string) {
	switch filename {
	case "package.json":
		characteristics["language"] = langJavaScriptTypeScript
		characteristics["type"] = "Node.js Project"
	case "go.mod":
		characteristics["language"] = langGo
		characteristics["type"] = "Go Module"
	case "Cargo.toml":
		characteristics["language"] = "Rust"
		characteristics["type"] = "Rust Project"
	case "pyproject.toml", "requirements.txt":
		characteristics["language"] = "Python"
		characteristics["type"] = "Python Project"
	case "Gemfile":
		characteristics["language"] = "Ruby"
		characteristics["type"] = "Ruby Project"
	case "composer.json":
		characteristics["language"] = "PHP"
		characteristics["type"] = "PHP Project"
	case "pom.xml":
		characteristics["language"] = "Java"
		characteristics["type"] = "Maven Project"
	case "build.gradle", "build.gradle.kts":
		characteristics["language"] = "Java/Kotlin"
		characteristics["type"] = "Gradle Project"
	}
}

// detectFrameworkFromFile detects framework from filename.
func (d *ProjectDetector) detectFrameworkFromFile(filename string, characteristics map[string]string) {
	switch filename {
	case "next.config.js":
		characteristics["framework"] = "Next.js"
	case "nuxt.config.js":
		characteristics["framework"] = "Nuxt.js"
	case "vue.config.js":
		characteristics["framework"] = "Vue.js"
	case "angular.json":
		characteristics["framework"] = "Angular"
	case "svelte.config.js":
		characteristics["framework"] = "Svelte"
	}
}

// setDefaultProjectType sets default project type if none detected.
func (d *ProjectDetector) setDefaultProjectType(characteristics map[string]string) {
	if characteristics["type"] == "" && len(d.getCurrentActionFiles()) > 0 {
		characteristics["type"] = "GitHub Action"
	}
}

// getCurrentActionFiles gets action files in current directory only.
func (d *ProjectDetector) getCurrentActionFiles() []string {
	actionFiles, _ := d.findActionFiles(d.currentDir, false)
	return actionFiles
}

// suggestConfiguration suggests configuration based on detected settings.
func (d *ProjectDetector) suggestConfiguration(settings *DetectedSettings) {
	d.suggestTheme(settings)
	d.suggestRunsOn(settings)
	d.suggestPermissions(settings)
}

// suggestTheme suggests an appropriate theme based on project characteristics.
func (d *ProjectDetector) suggestTheme(settings *DetectedSettings) {
	switch {
	case settings.HasCompositeAction:
		settings.SuggestedTheme = "professional"
	case settings.HasDockerfile:
		settings.SuggestedTheme = "github"
	case settings.Language == langGo:
		settings.SuggestedTheme = "minimal"
	case settings.Framework != "":
		settings.SuggestedTheme = "github"
	default:
		settings.SuggestedTheme = "default"
	}
}

// suggestRunsOn suggests appropriate runners based on language/framework.
func (d *ProjectDetector) suggestRunsOn(settings *DetectedSettings) {
	if len(settings.SuggestedRunsOn) != 1 || settings.SuggestedRunsOn[0] != "ubuntu-latest" {
		return
	}

	switch settings.Language {
	case langJavaScriptTypeScript:
		settings.SuggestedRunsOn = []string{"ubuntu-latest", "windows-latest", "macos-latest"}
	case langGo, "Python":
		settings.SuggestedRunsOn = []string{"ubuntu-latest"}
	}
}

// suggestPermissions suggests common permissions for GitHub Actions.
func (d *ProjectDetector) suggestPermissions(settings *DetectedSettings) {
	if settings.IsGitHubAction && len(settings.SuggestedPermissions) == 0 {
		settings.SuggestedPermissions = map[string]string{
			"contents": "read",
		}
	}
}
