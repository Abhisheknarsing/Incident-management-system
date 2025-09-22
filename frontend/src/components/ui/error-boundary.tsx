import React, { Component, ErrorInfo, ReactNode } from 'react'
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { ErrorInfo as ErrorInfoType } from '@/lib/errors'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
  errorInfo?: ErrorInfo
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
  }

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({ error, errorInfo })
    console.error('Uncaught error:', error, errorInfo)
    
    // Log to error tracking service
    const errorData: ErrorInfoType = {
      componentStack: errorInfo.componentStack,
      errorBoundary: 'GlobalErrorBoundary',
      eventType: 'react_error',
    }
    
    // In a real application, you would send this to an error tracking service
    console.error('Error boundary caught error:', { error, errorInfo: errorData })
  }

  private handleRetry = () => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined })
  }

  public render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <div className="min-h-screen flex items-center justify-center p-4">
          <Alert variant="destructive" className="max-w-2xl">
            <AlertTitle>Something went wrong</AlertTitle>
            <AlertDescription>
              <p className="mb-4">
                We're sorry, but something went wrong. Please try refreshing the page.
              </p>
              <details className="mb-4 text-sm whitespace-pre-wrap">
                <summary className="cursor-pointer font-medium">Error details</summary>
                <pre className="mt-2 p-2 bg-red-50 dark:bg-red-900/20 rounded">
                  {this.state.error?.toString()}
                  {this.state.errorInfo?.componentStack}
                </pre>
              </details>
              <div className="flex gap-2">
                <Button onClick={() => window.location.reload()}>
                  Refresh Page
                </Button>
                <Button variant="outline" onClick={this.handleRetry}>
                  Try Again
                </Button>
              </div>
            </AlertDescription>
          </Alert>
        </div>
      )
    }

    return this.props.children
  }
}