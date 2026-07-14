package internal

// Test-only wrappers. These were exported production functions that production
// code no longer calls (superseded by ProcessBatch / ConfigurationLoader). They
// remain as test helpers so the tests that use them to exercise the live
// internals (generateFromFile, loadAndUnmarshalConfig, the source map) keep
// their coverage without keeping dead code in the shipped binary. See
// docs/audit findings N-161.

import "github.com/ivuorinen/gh-action-readme/appconstants"

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

// NewConfigurationLoaderWithOptions creates a configuration loader with custom options.
func NewConfigurationLoaderWithOptions(opts ConfigurationOptions) *ConfigurationLoader {
	loader := &ConfigurationLoader{
		sources: make(map[appconstants.ConfigurationSource]bool),
	}

	if len(opts.EnabledSources) == 0 {
		opts.EnabledSources = []appconstants.ConfigurationSource{
			appconstants.SourceDefaults, appconstants.SourceGlobal, appconstants.SourceRepoOverride,
			appconstants.SourceRepoConfig, appconstants.SourceActionConfig, appconstants.SourceEnvironment,
		}
	}

	for _, source := range opts.EnabledSources {
		loader.sources[source] = true
	}

	return loader
}

// GetConfigurationSources returns the currently enabled configuration sources.
func (cl *ConfigurationLoader) GetConfigurationSources() []appconstants.ConfigurationSource {
	var sources []appconstants.ConfigurationSource
	for source, enabled := range cl.sources {
		if enabled {
			sources = append(sources, source)
		}
	}

	return sources
}

// EnableSource enables a specific configuration source.
func (cl *ConfigurationLoader) EnableSource(source appconstants.ConfigurationSource) {
	cl.sources[source] = true
}

// DisableSource disables a specific configuration source.
func (cl *ConfigurationLoader) DisableSource(source appconstants.ConfigurationSource) {
	cl.sources[source] = false
}
