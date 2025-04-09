package test

import (
	"testing"
	"time"

	"hawkling/pkg/aws"
)

func TestRoleIsUnused(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		lastUsed *time.Time
		days     int
		expected bool
	}{
		{
			name:     "never used role should be unused",
			lastUsed: nil,
			days:     90,
			expected: true,
		},
		{
			name:     "recently used role should not be unused",
			lastUsed: &now,
			days:     90,
			expected: false,
		},
		{
			name:     "role used 91 days ago should be unused with 90 days threshold",
			lastUsed: timePtr(now.AddDate(0, 0, -91)),
			days:     90,
			expected: true,
		},
		{
			name:     "role used 89 days ago should not be unused with 90 days threshold",
			lastUsed: timePtr(now.AddDate(0, 0, -89)),
			days:     90,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			role := aws.Role{
				Name:     "TestRole",
				LastUsed: test.lastUsed,
			}

			result := role.IsUnused(test.days)

			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

// Use timePtr from mock_iamclient.go
