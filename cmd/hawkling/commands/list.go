package commands

import (
	"context"
	"strings"
	"time"

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
	var filteredRoles []aws.Role

	// If both filters are enabled, return empty list (logical conflict)
	if c.options.OnlyUsed && c.options.OnlyUnused {
		roles = filteredRoles
	} else {
		filteredRoles = make([]aws.Role, 0, len(roles))

		// Apply filters
		for _, role := range roles {
			// OnlyUsed: 一度も使用されていないロール (LastUsed == nil) を除外
			if c.options.OnlyUsed && role.LastUsed == nil {
				continue
			}

			// OnlyUnused: 一度でも使用されたロール (LastUsed != nil) を除外
			if c.options.OnlyUnused && role.LastUsed != nil {
				continue
			}

			// Days フィルター: 指定された日数以内に使用されていない場合は除外
			if c.options.Days > 0 && role.LastUsed != nil {
				threshold := time.Now().AddDate(0, 0, -c.options.Days)
				isUnusedForDays := role.LastUsed.Before(threshold)
				if !isUnusedForDays {
					continue
				}
			}

			filteredRoles = append(filteredRoles, role)
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
