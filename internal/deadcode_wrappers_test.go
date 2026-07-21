package internal

// Test-only wrappers. These were exported production functions that production
// code no longer calls (superseded by ProcessBatch / ConfigurationLoader). They
// remain as test helpers so the tests that use them to exercise the live
// internals (generateFromFile, loadAndUnmarshalConfig) keep their coverage
// without keeping dead code in the shipped binary. See docs/audit findings N-161.

// GenerateFromFile processes a single action.yml file and generates documentation.
func (g *Generator) GenerateFromFile(actionPath string) error {
	return g.generateFromFile(actionPath, false)
}

// InitConfig initializes a configuration using Viper with XDG compliance.
func InitConfig(configFile string) (*AppConfig, error) {
	v, err := initializeViperInstance()
	if err != nil {
		return nil, err
	}

	return loadAndUnmarshalConfig(configFile, v)
}
