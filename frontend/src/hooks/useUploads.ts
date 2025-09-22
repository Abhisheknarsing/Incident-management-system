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
  
  return useQuery<Upload[]>({
    queryKey: QUERY_KEYS.uploads,
    queryFn: async () => {
      try {
        const data = await apiClient.uploads.list();
        // Ensure we always return an array
        return Array.isArray(data) ? data : [];
      } catch (error) {
        // Handle error and return empty array as fallback
        handleError(error, 'Failed to load uploads');
        return [];
      }
    },
    staleTime: 1000 * 60 * 2, // 2 minutes
    refetchInterval: 1000 * 30, // Refetch every 30 seconds
    // Ensure we always have an array, even if the API returns undefined
    initialData: [],
  })
}

export function useUploadFile(onProgress?: (progress: number) => void) {
  const queryClient = useQueryClient()
  const { handleError } = useErrorHandler()
  
  return useMutation({
    mutationFn: (file: File) => apiClient.uploads.upload(file, onProgress),
    onSuccess: (newUpload: Upload) => {
      // Update the uploads list with the new upload
      queryClient.setQueryData<Upload[]>(QUERY_KEYS.uploads, (old: Upload[] | undefined) => {
        // Handle case where old data might be undefined or null
        if (!old) {
          return [newUpload]
        }
        // Ensure old is actually an array before spreading
        if (!Array.isArray(old)) {
          return [newUpload]
        }
        return [newUpload, ...old]
      })
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