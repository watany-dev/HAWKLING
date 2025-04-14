package commands

import (
	"context"
	"fmt"
	"strings"

	"hawkling/pkg/aws"
	"hawkling/pkg/errors"
)

// PruneOptions contains options for the prune command
type PruneOptions struct {
	FilterOptions
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

	// Find roles based on the specified options
	filterOptions := aws.FilterOptions{
		Days:       c.options.FilterOptions.Days,
		OnlyUnused: c.options.FilterOptions.OnlyUnused,
		OnlyUsed:   c.options.FilterOptions.OnlyUsed,
	}

	filteredRoles := aws.FilterRoles(roles, filterOptions)

	if len(filteredRoles) == 0 {
		fmt.Println("No IAM roles found matching criteria")
		return nil
	}

	// Show filtered roles
	message := "Found %d IAM roles"
	if c.options.FilterOptions.OnlyUnused {
		message = "Found %d unused IAM roles (not used in the last %d days)"
	} else if c.options.FilterOptions.OnlyUsed {
		message = "Found %d used IAM roles"
		if c.options.FilterOptions.Days > 0 {
			message += " (not used in the last %d days)"
		}
	} else if c.options.FilterOptions.Days > 0 {
		message = "Found %d IAM roles (not used in the last %d days)"
	}

	if strings.Contains(message, "%d days") {
		fmt.Printf(message+":\n", len(filteredRoles), c.options.FilterOptions.Days)
	} else {
		fmt.Printf(message+":\n", len(filteredRoles))
	}
	for i, role := range filteredRoles {
		fmt.Printf("%d. %s\n", i+1, role.Name)
	}

	// If dry run, stop here
	if c.options.DryRun {
		fmt.Println("\nDRY RUN: No roles were deleted")
		return nil
	}

	// Confirm deletion if force flag is not set
	if !c.options.Force {
		prompt := fmt.Sprintf("\nAre you sure you want to delete %d roles? This cannot be undone. [y/N]: ", len(filteredRoles))
		confirmed, err := ConfirmAction(prompt)
		if err != nil {
			return errors.Wrap(err, "failed to read confirmation")
		}

		if !confirmed {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete the filtered roles
	var failedRoles []string
	for _, role := range filteredRoles {
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

	fmt.Printf("\nSuccessfully deleted %d IAM roles\n", len(filteredRoles))
	return nil
}
