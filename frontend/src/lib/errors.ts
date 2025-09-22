// Error types and utilities for frontend error handling

export interface APIError {
  code: string
  message: string
  details?: any
  validations?: ValidationError[]
  timestamp: string
  request_id: string
  path?: string
  method?: string
  user_message?: string
  suggestions?: string[]
  documentation?: string
}

export interface ValidationError {
  field: string
  value: any
  message: string
  row?: number
  column?: string
}

export interface ErrorState {
  hasError: boolean
  error: APIError | null
  isRetrying: boolean
  retryCount: number
  lastRetryAt?: Date
}

export interface RetryConfig {
  maxRetries: number
  baseDelay: number
  maxDelay: number
  backoffFactor: number
  retryableErrors: string[]
}

// Error codes that match backend
export const ErrorCodes = {
  // File and Upload Errors
  MISSING_FILE: 'MISSING_FILE',
  FILE_TOO_LARGE: 'FILE_TOO_LARGE',
  INVALID_FILE_FORMAT: 'INVALID_FILE_FORMAT',
  UPLOAD_NOT_FOUND: 'UPLOAD_NOT_FOUND',
  MISSING_UPLOAD_ID: 'MISSING_UPLOAD_ID',
  INVALID_STATUS: 'INVALID_STATUS',

  // Processing Errors
  PROCESSING_FAILED: 'PROCESSING_FAILED',
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  REQUIRED_FIELD_MISSING: 'REQUIRED_FIELD_MISSING',
  INVALID_DATE_FORMAT: 'INVALID_DATE_FORMAT',
  DUPLICATE_INCIDENT_ID: 'DUPLICATE_INCIDENT_ID',

  // Database Errors
  DATABASE_ERROR: 'DATABASE_ERROR',
  CONNECTION_FAILED: 'CONNECTION_FAILED',
  QUERY_TIMEOUT: 'QUERY_TIMEOUT',
  TRANSACTION_FAILED: 'TRANSACTION_FAILED',

  // API Errors
  INVALID_PARAMETER: 'INVALID_PARAMETER',
  MISSING_PARAMETER: 'MISSING_PARAMETER',
  UNAUTHORIZED: 'UNAUTHORIZED',
  FORBIDDEN: 'FORBIDDEN',
  RATE_LIMITED: 'RATE_LIMITED',

  // Export Errors
  EXPORT_FAILED: 'EXPORT_FAILED',
  UNSUPPORTED_FORMAT: 'UNSUPPORTED_FORMAT',
  EXPORT_TIMEOUT: 'EXPORT_TIMEOUT',

  // Performance Errors
  PERFORMANCE_DEGRADATION: 'PERFORMANCE_DEGRADATION',
  RESOURCE_EXHAUSTED: 'RESOURCE_EXHAUSTED',
  SERVICE_UNAVAILABLE: 'SERVICE_UNAVAILABLE',

  // Network Errors
  NETWORK_ERROR: 'NETWORK_ERROR',
  TIMEOUT_ERROR: 'TIMEOUT_ERROR',
  CONNECTION_ERROR: 'CONNECTION_ERROR',

  // Internal Errors
  INTERNAL_SERVER_ERROR: 'INTERNAL_SERVER_ERROR',
  NOT_IMPLEMENTED: 'NOT_IMPLEMENTED',
  CONFIGURATION_ERROR: 'CONFIGURATION_ERROR',
} as const

export type ErrorCode = typeof ErrorCodes[keyof typeof ErrorCodes]

// Default retry configuration
export const DEFAULT_RETRY_CONFIG: RetryConfig = {
  maxRetries: 3,
  baseDelay: 1000, // 1 second
  maxDelay: 10000, // 10 seconds
  backoffFactor: 2,
  retryableErrors: [
    ErrorCodes.DATABASE_ERROR,
    ErrorCodes.CONNECTION_FAILED,
    ErrorCodes.QUERY_TIMEOUT,
    ErrorCodes.SERVICE_UNAVAILABLE,
    ErrorCodes.PERFORMANCE_DEGRADATION,
    ErrorCodes.NETWORK_ERROR,
    ErrorCodes.TIMEOUT_ERROR,
    ErrorCodes.CONNECTION_ERROR,
  ],
}

// Error severity levels
export enum ErrorSeverity {
  LOW = 'low',
  MEDIUM = 'medium',
  HIGH = 'high',
  CRITICAL = 'critical',
}

// Get error severity based on error code
export function getErrorSeverity(error: APIError): ErrorSeverity {
  switch (error.code) {
    case ErrorCodes.INTERNAL_SERVER_ERROR:
    case ErrorCodes.DATABASE_ERROR:
    case ErrorCodes.CONNECTION_FAILED:
    case ErrorCodes.CONFIGURATION_ERROR:
      return ErrorSeverity.CRITICAL

    case ErrorCodes.PROCESSING_FAILED:
    case ErrorCodes.QUERY_TIMEOUT:
    case ErrorCodes.SERVICE_UNAVAILABLE:
    case ErrorCodes.EXPORT_FAILED:
      return ErrorSeverity.HIGH

    case ErrorCodes.VALIDATION_ERROR:
    case ErrorCodes.INVALID_PARAMETER:
    case ErrorCodes.UPLOAD_NOT_FOUND:
    case ErrorCodes.INVALID_FILE_FORMAT:
      return ErrorSeverity.MEDIUM

    case ErrorCodes.MISSING_PARAMETER:
    case ErrorCodes.INVALID_DATE_FORMAT:
    case ErrorCodes.MISSING_FILE:
      return ErrorSeverity.LOW

    default:
      return ErrorSeverity.MEDIUM
  }
}

// Check if an error is retryable
export function isRetryableError(error: APIError, config: RetryConfig = DEFAULT_RETRY_CONFIG): boolean {
  return config.retryableErrors.includes(error.code)
}

// Calculate retry delay with exponential backoff
export function calculateRetryDelay(retryCount: number, config: RetryConfig = DEFAULT_RETRY_CONFIG): number {
  const delay = config.baseDelay * Math.pow(config.backoffFactor, retryCount)
  return Math.min(delay, config.maxDelay)
}

// Get user-friendly error message
export function getUserFriendlyMessage(error: APIError): string {
  if (error.user_message) {
    return error.user_message
  }

  // Fallback messages based on error code
  switch (error.code) {
    case ErrorCodes.MISSING_FILE:
      return 'Please select a file to upload.'

    case ErrorCodes.FILE_TOO_LARGE:
      return 'The selected file is too large. Please choose a file smaller than 50MB.'

    case ErrorCodes.INVALID_FILE_FORMAT:
      return 'Invalid file format. Please upload an Excel file (.xlsx or .xls).'

    case ErrorCodes.PROCESSING_FAILED:
      return 'There was an error processing your file. Please check the data format and try again.'

    case ErrorCodes.VALIDATION_ERROR:
      return 'Some data in your file is invalid. Please review the errors and correct them.'

    case ErrorCodes.DATABASE_ERROR:
    case ErrorCodes.CONNECTION_FAILED:
      return 'We\'re experiencing technical difficulties. Please try again in a moment.'

    case ErrorCodes.SERVICE_UNAVAILABLE:
      return 'The service is temporarily unavailable. Please try again later.'

    case ErrorCodes.NETWORK_ERROR:
    case ErrorCodes.TIMEOUT_ERROR:
      return 'Network connection issue. Please check your internet connection and try again.'

    case ErrorCodes.UNAUTHORIZED:
      return 'You are not authorized to perform this action.'

    case ErrorCodes.FORBIDDEN:
      return 'Access denied. You don\'t have permission to access this resource.'

    case ErrorCodes.RATE_LIMITED:
      return 'Too many requests. Please wait a moment before trying again.'

    default:
      return error.message || 'An unexpected error occurred. Please try again.'
  }
}

// Get error suggestions
export function getErrorSuggestions(error: APIError): string[] {
  if (error.suggestions && error.suggestions.length > 0) {
    return error.suggestions
  }

  // Default suggestions based on error code
  switch (error.code) {
    case ErrorCodes.INVALID_FILE_FORMAT:
      return [
        'Ensure the file is in Excel format (.xlsx or .xls)',
        'Check that the file is not corrupted',
        'Try saving the file in a different Excel format',
      ]

    case ErrorCodes.FILE_TOO_LARGE:
      return [
        'Reduce the file size by removing unnecessary data',
        'Split large files into smaller chunks',
        'Compress the file before uploading',
      ]

    case ErrorCodes.PROCESSING_FAILED:
      return [
        'Check that all required fields are present',
        'Verify date formats are correct (YYYY-MM-DD)',
        'Ensure incident IDs are unique',
        'Remove special characters from text fields',
      ]

    case ErrorCodes.VALIDATION_ERROR:
      return [
        'Review the validation errors below',
        'Correct the highlighted fields',
        'Ensure data types match expected formats',
      ]

    case ErrorCodes.NETWORK_ERROR:
    case ErrorCodes.TIMEOUT_ERROR:
      return [
        'Check your internet connection',
        'Try refreshing the page',
        'Wait a moment and try again',
      ]

    case ErrorCodes.DATABASE_ERROR:
    case ErrorCodes.SERVICE_UNAVAILABLE:
      return [
        'Wait a few minutes and try again',
        'Contact support if the problem persists',
        'Check the system status page',
      ]

    default:
      return [
        'Try refreshing the page',
        'Check your internet connection',
        'Contact support if the problem continues',
      ]
  }
}

// Create a standardized error object
export function createError(
  code: ErrorCode,
  message: string,
  details?: any,
  userMessage?: string,
  suggestions?: string[]
): APIError {
  return {
    code,
    message,
    details,
    user_message: userMessage,
    suggestions,
    timestamp: new Date().toISOString(),
    request_id: crypto.randomUUID(),
  }
}

// Convert network errors to APIError
export function convertNetworkError(error: any): APIError {
  if (error.code) {
    // Already an APIError
    return error
  }

  // Handle Axios errors
  if (error.response) {
    // Server responded with error status
    return error.response.data || createError(
      ErrorCodes.INTERNAL_SERVER_ERROR,
      'Server error occurred',
      error.response.status
    )
  } else if (error.request) {
    // Network error
    return createError(
      ErrorCodes.NETWORK_ERROR,
      'Network connection failed',
      error.message
    )
  } else {
    // Other error
    return createError(
      ErrorCodes.INTERNAL_SERVER_ERROR,
      error.message || 'An unexpected error occurred'
    )
  }
}

// Error logging utility
export function logError(error: APIError, context?: Record<string, any>) {
  const logData = {
    error,
    context,
    timestamp: new Date().toISOString(),
    userAgent: navigator.userAgent,
    url: window.location.href,
  }

  console.error('Application Error:', logData)

  // In production, you might want to send this to an error tracking service
  // Example: Sentry, LogRocket, etc.
  if (import.meta.env.PROD) {
    // Send to error tracking service
    // errorTrackingService.captureError(logData)
  }
}

// Error recovery utilities
export class ErrorRecovery {
  static async withRetry<T>(
    operation: () => Promise<T>,
    config: RetryConfig = DEFAULT_RETRY_CONFIG
  ): Promise<T> {
    let lastError: APIError
    
    for (let attempt = 0; attempt <= config.maxRetries; attempt++) {
      try {
        return await operation()
      } catch (error) {
        lastError = convertNetworkError(error)
        
        if (attempt === config.maxRetries || !isRetryableError(lastError, config)) {
          throw lastError
        }
        
        const delay = calculateRetryDelay(attempt, config)
        await new Promise(resolve => setTimeout(resolve, delay))
      }
    }
    
    throw lastError!
  }

  static createRetryableOperation<T>(
    operation: () => Promise<T>,
    config?: Partial<RetryConfig>
  ) {
    const finalConfig = { ...DEFAULT_RETRY_CONFIG, ...config }
    return () => this.withRetry(operation, finalConfig)
  }
}

// Error boundary error info
export interface ErrorInfo {
  componentStack: string
  errorBoundary?: string
  eventType?: string
}

// Format error for display
export function formatErrorForDisplay(error: APIError): {
  title: string
  message: string
  suggestions: string[]
  severity: ErrorSeverity
  canRetry: boolean
} {
  const severity = getErrorSeverity(error)
  const canRetry = isRetryableError(error)
  
  return {
    title: getErrorTitle(error),
    message: getUserFriendlyMessage(error),
    suggestions: getErrorSuggestions(error),
    severity,
    canRetry,
  }
}

function getErrorTitle(error: APIError): string {
  switch (error.code) {
    case ErrorCodes.MISSING_FILE:
    case ErrorCodes.FILE_TOO_LARGE:
    case ErrorCodes.INVALID_FILE_FORMAT:
      return 'File Upload Error'

    case ErrorCodes.PROCESSING_FAILED:
    case ErrorCodes.VALIDATION_ERROR:
      return 'Data Processing Error'

    case ErrorCodes.DATABASE_ERROR:
    case ErrorCodes.CONNECTION_FAILED:
      return 'System Error'

    case ErrorCodes.NETWORK_ERROR:
    case ErrorCodes.TIMEOUT_ERROR:
      return 'Connection Error'

    case ErrorCodes.UNAUTHORIZED:
    case ErrorCodes.FORBIDDEN:
      return 'Access Error'

    case ErrorCodes.SERVICE_UNAVAILABLE:
      return 'Service Unavailable'

    default:
      return 'Error'
  }
}