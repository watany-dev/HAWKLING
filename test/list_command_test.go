package test

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"hawkling/cmd/hawkling/commands"
	"hawkling/pkg/aws"
	// "hawkling/pkg/formatter"
)

// simulatedError is already defined in mock_iamclient.go

// mockFormatter satisfies the formatter interface for testing
// type mockFormatter struct {
// 	output string
// }

// func (m *mockFormatter) FormatRoles(roles []aws.Role, format formatter.Format, showAllInfo bool) error {
// 	var output strings.Builder
// 	for _, role := range roles {
// 		output.WriteString(role.Name + "\n")
// 	}
// 	m.output = output.String()
// 	return nil
// }

func TestListCommandFilterFlags(t *testing.T) {
	// Set up mock AWS client
	mockClient := NewMockIAMClient()
	aws.SetTestClient(mockClient)
	// Clean up test client after test
	defer aws.ClearTestClient()

	tests := []struct {
		name             string
		options          commands.ListOptions
		expectedRoles    int      // Number of roles expected after filtering
		shouldContain    []string // Role names that should be in the output
		shouldNotContain []string // Role names that should not be in the output
	}{
		{
			name: "No filters - should return all roles",
			options: commands.ListOptions{
				Days:       90,
				Output:     "table",
				ShowAll:    false,
				OnlyUsed:   false,
				OnlyUnused: false,
			},
			expectedRoles:    3, // All 3 roles from the mock
			shouldContain:    []string{"ActiveRole", "InactiveRole", "NeverUsedRole"},
			shouldNotContain: []string{},
		},
		{
			name: "Used filter - should return only used roles",
			options: commands.ListOptions{
				Days:       90,
				Output:     "table",
				ShowAll:    false,
				OnlyUsed:   true,
				OnlyUnused: false,
			},
			expectedRoles:    1, // Only ActiveRole was used in the last 90 days
			shouldContain:    []string{"ActiveRole"},
			shouldNotContain: []string{"InactiveRole", "NeverUsedRole"},
		},
		{
			name: "Unused filter - should return only unused roles",
			options: commands.ListOptions{
				Days:       90,
				Output:     "table",
				ShowAll:    false,
				OnlyUsed:   false,
				OnlyUnused: true,
			},
			expectedRoles:    2, // InactiveRole and NeverUsedRole are unused
			shouldContain:    []string{"InactiveRole", "NeverUsedRole"},
			shouldNotContain: []string{"ActiveRole"},
		},
		{
			name: "Both filters - should return no roles (conflicting filters)",
			options: commands.ListOptions{
				Days:       90,
				Output:     "table",
				ShowAll:    false,
				OnlyUsed:   true,
				OnlyUnused: true,
			},
			expectedRoles:    0, // No roles match both used and unused
			shouldContain:    []string{},
			shouldNotContain: []string{"ActiveRole", "InactiveRole", "NeverUsedRole"},
		},
		{
			name: "Days filter - 3 days threshold should return different results",
			options: commands.ListOptions{
				Days:       3,
				Output:     "table",
				ShowAll:    false,
				OnlyUsed:   false,
				OnlyUnused: true,
			},
			expectedRoles:    3,                                                       // All should be unused with 3 day threshold (ActiveRole was used 5 days ago)
			shouldContain:    []string{"ActiveRole", "InactiveRole", "NeverUsedRole"}, // ActiveRole should be considered unused with 3 day threshold
			shouldNotContain: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Capture stdout
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run the command
			cmd := commands.NewListCommand("test-profile", "us-west-2", test.options)
			err := cmd.Execute(context.Background())
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			// Restore stdout and get output
			w.Close()
			os.Stdout = originalStdout
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			// Check expectations
			for _, expectedStr := range test.shouldContain {
				if !strings.Contains(output, expectedStr) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput: %s", expectedStr, output)
				}
			}

			for _, unexpectedStr := range test.shouldNotContain {
				if strings.Contains(output, unexpectedStr) {
					t.Errorf("Expected output to NOT contain %q, but it did.\nOutput: %s", unexpectedStr, output)
				}
			}
		})
	}
}
