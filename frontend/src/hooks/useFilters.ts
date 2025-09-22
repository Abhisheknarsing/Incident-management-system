import { useState, useEffect, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { FilterState } from '@/types'

interface UseFiltersOptions {
  persistToUrl?: boolean
  defaultFilters?: Partial<FilterState>
}

export function useFilterState(options: UseFiltersOptions = {}) {
  const { persistToUrl = true, defaultFilters = {} } = options
  const [searchParams, setSearchParams] = useSearchParams()
  
  // Initialize filters from URL params or defaults
  const initializeFilters = useCallback((): Partial<FilterState> => {
    if (!persistToUrl) {
      return defaultFilters
    }

    const filters: Partial<FilterState> = { ...defaultFilters }

    // Parse date range from URL
    const startDate = searchParams.get('start_date')
    const endDate = searchParams.get('end_date')
    if (startDate && endDate) {
      filters.dateRange = { start: startDate, end: endDate }
    }

    // Parse priorities from URL
    const priorities = searchParams.get('priorities')
    if (priorities) {
      filters.priorities = priorities.split(',').filter(Boolean)
    }

    // Parse applications from URL
    const applications = searchParams.get('applications')
    if (applications) {
      filters.applications = applications.split(',').filter(Boolean)
    }

    // Parse statuses from URL
    const statuses = searchParams.get('statuses')
    if (statuses) {
      filters.statuses = statuses.split(',').filter(Boolean)
    }

    return filters
  }, [searchParams, persistToUrl, defaultFilters])

  const [filters, setFilters] = useState<Partial<FilterState>>(initializeFilters)

  // Update URL when filters change
  useEffect(() => {
    if (!persistToUrl) return

    const newSearchParams = new URLSearchParams()

    // Add date range to URL
    if (filters.dateRange?.start && filters.dateRange?.end) {
      newSearchParams.set('start_date', filters.dateRange.start)
      newSearchParams.set('end_date', filters.dateRange.end)
    }

    // Add priorities to URL
    if (filters.priorities?.length) {
      newSearchParams.set('priorities', filters.priorities.join(','))
    }

    // Add applications to URL
    if (filters.applications?.length) {
      newSearchParams.set('applications', filters.applications.join(','))
    }

    // Add statuses to URL
    if (filters.statuses?.length) {
      newSearchParams.set('statuses', filters.statuses.join(','))
    }

    // Update URL without triggering navigation
    setSearchParams(newSearchParams, { replace: true })
  }, [filters, persistToUrl, setSearchParams])

  const updateFilters = useCallback((newFilters: Partial<FilterState>) => {
    setFilters(newFilters)
  }, [])

  const clearFilters = useCallback(() => {
    setFilters({})
  }, [])

  const setFilter = useCallback(<K extends keyof FilterState>(
    key: K,
    value: FilterState[K] | undefined
  ) => {
    setFilters(prev => {
      const newFilters = { ...prev }
      if (value === undefined) {
        delete newFilters[key]
      } else {
        newFilters[key] = value
      }
      return newFilters
    })
  }, [])

  const hasActiveFilters = useCallback(() => {
    return Object.keys(filters).some(key => {
      const filterKey = key as keyof FilterState
      const value = filters[filterKey]
      
      if (filterKey === 'dateRange' && value) {
        return true
      } else if (Array.isArray(value) && value.length > 0) {
        return true
      }
      return false
    })
  }, [filters])

  const getActiveFilterCount = useCallback(() => {
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
  }, [filters])

  // Convert filters to API parameters
  const getApiParams = useCallback(() => {
    const params: Record<string, any> = {}
    
    if (filters.dateRange?.start) {
      params.start_date = filters.dateRange.start
    }
    if (filters.dateRange?.end) {
      params.end_date = filters.dateRange.end
    }
    if (filters.priorities?.length) {
      params.priorities = filters.priorities.join(',')
    }
    if (filters.applications?.length) {
      params.applications = filters.applications.join(',')
    }
    if (filters.statuses?.length) {
      params.statuses = filters.statuses.join(',')
    }
    
    return Object.keys(params).length > 0 ? params : undefined
  }, [filters])

  return {
    filters,
    updateFilters,
    clearFilters,
    setFilter,
    hasActiveFilters: hasActiveFilters(),
    activeFilterCount: getActiveFilterCount(),
    apiParams: getApiParams()
  }
}