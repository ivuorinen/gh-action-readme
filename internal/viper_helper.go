package internal

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// initializeViperInstance creates and configures a new viper instance with standard settings.
// This includes XDG-compliant configuration paths, environment variable support,
// and standard search paths for configuration files.
func initializeViperInstance() (*viper.Viper, error) {
	v := viper.New()

	// Set configuration file name and type
	v.SetConfigName(appconstants.ConfigFileName)
	v.SetConfigType(appconstants.OutputFormatYAML)

	// Add XDG-compliant configuration directory
	configDir, err := xdg.ConfigFile(appconstants.AppName)
	if err != nil {
		return nil, fmt.Errorf(appconstants.ErrFailedToGetXDGConfigDir, err)
	}
	v.AddConfigPath(filepath.Dir(configDir))

	// Add additional search paths
	v.AddConfigPath(".")                         // current directory
	v.AddConfigPath(appconstants.PathHomeConfig) // fallback
	v.AddConfigPath(appconstants.PathEtcConfig)  // system-wide

	// Set environment variable prefix
	v.SetEnvPrefix(appconstants.EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	return v, nil
}

// setConfigDefaults sets all default configuration values in the viper instance.
// This ensures consistent default values across all configuration loading scenarios.
func setConfigDefaults(v *viper.Viper, defaults *AppConfig) {
	v.SetDefault(appconstants.ConfigKeyOrganization, defaults.Organization)
	v.SetDefault(appconstants.ConfigKeyRepository, defaults.Repository)
	v.SetDefault(appconstants.ConfigKeyVersion, defaults.Version)
	v.SetDefault(appconstants.ConfigKeyTheme, defaults.Theme)
	v.SetDefault(appconstants.ConfigKeyOutputFormat, defaults.OutputFormat)
	v.SetDefault(appconstants.ConfigKeyOutputDir, defaults.OutputDir)
	v.SetDefault(appconstants.ConfigKeyTemplate, defaults.Template)
	v.SetDefault(appconstants.ConfigKeyHeader, defaults.Header)
	v.SetDefault(appconstants.ConfigKeyFooter, defaults.Footer)
	v.SetDefault(appconstants.ConfigKeySchema, defaults.Schema)
	v.SetDefault(appconstants.ConfigKeyAnalyzeDependencies, defaults.AnalyzeDependencies)
	v.SetDefault(appconstants.ConfigKeyShowSecurityInfo, defaults.ShowSecurityInfo)
	v.SetDefault(appconstants.ConfigKeyVerbose, defaults.Verbose)
	v.SetDefault(appconstants.ConfigKeyQuiet, defaults.Quiet)
	v.SetDefault(appconstants.ConfigKeyDefaultsName, defaults.Defaults.Name)
	v.SetDefault(appconstants.ConfigKeyDefaultsDescription, defaults.Defaults.Description)
	v.SetDefault(appconstants.ConfigKeyDefaultsBrandingIcon, defaults.Defaults.Branding.Icon)
	v.SetDefault(appconstants.ConfigKeyDefaultsBrandingColor, defaults.Defaults.Branding.Color)
}

// loadConfigFromViper loads an AppConfig from a specified YAML config file using viper.
func loadConfigFromViper(configPath string) (*AppConfig, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType(appconstants.OutputFormatYAML)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", configPath, err)
	}

	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf(appconstants.ErrFailedToUnmarshalConfig, err)
	}

	return &config, nil
}
