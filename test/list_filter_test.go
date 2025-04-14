package test

import (
	"testing"
	"time"

	"hawkling/pkg/aws"
)

// Local timePtr for tests
func localTimePtr(t time.Time) *time.Time {
	return &t
}

func TestFilterRoles(t *testing.T) {
	now := time.Now()

	// Create test roles with different last used times
	roles := []aws.Role{
		{
			Name:        "ActiveRole",
			Arn:         "arn:aws:iam::123456789012:role/ActiveRole",
			CreateDate:  now.AddDate(-1, 0, 0),
			LastUsed:    localTimePtr(now.AddDate(0, 0, -5)), // Used 5 days ago
			Description: "Recently used role",
		},
		{
			Name:        "InactiveRole",
			Arn:         "arn:aws:iam::123456789012:role/InactiveRole",
			CreateDate:  now.AddDate(-2, 0, 0),
			LastUsed:    localTimePtr(now.AddDate(0, 0, -100)), // Used 100 days ago
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
			days:          0,
			onlyUsed:      false,
			onlyUnused:    false,
			expectedRoles: []string{"ActiveRole", "InactiveRole", "NeverUsedRole"},
		},
		{
			name:          "Used filter - should return only roles that were used at least once",
			days:          0,
			onlyUsed:      true,
			onlyUnused:    false,
			expectedRoles: []string{"ActiveRole", "InactiveRole"},
		},
		{
			name:          "Unused filter - should return only roles that were never used",
			days:          0,
			onlyUsed:      false,
			onlyUnused:    true,
			expectedRoles: []string{"NeverUsedRole"},
		},
		{
			name:          "Both filters - should return no roles (conflicting filters)",
			days:          0,
			onlyUsed:      true,
			onlyUnused:    true,
			expectedRoles: []string{},
		},
		{
			name:          "Days filter with 90 days - should return roles not used in 90 days",
			days:          90,
			onlyUsed:      false,
			onlyUnused:    false,
			expectedRoles: []string{"InactiveRole", "NeverUsedRole"},
		},
		{
			name:          "Days filter with 3 days - should return roles not used in 3 days",
			days:          3,
			onlyUsed:      false,
			onlyUnused:    false,
			expectedRoles: []string{"ActiveRole", "InactiveRole", "NeverUsedRole"},
		},
		{
			name:          "Days filter with Used filter - should return used roles not used in 90 days",
			days:          90,
			onlyUsed:      true,
			onlyUnused:    false,
			expectedRoles: []string{"InactiveRole"},
		},
		{
			name:          "Days filter with Unused filter - should return never used roles",
			days:          90,
			onlyUsed:      false,
			onlyUnused:    true,
			expectedRoles: []string{"NeverUsedRole"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Apply filter using unified filtering logic
			filterOptions := aws.FilterOptions{
				Days:       test.days,
				OnlyUsed:   test.onlyUsed,
				OnlyUnused: test.onlyUnused,
			}

			filteredRoles := aws.FilterRoles(roles, filterOptions)

			// Check if filtered roles match expected
			if len(filteredRoles) != len(test.expectedRoles) {
				t.Errorf("Expected %d roles, got %d", len(test.expectedRoles), len(filteredRoles))
				t.Logf("Expected: %v", test.expectedRoles)
				t.Logf("Got: %v", getRoleNames(filteredRoles))
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

func getRoleNames(roles []aws.Role) []string {
	names := make([]string, len(roles))
	for i, role := range roles {
		names[i] = role.Name
	}
	return names
}
