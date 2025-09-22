import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { FilterState } from '@/types'
import { DateRangePicker } from './DateRangePicker'
import { MultiSelectFilter } from './MultiSelectFilter'
import { X, Filter, RotateCcw } from 'lucide-react'

interface FilterPanelProps {
  filters: Partial<FilterState>
  onFiltersChange: (filters: Partial<FilterState>) => void
  availableOptions?: {
    priorities?: string[]
    applications?: string[]
    statuses?: string[]
  }
  className?: string
}

export function FilterPanel({
  filters,
  onFiltersChange,
  availableOptions = {},
  className = ""
}: FilterPanelProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  
  // Default options if not provided
  const defaultPriorities = ['P1', 'P2', 'P3', 'P4']
  const defaultStatuses = ['Open', 'In Progress', 'Resolved', 'Closed']
  
  const priorities = availableOptions.priorities || defaultPriorities
  const applications = availableOptions.applications || []
  const statuses = availableOptions.statuses || defaultStatuses

  const handleDateRangeChange = (dateRange: { start: string; end: string } | undefined) => {
    onFiltersChange({
      ...filters,
      dateRange
    })
  }

  const handlePriorityChange = (selectedPriorities: string[]) => {
    onFiltersChange({
      ...filters,
      priorities: selectedPriorities
    })
  }

  const handleApplicationChange = (selectedApplications: string[]) => {
    onFiltersChange({
      ...filters,
      applications: selectedApplications
    })
  }

  const handleStatusChange = (selectedStatuses: string[]) => {
    onFiltersChange({
      ...filters,
      statuses: selectedStatuses
    })
  }

  const clearAllFilters = () => {
    onFiltersChange({})
  }

  const removeFilter = (filterType: keyof FilterState, value?: string) => {
    const newFilters = { ...filters }
    
    if (filterType === 'dateRange') {
      delete newFilters.dateRange
    } else if (value && Array.isArray(newFilters[filterType])) {
      const currentValues = newFilters[filterType] as string[]
      newFilters[filterType] = currentValues.filter(v => v !== value)
      if (newFilters[filterType]?.length === 0) {
        delete newFilters[filterType]
      }
    } else {
      delete newFilters[filterType]
    }
    
    onFiltersChange(newFilters)
  }

  // Count active filters
  const activeFilterCount = Object.keys(filters).reduce((count, key) => {
    const filterKey = key as keyof FilterState
    const value = filters[filterKey]
    
    if (filterKey === 'dateRange' && value) {
      return count + 1
    } else if (Array.isArray(value) && value.length > 0) {
      return count + value.length
    }
    return count
  }, 0)

  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            <Filter className="h-5 w-5" />
            Filters
            {activeFilterCount > 0 && (
              <Badge variant="secondary" className="ml-2">
                {activeFilterCount}
              </Badge>
            )}
          </CardTitle>
          <div className="flex items-center gap-2">
            {activeFilterCount > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={clearAllFilters}
                className="text-muted-foreground hover:text-foreground"
              >
                <RotateCcw className="h-4 w-4 mr-1" />
                Clear All
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsExpanded(!isExpanded)}
            >
              {isExpanded ? 'Collapse' : 'Expand'}
            </Button>
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Active Filters Display */}
        {activeFilterCount > 0 && (
          <div className="space-y-2">
            <h4 className="text-sm font-medium text-muted-foreground">Active Filters:</h4>
            <div className="flex flex-wrap gap-2">
              {filters.dateRange && (
                <Badge variant="outline" className="gap-1">
                  Date: {filters.dateRange.start} to {filters.dateRange.end}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto p-0 ml-1"
                    onClick={() => removeFilter('dateRange')}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </Badge>
              )}
              
              {filters.priorities?.map(priority => (
                <Badge key={priority} variant="outline" className="gap-1">
                  Priority: {priority}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto p-0 ml-1"
                    onClick={() => removeFilter('priorities', priority)}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </Badge>
              ))}
              
              {filters.applications?.map(app => (
                <Badge key={app} variant="outline" className="gap-1">
                  App: {app}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto p-0 ml-1"
                    onClick={() => removeFilter('applications', app)}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </Badge>
              ))}
              
              {filters.statuses?.map(status => (
                <Badge key={status} variant="outline" className="gap-1">
                  Status: {status}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto p-0 ml-1"
                    onClick={() => removeFilter('statuses', status)}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* Filter Controls */}
        {(isExpanded || activeFilterCount === 0) && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Date Range</label>
              <DateRangePicker
                value={filters.dateRange}
                onChange={handleDateRangeChange}
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">Priority</label>
              <MultiSelectFilter
                options={priorities}
                selectedValues={filters.priorities || []}
                onChange={handlePriorityChange}
                placeholder="Select priorities..."
              />
            </div>

            {applications.length > 0 && (
              <div className="space-y-2">
                <label className="text-sm font-medium">Application</label>
                <MultiSelectFilter
                  options={applications}
                  selectedValues={filters.applications || []}
                  onChange={handleApplicationChange}
                  placeholder="Select applications..."
                />
              </div>
            )}

            <div className="space-y-2">
              <label className="text-sm font-medium">Status</label>
              <MultiSelectFilter
                options={statuses}
                selectedValues={filters.statuses || []}
                onChange={handleStatusChange}
                placeholder="Select statuses..."
              />
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}