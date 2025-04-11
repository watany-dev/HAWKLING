package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"hawkling/pkg/errors"
)

// ConfirmAction prompts the user for confirmation and returns their response
func ConfirmAction(prompt string) (bool, error) {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, errors.Wrap(err, "failed to read confirmation")
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// AddCommonFlags adds common flags to a command
func AddCommonFlags(cmd *cobra.Command, profile *string, region *string) {
	cmd.PersistentFlags().StringVarP(profile, "profile", "p", "", "AWS profile to use")
	cmd.PersistentFlags().StringVarP(region, "region", "r", "", "AWS region to use")
}

// AddFilterFlags adds filtering flags to a command
func AddFilterFlags(cmd *cobra.Command, days *int, onlyUsed *bool, onlyUnused *bool) {
	cmd.Flags().IntVarP(days, "days", "d", 0, "Number of days to consider for usage")
	cmd.Flags().BoolVar(onlyUsed, "used", false, "Show only used roles")
	cmd.Flags().BoolVar(onlyUnused, "unused", false, "Show only unused roles")
}

// AddOutputFlags adds output-related flags to a command
func AddOutputFlags(cmd *cobra.Command, output *string, showAll *bool) {
	cmd.Flags().StringVarP(output, "output", "o", "table", "Output format (table, json)")
	cmd.Flags().BoolVarP(showAll, "all", "a", false, "Show all information")
}

// AddDeletionFlags adds deletion-related flags to a command
func AddDeletionFlags(cmd *cobra.Command, dryRun *bool, force *bool) {
	cmd.Flags().BoolVar(dryRun, "dry-run", true, "Show what would be deleted without actually deleting")
	cmd.Flags().BoolVarP(force, "force", "f", false, "Skip confirmation prompts")
}

// AddPruneFlags adds pruning-related flags to a command
func AddPruneFlags(cmd *cobra.Command, days *int, dryRun *bool, force *bool) {
	cmd.Flags().IntVarP(days, "days", "d", 90, "Consider roles unused if not used in this many days")
	cmd.Flags().BoolVar(dryRun, "dry-run", true, "Show what would be deleted without actually deleting")
	cmd.Flags().BoolVarP(force, "force", "f", false, "Skip confirmation prompts")
}
