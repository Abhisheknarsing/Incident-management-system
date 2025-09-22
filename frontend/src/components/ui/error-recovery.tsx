import { Button } from '@/components/ui/button'
import { RotateCcw, X } from 'lucide-react'

interface ErrorRecoveryProps {
  onRetry: () => void
  onDismiss: () => void
  message?: string
}

export function ErrorRecovery({ onRetry, onDismiss, message }: ErrorRecoveryProps) {
  return (
    <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 flex items-center justify-between">
      <div>
        <p className="text-sm font-medium text-destructive">
          {message || 'Something went wrong'}
        </p>
      </div>
      <div className="flex space-x-2">
        <Button
          variant="outline"
          size="sm"
          onClick={onRetry}
          className="h-8 px-3 text-xs"
        >
          <RotateCcw className="h-3 w-3 mr-1" />
          Retry
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={onDismiss}
          className="h-8 px-3 text-xs"
        >
          <X className="h-3 w-3" />
        </Button>
      </div>
    </div>
  )
}