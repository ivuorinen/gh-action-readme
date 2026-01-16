package dependencies

import "github.com/ivuorinen/gh-action-readme/testutil"

// singleUpdateParams holds parameters for creating a test case with a single update.
type singleUpdateParams struct {
	name                                             string
	fixturePath                                      string
	oldUses, newUses, commitSHA, version, updateType string
	wantErr, validateBackup, checkRollback           bool
}

// createSingleUpdateTestCase creates a test case with a single PinnedUpdate.
// This helper reduces duplication for test cases that update a single dependency.
func createSingleUpdateTestCase(params singleUpdateParams) struct {
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
		name:          params.name,
		actionContent: testutil.MustReadFixture(params.fixturePath),
		updates: []PinnedUpdate{
			{
				FilePath:   "", // Will be set by test
				OldUses:    params.oldUses,
				NewUses:    params.newUses,
				CommitSHA:  params.commitSHA,
				Version:    params.version,
				UpdateType: params.updateType,
				LineNumber: 0,
			},
		},
		wantErr:        params.wantErr,
		validateBackup: params.validateBackup,
		checkRollback:  params.checkRollback,
	}
}
