import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api'

export interface FilterOptions {
  priorities: string[]
  applications: string[]
  statuses: string[]
}

export function useFilterOptions() {
  return useQuery({
    queryKey: ['filter-options'],
    queryFn: async (): Promise<FilterOptions> => {
      // For now, return default options
      // In a real implementation, this would fetch from the API
      return {
        priorities: ['P1', 'P2', 'P3', 'P4'],
        applications: [], // Will be populated from actual data
        statuses: ['Open', 'In Progress', 'Resolved', 'Closed']
      }
    },
    staleTime: 1000 * 60 * 10, // 10 minutes
    enabled: true,
  })
}

// Hook to get unique applications from analytics data
export function useAvailableApplications() {
  return useQuery({
    queryKey: ['available-applications'],
    queryFn: async (): Promise<string[]> => {
      try {
        const applications = await apiClient.analytics.getApplications()
        return applications.map(app => app.application_name).sort()
      } catch (error) {
        console.warn('Failed to fetch applications for filters:', error)
        return []
      }
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
    enabled: true,
  })
}