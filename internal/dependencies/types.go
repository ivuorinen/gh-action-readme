package dependencies

// CompositeStep represents a step in a composite action.
type CompositeStep struct {
	Name  string            `yaml:"name,omitempty"`
	Uses  string            `yaml:"uses,omitempty"`
	With  map[string]any    `yaml:"with,omitempty"`
	Run   string            `yaml:"run,omitempty"`
	Shell string            `yaml:"shell,omitempty"`
	Env   map[string]string `yaml:"env,omitempty"`
}

// CompositeRuns represents the runs section of a composite action.
type CompositeRuns struct {
	Using string          `yaml:"using"`
	Steps []CompositeStep `yaml:"steps"`
}

// ActionWithComposite represents an action.yml with composite steps support.
type ActionWithComposite struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Inputs      map[string]any `yaml:"inputs"`
	Outputs     map[string]any `yaml:"outputs"`
	Runs        CompositeRuns  `yaml:"runs"`
	Branding    any            `yaml:"branding,omitempty"`
}
