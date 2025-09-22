import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { AlertTriangle, RefreshCw } from 'lucide-react'
import { APIError, formatErrorForDisplay } from '@/lib/errors'

interface ErrorDisplayProps {
  error: APIError
  onRetry?: () => void
  className?: string
}

export function ErrorDisplay({ error, onRetry, className }: ErrorDisplayProps) {
  const errorInfo = formatErrorForDisplay(error)
  
  return (
    <Alert variant="destructive" className={className}>
      <AlertTriangle className="h-4 w-4" />
      <AlertTitle>{errorInfo.title}</AlertTitle>
      <AlertDescription>
        <p className="mb-3">{errorInfo.message}</p>
        
        {errorInfo.suggestions && errorInfo.suggestions.length > 0 && (
          <div className="mb-4">
            <p className="font-medium mb-2">Suggestions:</p>
            <ul className="list-disc list-inside space-y-1 text-sm">
              {errorInfo.suggestions.map((suggestion, index) => (
                <li key={index}>{suggestion}</li>
              ))}
            </ul>
          </div>
        )}
        
        {onRetry && errorInfo.canRetry && (
          <Button 
            onClick={onRetry}
            variant="outline" 
            size="sm"
            className="mt-2"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Try Again
          </Button>
        )}
        
        {error.details && (
          <details className="mt-3 text-xs">
            <summary className="cursor-pointer">Technical details</summary>
            <pre className="mt-2 p-2 bg-red-50 dark:bg-red-900/20 rounded overflow-x-auto">
              {JSON.stringify(error.details, null, 2)}
            </pre>
          </details>
        )}
      </AlertDescription>
    </Alert>
  )
}