package commands

import (
	"context"
	"fmt"
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
	// Implementation would go here
	fmt.Println("Listing IAM roles")
	return nil
}