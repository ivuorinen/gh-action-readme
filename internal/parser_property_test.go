package internal

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestPermissionMergingProperties verifies properties of permission merging.
//
//nolint:gocyclo // Property-based test with multiple properties
func TestPermissionMergingProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1: YAML always wins conflicts
	properties.Property("YAML permissions override comment permissions",
		prop.ForAll(
			func(key, yamlVal, commentVal string) bool {
				// Only test when values differ and key is non-empty
				if yamlVal == commentVal || yamlVal == "" || key == "" || commentVal == "" {
					return true
				}

				action := &ActionYML{
					Permissions: map[string]string{key: yamlVal},
				}

				commentPerms := map[string]string{key: commentVal}
				mergePermissions(action, commentPerms)

				// YAML value should be preserved
				return action.Permissions[key] == yamlVal
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	// Property 2: All unique keys preserved
	properties.Property("merge preserves all non-conflicting keys",
		prop.ForAll(
			func(yamlKey, commentKey, val string) bool {
				// Only test when keys are different and non-empty
				if yamlKey == commentKey || yamlKey == "" || commentKey == "" || val == "" {
					return true
				}

				action := &ActionYML{
					Permissions: map[string]string{yamlKey: val},
				}

				commentPerms := map[string]string{commentKey: val}
				mergePermissions(action, commentPerms)

				// Both keys should exist
				_, hasYaml := action.Permissions[yamlKey]
				_, hasComment := action.Permissions[commentKey]

				return hasYaml && hasComment
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	// Property 3: Identity - merging with nil preserves original
	properties.Property("merging with nil preserves original permissions",
		prop.ForAll(
			func(key, value string) bool {
				return verifyMergePreservesOriginal(key, value, nil)
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	// Property 4: Identity - merging with empty map preserves original
	properties.Property("merging with empty map preserves original permissions",
		prop.ForAll(
			func(key, value string) bool {
				return verifyMergePreservesOriginal(key, value, make(map[string]string))
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	// Property 5: Result size bounded
	properties.Property("merged permissions size bounded by sum of inputs",
		prop.ForAll(
			func(yamlKeys []string, commentKeys []string, value string) bool {
				if len(yamlKeys) == 0 || len(commentKeys) == 0 || value == "" {
					return true
				}

				// Build YAML permissions
				yamlPerms := make(map[string]string)
				for _, key := range yamlKeys {
					if key != "" {
						yamlPerms[key] = value
					}
				}

				// Build comment permissions
				commentPerms := make(map[string]string)
				for _, key := range commentKeys {
					if key != "" {
						commentPerms[key] = value
					}
				}

				action := &ActionYML{
					Permissions: yamlPerms,
				}

				mergePermissions(action, commentPerms)

				// Result size should be at most sum of input sizes
				// (could be less if there are duplicates)
				maxSize := len(yamlPerms) + len(commentPerms)

				return len(action.Permissions) <= maxSize
			},
			gen.SliceOf(gen.AlphaString()),
			gen.SliceOf(gen.AlphaString()),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// TestActionYMLNilPermissionsProperties verifies behavior when permissions is nil.
func TestActionYMLNilPermissionsProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1: Merging into nil permissions creates new map
	properties.Property("merging into nil permissions creates new map",
		prop.ForAll(
			func(key, value string) bool {
				if key == "" || value == "" {
					return true
				}

				action := &ActionYML{
					Permissions: nil,
				}

				commentPerms := map[string]string{key: value}
				mergePermissions(action, commentPerms)

				// Should create new map with comment permissions
				if action.Permissions == nil {
					return false
				}

				return action.Permissions[key] == value
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	// Property 2: Nil action permissions stays nil when merging with nil
	properties.Property("nil permissions stays nil when merging with nil",
		prop.ForAll(
			func() bool {
				action := &ActionYML{
					Permissions: nil,
				}

				mergePermissions(action, nil)

				// Should remain nil (no map created)
				return action.Permissions == nil
			},
		),
	)

	properties.TestingRun(t)
}

// TestCommentPermissionsOnlyProperties verifies behavior when only comment permissions exist.
//
//nolint:gocyclo // Property-based test with multiple properties
func TestCommentPermissionsOnlyProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)
	registerCommentPermissionsOnlyProperties(properties)
	properties.TestingRun(t)
}

func registerCommentPermissionsOnlyProperties(properties *gopter.Properties) {
	// Property: All comment permissions transferred when YAML is nil
	properties.Property("all comment permissions transferred when YAML is nil",
		prop.ForAll(
			verifyCommentPermissionsTransferred,
			gen.SliceOf(gen.AlphaString().SuchThat(func(s string) bool { return s != "" })),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)
}

func verifyCommentPermissionsTransferred(keys []string, value string) bool {
	if len(keys) == 0 || value == "" {
		return true
	}

	// Build comment permissions
	commentPerms := make(map[string]string)
	for _, key := range keys {
		if key != "" {
			commentPerms[key] = value
		}
	}

	if len(commentPerms) == 0 {
		return true
	}

	action := &ActionYML{
		Permissions: nil,
	}

	mergePermissions(action, commentPerms)

	// All comment permissions should be in action
	if action.Permissions == nil {
		return false
	}

	for key, val := range commentPerms {
		if action.Permissions[key] != val {
			return false
		}
	}

	return true
}

// verifyMergePreservesOriginal is a helper to test that merging with
// nil or empty permissions preserves the original permissions.
func verifyMergePreservesOriginal(key, value string, mergeWith map[string]string) bool {
	if key == "" || value == "" {
		return true
	}

	action := &ActionYML{
		Permissions: map[string]string{key: value},
	}

	// Make a copy to compare
	originalPerms := make(map[string]string)
	for k, v := range action.Permissions {
		originalPerms[k] = v
	}

	// Merge with provided map (nil or empty)
	mergePermissions(action, mergeWith)

	// Should be unchanged
	if len(action.Permissions) != len(originalPerms) {
		return false
	}

	for k, v := range originalPerms {
		if action.Permissions[k] != v {
			return false
		}
	}

	return true
}
