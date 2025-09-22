import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient, ProcessingStatus } from '@/lib/api'
import { Upload } from '@/types'

export const QUERY_KEYS = {
  uploads: ['uploads'] as const,
  uploadStatus: (id: string) => ['uploads', id, 'status'] as const,
}

export function useUploads() {
  return useQuery({
    queryKey: QUERY_KEYS.uploads,
    queryFn: apiClient.uploads.list,
    staleTime: 1000 * 60 * 2, // 2 minutes
    refetchInterval: 1000 * 30, // Refetch every 30 seconds
  })
}

export function useUploadFile() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: apiClient.uploads.upload,
    onSuccess: (newUpload: Upload) => {
      // Update the uploads list with the new upload
      queryClient.setQueryData(QUERY_KEYS.uploads, (old: Upload[] = []) => [
        newUpload,
        ...old,
      ])
    },
    onError: (error) => {
      console.error('Upload failed:', error)
    },
  })
}

export function useUploadStatus(uploadId: string, enabled = true) {
  return useQuery({
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
  })
}

export function useStartAnalysis() {
  const queryClient = useQueryClient()
  
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
      console.error('Analysis start failed:', error)
    },
  })
}