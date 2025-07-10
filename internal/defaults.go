// Package internal provides core logic for parsing, validating, and generating
// documentation for GitHub Actions.
package internal

// DefaultValues holds default values for action.yml fields, as loaded from config.yaml.
// Used to autofill missing fields in ActionYML.
type DefaultValues struct {
	Name        string
	Description string
	Runs        map[string]any
	Branding    Branding
	Version     string
}

// FillMissing sets missing fields in the given ActionYML struct using the provided DefaultValues.
// Fields filled: Name, Description, Runs, Branding.
// Note: Version is not set on ActionYML, only used for output.
func FillMissing(action *ActionYML, defs DefaultValues) {
	if action.Name == "" {
		action.Name = defs.Name
	}
	if action.Description == "" {
		action.Description = defs.Description
	}
	if len(action.Runs) == 0 && len(defs.Runs) > 0 {
		action.Runs = defs.Runs
	}
	if action.Branding == nil && defs.Branding.Icon != "" {
		action.Branding = &defs.Branding
	}
}
