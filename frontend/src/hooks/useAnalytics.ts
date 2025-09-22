import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api'
import { FilterState } from '@/types'

export const ANALYTICS_QUERY_KEYS = {
  timeline: (filters?: Record<string, any>) => ['analytics', 'timeline', filters] as const,
  priorities: (filters?: Record<string, any>) => ['analytics', 'priorities', filters] as const,
  applications: (filters?: Record<string, any>) => ['analytics', 'applications', filters] as const,
  sentiment: (filters?: Record<string, any>) => ['analytics', 'sentiment', filters] as const,
  resolution: (filters?: Record<string, any>) => ['analytics', 'resolution', filters] as const,
  automation: (filters?: Record<string, any>) => ['analytics', 'automation', filters] as const,
  dashboard: (filters?: Record<string, any>) => ['analytics', 'dashboard', filters] as const,
}

function filtersToParams(filters?: Partial<FilterState>) {
  if (!filters) return undefined
  
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
}

export function useTimelineData(filters?: Partial<FilterState>) {
  const params = filtersToParams(filters)
  
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.timeline(params),
    queryFn: () => apiClient.analytics.getTimeline(params),
    staleTime: 1000 * 60 * 5, // 5 minutes
    enabled: true,
  })
}

export function usePriorityAnalysis(filters?: Partial<FilterState>) {
  const params = filtersToParams(filters)
  
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.priorities(params),
    queryFn: () => apiClient.analytics.getPriorities(params),
    staleTime: 1000 * 60 * 5, // 5 minutes
    enabled: true,
  })
}

export function useApplicationAnalysis(filters?: Partial<FilterState>) {
  const params = filtersToParams(filters)
  
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.applications(params),
    queryFn: () => apiClient.analytics.getApplications(params),
    staleTime: 1000 * 60 * 5, // 5 minutes
    enabled: true,
  })
}

export function useSentimentAnalysis(filters?: Partial<FilterState>) {
  const params = filtersToParams(filters)
  
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.sentiment(params),
    queryFn: () => apiClient.analytics.getSentiment(params),
    staleTime: 1000 * 60 * 5, // 5 minutes
    enabled: true,
  })
}

export function useResolutionMetrics(filters?: Partial<FilterState>) {
  const params = filtersToParams(filters)
  
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.resolution(params),
    queryFn: () => apiClient.analytics.getResolution(params),
    staleTime: 1000 * 60 * 5, // 5 minutes
    enabled: true,
  })
}

export function useAutomationAnalysis(filters?: Partial<FilterState>) {
  const params = filtersToParams(filters)
  
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.automation(params),
    queryFn: () => apiClient.analytics.getAutomation(params),
    staleTime: 1000 * 60 * 5, // 5 minutes
    enabled: true,
  })
}

export function useDashboardData(filters?: Partial<FilterState>) {
  const params = filtersToParams(filters)
  
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.dashboard(params),
    queryFn: () => apiClient.analytics.getDashboard(params),
    staleTime: 1000 * 60 * 5, // 5 minutes
    enabled: true,
  })
}