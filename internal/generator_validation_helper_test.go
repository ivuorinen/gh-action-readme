package internal

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestAssertMessageCounts_Helper tests the assertMessageCounts helper function.
func TestAssertMessageCounts_Helper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		output *capturedOutput
		want   messageCountExpectations
	}{
		{
			name: "all counts zero",
			output: &capturedOutput{
				CapturedOutput: &testutil.CapturedOutput{
					BoldMessages:    []string{},
					SuccessMessages: []string{},
					WarningMessages: []string{},
					ErrorMessages:   []string{},
					InfoMessages:    []string{},
				},
			},
			want: messageCountExpectations{
				bold:    0,
				success: 0,
				warning: 0,
				error:   0,
				info:    0,
			},
		},
		{
			name: "some messages",
			output: &capturedOutput{
				CapturedOutput: &testutil.CapturedOutput{
					BoldMessages:    []string{"bold1", "bold2"},
					SuccessMessages: []string{"success1"},
					WarningMessages: []string{},
					ErrorMessages:   []string{"error1", "error2", "error3"},
					InfoMessages:    []string{"info1"},
				},
			},
			want: messageCountExpectations{
				bold:    2,
				success: 1,
				warning: 0,
				error:   3,
				info:    1,
			},
		},
		{
			name: "only bold and success",
			output: &capturedOutput{
				CapturedOutput: &testutil.CapturedOutput{
					BoldMessages:    []string{"bold"},
					SuccessMessages: []string{"success"},
					WarningMessages: []string{},
					ErrorMessages:   []string{},
					InfoMessages:    []string{},
				},
			},
			want: messageCountExpectations{
				bold:    1,
				success: 1,
				warning: 0,
				error:   0,
				info:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the helper - it validates message counts
			assertMessageCounts(t, tt.output, tt.want)
		})
	}
}
