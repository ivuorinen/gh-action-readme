// Package testutil provides testing fixtures and fixture management for gh-action-readme.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// fixtureCache provides thread-safe caching of fixture content.
var fixtureCache = struct {
	mu    sync.RWMutex
	cache map[string]string
}{
	cache: make(map[string]string),
}

// MustReadFixture reads a YAML fixture file from testdata/yaml-fixtures.
func MustReadFixture(filename string) string {
	return mustReadFixture(filename)
}

// mustReadFixture reads a YAML fixture file from testdata/yaml-fixtures with caching.
func mustReadFixture(filename string) string {
	// Try to get from cache first (read lock)
	fixtureCache.mu.RLock()
	if content, exists := fixtureCache.cache[filename]; exists {
		fixtureCache.mu.RUnlock()

		return content
	}
	fixtureCache.mu.RUnlock()

	// Not in cache, acquire write lock and read from disk
	fixtureCache.mu.Lock()
	defer fixtureCache.mu.Unlock()

	// Double-check in case another goroutine loaded it while we were waiting
	if content, exists := fixtureCache.cache[filename]; exists {
		return content
	}

	// Load from disk
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}

	// Get the project root (go up from testutil/fixtures.go to project root)
	projectRoot := filepath.Dir(filepath.Dir(currentFile))
	fixturePath := filepath.Join(projectRoot, "testdata", "yaml-fixtures", filename)

	contentBytes, err := os.ReadFile(fixturePath) // #nosec G304 -- test fixture path from project structure
	if err != nil {
		panic("failed to read fixture " + filename + ": " + err.Error())
	}

	content := string(contentBytes)

	// Store in cache
	fixtureCache.cache[filename] = content

	return content
}

// Constants for fixture management.
const (
	// YmlExtension represents the standard YAML file extension.
	YmlExtension = ".yml"
	// YamlExtension represents the alternative YAML file extension.
	YamlExtension = ".yaml"
)

// ActionType represents the type of GitHub Action being tested.
type ActionType string

const (
	// ActionTypeJavaScript represents JavaScript-based GitHub Actions that run on Node.js.
	ActionTypeJavaScript ActionType = "javascript"
	// ActionTypeComposite represents composite GitHub Actions that combine multiple steps.
	ActionTypeComposite ActionType = "composite"
	// ActionTypeDocker represents Docker-based GitHub Actions that run in containers.
	ActionTypeDocker ActionType = "docker"
	// ActionTypeInvalid represents invalid or malformed GitHub Actions for testing error scenarios.
	ActionTypeInvalid ActionType = "invalid"
	// ActionTypeMinimal represents minimal GitHub Actions with basic configuration.
	ActionTypeMinimal ActionType = "minimal"
)

// TestScenario represents a structured test scenario with metadata.
type TestScenario struct {
	ID          string         `yaml:"id"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	ActionType  ActionType     `yaml:"action_type"`
	Fixture     string         `yaml:"fixture"`
	ExpectValid bool           `yaml:"expect_valid"`
	ExpectError bool           `yaml:"expect_error"`
	Tags        []string       `yaml:"tags"`
	Metadata    map[string]any `yaml:"metadata,omitempty"`
}

// ActionFixture represents a loaded action YAML fixture with metadata.
type ActionFixture struct {
	Name       string
	Path       string
	Content    string
	ActionType ActionType
	IsValid    bool
	Scenario   *TestScenario
}

// ConfigFixture represents a loaded configuration YAML fixture.
type ConfigFixture struct {
	Name    string
	Path    string
	Content string
	Type    string
	IsValid bool
}

// FixtureManager manages test fixtures and scenarios.
type FixtureManager struct {
	basePath  string
	scenarios map[string]*TestScenario
	cache     map[string]*ActionFixture
	mu        sync.RWMutex // protects cache map
}

// GitHub API response fixtures for testing.

// GitHubReleaseResponse is a mock GitHub release API response.
const GitHubReleaseResponse = `{
	"id": 123456,
	"tag_name": "v4.1.1",
	"name": "v4.1.1",
	"body": "## What's Changed\n* Fix checkout bug\n* Improve performance",
	"draft": false,
	"prerelease": false,
	"created_at": "2023-11-01T10:00:00Z",
	"published_at": "2023-11-01T10:00:00Z",
	"tarball_url": "https://api.github.com/repos/actions/checkout/tarball/v4.1.1",
	"zipball_url": "https://api.github.com/repos/actions/checkout/zipball/v4.1.1"
}`

// GitHubTagResponse is a mock GitHub tag API response.
const GitHubTagResponse = `{
	"name": "v4.1.1",
	"zipball_url": "https://github.com/actions/checkout/zipball/v4.1.1",
	"tarball_url": "https://github.com/actions/checkout/tarball/v4.1.1",
	"commit": {
		"sha": "8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e",
		"url": "https://api.github.com/repos/actions/checkout/commits/8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e"
	},
	"node_id": "REF_kwDOAJy2KM9yZXJlZnMvdGFncy92NC4xLjE"
}`

// GitHubRepoResponse is a mock GitHub repository API response.
const GitHubRepoResponse = `{
	"id": 216219028,
	"name": "checkout",
	"full_name": "actions/checkout",
	"description": "Action for checking out a repo",
	"private": false,
	"html_url": "https://github.com/actions/checkout",
	"clone_url": "https://github.com/actions/checkout.git",
	"git_url": "git://github.com/actions/checkout.git",
	"ssh_url": "git@github.com:actions/checkout.git",
	"default_branch": "main",
	"created_at": "2019-10-16T19:40:57Z",
	"updated_at": "2023-11-01T10:00:00Z",
	"pushed_at": "2023-11-01T09:30:00Z",
	"stargazers_count": 4521,
	"watchers_count": 4521,
	"forks_count": 1234,
	"open_issues_count": 42,
	"topics": ["github-actions", "checkout", "git"]
}`

// GitHubCommitResponse is a mock GitHub commit API response.
const GitHubCommitResponse = `{
	"sha": "8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e",
	"node_id": "C_kwDOAJy2KNoAKDhmNGI3Zjg0YmQ1NzliOTVkN2YwYjkwZjhkOGI2ZTVkOWI4YTdmNmU",
	"commit": {
		"message": "Fix checkout bug and improve performance",
		"author": {
			"name": "GitHub Actions",
			"email": "actions@github.com",
			"date": "2023-11-01T09:30:00Z"
		},
		"committer": {
			"name": "GitHub Actions",
			"email": "actions@github.com",
			"date": "2023-11-01T09:30:00Z"
		}
	},
	"html_url": "https://github.com/actions/checkout/commit/8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e"
}`

// GitHubRateLimitResponse is a mock GitHub rate limit API response.
const GitHubRateLimitResponse = `{
	"resources": {
		"core": {
			"limit": 5000,
			"used": 1,
			"remaining": 4999,
			"reset": 1699027200
		},
		"search": {
			"limit": 30,
			"used": 0,
			"remaining": 30,
			"reset": 1699027200
		}
	},
	"rate": {
		"limit": 5000,
		"used": 1,
		"remaining": 4999,
		"reset": 1699027200
	}
}`

// SimpleTemplate is a basic template for testing.
const SimpleTemplate = `# {{ .Name }}

{{ .Description }}

## Installation

` + "```yaml" + `
uses: {{ gitOrg . }}/{{ gitRepo . }}@{{ actionVersion . }}
` + "```" + `

{{ if .Inputs }}
## Inputs

| Name | Description | Required | Default |
|------|-------------|----------|---------|
{{ range $key, $input := .Inputs -}}
| ` + "`{{ $key }}`" + ` | {{ $input.Description }} | {{ $input.Required }} | {{ $input.Default }} |
{{ end -}}
{{ end }}

{{ if .Outputs }}
## Outputs

| Name | Description |
|------|-------------|
{{ range $key, $output := .Outputs -}}
| ` + "`{{ $key }}`" + ` | {{ $output.Description }} |
{{ end -}}
{{ end }}
`

// GitHubErrorResponse is a mock GitHub error API response.
const GitHubErrorResponse = `{
	"message": "Not Found",
	"documentation_url": "https://docs.github.com/rest"
}`

// MockGitHubResponses returns a map of URL patterns to mock responses.
func MockGitHubResponses() map[string]string {
	return map[string]string{
		"GET https://api.github.com/repos/actions/checkout/releases/latest": GitHubReleaseResponse,
		"GET https://api.github.com/repos/actions/checkout/git/ref/tags/v4.1.1": `{
	"ref": "refs/tags/v4.1.1",
	"node_id": "REF_kwDOAJy2KM9yZXJlZnMvdGFncy92NC4xLjE",
	"url": "https://api.github.com/repos/actions/checkout/git/refs/tags/v4.1.1",
	"object": {
		"sha": "8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e",
		"type": "commit",
		"url": "https://api.github.com/repos/actions/checkout/git/commits/8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e"
	}
}`,
		"GET https://api.github.com/repos/actions/checkout/tags": `[` + GitHubTagResponse + `]`,
		"GET https://api.github.com/repos/actions/checkout":      GitHubRepoResponse,
		"GET https://api.github.com/repos/actions/checkout/commits/" +
			"8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e": GitHubCommitResponse,
		"GET https://api.github.com/rate_limit": GitHubRateLimitResponse,
		"GET https://api.github.com/repos/actions/setup-node/releases/latest": `{
	"id": 123457,
	"tag_name": "v4.0.0",
	"name": "v4.0.0",
	"body": "## What's Changed\n* Update Node.js versions\n* Fix compatibility issues",
	"draft": false,
	"prerelease": false,
	"created_at": "2023-10-15T10:00:00Z",
	"published_at": "2023-10-15T10:00:00Z"
}`,
		"GET https://api.github.com/repos/actions/setup-node/git/ref/tags/v4.0.0": `{
	"ref": "refs/tags/v4.0.0",
	"node_id": "REF_kwDOAJy2KM9yZXJlZnMvdGFncy92NC4wLjA",
	"url": "https://api.github.com/repos/actions/setup-node/git/refs/tags/v4.0.0",
	"object": {
		"sha": "1a4e6d7c9f8e5b2a3c4d5e6f7a8b9c0d1e2f3a4b",
		"type": "commit",
		"url": "https://api.github.com/repos/actions/setup-node/git/commits/1a4e6d7c9f8e5b2a3c4d5e6f7a8b9c0d1e2f3a4b"
	}
}`,
		"GET https://api.github.com/repos/actions/setup-node/tags": `[{
	"name": "v4.0.0",
	"commit": {
		"sha": "1a4e6d7c9f8e5b2a3c4d5e6f7a8b9c0d1e2f3a4b",
		"url": "https://api.github.com/repos/actions/setup-node/commits/1a4e6d7c9f8e5b2a3c4d5e6f7a8b9c0d1e2f3a4b"
	}
}]`,
	}
}

// GitIgnoreContent is a sample .gitignore file.
const GitIgnoreContent = `# Dependencies
node_modules/
*.log

# Build output
dist/
build/

# OS files
.DS_Store
Thumbs.db
`

// PackageJSONContent is a sample package.json file.
var PackageJSONContent = func() string {
	var result string
	result += "{\n"
	result += "  \"name\": \"test-action\",\n"
	result += "  \"version\": \"1.0.0\",\n"
	result += "  \"description\": \"Test GitHub Action\",\n"
	result += "  \"main\": \"index.js\",\n"
	result += "  \"scripts\": {\n"
	result += "    \"test\": \"jest\",\n"
	result += "    \"build\": \"webpack\"\n"
	result += "  },\n"
	result += "  \"dependencies\": {\n"
	result += "    \"@actions/core\": \"^1.10.0\",\n"
	result += "    \"@actions/github\": \"^5.1.1\"\n"
	result += "  },\n"
	result += "  \"devDependencies\": {\n"
	result += "    \"jest\": \"^29.0.0\",\n"
	result += "    \"webpack\": \"^5.0.0\"\n"
	result += "  }\n"
	result += "}\n"

	return result
}()

// NewFixtureManager creates a new fixture manager.
func NewFixtureManager() *FixtureManager {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}

	// Get the project root (go up from testutil/fixtures.go to project root)
	projectRoot := filepath.Dir(filepath.Dir(currentFile))
	basePath := filepath.Join(projectRoot, "testdata", "yaml-fixtures")

	return &FixtureManager{
		basePath:  basePath,
		scenarios: make(map[string]*TestScenario),
		cache:     make(map[string]*ActionFixture),
	}
}

// LoadScenarios loads test scenarios from the scenarios directory.
func (fm *FixtureManager) LoadScenarios() error {
	scenarioFile := filepath.Join(fm.basePath, "scenarios", "test-scenarios.yml")

	// Create default scenarios if file doesn't exist
	if _, err := os.Stat(scenarioFile); os.IsNotExist(err) {
		return fm.createDefaultScenarios(scenarioFile)
	}

	data, err := os.ReadFile(scenarioFile) // #nosec G304 -- test fixture path from project structure
	if err != nil {
		return fmt.Errorf("failed to read scenarios file: %w", err)
	}

	var scenarios struct {
		Scenarios []TestScenario `yaml:"scenarios"`
	}

	if err := yaml.Unmarshal(data, &scenarios); err != nil {
		return fmt.Errorf("failed to parse scenarios YAML: %w", err)
	}

	for i := range scenarios.Scenarios {
		scenario := &scenarios.Scenarios[i]
		fm.scenarios[scenario.ID] = scenario
	}

	return nil
}

// LoadActionFixture loads an action fixture with metadata.
func (fm *FixtureManager) LoadActionFixture(name string) (*ActionFixture, error) {
	// Check cache first with read lock
	fm.mu.RLock()
	if fixture, exists := fm.cache[name]; exists {
		fm.mu.RUnlock()

		return fixture, nil
	}
	fm.mu.RUnlock()

	// Determine fixture path based on naming convention
	fixturePath := fm.resolveFixturePath(name)

	content, err := os.ReadFile(fixturePath) // #nosec G304 -- test fixture path resolution
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture %s: %w", name, err)
	}

	fixture := &ActionFixture{
		Name:       name,
		Path:       fixturePath,
		Content:    string(content),
		ActionType: fm.determineActionType(name, string(content)),
		IsValid:    fm.validateFixtureContent(string(content)),
	}

	// Try to find associated scenario
	if scenario, exists := fm.scenarios[name]; exists {
		fixture.Scenario = scenario
	}

	// Cache the fixture with write lock
	fm.mu.Lock()
	// Double-check cache in case another goroutine cached it while we were loading
	if cachedFixture, exists := fm.cache[name]; exists {
		fm.mu.Unlock()

		return cachedFixture, nil
	}
	fm.cache[name] = fixture
	fm.mu.Unlock()

	return fixture, nil
}

// LoadConfigFixture loads a configuration fixture.
func (fm *FixtureManager) LoadConfigFixture(name string) (*ConfigFixture, error) {
	configPath := filepath.Join(fm.basePath, "configs", name)
	if !strings.HasSuffix(configPath, YmlExtension) && !strings.HasSuffix(configPath, YamlExtension) {
		configPath += YmlExtension
	}

	content, err := os.ReadFile(configPath) // #nosec G304 -- test fixture path from project structure
	if err != nil {
		return nil, fmt.Errorf("failed to read config fixture %s: %w", name, err)
	}

	return &ConfigFixture{
		Name:    name,
		Path:    configPath,
		Content: string(content),
		Type:    fm.determineConfigType(name),
		IsValid: fm.validateConfigContent(string(content)),
	}, nil
}

// GetFixturesByTag returns fixture names matching the specified tags.
func (fm *FixtureManager) GetFixturesByTag(tags ...string) []string {
	var matches []string

	for _, scenario := range fm.scenarios {
		if fm.scenarioMatchesTags(scenario, tags) {
			matches = append(matches, scenario.Fixture)
		}
	}

	return matches
}

// GetFixturesByActionType returns fixtures of a specific action type.
func (fm *FixtureManager) GetFixturesByActionType(actionType ActionType) []string {
	var matches []string

	for _, scenario := range fm.scenarios {
		if scenario.ActionType == actionType {
			matches = append(matches, scenario.Fixture)
		}
	}

	return matches
}

// GetValidFixtures returns all fixtures that should parse as valid actions.
func (fm *FixtureManager) GetValidFixtures() []string {
	var matches []string

	for _, scenario := range fm.scenarios {
		if scenario.ExpectValid {
			matches = append(matches, scenario.Fixture)
		}
	}

	return matches
}

// GetInvalidFixtures returns all fixtures that should be invalid.
func (fm *FixtureManager) GetInvalidFixtures() []string {
	var matches []string

	for _, scenario := range fm.scenarios {
		if !scenario.ExpectValid {
			matches = append(matches, scenario.Fixture)
		}
	}

	return matches
}

// resolveFixturePath determines the full path to a fixture file.
func (fm *FixtureManager) resolveFixturePath(name string) string {
	// If it's a direct path, use it
	if strings.Contains(name, "/") {
		return fm.ensureYamlExtension(filepath.Join(fm.basePath, name))
	}

	// Try to find the fixture in search directories
	if foundPath := fm.searchInDirectories(name); foundPath != "" {
		return foundPath
	}

	// Default to root level if not found
	return fm.ensureYamlExtension(filepath.Join(fm.basePath, name))
}

// ensureYamlExtension adds YAML extension if not present.
func (fm *FixtureManager) ensureYamlExtension(path string) string {
	if !strings.HasSuffix(path, YmlExtension) && !strings.HasSuffix(path, YamlExtension) {
		path += YmlExtension
	}

	return path
}

// searchInDirectories searches for fixture in predefined directories.
func (fm *FixtureManager) searchInDirectories(name string) string {
	searchDirs := []string{
		"actions/javascript",
		"actions/composite",
		"actions/docker",
		"actions/invalid",
		"", // root level
	}

	for _, dir := range searchDirs {
		path := fm.buildSearchPath(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// buildSearchPath constructs search path for a directory.
func (fm *FixtureManager) buildSearchPath(dir, name string) string {
	var path string
	if dir == "" {
		path = filepath.Join(fm.basePath, name)
	} else {
		path = filepath.Join(fm.basePath, dir, name)
	}

	return fm.ensureYamlExtension(path)
}

// determineActionType infers action type from fixture name and content.
func (fm *FixtureManager) determineActionType(name, content string) ActionType {
	// Check by name/path first
	if actionType := fm.determineActionTypeByName(name); actionType != ActionTypeMinimal {
		return actionType
	}

	// Fall back to content analysis
	return fm.determineActionTypeByContent(content)
}

// determineActionTypeByName infers action type from fixture name or path.
func (fm *FixtureManager) determineActionTypeByName(name string) ActionType {
	if strings.Contains(name, "javascript") || strings.Contains(name, "node") {
		return ActionTypeJavaScript
	}
	if strings.Contains(name, "composite") {
		return ActionTypeComposite
	}
	if strings.Contains(name, "docker") {
		return ActionTypeDocker
	}
	if strings.Contains(name, "invalid") {
		return ActionTypeInvalid
	}
	if strings.Contains(name, "minimal") {
		return ActionTypeMinimal
	}

	return ActionTypeMinimal
}

// determineActionTypeByContent infers action type from YAML content.
func (fm *FixtureManager) determineActionTypeByContent(content string) ActionType {
	if strings.Contains(content, `using: 'composite'`) || strings.Contains(content, `using: "composite"`) {
		return ActionTypeComposite
	}
	if strings.Contains(content, `using: 'docker'`) || strings.Contains(content, `using: "docker"`) {
		return ActionTypeDocker
	}
	if strings.Contains(content, `using: 'node`) {
		return ActionTypeJavaScript
	}

	return ActionTypeMinimal
}

// determineConfigType determines the type of configuration fixture.
func (fm *FixtureManager) determineConfigType(name string) string {
	if strings.Contains(name, "global") {
		return "global"
	}
	if strings.Contains(name, "repo") {
		return "repo-specific"
	}
	if strings.Contains(name, "user") {
		return "user-specific"
	}

	return "generic"
}

// validateFixtureContent performs basic validation on fixture content.
func (fm *FixtureManager) validateFixtureContent(content string) bool {
	// Basic YAML structure validation
	var data map[string]any
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return false
	}

	// Check for required fields for valid actions
	if _, hasName := data["name"]; !hasName {
		return false
	}
	if _, hasDescription := data["description"]; !hasDescription {
		return false
	}
	runs, hasRuns := data["runs"]
	if !hasRuns {
		return false
	}

	// Validate the runs section content more thoroughly
	runsMap, ok := runs.(map[string]any)
	if !ok {
		return false // runs field exists but is not a map
	}

	using, hasUsing := runsMap["using"]
	if !hasUsing {
		return false // runs section exists but has no using field
	}

	usingStr, ok := using.(string)
	if !ok {
		return false // using field exists but is not a string
	}

	// Use the same validation logic as ValidateActionYML
	if !isValidRuntime(usingStr) {
		return false
	}

	return true
}

// isValidRuntime checks if the given runtime is valid for GitHub Actions.
// This is duplicated from internal/validator.go to avoid import cycle.
func isValidRuntime(runtime string) bool {
	validRuntimes := []string{
		"node12",    // Legacy Node.js runtime (deprecated)
		"node16",    // Legacy Node.js runtime (deprecated)
		"node20",    // Current Node.js runtime
		"docker",    // Docker container runtime
		"composite", // Composite action runtime
	}

	runtime = strings.TrimSpace(strings.ToLower(runtime))
	for _, valid := range validRuntimes {
		if runtime == valid {
			return true
		}
	}

	return false
}

// validateConfigContent validates configuration fixture content.
func (fm *FixtureManager) validateConfigContent(content string) bool {
	var data map[string]any

	return yaml.Unmarshal([]byte(content), &data) == nil
}

// scenarioMatchesTags checks if a scenario matches any of the provided tags.
func (fm *FixtureManager) scenarioMatchesTags(scenario *TestScenario, tags []string) bool {
	if len(tags) == 0 {
		return true
	}

	for _, tag := range tags {
		for _, scenarioTag := range scenario.Tags {
			if tag == scenarioTag {
				return true
			}
		}
	}

	return false
}

// createDefaultScenarios creates a default scenarios file.
func (fm *FixtureManager) createDefaultScenarios(scenarioFile string) error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(scenarioFile), 0750); err != nil { // #nosec G301 -- test directory permissions
		return fmt.Errorf("failed to create scenarios directory: %w", err)
	}

	defaultScenarios := struct {
		Scenarios []TestScenario `yaml:"scenarios"`
	}{
		Scenarios: []TestScenario{
			{
				ID:          "simple-javascript",
				Name:        "Simple JavaScript Action",
				Description: "Basic JavaScript action with minimal configuration",
				ActionType:  ActionTypeJavaScript,
				Fixture:     "actions/javascript/simple.yml",
				ExpectValid: true,
				ExpectError: false,
				Tags:        []string{"javascript", "basic", "valid"},
			},
			{
				ID:          "composite-basic",
				Name:        "Basic Composite Action",
				Description: "Composite action with multiple steps",
				ActionType:  ActionTypeComposite,
				Fixture:     "actions/composite/basic.yml",
				ExpectValid: true,
				ExpectError: false,
				Tags:        []string{"composite", "basic", "valid"},
			},
			{
				ID:          "docker-basic",
				Name:        "Basic Docker Action",
				Description: "Docker-based action with Dockerfile",
				ActionType:  ActionTypeDocker,
				Fixture:     "actions/docker/basic.yml",
				ExpectValid: true,
				ExpectError: false,
				Tags:        []string{"docker", "basic", "valid"},
			},
			{
				ID:          "invalid-missing-description",
				Name:        "Invalid Action - Missing Description",
				Description: "Action missing required description field",
				ActionType:  ActionTypeInvalid,
				Fixture:     "actions/invalid/missing-description.yml",
				ExpectValid: false,
				ExpectError: true,
				Tags:        []string{"invalid", "validation", "error"},
			},
		},
	}

	data, err := yaml.Marshal(&defaultScenarios)
	if err != nil {
		return fmt.Errorf("failed to marshal default scenarios: %w", err)
	}

	if err := os.WriteFile(scenarioFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write scenarios file: %w", err)
	}

	// Load the scenarios we just created
	return fm.LoadScenarios()
}

// Global fixture manager instance.
var defaultFixtureManager *FixtureManager

// GetFixtureManager returns the global fixture manager instance.
func GetFixtureManager() *FixtureManager {
	if defaultFixtureManager == nil {
		defaultFixtureManager = NewFixtureManager()
		if err := defaultFixtureManager.LoadScenarios(); err != nil {
			panic(fmt.Sprintf("failed to load test scenarios: %v", err))
		}
	}

	return defaultFixtureManager
}

// Helper functions for backward compatibility and convenience

// LoadActionFixture loads an action fixture using the global fixture manager.
func LoadActionFixture(name string) (*ActionFixture, error) {
	return GetFixtureManager().LoadActionFixture(name)
}

// LoadConfigFixture loads a config fixture using the global fixture manager.
func LoadConfigFixture(name string) (*ConfigFixture, error) {
	return GetFixtureManager().LoadConfigFixture(name)
}

// GetFixturesByTag returns fixtures matching tags using the global fixture manager.
func GetFixturesByTag(tags ...string) []string {
	return GetFixtureManager().GetFixturesByTag(tags...)
}

// GetFixturesByActionType returns fixtures by action type using the global fixture manager.
func GetFixturesByActionType(actionType ActionType) []string {
	return GetFixtureManager().GetFixturesByActionType(actionType)
}

// GetValidFixtures returns all valid fixtures using the global fixture manager.
func GetValidFixtures() []string {
	return GetFixtureManager().GetValidFixtures()
}

// GetInvalidFixtures returns all invalid fixtures using the global fixture manager.
func GetInvalidFixtures() []string {
	return GetFixtureManager().GetInvalidFixtures()
}
