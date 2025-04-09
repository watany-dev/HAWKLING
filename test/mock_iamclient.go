package test

import (
	"context"
	"time"

	"mejiro/pkg/aws"
)

// Helper function to create a time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}

// MockIAMClient is a mock implementation of the IAMClient interface
type MockIAMClient struct {
	Roles    []aws.Role
	DeletedRoles []string
}

// NewMockIAMClient creates a new mock IAM client with predefined roles
func NewMockIAMClient() *MockIAMClient {
	now := time.Now()
	
	// Create some test roles with different last used times
	roles := []aws.Role{
		{
			Name:       "ActiveRole",
			Arn:        "arn:aws:iam::123456789012:role/ActiveRole",
			CreateDate: now.AddDate(-1, 0, 0),
			LastUsed:   timePtr(now.AddDate(0, 0, -5)), // Used 5 days ago
			Description: "Recently used role",
		},
		{
			Name:       "InactiveRole",
			Arn:        "arn:aws:iam::123456789012:role/InactiveRole",
			CreateDate: now.AddDate(-2, 0, 0),
			LastUsed:   timePtr(now.AddDate(0, 0, -100)), // Used 100 days ago
			Description: "Role unused for a long time",
		},
		{
			Name:       "NeverUsedRole",
			Arn:        "arn:aws:iam::123456789012:role/NeverUsedRole",
			CreateDate: now.AddDate(0, -6, 0),
			LastUsed:   nil, // Never used
			Description: "Role that was never used",
		},
	}
	
	return &MockIAMClient{
		Roles:    roles,
		DeletedRoles: []string{},
	}
}

// ListRoles returns all mock IAM roles
func (m *MockIAMClient) ListRoles(ctx context.Context) ([]aws.Role, error) {
	return m.Roles, nil
}

// GetRoleLastUsed returns the last used timestamp for a role
func (m *MockIAMClient) GetRoleLastUsed(ctx context.Context, roleName string) (*time.Time, error) {
	for _, role := range m.Roles {
		if role.Name == roleName {
			return role.LastUsed, nil
		}
	}
	return nil, nil
}

// DeleteRole simulates deleting an IAM role
func (m *MockIAMClient) DeleteRole(ctx context.Context, roleName string) error {
	m.DeletedRoles = append(m.DeletedRoles, roleName)
	
	// Remove the role from the list of roles
	var updatedRoles []aws.Role
	for _, role := range m.Roles {
		if role.Name != roleName {
			updatedRoles = append(updatedRoles, role)
		}
	}
	m.Roles = updatedRoles
	
	return nil
}