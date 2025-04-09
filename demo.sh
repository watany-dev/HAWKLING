#!/bin/bash
# HAWKLING Demo Script
# This script demonstrates the functionality of the hawkling command line tool
# with mock outputs for different commands and formatting options.

# Text formatting
BOLD='\033[1m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BOLD}HAWKLING DEMO${NC}"
echo -e "A CLI tool for managing AWS IAM roles\n"

# Function to simulate command execution
simulate_command() {
    echo -e "${BLUE}$ $1${NC}"
    sleep 1
}

# Function to display help
display_help() {
    simulate_command "hawkling --help"
    echo "A CLI tool for listing, detecting unused, and cleaning up AWS IAM roles.
Complete documentation is available at https://github.com/yourusername/hawkling

Usage:
  hawkling [command]

Available Commands:
  delete      Delete an IAM role
  help        Help about any command
  list        List IAM roles, optionally filtering for unused roles
  prune       Delete all unused IAM roles

Flags:
  -h, --help             help for hawkling
      --profile string   AWS profile to use
      --region string    AWS region to use (default \"us-east-1\")

Use \"hawkling [command] --help\" for more information about a command."
    echo
}

# List command help
display_list_help() {
    simulate_command "hawkling list --help"
    echo "List IAM roles, optionally filtering for unused roles

Usage:
  hawkling list [flags]

Flags:
      --all             Show all information including ARN and creation date
      --days int        Number of days to consider a role as unused (0 to list all roles)
  -h, --help            help for list
  -o, --output string   Output format (table or json) (default \"table\")
      --used            Show only roles that have been used at least once

Global Flags:
      --profile string   AWS profile to use
      --region string    AWS region to use (default \"us-east-1\")"
    echo
}

# Demo 1: List all roles in default simplified table format
demo_list_table_simple() {
    simulate_command "hawkling list --profile demo"
    echo "NAME                    LAST USED                DESCRIPTION"
    echo "AdminRole               2025-04-01T14:22:18Z     Administrator role for full access"
    echo "DataScientistRole       2025-03-29T09:41:05Z     Role for data science team with S3 and SageMaker access"
    echo "ReadOnlyRole            2025-04-01T10:11:35Z     Read-only access for auditing"
    echo "DeploymentRole          2025-03-15T16:30:22Z     Role for CI/CD deployment pipelines"
    echo "LambdaExecutionRole     Never                    Execution role for Lambda functions"
    echo "OldServiceRole          2023-10-05T13:45:17Z     Legacy role for deprecated service"
    echo
}

# Demo 2: List all roles in detailed table format (with --all)
demo_list_table_detailed() {
    simulate_command "hawkling list --profile demo --all"
    echo "NAME                    ARN                                             CREATED                  LAST USED                DESCRIPTION"
    echo "AdminRole               arn:aws:iam::123456789012:role/AdminRole       2023-01-15T10:30:45Z     2025-04-01T14:22:18Z     Administrator role for full access"
    echo "DataScientistRole       arn:aws:iam::123456789012:role/DataScientistRole 2023-04-22T08:15:32Z  2025-03-29T09:41:05Z     Role for data science team with S3 and SageMaker access"
    echo "ReadOnlyRole            arn:aws:iam::123456789012:role/ReadOnlyRole    2022-11-05T16:45:20Z     2025-04-01T10:11:35Z     Read-only access for auditing"
    echo "DeploymentRole          arn:aws:iam::123456789012:role/DeploymentRole  2024-01-30T11:20:15Z     2025-03-15T16:30:22Z     Role for CI/CD deployment pipelines"
    echo "LambdaExecutionRole     arn:aws:iam::123456789012:role/LambdaExecRole  2023-09-10T14:05:38Z     Never                    Execution role for Lambda functions"
    echo "OldServiceRole          arn:aws:iam::123456789012:role/OldServiceRole  2022-03-18T09:22:51Z     2023-10-05T13:45:17Z     Legacy role for deprecated service"
    echo
}

# Demo 3: List only used roles
demo_list_used() {
    simulate_command "hawkling list --profile demo --used"
    echo "NAME                    LAST USED                DESCRIPTION"
    echo "AdminRole               2025-04-01T14:22:18Z     Administrator role for full access"
    echo "DataScientistRole       2025-03-29T09:41:05Z     Role for data science team with S3 and SageMaker access"
    echo "ReadOnlyRole            2025-04-01T10:11:35Z     Read-only access for auditing"
    echo "DeploymentRole          2025-03-15T16:30:22Z     Role for CI/CD deployment pipelines"
    echo "OldServiceRole          2023-10-05T13:45:17Z     Legacy role for deprecated service"
    echo
}

# Demo 4: List all roles in JSON format
demo_list_json() {
    simulate_command "hawkling list --profile demo --output json"
    echo '[
  {
    "Name": "AdminRole",
    "Arn": "arn:aws:iam::123456789012:role/AdminRole",
    "Description": "Administrator role for full access",
    "CreateDate": "2023-01-15T10:30:45Z",
    "LastUsed": "2025-04-01T14:22:18Z"
  },
  {
    "Name": "DataScientistRole",
    "Arn": "arn:aws:iam::123456789012:role/DataScientistRole",
    "Description": "Role for data science team with S3 and SageMaker access",
    "CreateDate": "2023-04-22T08:15:32Z",
    "LastUsed": "2025-03-29T09:41:05Z"
  },
  {
    "Name": "ReadOnlyRole",
    "Arn": "arn:aws:iam::123456789012:role/ReadOnlyRole",
    "Description": "Read-only access for auditing",
    "CreateDate": "2022-11-05T16:45:20Z",
    "LastUsed": "2025-04-01T10:11:35Z"
  },
  {
    "Name": "DeploymentRole",
    "Arn": "arn:aws:iam::123456789012:role/DeploymentRole",
    "Description": "Role for CI/CD deployment pipelines",
    "CreateDate": "2024-01-30T11:20:15Z",
    "LastUsed": "2025-03-15T16:30:22Z"
  },
  {
    "Name": "LambdaExecutionRole",
    "Arn": "arn:aws:iam::123456789012:role/LambdaExecRole",
    "Description": "Execution role for Lambda functions",
    "CreateDate": "2023-09-10T14:05:38Z",
    "LastUsed": null
  },
  {
    "Name": "OldServiceRole",
    "Arn": "arn:aws:iam::123456789012:role/OldServiceRole",
    "Description": "Legacy role for deprecated service",
    "CreateDate": "2022-03-18T09:22:51Z",
    "LastUsed": "2023-10-05T13:45:17Z"
  }
]'
    echo
}


# Demo 5: List unused roles (over 180 days)
demo_list_unused() {
    simulate_command "hawkling list --profile demo --days 180"
    echo "NAME                    LAST USED                DESCRIPTION"
    echo "OldServiceRole          2023-10-05T13:45:17Z     Legacy role for deprecated service"
    echo "LambdaExecutionRole     Never                    Execution role for Lambda functions"
    echo
}

# Demo 6: Delete a specific role
demo_delete_role() {
    simulate_command "hawkling delete OldServiceRole --profile demo"
    echo "Are you sure you want to delete role OldServiceRole? [y/N]: y"
    sleep 0.5
    echo "Successfully deleted role: OldServiceRole"
    echo
}

# Demo 7: Delete with dry run
demo_delete_dryrun() {
    simulate_command "hawkling delete LambdaExecutionRole --profile demo --dry-run"
    echo "Would delete role: LambdaExecutionRole (dry run)"
    echo
}

# Demo 8: Prune unused roles
demo_prune() {
    simulate_command "hawkling prune --profile demo --days 180"
    echo "Found 2 unused roles (not used in the last 180 days):"
    echo "1. OldServiceRole"
    echo "2. LambdaExecutionRole"
    echo "Do you want to delete these roles? [y/N]: y"
    sleep 0.5
    echo "Deleting role: OldServiceRole"
    echo "Successfully deleted role: OldServiceRole"
    echo "Deleting role: LambdaExecutionRole"
    echo "Successfully deleted role: LambdaExecutionRole"
    echo
}

# Main demo
echo -e "${YELLOW}=== HAWKLING COMMANDS ===${NC}\n"
display_help
display_list_help

echo -e "${YELLOW}=== LISTING ROLES - TABLE FORMAT (DEFAULT/SIMPLIFIED) ===${NC}\n"
echo -e "This is the default output format, displaying roles in a simplified tabular format.\n"
demo_list_table_simple

echo -e "${YELLOW}=== LISTING ROLES - TABLE FORMAT WITH --ALL FLAG (DETAILED) ===${NC}\n"
echo -e "Adding the --all flag shows more detailed information including ARN and creation date.\n"
demo_list_table_detailed

echo -e "${YELLOW}=== LISTING ROLES - USED ROLES ONLY ===${NC}\n"
echo -e "Showing only roles that have been used at least once (excluding never-used roles).\n"
demo_list_used

echo -e "${YELLOW}=== LISTING ROLES - JSON FORMAT ===${NC}\n"
echo -e "JSON format is useful for programmatic access or piping to other tools.\n"
demo_list_json


echo -e "${YELLOW}=== LISTING UNUSED ROLES (OVER 180 DAYS) ===${NC}\n"
echo -e "Filter roles that haven't been used for a specified number of days.\n"
demo_list_unused

echo -e "${YELLOW}=== DELETING A SPECIFIC ROLE ===${NC}\n"
echo -e "Delete a specific IAM role by name.\n"
demo_delete_role

echo -e "${YELLOW}=== DELETING A ROLE (DRY RUN) ===${NC}\n"
echo -e "Simulate role deletion without actually deleting it.\n"
demo_delete_dryrun

echo -e "${YELLOW}=== PRUNING UNUSED ROLES ===${NC}\n"
echo -e "Delete multiple unused IAM roles at once.\n"
demo_prune

echo -e "${GREEN}Demo completed. Use the --all flag to show detailed information including ARN and creation date. Use the --used flag to show only roles that have been used at least once.${NC}"