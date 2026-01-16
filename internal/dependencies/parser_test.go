package dependencies

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
