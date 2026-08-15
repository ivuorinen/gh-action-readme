package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// runVersionLikeCmd wires cmd under a root carrying the same persistent
// --verbose flag main() registers, then executes it with args and returns
// everything written to stdout. Going through Execute rather than calling Run
// directly is what makes cmd.Flags().GetBool see the inherited persistent flag,
// which is how the command behaves in the real CLI.
func runVersionLikeCmd(t *testing.T, cmd *cobra.Command, args ...string) string {
	t.Helper()

	root := &cobra.Command{Use: "gh-action-readme"}
	root.PersistentFlags().BoolP(appconstants.ConfigKeyVerbose, "v", false, "verbose output")
	root.AddCommand(cmd)
	root.SetArgs(args)

	return testutil.CaptureStdout(func() {
		if err := root.Execute(); err != nil {
			t.Errorf("Execute() unexpected error: %v", err)
		}
	})
}

// setBuildInfo replaces the ldflags-injected build variables for the duration of
// a test, so assertions can name exact values instead of the "dev"/"unknown"
// defaults — which is what makes a swapped field detectable.
func setBuildInfo(t *testing.T, v, c, d, b string) {
	t.Helper()

	origV, origC, origD, origB := version, commit, date, builtBy
	t.Cleanup(func() { version, commit, date, builtBy = origV, origC, origD, origB })

	version, commit, date, builtBy = v, c, d, b
}

// TestVersionCmdPlain covers the default branch: the bare version and nothing
// else, so `gh-action-readme version` stays usable in a shell substitution.
func TestVersionCmdPlain(t *testing.T) {
	setBuildInfo(t, "1.2.3", "abc1234", "2026-01-02", "goreleaser")

	out := runVersionLikeCmd(t, newVersionCmd(), appconstants.CommandVersion)

	if strings.TrimSpace(out) != "1.2.3" {
		t.Errorf("plain version output = %q, want %q", strings.TrimSpace(out), "1.2.3")
	}
}

// TestVersionCmdVerbose covers the --verbose branch. Each build field is
// asserted against a distinct value: the four Printf lines are near-identical,
// which is exactly where a copy-paste swap hides.
func TestVersionCmdVerbose(t *testing.T) {
	setBuildInfo(t, "1.2.3", "abc1234", "2026-01-02", "goreleaser")

	out := runVersionLikeCmd(t, newVersionCmd(), appconstants.CommandVersion, "--verbose")

	for _, want := range []string{
		"gh-action-readme version 1.2.3",
		"commit: abc1234",
		"built at: 2026-01-02",
		"built by: goreleaser",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("verbose output missing %q; got:\n%s", want, out)
		}
	}
}

// TestAboutCmd covers the about command's single output line.
func TestAboutCmd(t *testing.T) {
	out := runVersionLikeCmd(t, newAboutCmd(), "about")

	if !strings.Contains(out, "Generates README.md and HTML for GitHub Actions") {
		t.Errorf("about output = %q, want the tool description", out)
	}
	if !strings.Contains(out, "MIT License") {
		t.Errorf("about output = %q, want the license named", out)
	}
}
