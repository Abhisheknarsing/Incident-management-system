import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { LoadingSpinner } from '@/components/ui/loading-spinner'
import { useUploads, useUploadStatus, useStartAnalysis } from '@/hooks/useUploads'
import { 
  FileSpreadsheet, 
  Play, 
  AlertCircle, 
  CheckCircle, 
  Clock, 
  RefreshCw,
  TrendingUp,
  Database,
  BarChart3,
  FileText
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Upload } from '@/types'






interface AnalysisConfirmationProps {
  upload: Upload
  onConfirm: () => void
  onCancel: () => void
  isLoading: boolean
}

function AnalysisConfirmation({ upload, onConfirm, onCancel, isLoading }: AnalysisConfirmationProps) {
  return (
    <Alert>
      <TrendingUp className="h-4 w-4" />
      <AlertDescription>
        <div className="space-y-3">
          <div>
            <p className="font-medium">Start Analysis for "{upload.original_filename}"?</p>
            <p className="text-sm text-muted-foreground mt-1">
              This will process {upload.record_count} records and generate sentiment analysis, 
              automation opportunities, and other insights. This process may take several minutes.
            </p>
          </div>
          <div className="flex space-x-2">
            <Button onClick={onConfirm} disabled={isLoading} size="sm">
              {isLoading && <LoadingSpinner className="mr-2 h-3 w-3" />}
              Start Analysis
            </Button>
            <Button variant="outline" onClick={onCancel} disabled={isLoading} size="sm">
              Cancel
            </Button>
          </div>
        </div>
      </AlertDescription>
    </Alert>
  )
}

interface UploadCardProps {
  upload: Upload
  onAnalyze: (upload: Upload) => void
  isAnalyzing: boolean
}

function UploadCard({ upload, onAnalyze, isAnalyzing }: UploadCardProps) {
  const { error: statusError } = useUploadStatus(
    upload.id, 
    upload.status === 'processing'
  )

  const getStatusIcon = () => {
    switch (upload.status) {
      case 'completed':
        return <CheckCircle className="h-5 w-5 text-green-600" />
      case 'failed':
        return <AlertCircle className="h-5 w-5 text-red-600" />
      case 'processing':
        return <LoadingSpinner className="h-5 w-5" />
      default:
        return <Clock className="h-5 w-5 text-yellow-600" />
    }
  }

  const getStatusColor = () => {
    switch (upload.status) {
      case 'completed':
        return 'default'
      case 'failed':
        return 'destructive'
      case 'processing':
        return 'secondary'
      default:
        return 'outline'
    }
  }

  const canAnalyze = upload.status === 'uploaded' && !isAnalyzing

  return (
    <Card className={cn(
      "transition-all duration-200",
      upload.status === 'processing' && "ring-2 ring-blue-200"
    )}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center space-x-3">
            <FileSpreadsheet className="h-6 w-6 text-muted-foreground" />
            <div>
              <CardTitle className="text-base">{upload.original_filename}</CardTitle>
              <CardDescription className="text-sm">
                Uploaded {new Date(upload.created_at).toLocaleDateString()} at{' '}
                {new Date(upload.created_at).toLocaleTimeString()}
              </CardDescription>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            {getStatusIcon()}
            <Badge variant={getStatusColor()}>{upload.status}</Badge>
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Upload Statistics */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
          <div className="flex items-center space-x-2">
            <Database className="h-4 w-4 text-muted-foreground" />
            <span className="text-muted-foreground">Records:</span>
            <span className="font-medium">{upload.record_count.toLocaleString()}</span>
          </div>
          
          {upload.processed_count > 0 && (
            <div className="flex items-center space-x-2">
              <BarChart3 className="h-4 w-4 text-muted-foreground" />
              <span className="text-muted-foreground">Processed:</span>
              <span className="font-medium">{upload.processed_count.toLocaleString()}</span>
            </div>
          )}
          
          {upload.error_count > 0 && (
            <div className="flex items-center space-x-2">
              <AlertCircle className="h-4 w-4 text-red-500" />
              <span className="text-muted-foreground">Errors:</span>
              <span className="font-medium text-red-600">{upload.error_count}</span>
            </div>
          )}
          
          {upload.processed_at && (
            <div className="flex items-center space-x-2">
              <CheckCircle className="h-4 w-4 text-green-600" />
              <span className="text-muted-foreground">Completed:</span>
              <span className="font-medium text-green-600">
                {new Date(upload.processed_at).toLocaleDateString()}
              </span>
            </div>
          )}
        </div>



        {/* Error Display */}
        {upload.status === 'failed' && upload.errors && upload.errors.length > 0 ? (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              <div className="space-y-1">
                <p className="font-medium">Processing failed:</p>
                <ul className="text-sm space-y-1">
                  {upload.errors.slice(0, 3).map((error, index) => (
                    <li key={index}>• {error}</li>
                  ))}
                  {upload.errors.length > 3 ? (
                    <li className="text-muted-foreground">
                      ... and {upload.errors.length - 3} more errors
                    </li>
                  ) : null}
                </ul>
              </div>
            </AlertDescription>
          </Alert>
        ) : null}

        {/* Status Error */}
        {statusError ? (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Failed to fetch processing status. Please refresh the page.
            </AlertDescription>
          </Alert>
        ) : null}

        {/* Action Buttons */}
        <div className="flex items-center justify-between pt-2">
          <div className="text-xs text-muted-foreground">
            ID: {upload.id}
          </div>
          
          <div className="flex space-x-2">
            {canAnalyze && (
              <Button 
                onClick={() => onAnalyze(upload)}
                disabled={isAnalyzing}
                size="sm"
              >
                <Play className="h-3 w-3 mr-1" />
                Analyze Data
              </Button>
            )}
            
            {upload.status === 'failed' && (
              <Button 
                variant="outline" 
                onClick={() => onAnalyze(upload)}
                disabled={isAnalyzing}
                size="sm"
              >
                <RefreshCw className="h-3 w-3 mr-1" />
                Retry Analysis
              </Button>
            )}
            
            {upload.status === 'completed' && (
              <Button variant="outline" size="sm" asChild>
                <a href="/dashboard">
                  <BarChart3 className="h-3 w-3 mr-1" />
                  View Dashboard
                </a>
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export function ProcessingPage() {
  const [selectedUpload, setSelectedUpload] = useState<Upload | null>(null)
  const { data: uploads, isLoading, error, refetch } = useUploads()
  const startAnalysisMutation = useStartAnalysis()

  const handleAnalyzeClick = (upload: Upload) => {
    setSelectedUpload(upload)
  }

  const handleConfirmAnalysis = async () => {
    if (!selectedUpload) return

    try {
      await startAnalysisMutation.mutateAsync(selectedUpload.id)
      setSelectedUpload(null)
    } catch (error) {
      console.error('Failed to start analysis:', error)
    }
  }

  const handleCancelAnalysis = () => {
    setSelectedUpload(null)
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold">Data Processing</h1>
          <p className="text-muted-foreground">
            Manage and monitor incident data processing
          </p>
        </div>
        <div className="flex items-center justify-center py-12">
          <LoadingSpinner className="h-8 w-8" />
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold">Data Processing</h1>
          <p className="text-muted-foreground">
            Manage and monitor incident data processing
          </p>
        </div>
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription className="flex items-center justify-between">
            <span>Failed to load uploads. Please try again.</span>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              <RefreshCw className="h-3 w-3 mr-1" />
              Retry
            </Button>
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Data Processing</h1>
        <p className="text-muted-foreground">
          Manage and monitor incident data processing
        </p>
      </div>

      {/* Analysis Confirmation */}
      {selectedUpload && (
        <AnalysisConfirmation
          upload={selectedUpload}
          onConfirm={handleConfirmAnalysis}
          onCancel={handleCancelAnalysis}
          isLoading={startAnalysisMutation.isPending}
        />
      )}

      {/* Upload List */}
      {uploads && uploads.length > 0 ? (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold">Uploaded Files</h2>
            <Button variant="outline" onClick={() => refetch()} size="sm">
              <RefreshCw className="h-3 w-3 mr-1" />
              Refresh
            </Button>
          </div>
          
          <div className="space-y-4">
            {uploads.map((upload) => (
              <UploadCard
                key={upload.id}
                upload={upload}
                onAnalyze={handleAnalyzeClick}
                isAnalyzing={startAnalysisMutation.isPending}
              />
            ))}
          </div>
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center text-center p-8">
          <FileText className="h-12 w-12 mb-4 text-muted-foreground" />
          <h3 className="text-lg font-semibold mb-2">No uploads found</h3>
          <p className="text-muted-foreground mb-4 max-w-md">
            Upload an Excel file to get started with incident data analysis.
          </p>
          <Button asChild>
            <a href="/">Upload File</a>
          </Button>
        </div>
      )}

      {/* Processing Statistics */}
      {uploads && uploads.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Processing Summary</CardTitle>
            <CardDescription>
              Overview of your uploaded files and processing status
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="text-center">
                <div className="text-2xl font-bold text-blue-600">
                  {uploads.length}
                </div>
                <div className="text-sm text-muted-foreground">Total Uploads</div>
              </div>
              
              <div className="text-center">
                <div className="text-2xl font-bold text-green-600">
                  {uploads.filter(u => u.status === 'completed').length}
                </div>
                <div className="text-sm text-muted-foreground">Completed</div>
              </div>
              
              <div className="text-center">
                <div className="text-2xl font-bold text-yellow-600">
                  {uploads.filter(u => u.status === 'processing').length}
                </div>
                <div className="text-sm text-muted-foreground">Processing</div>
              </div>
              
              <div className="text-center">
                <div className="text-2xl font-bold text-purple-600">
                  {uploads.reduce((sum, u) => sum + u.record_count, 0).toLocaleString()}
                </div>
                <div className="text-sm text-muted-foreground">Total Records</div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}