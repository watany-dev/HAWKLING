package performance

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"hawkling/pkg/aws"
)

// DelayedMockIAMClient is a mock implementation of IAM client with simulated API delays
type DelayedMockIAMClient struct {
	Roles        []aws.Role
	APIDelay     time.Duration // Simulated API delay
	DeletedRoles []string
}

// NewDelayedMockIAMClient creates a new mock IAM client with simulated delay
func NewDelayedMockIAMClient(roleCount int, delay time.Duration) *DelayedMockIAMClient {
	now := time.Now()
	roles := make([]aws.Role, roleCount)

	// Create the specified number of test roles
	for i := 0; i < roleCount; i++ {
		// Create a mix of roles with different last used times
		var lastUsed *time.Time
		switch i % 3 {
		case 0:
			// Recently used role
			t := now.AddDate(0, 0, -(i % 10))
			lastUsed = &t
		case 1:
			// Old role
			t := now.AddDate(0, 0, -100)
			lastUsed = &t
		case 2:
			// Never used role
			lastUsed = nil
		}

		roles[i] = aws.Role{
			Name:        fmt.Sprintf("Role-%d", i),
			Arn:         fmt.Sprintf("arn:aws:iam::123456789012:role/Role-%d", i),
			CreateDate:  now.AddDate(0, -(i % 12), 0),
			LastUsed:    lastUsed,
			Description: fmt.Sprintf("Test role %d", i),
		}
	}

	return &DelayedMockIAMClient{
		Roles:        roles,
		APIDelay:     delay,
		DeletedRoles: []string{},
	}
}

// ListRoles returns all mock IAM roles with simulated API delay
func (m *DelayedMockIAMClient) ListRoles(ctx context.Context) ([]aws.Role, error) {
	// Simulate API delay
	time.Sleep(m.APIDelay)
	return m.Roles, nil
}

// GetRoleLastUsed returns the last used timestamp with simulated API delay
func (m *DelayedMockIAMClient) GetRoleLastUsed(ctx context.Context, roleName string) (*time.Time, error) {
	// Simulate API delay
	time.Sleep(m.APIDelay)
	for _, role := range m.Roles {
		if role.Name == roleName {
			return role.LastUsed, nil
		}
	}
	return nil, nil
}

// DeleteRole simulates deleting an IAM role
func (m *DelayedMockIAMClient) DeleteRole(ctx context.Context, roleName string) error {
	// Simulate API delay
	time.Sleep(m.APIDelay)

	m.DeletedRoles = append(m.DeletedRoles, roleName)

	// Remove the role from the list of roles
	var updatedRoles []aws.Role
	for _, role := range m.Roles {
		if role.Name != roleName {
			updatedRoles = append(updatedRoles, role)
		}
	}
	m.Roles = updatedRoles

	return nil
}

// DetachRolePolicies mocks detaching all managed policies from a role
func (m *DelayedMockIAMClient) DetachRolePolicies(ctx context.Context, roleName string) error {
	// Simulate API delay
	time.Sleep(m.APIDelay)
	return nil
}

// DeleteInlinePolicies mocks deleting all inline policies from a role
func (m *DelayedMockIAMClient) DeleteInlinePolicies(ctx context.Context, roleName string) error {
	// Simulate API delay
	time.Sleep(m.APIDelay)
	return nil
}

// ListAllRolesSequential gets a list of all roles and their last used times sequentially
func ListAllRolesSequential(ctx context.Context, client aws.IAMClient) ([]aws.Role, error) {
	// Get all roles
	roles, err := client.ListRoles(ctx)
	if err != nil {
		return nil, err
	}

	// Get last used time for each role sequentially
	for i := range roles {
		lastUsed, err := client.GetRoleLastUsed(ctx, roles[i].Name)
		if err != nil {
			return nil, err
		}
		roles[i].LastUsed = lastUsed
	}

	return roles, nil
}

// ListAllRolesConcurrent gets a list of all roles and their last used times concurrently
func ListAllRolesConcurrent(ctx context.Context, client aws.IAMClient) ([]aws.Role, error) {
	// Get all roles
	roles, err := client.ListRoles(ctx)
	if err != nil {
		return nil, err
	}

	// Get last used time for each role concurrently
	var wg sync.WaitGroup
	wg.Add(len(roles))

	for i := range roles {
		// Capture the index in the closure to avoid race conditions
		i := i
		go func() {
			defer wg.Done()
			lastUsed, err := client.GetRoleLastUsed(ctx, roles[i].Name)
			if err == nil {
				roles[i].LastUsed = lastUsed
			}
			// Note: Error handling is simplified for this performance test
		}()
	}

	wg.Wait()
	return roles, nil
}

func TestIAMPerformance(t *testing.T) {
	// Configuration
	roleCounts := []int{10, 50, 100}
	apiDelay := 50 * time.Millisecond // Simulated API delay

	fmt.Println("IAM Performance Test")
	fmt.Println("===================")
	fmt.Printf("API Delay: %v\n\n", apiDelay)
	fmt.Println("Role Count | Sequential | Concurrent | Improvement")
	fmt.Println("----------|------------|------------|------------")

	for _, roleCount := range roleCounts {
		ctx := context.Background()

		// Create client with the specified number of roles and API delay
		client := NewDelayedMockIAMClient(roleCount, apiDelay)

		// Test sequential implementation
		seqStart := time.Now()
		seqRoles, err := ListAllRolesSequential(ctx, client)
		seqDuration := time.Since(seqStart)
		if err != nil {
			t.Errorf("Sequential implementation failed: %v", err)
			continue
		}

		// Test concurrent implementation
		concStart := time.Now()
		concRoles, err := ListAllRolesConcurrent(ctx, client)
		concDuration := time.Since(concStart)
		if err != nil {
			t.Errorf("Concurrent implementation failed: %v", err)
			continue
		}

		// Verify results are same length
		if len(seqRoles) != len(concRoles) {
			t.Errorf("Result mismatch: sequential=%d, concurrent=%d",
				len(seqRoles), len(concRoles))
		}

		// Calculate improvement factor
		improvement := float64(seqDuration) / float64(concDuration)

		fmt.Printf("%-10d | %-10s | %-10s | %.2fx\n",
			roleCount,
			seqDuration.Round(time.Millisecond),
			concDuration.Round(time.Millisecond),
			improvement,
		)
	}
}

// BenchmarkSequentialVsConcurrent runs benchmark comparisons
func BenchmarkSequentialVsConcurrent(b *testing.B) {
	// Configuration
	roleCounts := []int{10, 50, 100}
	apiDelay := 10 * time.Millisecond // Use shorter delay for benchmarks

	for _, roleCount := range roleCounts {
		client := NewDelayedMockIAMClient(roleCount, apiDelay)
		ctx := context.Background()

		b.Run(fmt.Sprintf("Sequential-%d", roleCount), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ListAllRolesSequential(ctx, client)
				if err != nil {
					b.Fatalf("error: %v", err)
				}
			}
		})

		b.Run(fmt.Sprintf("Concurrent-%d", roleCount), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ListAllRolesConcurrent(ctx, client)
				if err != nil {
					b.Fatalf("error: %v", err)
				}
			}
		})
	}
}
