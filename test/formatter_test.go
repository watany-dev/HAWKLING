package test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"hawkling/pkg/aws"
	"hawkling/pkg/formatter"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{
			input:    "short string",
			length:   20,
			expected: "short string",
		},
		{
			input:    "this is a longer string that should be truncated",
			length:   17,
			expected: "this is a long...",
		},
		{
			input:    "",
			length:   10,
			expected: "",
		},
	}

	for _, test := range tests {
		result := formatter.TruncateString(test.input, test.length)
		if result != test.expected {
			t.Errorf("expected %q, got %q", test.expected, result)
		}
	}
}

func TestFormatRolesAsJSON(t *testing.T) {
	now := time.Now()
	lastUsed := now.Add(-24 * time.Hour)
	
	roles := []aws.Role{
		{
			Name:        "Role1",
			Arn:         "arn:aws:iam::123456789012:role/Role1",
			Description: "Test role 1",
			CreateDate:  now.Add(-48 * time.Hour),
			LastUsed:    &lastUsed,
		},
		{
			Name:        "Role2",
			Arn:         "arn:aws:iam::123456789012:role/Role2",
			Description: "Test role 2",
			CreateDate:  now.Add(-72 * time.Hour),
			LastUsed:    nil,
		},
	}

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := formatter.FormatRolesAsJSON(roles)
	if err != nil {
		t.Fatalf("FormatRolesAsJSON() error = %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf = make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Parse the JSON output
	var parsedRoles []aws.Role
	err = json.Unmarshal([]byte(output), &parsedRoles)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Check if the parsed roles match the input roles
	if len(parsedRoles) != len(roles) {
		t.Errorf("expected %d roles, got %d", len(roles), len(parsedRoles))
	}
}