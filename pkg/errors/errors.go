package errors

import "fmt"

// WrapRoleError wraps an error with a role operation context
func WrapRoleError(op string, roleName string, err error) error {
	return fmt.Errorf("failed to %s role %s: %w", op, roleName, err)
}

// WrapAPIError wraps an error with an API operation context
func WrapAPIError(op string, err error) error {
	return fmt.Errorf("failed to %s: %w", op, err)
}

// ValidationError represents a user input validation error
type ValidationError struct {
	Message string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string) ValidationError {
	return ValidationError{Message: message}
}
