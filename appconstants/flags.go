// Package appconstants provides common constants used throughout the application.
package appconstants

// CLI flag and command names.
const (
	// FlagFormat is the format flag name.
	FlagFormat = "format"
	// FlagOutputDir is the output-dir flag name.
	FlagOutputDir = "output-dir"
	// FlagOutputFormat is the output-format flag name.
	FlagOutputFormat = "output-format"
	// FlagOutput is the output flag name.
	FlagOutput = "output"
	// FlagRecursive is the recursive flag name.
	FlagRecursive = "recursive"
	// FlagIgnoreDirs is the ignore-dirs flag name.
	FlagIgnoreDirs = "ignore-dirs"
	// FlagCI is the CI mode flag name.
	FlagCI = "ci"

	// CommandPin is the pin command name.
	CommandPin = "pin"
	// CommandGen is the gen command name.
	CommandGen = "gen"
	// CommandCache is the cache command name.
	CommandCache = "cache"
	// CommandStats is the stats subcommand name.
	CommandStats = "stats"
	// CommandConfig is the config command name.
	CommandConfig = "config"
	// CommandWizard is the wizard subcommand name.
	CommandWizard = "wizard"
	// CommandShow is the show subcommand name.
	CommandShow = "show"
	// CommandThemes is the themes subcommand name.
	CommandThemes = "themes"
	// CommandDeps is the deps command name.
	CommandDeps = "deps"
	// CommandList is the list subcommand name.
	CommandList = "list"
	// CommandValidate is the validate command name.
	CommandValidate = "validate"
	// CommandSchema is the schema command name.
	CommandSchema = "schema"
	// CommandVersion is the version command name.
	CommandVersion = "version"

	// CacheStatsKeyDir is the cache stats key for directory.
	CacheStatsKeyDir = "cache_dir"

	// CachePathUnknown is the placeholder shown when the cache directory cannot be
	// resolved from cache stats. It is not a real filesystem path and must not be
	// passed to os.Stat.
	CachePathUnknown = "<unknown>"
)
