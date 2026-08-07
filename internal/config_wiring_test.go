package internal

import (
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"

	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestResolvePermissionsFallback pins the config-level `permissions:` semantics: the
// action's own declaration wins, and the config block is a fallback for actions that
// declare none. Before this, the config key was accepted, merged, and then read by
// nothing.
func TestResolvePermissionsFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		actionPerms  PermissionMap
		configPerms  map[string]string
		wantScope    string
		wantLevel    string
		wantEmptySet bool
	}{
		{
			name:        "action declaration wins",
			actionPerms: PermissionMap{testutil.PermissionContents: appconstants.PermissionWrite},
			configPerms: map[string]string{
				testutil.PermissionContents: appconstants.PermissionRead,
				"issues":                    appconstants.PermissionRead,
			},
			wantScope: testutil.PermissionContents,
			wantLevel: appconstants.PermissionWrite,
		},
		{
			name:        "config fills in when action declares none",
			actionPerms: nil,
			configPerms: map[string]string{testutil.PermissionContents: appconstants.PermissionRead},
			wantScope:   testutil.PermissionContents,
			wantLevel:   appconstants.PermissionRead,
		},
		{
			name:         "neither declares any",
			actionPerms:  nil,
			configPerms:  nil,
			wantEmptySet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resolvePermissions(
				&ActionYML{Permissions: tt.actionPerms},
				&AppConfig{Permissions: tt.configPerms},
			)
			if tt.wantEmptySet {
				testutil.AssertEqual(t, 0, len(got))

				return
			}
			testutil.AssertEqual(t, tt.wantLevel, got[tt.wantScope])
		})
	}
}

// TestResolvePermissionsDoesNotMutateAction guards the shadowing choice: the resolved
// set must not be written back onto the parsed action, or the fallback would leak
// across output formats within one run.
func TestResolvePermissionsDoesNotMutateAction(t *testing.T) {
	t.Parallel()

	action := &ActionYML{}
	config := &AppConfig{Permissions: map[string]string{testutil.PermissionContents: appconstants.PermissionRead}}

	resolved := resolvePermissions(action, config)
	testutil.AssertEqual(t, appconstants.PermissionRead, resolved[testutil.PermissionContents])
	if len(action.Permissions) != 0 {
		t.Errorf("action.Permissions was mutated: %v", action.Permissions)
	}

	// The copy must be independent of the config map too.
	resolved[testutil.PermissionContents] = appconstants.PermissionWrite
	testutil.AssertEqual(t, appconstants.PermissionRead, config.Permissions[testutil.PermissionContents])
}

// TestRunsOnValue covers the `runs_on` config key reaching the workflow examples.
func TestRunsOnValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		runsOn []string
		want   string
	}{
		{"empty falls back to the default runner", nil, appconstants.RunnerUbuntuLatest},
		{"single runner renders as a scalar", []string{testutil.RunnerMacosLatest}, testutil.RunnerMacosLatest},
		{
			"multiple runners render as a YAML flow sequence",
			[]string{appconstants.RunnerUbuntuLatest, testutil.RunnerMacosLatest},
			"[" + appconstants.RunnerUbuntuLatest + ", " + testutil.RunnerMacosLatest + "]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := DefaultAppConfig()
			cfg.RunsOn = tt.runsOn
			testutil.AssertEqual(t, tt.want, runsOnValue(&TemplateData{Config: cfg}))
		})
	}
}

// TestConfigVariable covers the `variables:` config key becoming reachable from
// templates via the var function.
func TestConfigVariable(t *testing.T) {
	t.Parallel()

	cfg := DefaultAppConfig()
	cfg.Variables = map[string]string{"support_url": "https://example.com/support"}
	td := &TemplateData{Config: cfg}

	testutil.AssertEqual(t, "https://example.com/support", configVariable(td, "support_url"))
	// An unknown key yields "" so {{with var . "k"}} can guard the block.
	testutil.AssertEqual(t, "", configVariable(td, "absent"))
	// Non-TemplateData input must not panic.
	testutil.AssertEqual(t, "", configVariable("not template data", "support_url"))
}

// TestLicenseBadgeRendersWithoutBranding guards the professional theme's badge
// container: the license image used to sit inside {{if .Branding}}, so a licensed
// action with no branding silently lost its badge. Rendered directly rather than
// through the CLI because FillMissing injects a default branding icon, which masks
// the coupling end-to-end.
func TestLicenseBadgeRendersWithoutBranding(t *testing.T) {
	t.Parallel()

	td := &TemplateData{
		ActionYML: &ActionYML{Name: "NoBrand", Description: "d"},
		Config:    DefaultAppConfig(),
		License:   appconstants.SPDXMIT,
	}

	out, err := RenderReadme(td, TemplateOptions{
		TemplatePath: "templates/themes/professional/readme.tmpl",
		Format:       "md",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if !strings.Contains(out, "badge/license-MIT") {
		t.Errorf("license badge missing for a licensed action without branding:\n%s", out[:400])
	}
	if strings.Contains(out, "badge/icon-") {
		t.Error("icon badge rendered for an action with no branding")
	}
}

// TestShowSecurityInfoRendersSection covers the `show_security_info` config key: it
// gates a Security section that reports dependency pinning, and is silent when off.
func TestShowSecurityInfoRendersSection(t *testing.T) {
	t.Parallel()

	deps := []dependencies.Dependency{
		{Name: "actions/checkout", Uses: "actions/checkout@v4", IsPinned: false},
		{Name: "actions/cache", Uses: "actions/cache@v3.2.1", IsPinned: true},
	}

	for _, show := range []bool{true, false} {
		cfg := DefaultAppConfig()
		cfg.ShowSecurityInfo = show

		td := &TemplateData{
			ActionYML:    &ActionYML{Name: "Sec", Description: "d"},
			Config:       cfg,
			Dependencies: deps,
		}

		out, err := RenderReadme(td, TemplateOptions{
			TemplatePath: "templates/themes/github/readme.tmpl",
			Format:       "md",
		})
		if err != nil {
			t.Fatalf("render (show=%v): %v", show, err)
		}

		hasSection := strings.Contains(out, "Security")
		if hasSection != show {
			t.Errorf("show_security_info=%v produced Security section=%v, want %v", show, hasSection, show)
		}
		if show && !strings.Contains(out, "1 of 2 dependencies are pinned") {
			t.Errorf("security section missing the pinned/total count:\n%s", out)
		}
	}
}
