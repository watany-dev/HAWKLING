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
				Days:       90,
				OnlyUsed:   false,
				OnlyUnused: false,
			},
			expectedCount: 3,
			expectedRoles: []string{"ActiveRole", "InactiveRole", "NeverUsedRole"},
		},
		{
			name: "Only used roles - 90 days",
			options: commands.ListOptions{
				Days:       90,
				OnlyUsed:   true,
				OnlyUnused: false,
			},
			expectedCount: 1,
			expectedRoles: []string{"ActiveRole"},
		},
		{
			name: "Only unused roles - 90 days",
			options: commands.ListOptions{
				Days:       90,
				OnlyUsed:   false,
				OnlyUnused: true,
			},
			expectedCount: 2,
			expectedRoles: []string{"InactiveRole", "NeverUsedRole"},
		},
		{
			name: "Only used roles - 3 days",
			options: commands.ListOptions{
				Days:       3,
				OnlyUsed:   true,
				OnlyUnused: false,
			},
			expectedCount: 0, // None were used within 3 days
			expectedRoles: []string{},
		},
		{
			name: "Only unused roles - 3 days",
			options: commands.ListOptions{
				Days:       3,
				OnlyUsed:   false,
				OnlyUnused: true,
			},
			expectedCount: 3, // All were unused for 3 days
			expectedRoles: []string{"ActiveRole", "InactiveRole", "NeverUsedRole"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var filteredRoles []aws.Role

			// Apply the improved filtering logic
			for _, role := range roles {
				isUnusedForDays := role.IsUnused(test.options.Days)

				// OnlyUsed: Show only roles that have been used (LastUsed != nil) and were used within the specified days
				if test.options.OnlyUsed && (role.LastUsed == nil || isUnusedForDays) {
					continue
				}

				// OnlyUnused: Show only roles that have not been used within the specified days
				if test.options.OnlyUnused && !isUnusedForDays {
					continue
				}

				filteredRoles = append(filteredRoles, role)
			}

			// Check count
			if len(filteredRoles) != test.expectedCount {
				t.Errorf("Expected %d roles, got %d", test.expectedCount, len(filteredRoles))
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
