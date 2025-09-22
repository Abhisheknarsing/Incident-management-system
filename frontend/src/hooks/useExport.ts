import { useMutation } from '@tanstack/react-query'
import { apiClient } from '@/lib/api'
import { useNotifications } from './useAppState'

interface ExportOptions {
  format: 'csv' | 'pdf'
  dataType: string
  filters?: Record<string, any>
}

export function useExport() {
  const { addNotification } = useNotifications()
  
  return useMutation({
    mutationFn: (options: ExportOptions) => apiClient.export.request(options),
    onSuccess: (data) => {
      // Trigger download
      const link = document.createElement('a')
      link.href = data.download_url
      link.download = ''
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      
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