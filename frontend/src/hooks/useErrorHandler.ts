import { useState, useCallback } from 'react'
import { useNotifications } from '@/hooks/useAppState'
import { APIError, ErrorRecovery, isRetryableError, getUserFriendlyMessage, getErrorSuggestions } from '@/lib/errors'
import { AxiosError } from 'axios'

interface UseErrorHandlerReturn {
  handleError: (error: any, context?: string) => void
  isRetrying: boolean
  retryCount: number
  clearError: () => void
}

export function useErrorHandler(): UseErrorHandlerReturn {
  const { addNotification } = useNotifications()
  const [isRetrying, setIsRetrying] = useState(false)
  const [retryCount, setRetryCount] = useState(0)

  const handleError = useCallback((error: any, context?: string) => {
    console.error('Error caught:', error)
    
    // Convert to standard API error format
    let apiError: APIError
    
    if (error.code) {
      // Already an APIError
      apiError = error
    } else if (error.response?.data) {
      // Axios error with response data
      apiError = {
        code: error.response.data.code || 'UNKNOWN_ERROR',
        message: error.response.data.message || error.message || 'An unexpected error occurred',
        details: error.response.data.details,
        validations: error.response.data.validations,
        timestamp: error.response.data.timestamp || new Date().toISOString(),
        request_id: error.response.data.request_id || error.config?.headers?.['X-Request-ID'] || 'unknown',
        path: error.response.data.path,
        method: error.response.data.method,
        user_message: error.response.data.user_message,
        suggestions: error.response.data.suggestions,
        documentation: error.response.data.documentation,
      }
    } else if (error.request) {
      // Network error
      apiError = {
        code: 'NETWORK_ERROR',
        message: 'Network connection failed',
        timestamp: new Date().toISOString(),
        request_id: error.config?.headers?.['X-Request-ID'] || 'unknown',
      }
    } else {
      // Other error
      apiError = {
        code: 'INTERNAL_ERROR',
        message: error.message || 'An unexpected error occurred',
        timestamp: new Date().toISOString(),
        request_id: 'unknown',
      }
    }
    
    // Log the error
    console.error('Processed error:', apiError)
    
    // Check if error is retryable
    const retryable = isRetryableError(apiError)
    
    // Create user-friendly message
    const userMessage = getUserFriendlyMessage(apiError)
    const suggestions = getErrorSuggestions(apiError)
    
    // Show notification
    addNotification({
      type: 'error',
      title: context || 'Error',
      message: userMessage,
    })
    
    // If retryable and we haven't exceeded retry attempts, automatically retry
    if (retryable && retryCount < 3) {
      setIsRetrying(true)
      
      // Calculate delay with exponential backoff
      const delay = Math.min(1000 * Math.pow(2, retryCount), 10000) // Max 10 seconds
      
      setTimeout(() => {
        setRetryCount(prev => prev + 1)
        setIsRetrying(false)
        // The actual retry logic would be handled by the calling component
      }, delay)
    }
  }, [addNotification, retryCount])

  const clearError = useCallback(() => {
    setIsRetrying(false)
    setRetryCount(0)
  }, [])

  return {
    handleError,
    isRetrying,
    retryCount,
    clearError,
  }
}