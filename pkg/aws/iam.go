package aws

import (
	"context"
	"time"
)

// IAMClient defines the interface for AWS IAM operations
type IAMClient interface {
	// ListRoles returns all IAM roles
	ListRoles(ctx context.Context) ([]Role, error)
	
	// GetRoleLastUsed returns the last used timestamp for a role
	GetRoleLastUsed(ctx context.Context, roleName string) (*time.Time, error)
	
	// DeleteRole deletes an IAM role
	DeleteRole(ctx context.Context, roleName string) error
}

// Role represents an AWS IAM role
type Role struct {
	Name        string
	Arn         string
	Description string
	CreateDate  time.Time
	LastUsed    *time.Time
}

// IsUnused checks if a role is unused for the specified number of days
func (r *Role) IsUnused(days int) bool {
	if r.LastUsed == nil {
		return true
	}
	
	threshold := time.Now().AddDate(0, 0, -days)
	return r.LastUsed.Before(threshold)
}