package internal

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// newHeaderFooterGenerator builds a generator whose config starts from the real
// defaults, so the "configured value equals the built-in default" case — which is
// what distinguishes an inherited default from a deliberate override — is exercised
// exactly as production sees it.
func newHeaderFooterGenerator(theme, header, footer string) *Generator {
	cfg := DefaultAppConfig()
	cfg.Theme = theme
	if header != "" {
		cfg.Header = header
	}
	if footer != "" {
		cfg.Footer = footer
	}

	return NewGenerator(cfg)
}

// TestResolveHeaderFooterHTMLKeepsBuiltinDefaults guards the pre-existing HTML
// behavior: with no explicit override, HTML output still gets the bundled HTML
// scaffolding partials.
func TestResolveHeaderFooterHTMLKeepsBuiltinDefaults(t *testing.T) {
	t.Parallel()

	defaults := DefaultAppConfig()
	g := newHeaderFooterGenerator(appconstants.ThemeGitHub, "", "")

	header, footer := g.resolveHeaderFooter(appconstants.OutputFormatHTML)
	testutil.AssertEqual(t, defaults.Header, header)
	testutil.AssertEqual(t, defaults.Footer, footer)
}

// TestResolveHeaderFooterMarkdownRejectsHTMLDefaults is the regression guard for the
// most dangerous edge of making header/footer apply beyond HTML: the bundled partials
// are an HTML document's <head>/<body> scaffolding, and DefaultAppConfig seeds
// Header/Footer with them. They must never be injected into Markdown.
func TestResolveHeaderFooterMarkdownRejectsHTMLDefaults(t *testing.T) {
	t.Parallel()

	g := newHeaderFooterGenerator(appconstants.ThemeGitHub, "", "")

	for _, format := range []string{appconstants.OutputFormatMarkdown, appconstants.OutputFormatASCIIDoc} {
		header, footer := g.resolveHeaderFooter(format)
		if header != "" {
			t.Errorf("%s header = %q, want empty (built-in HTML partial must not leak)", format, header)
		}
		if footer != "" {
			t.Errorf("%s footer = %q, want empty (built-in HTML partial must not leak)", format, footer)
		}
	}
}

// TestResolveHeaderFooterExplicitOverrideAppliesToAllFormats covers the requested
// behavior: an explicitly configured partial is honored for Markdown and AsciiDoc,
// not only HTML.
func TestResolveHeaderFooterExplicitOverrideAppliesToAllFormats(t *testing.T) {
	t.Parallel()

	const custom = "custom/my-footer.tmpl"
	g := newHeaderFooterGenerator(appconstants.ThemeGitHub, "", custom)

	for _, format := range []string{
		appconstants.OutputFormatMarkdown,
		appconstants.OutputFormatASCIIDoc,
		appconstants.OutputFormatHTML,
	} {
		_, footer := g.resolveHeaderFooter(format)
		testutil.AssertEqual(t, custom, footer)
	}
}

// TestResolveHeaderFooterPartialOverrideIsIndependent is the "partial overwrite"
// requirement: overriding only the footer must leave the header resolution alone.
func TestResolveHeaderFooterPartialOverrideIsIndependent(t *testing.T) {
	t.Parallel()

	defaults := DefaultAppConfig()
	const customFooter = "custom/only-footer.tmpl"
	g := newHeaderFooterGenerator(appconstants.ThemeGitHub, "", customFooter)

	header, footer := g.resolveHeaderFooter(appconstants.OutputFormatHTML)
	testutil.AssertEqual(t, customFooter, footer)
	// Header untouched by the footer override.
	testutil.AssertEqual(t, defaults.Header, header)
}

// TestResolveThemePartialAbsentTheme confirms an unknown theme and a theme with no
// partials both resolve to "" rather than a bogus path.
func TestResolveThemePartialAbsentTheme(t *testing.T) {
	t.Parallel()

	testutil.AssertEqual(t, "", resolveThemePartial("no-such-theme", appconstants.ThemePartialHeader))
	// No bundled theme currently ships partials, so this is "" today; the assertion
	// documents that absence is handled, not that partials are unsupported.
	testutil.AssertEqual(t, "", resolveThemePartial(appconstants.ThemeGitHub, appconstants.ThemePartialHeader))
}
