// Package internal provides core logic for parsing, validating, and rendering
// documentation for GitHub Actions.
package internal

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// ActionYML represents a parsed GitHub Action's action.yml file.
// Fields correspond to the official action.yml schema.
type ActionYML struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	// LongDescription holds documentation lines parsed from comment blocks
	// between `# docs:start` and `# docs:end` in the action.yml file.
	LongDescription string                  `yaml:"-"`
	Inputs          map[string]ActionInput  `yaml:"inputs"`
	Outputs         map[string]ActionOutput `yaml:"outputs"`
	Runs            map[string]any          `yaml:"runs"`
	Branding        *Branding               `yaml:"branding,omitempty"`
	Dependencies    []ActionDependency      `yaml:"-"`
}

// ActionInput represents a single input parameter for a GitHub Action.
type ActionInput struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     any    `yaml:"default"`
}

// ActionOutput represents a single output parameter for a GitHub Action.
type ActionOutput struct {
	Description string `yaml:"description"`
}

// Branding represents the branding section of a GitHub Action.
type Branding struct {
	Icon  string `yaml:"icon"`
	Color string `yaml:"color"`
}

// ActionDependency represents an external action used in a composite action step.
type ActionDependency struct {
	Name    string
	Version string
	Ref     string
	Pinned  bool
}

// ParseActionYML parses the action.yml file at the given path and returns a pointer to ActionYML.
func ParseActionYML(path string) (*ActionYML, error) {
	cleanPath := filepath.Clean(path)
	f, err := os.Open(cleanPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := f.Close()
		if cerr != nil {
			logrus.Error("Failed to close file:", cleanPath)
		}
	}()
	var a ActionYML
	dec := yaml.NewDecoder(f)
	if decErr := dec.Decode(&a); decErr != nil {
		return nil, decErr
	}

	// Also parse optional documentation comments from the file.
	if docs, docErr := parseDocsFromFile(cleanPath); docErr == nil {
		a.LongDescription = docs
	}

	if deps, depErr := parseDependenciesFromFile(cleanPath); depErr == nil {
		a.Dependencies = deps
	}

	return &a, nil
}

// parseDocsFromFile reads the file and returns text between '# docs:start' and
// '# docs:end' comment markers. If no such block exists, an empty string is
// returned.
func parseDocsFromFile(path string) (string, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path validated by caller
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	var block []string
	inBlock := false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(trimmed, "# docs:start"):
			inBlock = true
		case strings.HasPrefix(trimmed, "# docs:end"):
			inBlock = false
			if len(block) > 0 {
				return strings.Join(block, "\n"), nil
			}
		default:
			if inBlock && strings.HasPrefix(trimmed, "#") {
				block = append(block, strings.TrimSpace(strings.TrimPrefix(trimmed, "#")))
			}
		}
	}

	if len(block) > 0 {
		return strings.Join(block, "\n"), nil
	}

	return "", nil
}

// parseDependenciesFromFile scans the action.yml file for external action dependencies.
// It looks for lines like `uses: actions/checkout@<ref> # <version>`.
func parseDependenciesFromFile(path string) ([]ActionDependency, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path validated by caller
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?i)uses:\s*(\S+)(?:\s*#\s*(.+))?`)
	shaRE := regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	lines := strings.Split(string(data), "\n")
	deps := make([]ActionDependency, 0)
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if !strings.Contains(trimmed, "uses:") {
			continue
		}
		m := re.FindStringSubmatch(trimmed)
		if len(m) < 2 {
			continue
		}
		spec := m[1]
		comment := ""
		if len(m) > 2 {
			comment = strings.TrimSpace(m[2])
		}
		if strings.HasPrefix(spec, "./") ||
			strings.HasPrefix(spec, "../") ||
			strings.HasPrefix(spec, "/") {
			continue // local action reference
		}
		parts := strings.SplitN(spec, "@", 2)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		ref := parts[1]
		pinned := shaRE.MatchString(ref)
		version := comment
		if version == "" {
			version = ref
		}
		deps = append(deps, ActionDependency{
			Name:    name,
			Version: version,
			Ref:     ref,
			Pinned:  pinned,
		})
	}

	return deps, nil
}
