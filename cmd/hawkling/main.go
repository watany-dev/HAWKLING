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
	profile string
	region  string
	days    int
	dryRun  bool
	force   bool
	output  string
)

func main() {
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

	rootCmd.AddCommand(listCmd, deleteCmd, pruneCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
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

	format := formatter.TableFormat
	if output == "json" {
		format = formatter.JSONFormat
	}

	return formatter.FormatRoles(roles, format)
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
