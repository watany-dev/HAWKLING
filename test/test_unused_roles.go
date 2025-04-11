package test

import (
	"testing"
	"time"

	"hawkling/pkg/aws"
)

// TestIsUnused tests the IsUnused method of Role
func TestIsUnused(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		role     aws.Role
		days     int
		expected bool
	}{
		{
			name: "Role used recently - should not be considered unused",
			role: aws.Role{
				Name:     "RecentlyUsedRole",
				LastUsed: timePtr(now.AddDate(0, 0, -5)), // Used 5 days ago
			},
			days:     90,
			expected: false, // Not unused for 90 days
		},
		{
			name: "Role not used for a long time - should be considered unused",
			role: aws.Role{
				Name:     "OldRole",
				LastUsed: timePtr(now.AddDate(0, 0, -100)), // Used 100 days ago
			},
			days:     90,
			expected: true, // Unused for 90 days
		},
		{
			name: "Role never used - should be considered unused",
			role: aws.Role{
				Name:     "NeverUsedRole",
				LastUsed: nil, // Never used
			},
			days:     90,
			expected: true, // Never used, so unused for any days
		},
		{
			name: "Edge case - Role used exactly at threshold",
			role: aws.Role{
				Name:     "ThresholdRole",
				LastUsed: timePtr(now.AddDate(0, 0, -90)), // Used 90 days ago
			},
			days:     90,
			expected: true, // At threshold, should be considered unused
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.role.IsUnused(test.days)
			if result != test.expected {
				t.Errorf("IsUnused(%d) = %v; want %v", test.days, result, test.expected)
			}
		})
	}
}
