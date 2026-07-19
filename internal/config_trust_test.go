package internal

import "testing"

// TestMergeConfigsUntrustedPathFields verifies the S1 trust-gate fix: an untrusted
// source (processed repo / action config, allowTokens=false) must NOT be able to set
// the local-file-path fields Template/Header/Footer, since ReadTemplate later opens
// them. Benign fields such as Theme still merge from any source.
func TestMergeConfigsUntrustedPathFields(t *testing.T) {
	t.Parallel()

	src := &AppConfig{
		Theme:    "github",
		Template: "/etc/hosts",
		Header:   "/etc/hosts",
		Footer:   "/home/victim/.ssh/id_rsa",
	}

	t.Run("untrusted source drops path fields", func(t *testing.T) {
		t.Parallel()
		dst := &AppConfig{}
		MergeConfigs(dst, src, false)

		if dst.Theme != "github" {
			t.Errorf("Theme should merge from any source: got %q", dst.Theme)
		}
		if dst.Template != "" || dst.Header != "" || dst.Footer != "" {
			t.Errorf("untrusted path fields must not merge: Template=%q Header=%q Footer=%q",
				dst.Template, dst.Header, dst.Footer)
		}
	})

	t.Run("trusted source keeps path fields", func(t *testing.T) {
		t.Parallel()
		dst := &AppConfig{}
		MergeConfigs(dst, src, true)

		if dst.Template != src.Template || dst.Header != src.Header || dst.Footer != src.Footer {
			t.Errorf("trusted path fields must merge: Template=%q Header=%q Footer=%q",
				dst.Template, dst.Header, dst.Footer)
		}
	})
}
