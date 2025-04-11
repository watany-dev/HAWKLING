package commands

import (
	"context"
	"strings"

	"hawkling/pkg/aws"
	"hawkling/pkg/errors"
	"hawkling/pkg/formatter"
)

// ListOptions contains options for the list command
type ListOptions struct {
	Days       int
	Output     string
	ShowAll    bool
	OnlyUsed   bool
	OnlyUnused bool
}

// ListCommand represents the list command
type ListCommand struct {
	profile string
	region  string
	options ListOptions
}

// NewListCommand creates a new list command
func NewListCommand(profile, region string, options ListOptions) *ListCommand {
	return &ListCommand{
		profile: profile,
		region:  region,
		options: options,
	}
}

// Execute runs the list command
func (c *ListCommand) Execute(ctx context.Context) error {
	client, err := aws.NewAWSClient(ctx, c.profile, c.region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS client")
	}

	roles, err := client.ListRoles(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list roles")
	}

	// Filter roles if needed
	if c.options.OnlyUsed || c.options.OnlyUnused {
		var filteredRoles []aws.Role
		for _, role := range roles {
			isUnused := role.IsUnused(c.options.Days)

			if (c.options.OnlyUsed && !isUnused) || (c.options.OnlyUnused && isUnused) {
				filteredRoles = append(filteredRoles, role)
			}
		}
		roles = filteredRoles
	}

	// Format output
	format := formatter.Format(strings.ToLower(c.options.Output))
	if err := formatter.FormatRoles(roles, format, c.options.ShowAll); err != nil {
		return errors.Wrap(err, "failed to format output")
	}

	return nil
}
