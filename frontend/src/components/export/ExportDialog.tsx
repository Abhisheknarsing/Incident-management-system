import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  Dialog, 
  DialogContent, 
  DialogDescription, 
  DialogHeader, 
  DialogTitle, 
  DialogTrigger 
} from '@/components/ui/dialog'
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { useExport } from '@/hooks/useExport'
import { FilterState } from '@/types'
import { 
  Download, 
  FileText, 
  FileSpreadsheet, 
  X, 
  AlertCircle,
  CheckCircle,
  Loader2
} from 'lucide-react'

interface ExportDialogProps {
  trigger?: React.ReactNode
  dataType?: string
  filters?: Partial<FilterState>
  className?: string
}

const dataTypeOptions = [
  { value: 'timeline', label: 'Timeline Data', description: 'Incident trends over time' },
  { value: 'priority', label: 'Priority Analysis', description: 'Priority distribution breakdown' },
  { value: 'application', label: 'Application Analysis', description: 'Application-wise incident data' },
  { value: 'sentiment', label: 'Sentiment Analysis', description: 'Sentiment breakdown and scores' },
  { value: 'resolution', label: 'Resolution Metrics', description: 'Resolution times and trends' },
  { value: 'automation', label: 'Automation Analysis', description: 'Automation opportunities' },
  { value: 'all', label: 'Complete Dashboard', description: 'All analytics data combined' }
]

const formatOptions = [
  { 
    value: 'csv', 
    label: 'CSV', 
    description: 'Comma-separated values for spreadsheet applications',
    icon: FileSpreadsheet
  },
  { 
    value: 'pdf', 
    label: 'PDF', 
    description: 'Formatted report with charts and summaries',
    icon: FileText
  }
]

export function ExportDialog({ 
  trigger, 
  dataType: defaultDataType = 'all',
  filters = {},
  className = ""
}: ExportDialogProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [selectedFormat, setSelectedFormat] = useState<'csv' | 'pdf'>('csv')
  const [selectedDataType, setSelectedDataType] = useState(defaultDataType)
  const [includeFilters, setIncludeFilters] = useState(true)
  
  const { exportData, isExporting, progress, error, cancelExport, isLoading } = useExport()

  const handleExport = () => {
    const exportOptions = {
      format: selectedFormat,
      dataType: selectedDataType,
      filters: includeFilters ? filters : undefined
    }
    
    exportData(exportOptions)
  }

  const handleCancel = () => {
    cancelExport()
    setIsOpen(false)
  }

  const hasActiveFilters = Object.keys(filters).some(key => {
    const filterKey = key as keyof FilterState
    const value = filters[filterKey]
    
    if (filterKey === 'dateRange' && value) {
      return true
    } else if (Array.isArray(value) && value.length > 0) {
      return true
    }
    return false
  })

  const getActiveFilterCount = () => {
    return Object.keys(filters).reduce((count, key) => {
      const filterKey = key as keyof FilterState
      const value = filters[filterKey]
      
      if (filterKey === 'dateRange' && value) {
        return count + 1
      } else if (Array.isArray(value) && value.length > 0) {
        return count + value.length
      }
      return count
    }, 0)
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        {trigger || (
          <Button variant="outline" size="sm" className={className}>
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Download className="h-5 w-5" />
            Export Data
          </DialogTitle>
          <DialogDescription>
            Configure your export settings and download your analytics data.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Export Progress */}
          {isExporting && (
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm flex items-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Exporting Data...
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <Progress value={progress} className="w-full" />
                <div className="flex items-center justify-between text-sm text-muted-foreground">
                  <span>{progress}% complete</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleCancel}
                    className="text-muted-foreground hover:text-foreground"
                  >
                    <X className="h-4 w-4 mr-1" />
                    Cancel
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Error Display */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Export Configuration */}
          {!isExporting && (
            <>
              {/* Data Type Selection */}
              <div className="space-y-3">
                <Label className="text-sm font-medium">Data Type</Label>
                <Select value={selectedDataType} onValueChange={setSelectedDataType}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select data type to export" />
                  </SelectTrigger>
                  <SelectContent>
                    {dataTypeOptions.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        <div>
                          <div className="font-medium">{option.label}</div>
                          <div className="text-xs text-muted-foreground">
                            {option.description}
                          </div>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Format Selection */}
              <div className="space-y-3">
                <Label className="text-sm font-medium">Export Format</Label>
                <div className="grid grid-cols-2 gap-3">
                  {formatOptions.map((format) => {
                    const Icon = format.icon
                    return (
                      <Card
                        key={format.value}
                        className={`cursor-pointer transition-colors ${
                          selectedFormat === format.value
                            ? 'border-primary bg-primary/5'
                            : 'hover:border-primary/50'
                        }`}
                        onClick={() => setSelectedFormat(format.value as 'csv' | 'pdf')}
                      >
                        <CardContent className="p-4">
                          <div className="flex items-center gap-3">
                            <Icon className="h-5 w-5 text-primary" />
                            <div className="flex-1">
                              <div className="font-medium">{format.label}</div>
                              <div className="text-xs text-muted-foreground">
                                {format.description}
                              </div>
                            </div>
                            {selectedFormat === format.value && (
                              <CheckCircle className="h-4 w-4 text-primary" />
                            )}
                          </div>
                        </CardContent>
                      </Card>
                    )
                  })}
                </div>
              </div>

              {/* Filter Options */}
              <div className="space-y-3">
                <Label className="text-sm font-medium">Filter Options</Label>
                <div className="space-y-3">
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="include-filters"
                      checked={includeFilters}
                      onCheckedChange={(checked) => setIncludeFilters(checked === true)}
                    />
                    <Label htmlFor="include-filters" className="text-sm">
                      Apply current filters to export
                    </Label>
                  </div>
                  
                  {hasActiveFilters && (
                    <div className="ml-6 space-y-2">
                      <div className="text-xs text-muted-foreground">
                        Active filters ({getActiveFilterCount()}):
                      </div>
                      <div className="flex flex-wrap gap-1">
                        {filters.dateRange && (
                          <Badge variant="outline" className="text-xs">
                            Date: {filters.dateRange.start} to {filters.dateRange.end}
                          </Badge>
                        )}
                        {filters.priorities?.map(priority => (
                          <Badge key={priority} variant="outline" className="text-xs">
                            Priority: {priority}
                          </Badge>
                        ))}
                        {filters.applications?.map(app => (
                          <Badge key={app} variant="outline" className="text-xs">
                            App: {app}
                          </Badge>
                        ))}
                        {filters.statuses?.map(status => (
                          <Badge key={status} variant="outline" className="text-xs">
                            Status: {status}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  )}
                  
                  {!hasActiveFilters && (
                    <div className="ml-6 text-xs text-muted-foreground">
                      No active filters - all data will be exported
                    </div>
                  )}
                </div>
              </div>

              {/* Export Actions */}
              <div className="flex gap-3 pt-4">
                <Button
                  onClick={handleExport}
                  disabled={isLoading}
                  className="flex-1"
                >
                  {isLoading ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Starting Export...
                    </>
                  ) : (
                    <>
                      <Download className="h-4 w-4 mr-2" />
                      Export {selectedFormat.toUpperCase()}
                    </>
                  )}
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setIsOpen(false)}
                  disabled={isLoading}
                >
                  Cancel
                </Button>
              </div>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}