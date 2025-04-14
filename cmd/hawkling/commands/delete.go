package commands

import (
	"context"
	"fmt"

	"hawkling/pkg/aws"
	"hawkling/pkg/errors"
)

// DeleteOptions contains options for the delete command
type DeleteOptions struct {
	DryRun bool
	Force  bool
}

// DeleteCommand represents the delete command
type DeleteCommand struct {
	profile  string
	region   string
	roleName string
	options  DeleteOptions
}

// NewDeleteCommand creates a new delete command
func NewDeleteCommand(profile, region, roleName string, options DeleteOptions) *DeleteCommand {
	return &DeleteCommand{
		profile:  profile,
		region:   region,
		roleName: roleName,
		options:  options,
	}
}

// Execute runs the delete command
func (c *DeleteCommand) Execute(ctx context.Context) error {
	client, err := aws.NewAWSClient(ctx, c.profile, c.region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS client")
	}

	// Get the role to ensure it exists
	roles, err := client.ListRoles(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list roles")
	}

	var targetRole *aws.Role
	for i, role := range roles {
		if role.Name == c.roleName {
			targetRole = &roles[i]
			break
		}
	}

	if targetRole == nil {
		return errors.Errorf("role '%s' not found", c.roleName)
	}

	// If dry run, just show what would be deleted
	if c.options.DryRun {
		fmt.Printf("DRY RUN: Would delete IAM role: %s\n", c.roleName)
		return nil
	}

	// Confirm deletion if force flag is not set
	if !c.options.Force {
		prompt := fmt.Sprintf("Are you sure you want to delete role '%s'? This cannot be undone. [y/N]: ", c.roleName)
		confirmed, err := ConfirmAction(prompt)
		if err != nil {
			return errors.Wrap(err, "failed to read confirmation")
		}

		if !confirmed {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete the role
	if err := client.DeleteRole(ctx, c.roleName); err != nil {
		return errors.Wrap(err, "failed to delete role")
	}

	fmt.Printf("Successfully deleted IAM role: %s\n", c.roleName)
	return nil
}
