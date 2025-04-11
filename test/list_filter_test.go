package test

import (
	"testing"
	"time"

	"hawkling/pkg/aws"
)

// timePtr is defined in mock_iamclient.go

func TestFilterRoles(t *testing.T) {
	now := time.Now()

	// Create test roles with different last used times
	roles := []aws.Role{
		{
			Name:        "ActiveRole",
			Arn:         "arn:aws:iam::123456789012:role/ActiveRole",
			CreateDate:  now.AddDate(-1, 0, 0),
			LastUsed:    timePtr(now.AddDate(0, 0, -5)), // Used 5 days ago
			Description: "Recently used role",
		},
		{
			Name:        "InactiveRole",
			Arn:         "arn:aws:iam::123456789012:role/InactiveRole",
			CreateDate:  now.AddDate(-2, 0, 0),
			LastUsed:    timePtr(now.AddDate(0, 0, -100)), // Used 100 days ago
			Description: "Role unused for a long time",
		},
		{
			Name:        "NeverUsedRole",
			Arn:         "arn:aws:iam::123456789012:role/NeverUsedRole",
			CreateDate:  now.AddDate(0, -6, 0),
			LastUsed:    nil, // Never used
			Description: "Role that was never used",
		},
	}

	tests := []struct {
		name          string
		days          int
		onlyUsed      bool
		onlyUnused    bool
		expectedRoles []string // Names of roles that should be in the result
	}{
		{
			name:          "No filters - should return all roles",
			days:          90,
			onlyUsed:      false,
			onlyUnused:    false,
			expectedRoles: []string{"ActiveRole", "InactiveRole", "NeverUsedRole"},
		},
		{
			name:          "Used filter with 90 days - should return only used roles",
			days:          90,
			onlyUsed:      true,
			onlyUnused:    false,
			expectedRoles: []string{"ActiveRole"},
		},
		{
			name:          "Unused filter with 90 days - should return only unused roles",
			days:          90,
			onlyUsed:      false,
			onlyUnused:    true,
			expectedRoles: []string{"InactiveRole", "NeverUsedRole"},
		},
		{
			name:          "Both filters - should return no roles (conflicting filters)",
			days:          90,
			onlyUsed:      true,
			onlyUnused:    true,
			expectedRoles: []string{},
		},
		{
			name:          "Days filter - 3 days threshold should return different results",
			days:          3,
			onlyUsed:      false,
			onlyUnused:    true,
			expectedRoles: []string{"ActiveRole", "InactiveRole", "NeverUsedRole"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Apply filter
			var filteredRoles []aws.Role

			// If both filters are enabled, return empty list (logical conflict)
			if test.onlyUsed && test.onlyUnused {
				// Leave filteredRoles empty
			} else if test.onlyUsed || test.onlyUnused {
				for _, role := range roles {
					isUnused := role.IsUnused(test.days)

					if (test.onlyUsed && !isUnused) || (test.onlyUnused && isUnused) {
						filteredRoles = append(filteredRoles, role)
					}
				}
			} else {
				filteredRoles = roles
			}

			// Check if filtered roles match expected
			if len(filteredRoles) != len(test.expectedRoles) {
				t.Errorf("Expected %d roles, got %d", len(test.expectedRoles), len(filteredRoles))
			}

			// Check if each expected role is in the filtered roles
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
		})
	}
}
