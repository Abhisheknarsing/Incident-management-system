import { useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api'
import { useNotifications } from './useAppState'


interface ExportOptions {
  format: 'csv' | 'pdf'
  dataType: string
  filters?: Record<string, any>
}

interface ExportState {
  isExporting: boolean
  progress: number
  jobId: string | null
  error: string | null
}

export function useExport() {
  const { addNotification } = useNotifications()
  const [exportState, setExportState] = useState<ExportState>({
    isExporting: false,
    progress: 0,
    jobId: null,
    error: null
  })

  // Poll export status
  useQuery({
    queryKey: ['export-status', exportState.jobId],
    queryFn: () => exportState.jobId ? apiClient.export.getStatus(exportState.jobId) : null,
    enabled: !!exportState.jobId && exportState.isExporting,
    refetchInterval: (data) => {
      if (data?.status === 'completed' || data?.status === 'failed') {
        return false
      }
      return 1000 // Poll every second
    },
    onSuccess: (data) => {
      if (!data) return

      setExportState(prev => ({
        ...prev,
        progress: data.progress
      }))

      if (data.status === 'completed' && data.download_url) {
        handleDownload(data.download_url)
        setExportState(prev => ({
          ...prev,
          isExporting: false,
          jobId: null
        }))
        addNotification({
          type: 'success',
          title: 'Export Complete',
          message: 'Your data has been exported successfully.',
        })
      } else if (data.status === 'failed') {
        setExportState(prev => ({
          ...prev,
          isExporting: false,
          error: data.error || 'Export failed',
          jobId: null
        }))
        addNotification({
          type: 'error',
          title: 'Export Failed',
          message: data.error || 'Failed to export data. Please try again.',
        })
      }
    }
  })

  const exportMutation = useMutation({
    mutationFn: (options: ExportOptions) => apiClient.export.request(options),
    onSuccess: (data) => {
      setExportState({
        isExporting: true,
        progress: 0,
        jobId: data.job_id,
        error: null
      })
      addNotification({
        type: 'info',
        title: 'Export Started',
        message: 'Your export is being processed. You will be notified when it\'s ready.',
      })
    },
    onError: (error: any) => {
      setExportState(prev => ({
        ...prev,
        isExporting: false,
        error: error.message || 'Failed to start export'
      }))
      addNotification({
        type: 'error',
        title: 'Export Failed',
        message: error.message || 'Failed to start export. Please try again.',
      })
    },
  })

  const handleDownload = async (downloadUrl: string) => {
    try {
      const blob = await apiClient.export.download(downloadUrl)
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = '' // Let browser determine filename from headers
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
    } catch (error) {
      addNotification({
        type: 'error',
        title: 'Download Failed',
        message: 'Failed to download the exported file. Please try again.',
      })
    }
  }

  const cancelExport = () => {
    setExportState({
      isExporting: false,
      progress: 0,
      jobId: null,
      error: null
    })
  }

  return {
    exportData: exportMutation.mutate,
    isExporting: exportState.isExporting,
    progress: exportState.progress,
    error: exportState.error,
    cancelExport,
    isLoading: exportMutation.isPending
  }
}

// Hook for quick export without progress tracking (for simple cases)
export function useQuickExport() {
  const { addNotification } = useNotifications()
  
  return useMutation({
    mutationFn: async (options: ExportOptions) => {
      const response = await apiClient.export.request(options)
      // For quick export, assume immediate download URL
      if (response.download_url) {
        const blob = await apiClient.export.download(response.download_url)
        return blob
      }
      throw new Error('No download URL provided')
    },
    onSuccess: (blob) => {
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = ''
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
      
      addNotification({
        type: 'success',
        title: 'Export Complete',
        message: 'Your data has been exported successfully.',
      })
    },
    onError: (error: any) => {
      addNotification({
        type: 'error',
        title: 'Export Failed',
        message: error.message || 'Failed to export data. Please try again.',
      })
    },
  })
}