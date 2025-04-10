package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"hawkling/cmd/hawkling/commands"
)

var (
	// Global flags
	profile string
	region  string

	// Command-specific flags
	dryRun      bool
	force       bool
	output      string
	showAllInfo bool
	onlyUsed    bool
	onlyUnused  bool
)

func main() {
	// Normal CLI execution
	rootCmd := createRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// createRootCommand sets up the command structure
func createRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "hawkling",
		Short: "Hawkling is a tool for managing AWS IAM roles",
		Long: `A CLI tool for listing, detecting unused, and cleaning up AWS IAM roles.
Complete documentation is available at https://github.com/watany-dev/hawkling`,
		SilenceUsage: true,
	}

	// Global flags
	commands.AddCommonFlags(rootCmd, &profile, &region)

	// List command
	var listDays int
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List IAM roles, optionally filtering for unused roles",
		RunE: func(cmd *cobra.Command, args []string) error {
			listOptions := commands.ListOptions{
				Days:       listDays,
				Output:     output,
				ShowAll:    showAllInfo,
				OnlyUsed:   onlyUsed,
				OnlyUnused: onlyUnused,
			}

			listCmd := commands.NewListCommand(profile, region, listOptions)
			return listCmd.Execute(context.Background())
		},
	}
	commands.AddFilterFlags(listCmd, &listDays, &onlyUsed, &onlyUnused)
	commands.AddOutputFlags(listCmd, &output, &showAllInfo)

	// Delete command
	deleteCmd := &cobra.Command{
		Use:   "delete [role-name]",
		Short: "Delete an IAM role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			roleName := args[0]
			deleteOptions := commands.DeleteOptions{
				DryRun: dryRun,
				Force:  force,
			}

			deleteCmd := commands.NewDeleteCommand(profile, region, roleName, deleteOptions)
			return deleteCmd.Execute(context.Background())
		},
	}
	commands.AddDeletionFlags(deleteCmd, &dryRun, &force)

	// Prune command
	var pruneDays int
	pruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "Delete all unused IAM roles",
		RunE: func(cmd *cobra.Command, args []string) error {
			pruneOptions := commands.PruneOptions{
				Days:   pruneDays,
				DryRun: dryRun,
				Force:  force,
			}

			pruneCmd := commands.NewPruneCommand(profile, region, pruneOptions)
			return pruneCmd.Execute(context.Background())
		},
	}
	commands.AddPruneFlags(pruneCmd, &pruneDays, &dryRun, &force)

	return rootCmd
}
