package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"hawkling/pkg/aws"
	"hawkling/pkg/formatter"
)

var (
	profile     string
	region      string
	days        int
	dryRun      bool
	force       bool
	output      string
	showAllInfo bool
	onlyUsed    bool
	onlyUnused  bool
)

func main() {
	// Check if demo mode is requested
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		Demo()
		return
	}

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
		RunE:  listRoles,
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
		RunE:  deleteRole,
	}
	deleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate deletion without actually deleting")
	deleteCmd.Flags().BoolVar(&force, "force", false, "Delete without confirmation")

	// Prune command
	pruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "Delete all unused IAM roles",
		RunE:  pruneRoles,
	}
	pruneCmd.Flags().IntVar(&days, "days", 90, "Number of days to consider a role as unused")
	pruneCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate deletion without actually deleting")
	pruneCmd.Flags().BoolVar(&force, "force", false, "Delete without confirmation")

	// Demo command
	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "Run a demonstration of output formats",
		Run: func(cmd *cobra.Command, args []string) {
			Demo()
		},
	}

	rootCmd.AddCommand(listCmd, deleteCmd, pruneCmd, demoCmd)

	return rootCmd
}

func createClient(ctx context.Context) (aws.IAMClient, error) {
	client, err := aws.NewAWSClient(ctx, profile, region)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client: %w", err)
	}
	return client, nil
}

func listRoles(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	client, err := createClient(ctx)
	if err != nil {
		return err
	}

	// Check for conflicting options
	if onlyUsed && onlyUnused {
		return fmt.Errorf("--used and --unused flags cannot be used together")
	}

	// Only validate if days flag was actually specified by user
	if onlyUnused && cmd.Flags().Changed("days") {
		return fmt.Errorf("--unused and --days flags cannot be used together")
	}

	roles, err := client.ListRoles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	// Filter roles if days flag is provided (>0)
	if days > 0 {
		var unusedRoles []aws.Role
		for _, role := range roles {
			if role.IsUnused(days) {
				unusedRoles = append(unusedRoles, role)
			}
		}
		roles = unusedRoles
	}

	// Filter for only used roles if --used flag is provided
	if onlyUsed {
		var usedRoles []aws.Role
		for _, role := range roles {
			if role.LastUsed != nil {
				usedRoles = append(usedRoles, role)
			}
		}
		roles = usedRoles
	}

	// Filter for only never used roles if --unused flag is provided
	if onlyUnused {
		var neverUsedRoles []aws.Role
		for _, role := range roles {
			if role.LastUsed == nil {
				neverUsedRoles = append(neverUsedRoles, role)
			}
		}
		roles = neverUsedRoles
	}

	var format formatter.Format
	switch output {
	case "json":
		format = formatter.JSONFormat
	default:
		format = formatter.TableFormat
	}

	return formatter.FormatRoles(roles, format, showAllInfo)
}

func deleteRole(cmd *cobra.Command, args []string) error {
	roleName := args[0]
	ctx := context.Background()
	client, err := createClient(ctx)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("Would delete role: %s (dry run)\n", roleName)
		return nil
	}

	if !force {
		fmt.Printf("Are you sure you want to delete role %s? [y/N]: ", roleName)
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil || (response != "y" && response != "Y") {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	if err := client.DeleteRole(ctx, roleName); err != nil {
		return fmt.Errorf("failed to delete role %s: %w", roleName, err)
	}

	fmt.Printf("Successfully deleted role: %s\n", roleName)
	return nil
}

func pruneRoles(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	client, err := createClient(ctx)
	if err != nil {
		return err
	}

	roles, err := client.ListRoles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	var unusedRoles []aws.Role
	for _, role := range roles {
		if role.IsUnused(days) {
			unusedRoles = append(unusedRoles, role)
		}
	}

	if len(unusedRoles) == 0 {
		fmt.Println("No unused roles found")
		return nil
	}

	fmt.Printf("Found %d unused roles (not used in the last %d days):\n", len(unusedRoles), days)
	for i, role := range unusedRoles {
		fmt.Printf("%d. %s\n", i+1, role.Name)
	}

	if dryRun {
		fmt.Println("Dry run: no roles will be deleted")
		return nil
	}

	if !force {
		fmt.Print("Do you want to delete these roles? [y/N]: ")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil || (response != "y" && response != "Y") {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	for _, role := range unusedRoles {
		fmt.Printf("Deleting role: %s\n", role.Name)
		if err := client.DeleteRole(ctx, role.Name); err != nil {
			fmt.Printf("Error deleting role %s: %s\n", role.Name, err)
		} else {
			fmt.Printf("Successfully deleted role: %s\n", role.Name)
		}
	}

	return nil
}
