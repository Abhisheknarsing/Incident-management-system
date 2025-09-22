import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient, ProcessingStatus } from '@/lib/api'
import { Upload } from '@/types'
import { useErrorHandler } from '@/hooks/useErrorHandler'

export const QUERY_KEYS = {
  uploads: ['uploads'] as const,
  uploadStatus: (id: string) => ['uploads', id, 'status'] as const,
}

export function useUploads() {
  const { handleError } = useErrorHandler()
  
  return useQuery({
    queryKey: QUERY_KEYS.uploads,
    queryFn: apiClient.uploads.list,
    staleTime: 1000 * 60 * 2, // 2 minutes
    refetchInterval: 1000 * 30, // Refetch every 30 seconds
    onError: (error) => {
      handleError(error, 'Failed to load uploads')
    },
  })
}

export function useUploadFile(onProgress?: (progress: number) => void) {
  const queryClient = useQueryClient()
  const { handleError } = useErrorHandler()
  
  return useMutation({
    mutationFn: (file: File) => apiClient.uploads.upload(file, onProgress),
    onSuccess: (newUpload: Upload) => {
      // Update the uploads list with the new upload
      queryClient.setQueryData(QUERY_KEYS.uploads, (old: Upload[] = []) => [
        newUpload,
        ...old,
      ])
    },
    onError: (error) => {
      handleError(error, 'File upload failed')
    },
  })
}

export function useUploadStatus(uploadId: string, enabled = true) {
  const { handleError } = useErrorHandler()
  
  return useQuery<ProcessingStatus>({
    queryKey: QUERY_KEYS.uploadStatus(uploadId),
    queryFn: () => apiClient.uploads.getStatus(uploadId),
    enabled: enabled && !!uploadId,
    refetchInterval: (data: ProcessingStatus | undefined) => {
      // Stop polling if processing is complete or failed
      if (data?.status === 'completed' || data?.status === 'failed') {
        return false
      }
      // Poll every 2 seconds while processing
      return 2000
    },
    staleTime: 0, // Always fetch fresh status
    onError: (error) => {
      handleError(error, 'Failed to get upload status')
    },
  })
}

export function useStartAnalysis() {
  const queryClient = useQueryClient()
  const { handleError } = useErrorHandler()
  
  return useMutation({
    mutationFn: apiClient.uploads.startAnalysis,
    onSuccess: (_, uploadId) => {
      // Invalidate upload status to start polling
      queryClient.invalidateQueries({
        queryKey: QUERY_KEYS.uploadStatus(uploadId),
      })
      // Invalidate uploads list to update status
      queryClient.invalidateQueries({
        queryKey: QUERY_KEYS.uploads,
      })
    },
    onError: (error) => {
      handleError(error, 'Failed to start analysis')
    },
  })
}