package dependencies

import (
	"strings"
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid relative path",
			path:    "testdata/action.yml",
			wantErr: false,
		},
		{
			name:    "valid absolute path",
			path:    "/tmp/action.yml",
			wantErr: false,
		},
		{
			name:    "traversal with double dots",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "traversal in middle of path",
			path:    "foo/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "clean path with dot slash",
			path:    "./foo/bar",
			wantErr: false,
		},
		{
			name:    "valid nested path",
			path:    "internal/testdata/fixtures/action.yml",
			wantErr: false,
		},
		{
			name:    "path with trailing slash",
			path:    "testdata/action.yml/",
			wantErr: false,
		},
		{
			// filepath.Clean collapses this to "/etc/passwd"; a cleaned-path
			// check would miss it, so the original path must be inspected.
			name:    "absolute traversal collapsed by Clean",
			path:    "/tmp/x/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "relative traversal that resolves in-bounds is still rejected",
			path:    "actions/build/../build/action.yml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
			// For rejected paths, assert the rejection is specifically for
			// traversal — not some incidental error — so a regression that
			// changes WHY a path is rejected does not slip past this test.
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), "traversal") {
				t.Errorf("validateFilePath() error = %q, want it to mention traversal", err.Error())
			}
		})
	}
}
