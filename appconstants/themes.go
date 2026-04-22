// Package appconstants provides common constants used throughout the application.
package appconstants

// Theme identifier constants.
const (
	// ThemeGitHub is the GitHub theme identifier.
	ThemeGitHub = "github"
	// ThemeGitLab is the GitLab theme identifier.
	ThemeGitLab = "gitlab"
	// ThemeMinimal is the minimal theme identifier.
	ThemeMinimal = "minimal"
	// ThemeProfessional is the professional theme identifier.
	ThemeProfessional = "professional"
	// ThemeDefault is the default theme identifier.
	ThemeDefault = "default"
)

// supportedThemes lists all available theme names (unexported to prevent modification).
var supportedThemes = []string{
	ThemeDefault,
	ThemeGitHub,
	ThemeGitLab,
	ThemeMinimal,
	ThemeProfessional,
}

// GetSupportedThemes returns a copy of the supported theme names.
// Returns a new slice to prevent external modification of the internal list.
func GetSupportedThemes() []string {
	themes := make([]string, len(supportedThemes))
	copy(themes, supportedThemes)

	return themes
}

// Template path constants.
const (
	// TemplatePathDefault is the default template path.
	TemplatePathDefault = "templates/readme.tmpl"
	// TemplatePathGitHub is the GitHub theme template path.
	TemplatePathGitHub = "templates/themes/github/readme.tmpl"
	// TemplatePathGitLab is the GitLab theme template path.
	TemplatePathGitLab = "templates/themes/gitlab/readme.tmpl"
	// TemplatePathMinimal is the minimal theme template path.
	TemplatePathMinimal = "templates/themes/minimal/readme.tmpl"
	// TemplatePathProfessional is the professional theme template path.
	TemplatePathProfessional = "templates/themes/professional/readme.tmpl"
	// TemplateNameReadme is the template name used in template.New().
	TemplateNameReadme = "readme"
)

// Output format constants.
const (
	// OutputFormatMarkdown is the Markdown output format.
	OutputFormatMarkdown = "md"
	// OutputFormatHTML is the HTML output format.
	OutputFormatHTML = "html"
	// OutputFormatJSON is the JSON output format.
	OutputFormatJSON = "json"
	// OutputFormatYAML is the YAML output format.
	OutputFormatYAML = "yaml"
	// OutputFormatTOML is the TOML output format.
	OutputFormatTOML = "toml"
	// OutputFormatASCIIDoc is the AsciiDoc output format.
	OutputFormatASCIIDoc = "asciidoc"
)

// supportedOutputFormats lists all available output format names (unexported to prevent modification).
var supportedOutputFormats = []string{
	OutputFormatMarkdown,
	OutputFormatHTML,
	OutputFormatJSON,
	OutputFormatASCIIDoc,
}

// GetSupportedOutputFormats returns a copy of the supported output format names.
// Returns a new slice to prevent external modification of the internal list.
func GetSupportedOutputFormats() []string {
	formats := make([]string, len(supportedOutputFormats))
	copy(formats, supportedOutputFormats)

	return formats
}

// UI and display format constants.
const (
	// FormatKeyValue is the key-value format string.
	FormatKeyValue = "%s: %s"
	// FormatDetailKeyValue is the detailed key-value format string.
	FormatDetailKeyValue = "  %s: %s"
	// FormatPrompt is the prompt format string.
	FormatPrompt = "%s: "
	// FormatPromptDefault is the prompt with default format string.
	FormatPromptDefault = "%s [%s]: "
	// FormatEnvVar is the environment variable format string.
	FormatEnvVar = "%s = %q\n"
)

// UI display constants.
const (
	// SymbolArrow is the arrow symbol for UI.
	SymbolArrow = "►"
)
