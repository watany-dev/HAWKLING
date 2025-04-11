package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"hawkling/pkg/aws"
	"hawkling/pkg/errors"
)

// PruneOptions contains options for the prune command
type PruneOptions struct {
	Days   int
	DryRun bool
	Force  bool
}

// PruneCommand represents the prune command
type PruneCommand struct {
	profile string
	region  string
	options PruneOptions
}

// NewPruneCommand creates a new prune command
func NewPruneCommand(profile, region string, options PruneOptions) *PruneCommand {
	return &PruneCommand{
		profile: profile,
		region:  region,
		options: options,
	}
}

// Execute runs the prune command
func (c *PruneCommand) Execute(ctx context.Context) error {
	client, err := aws.NewAWSClient(ctx, c.profile, c.region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS client")
	}

	// Get all roles
	roles, err := client.ListRoles(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list roles")
	}

	// Find unused roles based on the specified days
	var unusedRoles []aws.Role
	for _, role := range roles {
		if role.IsUnused(c.options.Days) {
			unusedRoles = append(unusedRoles, role)
		}
	}

	if len(unusedRoles) == 0 {
		fmt.Println("No unused IAM roles found")
		return nil
	}

	// Show unused roles
	fmt.Printf("Found %d unused IAM roles (not used in the last %d days):\n", len(unusedRoles), c.options.Days)
	for i, role := range unusedRoles {
		fmt.Printf("%d. %s\n", i+1, role.Name)
	}

	// If dry run, stop here
	if c.options.DryRun {
		fmt.Println("\nDRY RUN: No roles were deleted")
		return nil
	}

	// Confirm deletion if force flag is not set
	if !c.options.Force {
		fmt.Printf("\nAre you sure you want to delete %d unused roles? This cannot be undone. [y/N]: ", len(unusedRoles))

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return errors.Wrap(err, "failed to read confirmation")
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete the unused roles
	var failedRoles []string
	for _, role := range unusedRoles {
		if err := client.DeleteRole(ctx, role.Name); err != nil {
			failedRoles = append(failedRoles, role.Name)
			fmt.Printf("Failed to delete role %s: %v\n", role.Name, err)
		} else {
			fmt.Printf("Deleted role: %s\n", role.Name)
		}
	}

	if len(failedRoles) > 0 {
		fmt.Printf("\nFailed to delete %d roles: %s\n", len(failedRoles), strings.Join(failedRoles, ", "))
		return errors.Errorf("failed to delete %d roles", len(failedRoles))
	}

	fmt.Printf("\nSuccessfully deleted %d unused IAM roles\n", len(unusedRoles))
	return nil
}
