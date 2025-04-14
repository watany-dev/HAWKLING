package test

import (
	"testing"

	"hawkling/cmd/hawkling/commands"
	"hawkling/pkg/aws"
)

func TestFilteringLogic(t *testing.T) {
	// Create test roles - reusing the MockIAMClient setup
	mockClient := NewMockIAMClient()
	roles := mockClient.Roles

	tests := []struct {
		name          string
		options       commands.ListOptions
		expectedCount int
		expectedRoles []string
	}{
		{
			name: "No filtering",
			options: commands.ListOptions{
				FilterOptions: commands.FilterOptions{
					Days:       0, // Changed from 90 to 0 - no days filter
					OnlyUsed:   false,
					OnlyUnused: false,
				},
			},
			expectedCount: 3,
			expectedRoles: []string{"ActiveRole", "InactiveRole", "NeverUsedRole"},
		},
		{
			name: "Only used roles - no days",
			options: commands.ListOptions{
				FilterOptions: commands.FilterOptions{
					Days:       0,
					OnlyUsed:   true,
					OnlyUnused: false,
				},
			},
			expectedCount: 2,
			expectedRoles: []string{"ActiveRole", "InactiveRole"},
		},
		{
			name: "Only unused roles - no days",
			options: commands.ListOptions{
				FilterOptions: commands.FilterOptions{
					Days:       0,
					OnlyUsed:   false,
					OnlyUnused: true,
				},
			},
			expectedCount: 1,
			expectedRoles: []string{"NeverUsedRole"},
		},
		{
			name: "Days filter with 90 days",
			options: commands.ListOptions{
				FilterOptions: commands.FilterOptions{
					Days:       90,
					OnlyUsed:   false,
					OnlyUnused: false,
				},
			},
			expectedCount: 2,
			expectedRoles: []string{"InactiveRole", "NeverUsedRole"},
		},
		{
			name: "Days filter with 3 days",
			options: commands.ListOptions{
				FilterOptions: commands.FilterOptions{
					Days:       3,
					OnlyUsed:   false,
					OnlyUnused: false,
				},
			},
			expectedCount: 3,
			expectedRoles: []string{"ActiveRole", "InactiveRole", "NeverUsedRole"},
		},
		{
			name: "Days filter with Used filter - 90 days",
			options: commands.ListOptions{
				FilterOptions: commands.FilterOptions{
					Days:       90,
					OnlyUsed:   true,
					OnlyUnused: false,
				},
			},
			expectedCount: 1,
			expectedRoles: []string{"InactiveRole"},
		},
		{
			name: "Days filter with Unused filter",
			options: commands.ListOptions{
				FilterOptions: commands.FilterOptions{
					Days:       90,
					OnlyUsed:   false,
					OnlyUnused: true,
				},
			},
			expectedCount: 1,
			expectedRoles: []string{"NeverUsedRole"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Apply filter using common filtering logic
			filterOptions := aws.FilterOptions{
				Days:       test.options.FilterOptions.Days,
				OnlyUsed:   test.options.FilterOptions.OnlyUsed,
				OnlyUnused: test.options.FilterOptions.OnlyUnused,
			}

			filteredRoles := aws.FilterRoles(roles, filterOptions)

			// Check count
			if len(filteredRoles) != test.expectedCount {
				t.Errorf("Expected %d roles, got %d", test.expectedCount, len(filteredRoles))
				t.Logf("Expected: %v", test.expectedRoles)
				t.Logf("Got: %v", getRoleNames(filteredRoles))
			}

			// Check if expected roles are in the result
			for _, expectedName := range test.expectedRoles {
				found := false
				for _, role := range filteredRoles {
					if role.Name == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected role %q not found in filtered roles", expectedName)
				}
			}

			// Check if there are no unexpected roles
			if len(filteredRoles) > 0 {
				for _, role := range filteredRoles {
					found := false
					for _, expectedName := range test.expectedRoles {
						if role.Name == expectedName {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Unexpected role %q found in filtered roles", role.Name)
					}
				}
			}
		})
	}
}
