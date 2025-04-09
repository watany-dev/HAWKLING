package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// AWSclient implements the IAMClient interface
type AWSClient struct {
	iamClient *iam.Client
}

// NewAWSClient creates a new AWS client with the specified profile and region
func NewAWSClient(ctx context.Context, profile, region string) (*AWSClient, error) {
	var cfg aws.Config
	var err error

	if profile != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithSharedConfigProfile(profile),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AWSClient{
		iamClient: iam.NewFromConfig(cfg),
	}, nil
}

// ListRoles returns all IAM roles
func (c *AWSClient) ListRoles(ctx context.Context) ([]Role, error) {
	var roles []Role

	paginator := iam.NewListRolesPaginator(c.iamClient, &iam.ListRolesInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list roles: %w", err)
		}

		for _, r := range output.Roles {
			role := Role{
				Name:       *r.RoleName,
				Arn:        *r.Arn,
				CreateDate: *r.CreateDate,
			}

			if r.Description != nil {
				role.Description = *r.Description
			}

			// We'll get last used info separately
			roles = append(roles, role)
		}
	}

	// Concurrently fetch last used info for each role
	type roleResult struct {
		index    int
		lastUsed *time.Time
		err      error
	}

	// Use a semaphore to limit concurrency
	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)
	results := make(chan roleResult, len(roles))

	for i := range roles {
		go func(i int) {
			sem <- struct{}{} // Acquire semaphore
			defer func() {
				<-sem // Release semaphore
			}()

			lastUsed, err := c.GetRoleLastUsed(ctx, roles[i].Name)
			results <- roleResult{
				index:    i,
				lastUsed: lastUsed,
				err:      err,
			}
		}(i)
	}

	// Collect results
	for i := 0; i < len(roles); i++ {
		result := <-results
		if result.err != nil {
			return nil, result.err
		}
		roles[result.index].LastUsed = result.lastUsed
	}

	return roles, nil
}

// GetRoleLastUsed returns the last used timestamp for a role
func (c *AWSClient) GetRoleLastUsed(ctx context.Context, roleName string) (*time.Time, error) {
	resp, err := c.iamClient.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get role %s: %w", roleName, err)
	}

	if resp.Role.RoleLastUsed.LastUsedDate == nil {
		return nil, nil // Role never used
	}

	return resp.Role.RoleLastUsed.LastUsedDate, nil
}

// DeleteRole deletes an IAM role
func (c *AWSClient) DeleteRole(ctx context.Context, roleName string) error {
	// First detach all policies
	if err := c.detachRolePolicies(ctx, roleName); err != nil {
		return err
	}

	// Delete all inline policies
	if err := c.deleteInlinePolicies(ctx, roleName); err != nil {
		return err
	}

	// Delete the role
	_, err := c.iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete role %s: %w", roleName, err)
	}

	return nil
}

// detachRolePolicies detaches all managed policies from a role
func (c *AWSClient) detachRolePolicies(ctx context.Context, roleName string) error {
	paginator := iam.NewListAttachedRolePoliciesPaginator(c.iamClient, &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list attached policies for role %s: %w", roleName, err)
		}

		for _, policy := range output.AttachedPolicies {
			_, err := c.iamClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
				RoleName:  aws.String(roleName),
				PolicyArn: policy.PolicyArn,
			})
			if err != nil {
				return fmt.Errorf("failed to detach policy %s from role %s: %w", *policy.PolicyArn, roleName, err)
			}
		}
	}

	return nil
}

// deleteInlinePolicies deletes all inline policies from a role
func (c *AWSClient) deleteInlinePolicies(ctx context.Context, roleName string) error {
	paginator := iam.NewListRolePoliciesPaginator(c.iamClient, &iam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list inline policies for role %s: %w", roleName, err)
		}

		for _, policyName := range output.PolicyNames {
			_, err := c.iamClient.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
				RoleName:   aws.String(roleName),
				PolicyName: aws.String(policyName),
			})
			if err != nil {
				return fmt.Errorf("failed to delete inline policy %s from role %s: %w", policyName, roleName, err)
			}
		}
	}

	return nil
}
