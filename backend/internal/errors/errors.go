package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// File and Upload Errors
	ErrMissingFile        ErrorCode = "MISSING_FILE"
	ErrFileTooLarge      ErrorCode = "FILE_TOO_LARGE"
	ErrInvalidFileFormat ErrorCode = "INVALID_FILE_FORMAT"
	ErrUploadNotFound    ErrorCode = "UPLOAD_NOT_FOUND"
	ErrMissingUploadID   ErrorCode = "MISSING_UPLOAD_ID"
	ErrInvalidStatus     ErrorCode = "INVALID_STATUS"

	// Processing Errors
	ErrProcessingFailed   ErrorCode = "PROCESSING_FAILED"
	ErrValidationError    ErrorCode = "VALIDATION_ERROR"
	ErrRequiredFieldMissing ErrorCode = "REQUIRED_FIELD_MISSING"
	ErrInvalidDateFormat  ErrorCode = "INVALID_DATE_FORMAT"
	ErrDuplicateIncidentID ErrorCode = "DUPLICATE_INCIDENT_ID"

	// Database Errors
	ErrDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrConnectionFailed   ErrorCode = "CONNECTION_FAILED"
	ErrQueryTimeout       ErrorCode = "QUERY_TIMEOUT"
	ErrTransactionFailed  ErrorCode = "TRANSACTION_FAILED"

	// API Errors
	ErrInvalidParameter   ErrorCode = "INVALID_PARAMETER"
	ErrMissingParameter   ErrorCode = "MISSING_PARAMETER"
	ErrUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrForbidden          ErrorCode = "FORBIDDEN"
	ErrRateLimited        ErrorCode = "RATE_LIMITED"

	// Export Errors
	ErrExportFailed       ErrorCode = "EXPORT_FAILED"
	ErrUnsupportedFormat  ErrorCode = "UNSUPPORTED_FORMAT"
	ErrExportTimeout      ErrorCode = "EXPORT_TIMEOUT"

	// Performance Errors
	ErrPerformanceDegradation ErrorCode = "PERFORMANCE_DEGRADATION"
	ErrResourceExhausted      ErrorCode = "RESOURCE_EXHAUSTED"
	ErrServiceUnavailable     ErrorCode = "SERVICE_UNAVAILABLE"

	// Internal Errors
	ErrInternalServer     ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrNotImplemented     ErrorCode = "NOT_IMPLEMENTED"
	ErrConfigurationError ErrorCode = "CONFIGURATION_ERROR"
)

// ValidationError represents a field-level validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
	Row     *int        `json:"row,omitempty"`
	Column  *string     `json:"column,omitempty"`
}

// APIError represents a standardized API error response
type APIError struct {
	Code         ErrorCode          `json:"code"`
	Message      string             `json:"message"`
	Details      interface{}        `json:"details,omitempty"`
	Validations  []ValidationError  `json:"validations,omitempty"`
	Timestamp    time.Time          `json:"timestamp"`
	RequestID    string             `json:"request_id"`
	Path         string             `json:"path,omitempty"`
	Method       string             `json:"method,omitempty"`
	UserMessage  string             `json:"user_message,omitempty"`
	Suggestions  []string           `json:"suggestions,omitempty"`
	Documentation string            `json:"documentation,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string) *APIError {
	return &APIError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// NewValidationError creates a new validation error
func NewValidationError(code ErrorCode, message string, validations []ValidationError) *APIError {
	return &APIError{
		Code:        code,
		Message:     message,
		Validations: validations,
		Timestamp:   time.Now(),
	}
}

// WithDetails adds details to the error
func (e *APIError) WithDetails(details interface{}) *APIError {
	e.Details = details
	return e
}

// WithRequestID adds request ID to the error
func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// WithPath adds request path to the error
func (e *APIError) WithPath(path string) *APIError {
	e.Path = path
	return e
}

// WithMethod adds request method to the error
func (e *APIError) WithMethod(method string) *APIError {
	e.Method = method
	return e
}

// WithUserMessage adds a user-friendly message
func (e *APIError) WithUserMessage(message string) *APIError {
	e.UserMessage = message
	return e
}

// WithSuggestions adds suggestions for fixing the error
func (e *APIError) WithSuggestions(suggestions []string) *APIError {
	e.Suggestions = suggestions
	return e
}

// WithDocumentation adds documentation link
func (e *APIError) WithDocumentation(doc string) *APIError {
	e.Documentation = doc
	return e
}

// ToJSON converts the error to JSON
func (e *APIError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// GetHTTPStatus returns the appropriate HTTP status code for the error
func (e *APIError) GetHTTPStatus() int {
	switch e.Code {
	case ErrMissingFile, ErrFileTooLarge, ErrInvalidFileFormat, ErrMissingUploadID,
		 ErrInvalidStatus, ErrInvalidParameter, ErrMissingParameter, ErrValidationError,
		 ErrRequiredFieldMissing, ErrInvalidDateFormat, ErrDuplicateIncidentID,
		 ErrUnsupportedFormat:
		return http.StatusBadRequest
	case ErrUploadNotFound:
		return http.StatusNotFound
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrRateLimited:
		return http.StatusTooManyRequests
	case ErrQueryTimeout, ErrExportTimeout:
		return http.StatusRequestTimeout
	case ErrServiceUnavailable, ErrPerformanceDegradation:
		return http.StatusServiceUnavailable
	case ErrNotImplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

// Common error constructors
func BadRequest(message string) *APIError {
	return NewAPIError(ErrInvalidParameter, message)
}

func NotFound(resource string) *APIError {
	return NewAPIError(ErrUploadNotFound, fmt.Sprintf("%s not found", resource))
}

func InternalServer(message string) *APIError {
	return NewAPIError(ErrInternalServer, message)
}

func DatabaseError(operation string, err error) *APIError {
	return NewAPIError(ErrDatabaseError, fmt.Sprintf("Database operation failed: %s", operation)).
		WithDetails(err.Error())
}

func ValidationFailed(validations []ValidationError) *APIError {
	return NewValidationError(ErrValidationError, "Validation failed", validations)
}

func ProcessingFailed(details string) *APIError {
	return NewAPIError(ErrProcessingFailed, "Data processing failed").
		WithDetails(details).
		WithUserMessage("There was an error processing your file. Please check the data format and try again.").
		WithSuggestions([]string{
			"Ensure all required fields are present",
			"Check date formats (YYYY-MM-DD)",
			"Verify incident IDs are unique",
			"Remove any special characters from text fields",
		})
}

func FileUploadError(reason string) *APIError {
	suggestions := []string{
		"Ensure the file is in Excel format (.xlsx or .xls)",
		"Check that the file size is under 50MB",
		"Verify the file is not corrupted",
	}
	
	var code ErrorCode
	var userMessage string
	
	switch reason {
	case "file_too_large":
		code = ErrFileTooLarge
		userMessage = "The uploaded file is too large. Please use a file smaller than 50MB."
	case "invalid_format":
		code = ErrInvalidFileFormat
		userMessage = "The uploaded file format is not supported. Please upload an Excel file (.xlsx or .xls)."
	default:
		code = ErrInvalidFileFormat
		userMessage = "There was an error with the uploaded file. Please try again."
	}
	
	return NewAPIError(code, reason).
		WithUserMessage(userMessage).
		WithSuggestions(suggestions)
}

// ErrorHandler is a Gin middleware for centralized error handling
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			var apiError *APIError
			
			// Check if it's already an APIError
			if ae, ok := err.Err.(*APIError); ok {
				apiError = ae
			} else {
				// Convert generic error to APIError
				apiError = InternalServer(err.Error())
			}
			
			// Add request context
			apiError.WithRequestID(c.GetString("request_id")).
				WithPath(c.Request.URL.Path).
				WithMethod(c.Request.Method)
			
			// Send error response
			c.JSON(apiError.GetHTTPStatus(), apiError)
			return
		}
	}
}

// SendError sends a standardized error response
func SendError(c *gin.Context, err *APIError) {
	// Add request context if not already present
	if err.RequestID == "" {
		err.WithRequestID(c.GetString("request_id"))
	}
	if err.Path == "" {
		err.WithPath(c.Request.URL.Path)
	}
	if err.Method == "" {
		err.WithMethod(c.Request.Method)
	}
	
	c.JSON(err.GetHTTPStatus(), err)
}

// AbortWithError aborts the request with an error
func AbortWithError(c *gin.Context, err *APIError) {
	SendError(c, err)
	c.Abort()
}

// RecoveryHandler is a Gin middleware for panic recovery
func RecoveryHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		err := InternalServer("Internal server error occurred").
			WithRequestID(c.GetString("request_id")).
			WithPath(c.Request.URL.Path).
			WithMethod(c.Request.Method).
			WithDetails(fmt.Sprintf("Panic: %v", recovered))
		
		c.JSON(err.GetHTTPStatus(), err)
		c.Abort()
	})
}

// WrapDatabaseError wraps database errors with context
func WrapDatabaseError(operation string, err error) error {
	if err == nil {
		return nil
	}
	
	return DatabaseError(operation, err)
}

// WrapValidationErrors converts validation errors to APIError
func WrapValidationErrors(errors []ValidationError) *APIError {
	if len(errors) == 0 {
		return nil
	}
	
	return ValidationFailed(errors).
		WithUserMessage("Please correct the validation errors and try again.")
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		switch apiErr.Code {
		case ErrDatabaseError, ErrConnectionFailed, ErrQueryTimeout, 
			 ErrServiceUnavailable, ErrPerformanceDegradation:
			return true
		}
	}
	return false
}

// GetErrorSeverity returns the severity level of an error
func GetErrorSeverity(err error) string {
	if apiErr, ok := err.(*APIError); ok {
		switch apiErr.Code {
		case ErrInternalServer, ErrDatabaseError, ErrConnectionFailed, ErrConfigurationError:
			return "critical"
		case ErrProcessingFailed, ErrQueryTimeout, ErrServiceUnavailable:
			return "high"
		case ErrValidationError, ErrInvalidParameter, ErrUploadNotFound:
			return "medium"
		case ErrMissingParameter, ErrInvalidDateFormat:
			return "low"
		}
	}
	return "unknown"
}