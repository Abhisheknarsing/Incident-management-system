import { useState, useEffect } from 'react'
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Loader2, RefreshCw } from 'lucide-react'
import { APIError } from '@/lib/errors'

interface ErrorRecoveryProps {
  error: APIError
  onRetry: () => void
  isRetrying?: boolean
  retryCount?: number
  maxRetries?: number
}

export function ErrorRecovery({ 
  error, 
  onRetry, 
  isRetrying = false, 
  retryCount = 0,
  maxRetries = 3 
}: ErrorRecoveryProps) {
  const [countdown, setCountdown] = useState(0)
  
  // Auto-retry with countdown
  useEffect(() => {
    if (retryCount > 0 && retryCount <= maxRetries && !isRetrying) {
      setCountdown(5) // 5 second countdown
      
      const timer = setInterval(() => {
        setCountdown(prev => {
          if (prev <= 1) {
            clearInterval(timer)
            onRetry()
            return 0
          }
          return prev - 1
        })
      }, 1000)
      
      return () => clearInterval(timer)
    }
  }, [retryCount, maxRetries, isRetrying, onRetry])
  
  const canRetry = retryCount < maxRetries
  
  return (
    <Alert variant="warning" className="max-w-2xl mx-auto">
      <AlertTitle>Connection Issue</AlertTitle>
      <AlertDescription>
        <p className="mb-3">
          We're having trouble connecting to the server. This might be due to:
        </p>
        <ul className="list-disc list-inside space-y-1 mb-4">
          <li>Network connectivity issues</li>
          <li>Server maintenance</li>
          <li>Temporary service disruption</li>
        </ul>
        
        {isRetrying ? (
          <div className="flex items-center justify-center py-4">
            <Loader2 className="h-6 w-6 animate-spin mr-2" />
            <span>Retrying...</span>
          </div>
        ) : countdown > 0 ? (
          <div className="text-center py-2">
            <p className="mb-2">Automatically retrying in {countdown} seconds...</p>
            <Button onClick={onRetry} variant="outline">
              Retry Now
            </Button>
          </div>
        ) : (
          <div className="flex flex-col sm:flex-row gap-2">
            <Button 
              onClick={onRetry} 
              disabled={!canRetry || isRetrying}
              className="flex-1"
            >
              <RefreshCw className={`h-4 w-4 mr-2 ${isRetrying ? 'animate-spin' : ''}`} />
              {isRetrying ? 'Retrying...' : 'Retry Connection'}
            </Button>
            
            {!canRetry && (
              <Button variant="outline" onClick={() => window.location.reload()}>
                Refresh Page
              </Button>
            )}
          </div>
        )}
        
        {!canRetry && (
          <p className="text-sm text-muted-foreground mt-3">
            Maximum retry attempts reached. Please refresh the page or try again later.
          </p>
        )}
      </AlertDescription>
    </Alert>
  )
}