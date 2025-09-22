import { useState, useCallback, useRef } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { LoadingSpinner } from '@/components/ui/loading-spinner'
import { useUploadFile, useUploads } from '@/hooks/useUploads'
import { Upload, Cloud, FileSpreadsheet, AlertCircle, CheckCircle, X, RefreshCw } from 'lucide-react'
import { cn } from '@/lib/utils'

interface FileUploadState {
  file: File | null
  isDragOver: boolean
  uploadProgress: number
  isUploading: boolean
  error: string | null
  preview: {
    name: string
    size: string
    type: string
  } | null
}

export function UploadPage() {
  const [uploadState, setUploadState] = useState<FileUploadState>({
    file: null,
    isDragOver: false,
    uploadProgress: 0,
    isUploading: false,
    error: null,
    preview: null,
  })

  const fileInputRef = useRef<HTMLInputElement>(null)
  const uploadMutation = useUploadFile((progress) => {
    setUploadState(prev => ({ ...prev, uploadProgress: progress }))
  })
  const { data: uploads, refetch: refetchUploads } = useUploads()

  // File validation
  const validateFile = useCallback((file: File): string | null => {
    const allowedTypes = [
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', // .xlsx
      'application/vnd.ms-excel', // .xls
    ]
    
    const allowedExtensions = ['.xlsx', '.xls']
    const fileExtension = file.name.toLowerCase().substring(file.name.lastIndexOf('.'))
    
    if (!allowedTypes.includes(file.type) && !allowedExtensions.includes(fileExtension)) {
      return 'Please select a valid Excel file (.xlsx or .xls)'
    }
    
    if (file.size > 50 * 1024 * 1024) { // 50MB limit
      return 'File size must be less than 50MB'
    }
    
    return null
  }, [])

  // Format file size
  const formatFileSize = useCallback((bytes: number): string => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }, [])

  // Handle file selection
  const handleFileSelect = useCallback((file: File) => {
    const error = validateFile(file)
    
    if (error) {
      setUploadState(prev => ({
        ...prev,
        error,
        file: null,
        preview: null,
      }))
      return
    }

    setUploadState(prev => ({
      ...prev,
      file,
      error: null,
      preview: {
        name: file.name,
        size: formatFileSize(file.size),
        type: file.type,
      },
    }))
  }, [validateFile, formatFileSize])

  // Handle drag events
  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setUploadState(prev => ({ ...prev, isDragOver: true }))
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setUploadState(prev => ({ ...prev, isDragOver: false }))
  }, [])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setUploadState(prev => ({ ...prev, isDragOver: false }))
    
    const files = Array.from(e.dataTransfer.files)
    if (files.length > 0) {
      handleFileSelect(files[0])
    }
  }, [handleFileSelect])

  // Handle file input change
  const handleFileInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (files && files.length > 0) {
      handleFileSelect(files[0])
    }
  }, [handleFileSelect])

  // Handle upload
  const handleUpload = useCallback(async () => {
    if (!uploadState.file) return

    setUploadState(prev => ({ 
      ...prev, 
      isUploading: true, 
      uploadProgress: 0,
      error: null 
    }))

    try {
      await uploadMutation.mutateAsync(uploadState.file)
      setUploadState(prev => ({ ...prev, uploadProgress: 100, isUploading: false }))
      
      // Refresh uploads list to show the newly uploaded file
      refetchUploads()
      
    } catch (error: any) {
      setUploadState(prev => ({
        ...prev,
        isUploading: false,
        uploadProgress: 0,
        error: error.message || 'Upload failed. Please try again.',
      }))
    }
  }, [uploadState.file, uploadMutation, refetchUploads])

  // Clear file selection
  const clearFile = useCallback(() => {
    setUploadState(prev => ({
      ...prev,
      file: null,
      preview: null,
      error: null,
    }))
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }, [])

  // Retry upload
  const retryUpload = useCallback(() => {
    setUploadState(prev => ({
      ...prev,
      error: null,
      uploadProgress: 0,
    }))
    handleUpload()
  }, [handleUpload])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Upload Incident Data</h1>
        <p className="text-muted-foreground">
          Upload Excel files containing incident data for analysis
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>File Upload</CardTitle>
          <CardDescription>
            Select an Excel file (.xlsx, .xls) containing incident data. Maximum file size: 50MB.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Upload Area */}
          <div
            className={cn(
              "border-2 border-dashed rounded-lg p-8 text-center transition-colors cursor-pointer",
              uploadState.isDragOver
                ? "border-primary bg-primary/5"
                : "border-muted-foreground/25 hover:border-muted-foreground/50",
              uploadState.isUploading && "pointer-events-none opacity-50"
            )}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept=".xlsx,.xls,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.ms-excel"
              onChange={handleFileInputChange}
              className="hidden"
              disabled={uploadState.isUploading}
            />
            
            {uploadState.isUploading ? (
              <div className="space-y-4">
                <LoadingSpinner className="mx-auto" />
                <div className="space-y-2">
                  <p className="text-sm font-medium">Uploading...</p>
                  <Progress value={uploadState.uploadProgress} className="w-full max-w-xs mx-auto" />
                  <p className="text-xs text-muted-foreground">{uploadState.uploadProgress}% complete</p>
                </div>
              </div>
            ) : uploadState.preview ? (
              <div className="space-y-4">
                <div className="flex items-center justify-center space-x-2">
                  <FileSpreadsheet className="h-8 w-8 text-green-600" />
                  <CheckCircle className="h-5 w-5 text-green-600" />
                </div>
                <div className="space-y-2">
                  <p className="font-medium">{uploadState.preview.name}</p>
                  <p className="text-sm text-muted-foreground">{uploadState.preview.size}</p>
                  <Badge variant="secondary">Excel File</Badge>
                </div>
                <div className="flex items-center justify-center space-x-2">
                  <Button onClick={handleUpload} disabled={uploadState.isUploading}>
                    <Upload className="h-4 w-4 mr-2" />
                    Upload File
                  </Button>
                  <Button variant="outline" onClick={clearFile}>
                    <X className="h-4 w-4 mr-2" />
                    Clear
                  </Button>
                </div>
              </div>
            ) : (
              <div className="space-y-4">
                <Cloud className="h-12 w-12 mx-auto text-muted-foreground" />
                <div className="space-y-2">
                  <p className="text-lg font-medium">Drop your Excel file here</p>
                  <p className="text-sm text-muted-foreground">
                    or click to browse and select a file
                  </p>
                </div>
                <div className="flex items-center justify-center space-x-4 text-xs text-muted-foreground">
                  <span>Supported: .xlsx, .xls</span>
                  <span>•</span>
                  <span>Max size: 50MB</span>
                </div>
              </div>
            )}
          </div>

          {/* Error Display */}
          {uploadState.error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription className="flex items-center justify-between">
                <span>{uploadState.error}</span>
                {uploadState.file && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={retryUpload}
                    className="ml-2"
                  >
                    <RefreshCw className="h-3 w-3 mr-1" />
                    Retry
                  </Button>
                )}
              </AlertDescription>
            </Alert>
          )}

          {/* Upload Success */}
          {uploadState.uploadProgress === 100 && !uploadState.error && (
            <Alert>
              <CheckCircle className="h-4 w-4" />
              <AlertDescription className="flex items-center justify-between">
                <span>File uploaded successfully! You can now analyze the data from the Processing page.</span>
                <div className="space-x-2">
                  <Button variant="outline" size="sm" asChild>
                    <a href="/processing">Go to Processing</a>
                  </Button>
                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={() => {
                      setUploadState({
                        file: null,
                        isDragOver: false,
                        uploadProgress: 0,
                        isUploading: false,
                        error: null,
                        preview: null,
                      })
                      if (fileInputRef.current) {
                        fileInputRef.current.value = ''
                      }
                    }}
                  >
                    Upload Another
                  </Button>
                </div>
              </AlertDescription>
            </Alert>
          )}

          {/* File Requirements */}
          <div className="bg-muted/50 rounded-lg p-4 space-y-2">
            <h4 className="font-medium text-sm">File Requirements:</h4>
            <ul className="text-xs text-muted-foreground space-y-1">
              <li>• Excel format (.xlsx or .xls)</li>
              <li>• Required columns: incident_id, report_date, brief_description, application_name, resolution_group, resolved_person, priority</li>
              <li>• Optional columns: resolve_date, last_resolve_date, description, category, subcategory, impact, urgency, status</li>
              <li>• Maximum file size: 50MB</li>
              <li>• First row should contain column headers</li>
            </ul>
          </div>
        </CardContent>
      </Card>

      {/* Recent Uploads */}
      {uploads && Array.isArray(uploads) && uploads.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Recent Uploads</CardTitle>
            <CardDescription>
              Your recently uploaded files
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {uploads
                .slice(0, 5)
                .filter((upload) => upload && upload.id) // Filter out invalid uploads
                .map((upload) => (
                  <div
                    key={upload.id} // Ensure each upload has a unique key
                    className="flex items-center justify-between p-3 border rounded-lg"
                  >
                    <div className="flex items-center space-x-3">
                      <FileSpreadsheet className="h-5 w-5 text-muted-foreground" />
                      <div>
                        <p className="font-medium text-sm">{upload.original_filename}</p>
                        <p className="text-xs text-muted-foreground">
                          {new Date(upload.created_at).toLocaleDateString()} • {upload.record_count} records
                        </p>
                      </div>
                    </div>
                    <Badge
                      variant={
                        upload.status === 'completed'
                          ? 'default'
                          : upload.status === 'failed'
                          ? 'destructive'
                          : upload.status === 'processing'
                          ? 'secondary'
                          : 'outline'
                      }
                    >
                      {upload.status}
                    </Badge>
                  </div>
                ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}