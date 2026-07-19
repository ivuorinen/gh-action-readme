package internal

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// permissionScopeGen draws a real per-scope permission key, excluding the
// reserved `all` global scope (which the scalar shorthand owns and which
// mergePermissions treats as authoritative, not a mergeable per-scope key).
func permissionScopeGen() gopter.Gen {
	return gen.OneConstOf("contents", "issues", "pull-requests", "actions", "packages", "deployments")
}

// permissionLevelGen draws a real permission level, matching the domain the
// parser actually accepts rather than arbitrary alpha strings.
func permissionLevelGen() gopter.Gen {
	return gen.OneConstOf(appconstants.PermissionRead, appconstants.PermissionWrite, "none")
}

// TestPermissionMergingProperties verifies properties of permission merging.
func TestPermissionMergingProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)
	registerPermissionMergingProperties(properties)
	properties.TestingRun(t)
}

// registerPermissionMergingProperties registers all permission merging property tests.
func registerPermissionMergingProperties(properties *gopter.Properties) {
	registerYAMLOverridesProperty(properties)
	registerNonConflictingKeysProperty(properties)
	registerNilPreservesOriginalProperty(properties)
	registerEmptyMapPreservesOriginalProperty(properties)
	registerResultSizeBoundedProperty(properties)
}

// registerYAMLOverridesProperty tests that YAML permissions override comment permissions.
func registerYAMLOverridesProperty(properties *gopter.Properties) {
	properties.Property("YAML permissions override comment permissions",
		prop.ForAll(
			func(key, yamlVal, commentVal string) bool {
				if yamlVal == commentVal || yamlVal == "" || key == "" || commentVal == "" {
					return true
				}
				action := &ActionYML{Permissions: map[string]string{key: yamlVal}}
				mergePermissions(action, map[string]string{key: commentVal})

				return action.Permissions[key] == yamlVal
			},
			permissionScopeGen(),
			permissionLevelGen(),
			permissionLevelGen(),
		),
	)
}

// registerNonConflictingKeysProperty tests that non-conflicting keys are preserved.
func registerNonConflictingKeysProperty(properties *gopter.Properties) {
	properties.Property("merge preserves all non-conflicting keys",
		prop.ForAll(
			func(yamlKey, commentKey, val string) bool {
				if yamlKey == commentKey || yamlKey == "" || commentKey == "" || val == "" {
					return true
				}
				action := &ActionYML{Permissions: map[string]string{yamlKey: val}}
				mergePermissions(action, map[string]string{commentKey: val})
				_, hasYaml := action.Permissions[yamlKey]
				_, hasComment := action.Permissions[commentKey]

				return hasYaml && hasComment
			},
			permissionScopeGen(),
			permissionScopeGen(),
			permissionLevelGen(),
		),
	)
}

// registerNilPreservesOriginalProperty tests merging with nil preserves original.
func registerNilPreservesOriginalProperty(properties *gopter.Properties) {
	properties.Property("merging with nil preserves original permissions",
		prop.ForAll(
			func(key, value string) bool {
				return verifyMergePreservesOriginal(key, value, nil)
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)
}

// registerEmptyMapPreservesOriginalProperty tests merging with empty map preserves original.
func registerEmptyMapPreservesOriginalProperty(properties *gopter.Properties) {
	properties.Property("merging with empty map preserves original permissions",
		prop.ForAll(
			func(key, value string) bool {
				return verifyMergePreservesOriginal(key, value, make(map[string]string))
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)
}

// registerResultSizeBoundedProperty tests result size is bounded by sum of inputs.
func registerResultSizeBoundedProperty(properties *gopter.Properties) {
	properties.Property("merged permissions size bounded by sum of inputs",
		prop.ForAll(
			verifyMergedSizeBounded,
			gen.SliceOf(gen.AlphaString()),
			gen.SliceOf(gen.AlphaString()),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)
}

// verifyMergedSizeBounded checks that merged result size is bounded.
func verifyMergedSizeBounded(yamlKeys, commentKeys []string, value string) bool {
	if len(yamlKeys) == 0 || len(commentKeys) == 0 || value == "" {
		return true
	}
	yamlPerms := make(map[string]string)
	for _, key := range yamlKeys {
		if key != "" {
			yamlPerms[key] = value
		}
	}
	commentPerms := make(map[string]string)
	for _, key := range commentKeys {
		if key != "" {
			commentPerms[key] = value
		}
	}
	action := &ActionYML{Permissions: yamlPerms}
	mergePermissions(action, commentPerms)

	return len(action.Permissions) <= len(yamlPerms)+len(commentPerms)
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
