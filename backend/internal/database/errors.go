package database

import (
	"errors"
	"fmt"
	"strings"
)

// DatabaseError represents a database-specific error
type DatabaseError struct {
	Operation string
	Err       error
	Code      string
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error in %s: %v", e.Operation, e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// Error codes for different database error types
const (
	ErrCodeConnection     = "CONNECTION_ERROR"
	ErrCodeSchema         = "SCHEMA_ERROR"
	ErrCodeConstraint     = "CONSTRAINT_ERROR"
	ErrCodeTransaction    = "TRANSACTION_ERROR"
	ErrCodeQuery          = "QUERY_ERROR"
	ErrCodeTimeout        = "TIMEOUT_ERROR"
	ErrCodeDuplicateKey   = "DUPLICATE_KEY_ERROR"
	ErrCodeForeignKey     = "FOREIGN_KEY_ERROR"
	ErrCodeNotFound       = "NOT_FOUND_ERROR"
)

// Common database errors
var (
	ErrConnectionNotReady = errors.New("database connection not ready")
	ErrSchemaNotFound     = errors.New("database schema not found")
	ErrTransactionFailed  = errors.New("database transaction failed")
	ErrQueryTimeout       = errors.New("database query timeout")
)

// WrapError wraps an error with database context
func WrapError(operation string, err error) error {
	if err == nil {
		return nil
	}

	code := classifyError(err)
	return &DatabaseError{
		Operation: operation,
		Err:       err,
		Code:      code,
	}
}

// classifyError determines the error code based on the error message
func classifyError(err error) string {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "connection"):
		return ErrCodeConnection
	case strings.Contains(errMsg, "constraint"):
		return ErrCodeConstraint
	case strings.Contains(errMsg, "duplicate"):
		return ErrCodeDuplicateKey
	case strings.Contains(errMsg, "foreign key"):
		return ErrCodeForeignKey
	case strings.Contains(errMsg, "timeout"):
		return ErrCodeTimeout
	case strings.Contains(errMsg, "not found"):
		return ErrCodeNotFound
	case strings.Contains(errMsg, "transaction"):
		return ErrCodeTransaction
	case strings.Contains(errMsg, "schema"):
		return ErrCodeSchema
	default:
		return ErrCodeQuery
	}
}

// IsConnectionError checks if the error is a connection-related error
func IsConnectionError(err error) bool {
	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		return dbErr.Code == ErrCodeConnection
	}
	return false
}

// IsConstraintError checks if the error is a constraint violation
func IsConstraintError(err error) bool {
	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		return dbErr.Code == ErrCodeConstraint
	}
	return false
}

// IsDuplicateKeyError checks if the error is a duplicate key violation
func IsDuplicateKeyError(err error) bool {
	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		return dbErr.Code == ErrCodeDuplicateKey
	}
	return false
}

// IsTimeoutError checks if the error is a timeout error
func IsTimeoutError(err error) bool {
	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		return dbErr.Code == ErrCodeTimeout
	}
	return false
}

// RetryableError checks if an error is retryable
func RetryableError(err error) bool {
	return IsConnectionError(err) || IsTimeoutError(err)
}