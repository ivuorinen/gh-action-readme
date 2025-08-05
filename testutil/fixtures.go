// Package testutil provides testing fixtures for gh-action-readme.
package testutil

import (
	"os"
	"path/filepath"
	"runtime"
)

// mustReadFixture reads a YAML fixture file from testdata/yaml-fixtures.
func mustReadFixture(filename string) string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}

	// Get the project root (go up from testutil/fixtures.go to project root)
	projectRoot := filepath.Dir(filepath.Dir(currentFile))
	fixturePath := filepath.Join(projectRoot, "testdata", "yaml-fixtures", filename)

	content, err := os.ReadFile(fixturePath)
	if err != nil {
		panic("failed to read fixture " + filename + ": " + err.Error())
	}

	return string(content)
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

// Sample action.yml files for testing.

// SimpleActionYML is a basic GitHub Action YAML.
var SimpleActionYML = mustReadFixture("simple-action.yml")

// CompositeActionYML is a composite GitHub Action with dependencies.
var CompositeActionYML = mustReadFixture("composite-action.yml")

// DockerActionYML is a Docker-based GitHub Action.
var DockerActionYML = mustReadFixture("docker-action.yml")

// InvalidActionYML is an invalid action.yml for error testing.
var InvalidActionYML = mustReadFixture("invalid-action.yml")

// MinimalActionYML is a minimal valid action.yml.
var MinimalActionYML = mustReadFixture("minimal-action.yml")

// TestProjectActionYML is used for integration tests.
var TestProjectActionYML = mustReadFixture("test-project-action.yml")

// Configuration file fixtures.

// DefaultConfigYAML is a default configuration file.
const DefaultConfigYAML = `theme: github
output_format: md
output_dir: .
verbose: false
quiet: false
`

// CustomConfigYAML is a custom configuration file.
const CustomConfigYAML = `theme: professional
output_format: html
output_dir: docs
template: custom-template.tmpl
schema: custom-schema.json
verbose: true
quiet: false
github_token: test-token-from-config
`

// RepoSpecificConfigYAML is a repository-specific configuration.
var RepoSpecificConfigYAML = mustReadFixture("repo-config.yml")

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
var PackageJSONContent = mustReadFixture("package.json")
