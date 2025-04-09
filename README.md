# Hawkling

Hawkling is a command-line tool for managing AWS IAM roles, with a focus on identifying and cleaning up unused roles. It provides functionality for listing all IAM roles, detecting unused roles, and safely deleting them either individually or in bulk.

## Features

- List all IAM roles in your AWS account
- Identify roles that haven't been used for a specified period
- Safely delete individual roles with confirmation prompts
- Bulk delete unused roles with optional dry-run mode
- Support for different output formats (table or JSON)

## Installation

### Using Go

```bash
go install github.com/yourusername/hawkling/cmd/hawkling@latest
```

### From Source

```bash
git clone https://github.com/yourusername/hawkling.git
cd hawkling
go build -o hawkling ./cmd/hawkling
```

## Usage

Hawkling offers several commands with various options:

### Global Options

- `--profile` - AWS profile to use (optional)
- `--region` - AWS region (defaults to us-east-1)

### Commands

#### List all IAM roles

```bash
hawkling list -o table --profile myprofile --region us-east-1
```

Options:
- `-o, --output` - Output format: `table` or `json` (default: table)

#### Find unused IAM roles

```bash
hawkling unused --days 90
```

Options:
- `--days` - Number of days to consider a role as unused (default: 90)
- `-o, --output` - Output format: `table` or `json` (default: table)

#### Delete a specific role

```bash
hawkling delete MyUnusedRole --dry-run
```

Options:
- `--dry-run` - Simulate deletion without actually deleting
- `--force` - Delete without confirmation

#### Prune (bulk delete) unused roles

```bash
hawkling prune --days 90
hawkling prune --days 90 --force
```

Options:
- `--days` - Number of days to consider a role as unused (default: 90)
- `--dry-run` - Simulate deletion without actually deleting
- `--force` - Delete without confirmation

## Examples

### List all roles in a specific AWS account

```bash
hawkling list --profile production
```

### Find roles not used in the last 180 days

```bash
hawkling unused --days 180
```

### Delete an unused role (with confirmation)

```bash
hawkling delete OldServiceRole
```

### Delete an unused role (without confirmation)

```bash
hawkling delete OldServiceRole --force
```

### Prune all roles not used in the last 30 days (dry run mode)

```bash
hawkling prune --days 30 --dry-run
```

## Security Considerations

Hawkling requires IAM permissions to list and delete roles. It's recommended to use it with an IAM user or role that has appropriate permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iam:ListRoles",
                "iam:GetRole",
                "iam:DeleteRole",
                "iam:ListRolePolicies",
                "iam:DeleteRolePolicy",
                "iam:ListAttachedRolePolicies",
                "iam:DetachRolePolicy"
            ],
            "Resource": "*"
        }
    ]
}
```

## Development

### Requirements

- Go 1.19 or higher
- AWS SDK for Go v2

### Building from source

```bash
go build -o hawkling ./cmd/hawkling
```

### Running tests

```bash
go test ./...
```

## License

[MIT License](LICENSE)