package internal

// Test-only wrappers. These were exported production functions that production
// code no longer calls (superseded by ProcessBatch / ConfigurationLoader). They
// remain as test helpers so the tests that use them to exercise the live
// internals (generateFromFile, loadAndUnmarshalConfig, buildTemplateData) keep
// their coverage without keeping dead code in the shipped binary. See docs/audit
// findings N-161 and unwired-8e98592a.
//
// Adding one here rather than in production code keeps `deadcode -test=false ./...`
// empty outside testutil, so that command stays usable as a wiring gate.

// GenerateFromFile processes a single action.yml file and generates documentation.
func (g *Generator) GenerateFromFile(actionPath string) error {
	return g.generateFromFile(actionPath, false)
}

// BuildTemplateData constructs comprehensive template data from action and
// configuration, using standalone (non-shared) dependency resources.
func BuildTemplateData(action *ActionYML, config *AppConfig, repoRoot, actionPath string) *TemplateData {
	return buildTemplateData(action, config, repoRoot, actionPath, nil)
}

// InitConfig initializes a configuration using Viper with XDG compliance.
func InitConfig(configFile string) (*AppConfig, error) {
	v, err := initializeViperInstance()
	if err != nil {
		return nil, err
	}

	return loadAndUnmarshalConfig(configFile, v)
}
