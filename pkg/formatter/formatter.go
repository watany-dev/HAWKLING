package formatter

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"hawkling/pkg/aws"
)

// Format represents the output format
type Format string

const (
	// TableFormat outputs data in a tabular format
	TableFormat Format = "table"

	// JSONFormat outputs data in JSON format
	JSONFormat Format = "json"
)

// FormatRoles formats the roles according to the specified format
func FormatRoles(roles []aws.Role, format Format, showAllInfo bool) error {
	switch format {
	case TableFormat:
		return formatRolesAsTable(roles, showAllInfo)
	case JSONFormat:
		return FormatRolesAsJSON(roles)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// formatRolesAsTable prints roles in tabular format
func formatRolesAsTable(roles []aws.Role, showAllInfo bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if showAllInfo {
		fmt.Fprintln(w, "NAME\tARN\tCREATED\tLAST USED\tDESCRIPTION")
	} else {
		fmt.Fprintln(w, "NAME\tLAST USED\tDESCRIPTION")
	}

	for _, role := range roles {
		lastUsed := "Never"
		if role.LastUsed != nil {
			lastUsed = role.LastUsed.Format(time.RFC3339)
		}

		if showAllInfo {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				role.Name,
				role.Arn,
				role.CreateDate.Format(time.RFC3339),
				lastUsed,
				TruncateString(role.Description, 50),
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				role.Name,
				lastUsed,
				TruncateString(role.Description, 50),
			)
		}
	}

	return w.Flush()
}

// FormatRolesAsJSON prints roles in JSON format
func FormatRolesAsJSON(roles []aws.Role) error {
	data, err := json.MarshalIndent(roles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal roles to JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// TruncateString truncates a string if it's longer than the specified length
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}

	return s[:length-3] + "..."
}
