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
	days        int
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
Complete documentation is available at https://github.com/yourusername/hawkling`,
		SilenceUsage: true,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "AWS profile to use")
	rootCmd.PersistentFlags().StringVar(&region, "region", "us-east-1", "AWS region to use")

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List IAM roles, optionally filtering for unused roles",
		RunE: func(cmd *cobra.Command, args []string) error {
			listOptions := commands.ListOptions{
				Days:       days,
				Output:     output,
				ShowAll:    showAllInfo,
				OnlyUsed:   onlyUsed,
				OnlyUnused: onlyUnused,
			}

			listCmd := commands.NewListCommand(profile, region, listOptions)
			return listCmd.Execute(context.Background())
		},
	}
	listCmd.Flags().IntVar(&days, "days", 0, "Number of days to consider a role as unused (0 to list all roles)")
	listCmd.Flags().StringVarP(&output, "output", "o", "table", "Output format (table or json)")
	listCmd.Flags().BoolVar(&showAllInfo, "all", false, "Show all information including ARN and creation date")
	listCmd.Flags().BoolVar(&onlyUsed, "used", false, "Show only roles that have been used at least once")
	listCmd.Flags().BoolVar(&onlyUnused, "unused", false, "Show only roles that have never been used")

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
	deleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate deletion without actually deleting")
	deleteCmd.Flags().BoolVar(&force, "force", false, "Delete without confirmation")

	// Prune command
	pruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "Delete all unused IAM roles",
		RunE: func(cmd *cobra.Command, args []string) error {
			pruneOptions := commands.PruneOptions{
				Days:   days,
				DryRun: dryRun,
				Force:  force,
			}

			pruneCmd := commands.NewPruneCommand(profile, region, pruneOptions)
			return pruneCmd.Execute(context.Background())
		},
	}
	pruneCmd.Flags().IntVar(&days, "days", 90, "Number of days to consider a role as unused")
	pruneCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate deletion without actually deleting")
	pruneCmd.Flags().BoolVar(&force, "force", false, "Delete without confirmation")

	// Demo command
	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "Run a demonstration of output formats",
		RunE: func(cmd *cobra.Command, args []string) error {
			demoCmd := commands.NewDemoCommand()
			return demoCmd.Execute(context.Background())
		},
	}

	rootCmd.AddCommand(listCmd, deleteCmd, pruneCmd, demoCmd)

	return rootCmd
}
