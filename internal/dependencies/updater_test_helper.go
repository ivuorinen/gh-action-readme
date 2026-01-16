package dependencies

import "github.com/ivuorinen/gh-action-readme/testutil"

// createSingleUpdateTestCase creates a test case with a single PinnedUpdate.
// This helper reduces duplication for test cases that update a single dependency.
func createSingleUpdateTestCase(
	name, fixturePath string,
	oldUses, newUses, commitSHA, version, updateType string,
	wantErr, validateBackup, checkRollback bool,
) struct {
	name           string
	actionContent  string
	updates        []PinnedUpdate
	wantErr        bool
	validateBackup bool
	checkRollback  bool
} {
	return struct {
		name           string
		actionContent  string
		updates        []PinnedUpdate
		wantErr        bool
		validateBackup bool
		checkRollback  bool
	}{
		name:          name,
		actionContent: testutil.MustReadFixture(fixturePath),
		updates: []PinnedUpdate{
			{
				FilePath:   "", // Will be set by test
				OldUses:    oldUses,
				NewUses:    newUses,
				CommitSHA:  commitSHA,
				Version:    version,
				UpdateType: updateType,
				LineNumber: 0,
			},
		},
		wantErr:        wantErr,
		validateBackup: validateBackup,
		checkRollback:  checkRollback,
	}
}
