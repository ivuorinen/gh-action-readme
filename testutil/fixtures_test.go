package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const testVersion = "v4.1.1"

func TestMustReadFixture(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "valid fixture file",
			filename: "simple-action.yml",
			wantErr:  false,
		},
		{
			name:     "another valid fixture",
			filename: "composite-action.yml",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic but got none")
					}
				}()
			}

			content := mustReadFixture(tt.filename)
			if !tt.wantErr {
				if content == "" {
					t.Error("expected non-empty content")
				}
				// Verify it's valid YAML
				var yamlContent map[string]any
				if err := yaml.Unmarshal([]byte(content), &yamlContent); err != nil {
					t.Errorf("fixture content is not valid YAML: %v", err)
				}
			}
		})
	}
}

func TestMustReadFixture_Panic(t *testing.T) {
	t.Run("missing file panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic but got none")
			} else {
				errStr, ok := r.(string)
				if !ok {
					t.Errorf("expected panic to contain string message, got: %T", r)

					return
				}
				if !strings.Contains(errStr, "failed to read fixture") {
					t.Errorf("expected panic message about fixture reading, got: %v", r)
				}
			}
		}()

		mustReadFixture("nonexistent-file.yml")
	})
}

func TestGitHubAPIResponses(t *testing.T) {
	t.Run("GitHubReleaseResponse", func(t *testing.T) {
		testGitHubReleaseResponse(t)
	})
	t.Run("GitHubTagResponse", func(t *testing.T) {
		testGitHubTagResponse(t)
	})
	t.Run("GitHubRepoResponse", func(t *testing.T) {
		testGitHubRepoResponse(t)
	})
	t.Run("GitHubCommitResponse", func(t *testing.T) {
		testGitHubCommitResponse(t)
	})
	t.Run("GitHubRateLimitResponse", func(t *testing.T) {
		testGitHubRateLimitResponse(t)
	})
	t.Run("GitHubErrorResponse", func(t *testing.T) {
		testGitHubErrorResponse(t)
	})
}

// testGitHubReleaseResponse validates the GitHub release response format.
func testGitHubReleaseResponse(t *testing.T) {
	data := parseJSONResponse(t, GitHubReleaseResponse)

	if data["id"] == nil {
		t.Error("expected id field")
	}
	if data["tag_name"] != testVersion {
		t.Errorf("expected tag_name %s, got %v", testVersion, data["tag_name"])
	}
	if data["name"] != testVersion {
		t.Errorf("expected name %s, got %v", testVersion, data["name"])
	}
}

// testGitHubTagResponse validates the GitHub tag response format.
func testGitHubTagResponse(t *testing.T) {
	data := parseJSONResponse(t, GitHubTagResponse)

	if data["name"] != testVersion {
		t.Errorf("expected name %s, got %v", testVersion, data["name"])
	}
	if data["commit"] == nil {
		t.Error("expected commit field")
	}
}

// testGitHubRepoResponse validates the GitHub repository response format.
func testGitHubRepoResponse(t *testing.T) {
	data := parseJSONResponse(t, GitHubRepoResponse)

	if data["name"] != "checkout" {
		t.Errorf("expected name checkout, got %v", data["name"])
	}
	if data["full_name"] != "actions/checkout" {
		t.Errorf("expected full_name actions/checkout, got %v", data["full_name"])
	}
}

// testGitHubCommitResponse validates the GitHub commit response format.
func testGitHubCommitResponse(t *testing.T) {
	data := parseJSONResponse(t, GitHubCommitResponse)

	if data["sha"] == nil {
		t.Error("expected sha field")
	}
	if data["commit"] == nil {
		t.Error("expected commit field")
	}
}

// testGitHubRateLimitResponse validates the GitHub rate limit response format.
func testGitHubRateLimitResponse(t *testing.T) {
	data := parseJSONResponse(t, GitHubRateLimitResponse)

	if data["resources"] == nil {
		t.Error("expected resources field")
	}
	if data["rate"] == nil {
		t.Error("expected rate field")
	}
}

// testGitHubErrorResponse validates the GitHub error response format.
func testGitHubErrorResponse(t *testing.T) {
	data := parseJSONResponse(t, GitHubErrorResponse)

	if data["message"] != "Not Found" {
		t.Errorf("expected message 'Not Found', got %v", data["message"])
	}
}

// parseJSONResponse parses a JSON response string and returns the data map.
func parseJSONResponse(t *testing.T, response string) map[string]any {
	var data map[string]any
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	return data
}

func TestSimpleTemplate(t *testing.T) {
	template := SimpleTemplate

	// Check that template contains expected sections
	expectedSections := []string{
		"# {{ .Name }}",
		"{{ .Description }}",
		"## Installation",
		"uses: {{ gitOrg . }}/{{ gitRepo . }}@{{ actionVersion . }}",
		"## Inputs",
		"## Outputs",
	}

	for _, section := range expectedSections {
		if !strings.Contains(template, section) {
			t.Errorf("template missing expected section: %s", section)
		}
	}

	// Verify template has proper structure
	if !strings.Contains(template, "```yaml") {
		t.Error("template should contain YAML code blocks")
	}

	if !strings.Contains(template, "| Name | Description |") {
		t.Error("template should contain table headers")
	}
}

func TestMockGitHubResponses(t *testing.T) {
	responses := MockGitHubResponses()

	// Test that all expected endpoints are present
	expectedEndpoints := []string{
		"GET https://api.github.com/repos/actions/checkout/releases/latest",
		"GET https://api.github.com/repos/actions/checkout/git/ref/tags/v4.1.1",
		"GET https://api.github.com/repos/actions/checkout/tags",
		"GET https://api.github.com/repos/actions/checkout",
		"GET https://api.github.com/rate_limit",
		"GET https://api.github.com/repos/actions/setup-node/releases/latest",
	}

	for _, endpoint := range expectedEndpoints {
		if _, exists := responses[endpoint]; !exists {
			t.Errorf("missing endpoint: %s", endpoint)
		}
	}

	// Test that all responses are valid JSON
	for endpoint, response := range responses {
		var data any
		if err := json.Unmarshal([]byte(response), &data); err != nil {
			t.Errorf("invalid JSON for endpoint %s: %v", endpoint, err)
		}
	}

	// Test specific response structures
	t.Run("checkout releases response", func(t *testing.T) {
		response := responses["GET https://api.github.com/repos/actions/checkout/releases/latest"]
		var release map[string]any
		if err := json.Unmarshal([]byte(response), &release); err != nil {
			t.Fatalf("failed to parse release response: %v", err)
		}

		if release["tag_name"] == nil {
			t.Error("release response missing tag_name")
		}
	})
}

func TestFixtureConstants(t *testing.T) {
	// Test that all fixture variables are properly loaded
	fixtures := map[string]string{
		"SimpleActionYML":        MustReadFixture("actions/javascript/simple.yml"),
		"CompositeActionYML":     MustReadFixture("actions/composite/basic.yml"),
		"DockerActionYML":        MustReadFixture("actions/docker/basic.yml"),
		"InvalidActionYML":       MustReadFixture("actions/invalid/missing-description.yml"),
		"MinimalActionYML":       MustReadFixture("minimal-action.yml"),
		"TestProjectActionYML":   MustReadFixture("test-project-action.yml"),
		"RepoSpecificConfigYAML": MustReadFixture("repo-config.yml"),
		"PackageJSONContent":     PackageJSONContent,
	}

	for name, content := range fixtures {
		t.Run(name, func(t *testing.T) {
			if content == "" {
				t.Errorf("%s is empty", name)
			}

			// For YAML fixtures, verify they're valid YAML (except InvalidActionYML)
			if strings.HasSuffix(name, "YML") || strings.HasSuffix(name, "YAML") {
				if name != "InvalidActionYML" {
					var yamlContent map[string]any
					if err := yaml.Unmarshal([]byte(content), &yamlContent); err != nil {
						t.Errorf("%s contains invalid YAML: %v", name, err)
					}
				}
			}

			// For JSON fixtures, verify they're valid JSON
			if strings.Contains(name, "JSON") {
				var jsonContent any
				if err := json.Unmarshal([]byte(content), &jsonContent); err != nil {
					t.Errorf("%s contains invalid JSON: %v", name, err)
				}
			}
		})
	}
}

func TestGitIgnoreContent(t *testing.T) {
	content := GitIgnoreContent

	expectedPatterns := []string{
		"node_modules/",
		"*.log",
		"dist/",
		"build/",
		".DS_Store",
		"Thumbs.db",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(content, pattern) {
			t.Errorf("GitIgnoreContent missing pattern: %s", pattern)
		}
	}

	// Verify it has comments
	if !strings.Contains(content, "# Dependencies") {
		t.Error("GitIgnoreContent should contain section comments")
	}
}

// Test helper functions that interact with the filesystem.
func TestFixtureFileSystem(t *testing.T) {
	// Verify that the fixture files actually exist
	fixtureFiles := []string{
		"simple-action.yml",
		"composite-action.yml",
		"docker-action.yml",
		"invalid-action.yml",
		"minimal-action.yml",
		"test-project-action.yml",
		"repo-config.yml",
		"package.json",
		"dynamic-action-template.yml",
		"composite-template.yml",
	}

	// Get the testdata directory path
	projectRoot := func() string {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get working directory: %v", err)
		}

		return filepath.Dir(wd) // Go up from testutil to project root
	}()

	fixturesDir := filepath.Join(projectRoot, "testdata", "yaml-fixtures")

	for _, filename := range fixtureFiles {
		t.Run(filename, func(t *testing.T) {
			path := filepath.Join(fixturesDir, filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("fixture file does not exist: %s", path)
			}
		})
	}
}

// Tests for FixtureManager functionality (consolidated from scenarios.go tests)

func TestNewFixtureManager(t *testing.T) {
	fm := NewFixtureManager()
	if fm == nil {
		t.Fatal("expected fixture manager to be created")
	}

	if fm.basePath == "" {
		t.Error("expected basePath to be set")
	}

	if fm.scenarios == nil {
		t.Error("expected scenarios map to be initialized")
	}

	if fm.cache == nil {
		t.Error("expected cache map to be initialized")
	}
}

func TestFixtureManagerLoadScenarios(t *testing.T) {
	fm := NewFixtureManager()

	// Test loading scenarios (will create default if none exist)
	err := fm.LoadScenarios()
	if err != nil {
		t.Fatalf("failed to load scenarios: %v", err)
	}

	// Should have some default scenarios
	if len(fm.scenarios) == 0 {
		t.Error("expected default scenarios to be loaded")
	}
}

func TestFixtureManagerActionTypes(t *testing.T) {
	fm := NewFixtureManager()

	tests := []struct {
		name     string
		content  string
		expected ActionType
	}{
		{
			name:     "javascript action",
			content:  "using: 'node20'",
			expected: ActionTypeJavaScript,
		},
		{
			name:     "composite action",
			content:  "using: 'composite'",
			expected: ActionTypeComposite,
		},
		{
			name:     "docker action",
			content:  "using: 'docker'",
			expected: ActionTypeDocker,
		},
		{
			name:     "minimal action",
			content:  "name: test",
			expected: ActionTypeMinimal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualType := fm.determineActionTypeByContent(tt.content)
			if actualType != tt.expected {
				t.Errorf("expected action type %s, got %s", tt.expected, actualType)
			}
		})
	}
}

func TestFixtureManagerValidation(t *testing.T) {
	fm := NewFixtureManager()

	tests := []struct {
		name     string
		fixture  string
		expected bool
	}{
		{
			name:     "valid action",
			fixture:  "validation/valid-action.yml",
			expected: true,
		},
		{
			name:     "missing name",
			fixture:  "validation/missing-name.yml",
			expected: false,
		},
		{
			name:     "missing description",
			fixture:  "validation/missing-description.yml",
			expected: false,
		},
		{
			name:     "missing runs",
			fixture:  "validation/missing-runs.yml",
			expected: false,
		},
		{
			name:     "invalid yaml",
			fixture:  "validation/invalid-yaml.yml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := MustReadFixture(tt.fixture)
			isValid := fm.validateFixtureContent(content)
			if isValid != tt.expected {
				t.Errorf("expected validation result %v, got %v", tt.expected, isValid)
			}
		})
	}
}

func TestGetFixtureManager(t *testing.T) {
	// Test singleton behavior
	fm1 := GetFixtureManager()
	fm2 := GetFixtureManager()

	if fm1 != fm2 {
		t.Error("expected GetFixtureManager to return same instance")
	}

	if fm1 == nil {
		t.Fatal("expected fixture manager to be created")
	}
}

func TestActionFixtureLoading(t *testing.T) {
	// Test loading a fixture that should exist
	fixture, err := LoadActionFixture("simple-action.yml")
	if err != nil {
		t.Fatalf("failed to load simple action fixture: %v", err)
	}

	if fixture == nil {
		t.Fatal("expected fixture to be loaded")
	}

	if fixture.Name == "" {
		t.Error("expected fixture name to be set")
	}

	if fixture.Content == "" {
		t.Error("expected fixture content to be loaded")
	}

	if fixture.ActionType == "" {
		t.Error("expected action type to be determined")
	}
}

// Test helper functions for other components

func TestHelperFunctions(t *testing.T) {
	t.Run("GetValidFixtures", func(t *testing.T) {
		validFixtures := GetValidFixtures()
		if len(validFixtures) == 0 {
			t.Skip("no valid fixtures available")
		}

		for _, fixture := range validFixtures {
			if fixture == "" {
				t.Error("fixture name should not be empty")
			}
		}
	})

	t.Run("GetInvalidFixtures", func(t *testing.T) {
		invalidFixtures := GetInvalidFixtures()
		// It's okay if there are no invalid fixtures for testing

		for _, fixture := range invalidFixtures {
			if fixture == "" {
				t.Error("fixture name should not be empty")
			}
		}
	})

	t.Run("GetFixturesByActionType", func(_ *testing.T) {
		javascriptFixtures := GetFixturesByActionType(ActionTypeJavaScript)
		compositeFixtures := GetFixturesByActionType(ActionTypeComposite)
		dockerFixtures := GetFixturesByActionType(ActionTypeDocker)

		// We don't require specific fixtures to exist, just test the function works
		_ = javascriptFixtures
		_ = compositeFixtures
		_ = dockerFixtures
	})

	t.Run("GetFixturesByTag", func(_ *testing.T) {
		validTaggedFixtures := GetFixturesByTag("valid")
		invalidTaggedFixtures := GetFixturesByTag("invalid")
		basicTaggedFixtures := GetFixturesByTag("basic")

		// We don't require specific fixtures to exist, just test the function works
		_ = validTaggedFixtures
		_ = invalidTaggedFixtures
		_ = basicTaggedFixtures
	})
}
