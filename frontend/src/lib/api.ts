import axios, { AxiosError, AxiosResponse } from 'axios'
import { Upload, DashboardData, TimelineData, PriorityAnalysis, ApplicationAnalysis, SentimentAnalysis, ResolutionMetrics, AutomationAnalysis } from '@/types'
import { APIError } from '@/lib/errors'

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api'

export interface ValidationError {
  field: string
  value: string
  message: string
  row?: number
}

export interface ProcessingStatus {
  upload_id: string
  status: 'pending' | 'processing' | 'completed' | 'failed'
  progress: number
  message?: string
  errors?: ValidationError[]
}

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000, // 30 seconds timeout
})

// Request interceptor for adding auth tokens if needed
api.interceptors.request.use(
  (config) => {
    // Add auth token here if needed
    // Add request ID for tracking
    config.headers['X-Request-ID'] = crypto.randomUUID()
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for handling errors
api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    // Handle common errors here
    const responseData = error.response?.data as any
    
    const apiError: APIError = {
      code: responseData?.code || 'UNKNOWN_ERROR',
      message: responseData?.message || error.message || 'An unexpected error occurred',
      details: responseData?.details,
      validations: responseData?.validations,
      timestamp: new Date().toISOString(),
      request_id: error.config?.headers?.['X-Request-ID'] as string || 'unknown',
      path: error.config?.url,
      method: error.config?.method?.toUpperCase(),
      user_message: responseData?.user_message,
      suggestions: responseData?.suggestions,
      documentation: responseData?.documentation,
    }

    console.error('API Error:', apiError)
    return Promise.reject(apiError)
  }
)

// API Client functions
export const apiClient = {
  // Upload endpoints
  uploads: {
    list: (): Promise<Upload[]> => 
      api.get('/uploads').then(res => res.data),
    
    upload: (file: File, onProgress?: (progress: number) => void): Promise<Upload> => {
      const formData = new FormData()
      formData.append('file', file)
      return api.post('/uploads', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: (progressEvent) => {
          if (onProgress && progressEvent.total) {
            const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
            onProgress(progress)
          }
        },
      }).then(res => res.data)
    },
    
    getStatus: (uploadId: string): Promise<ProcessingStatus> =>
      api.get(`/uploads/${uploadId}/status`).then(res => res.data),
    
    startAnalysis: (uploadId: string): Promise<{ message: string }> =>
      api.post(`/uploads/${uploadId}/analyze`).then(res => res.data),
  },

  // Analytics endpoints
  analytics: {
    getTimeline: (filters?: Record<string, any>): Promise<TimelineData[]> =>
      api.get('/analytics/timeline', { params: filters }).then(res => res.data),
    
    getPriorities: (filters?: Record<string, any>): Promise<PriorityAnalysis[]> =>
      api.get('/analytics/priorities', { params: filters }).then(res => res.data),
    
    getApplications: (filters?: Record<string, any>): Promise<ApplicationAnalysis[]> =>
      api.get('/analytics/applications', { params: filters }).then(res => res.data),
    
    getSentiment: (filters?: Record<string, any>): Promise<SentimentAnalysis> =>
      api.get('/analytics/sentiment', { params: filters }).then(res => res.data),
    
    getResolution: (filters?: Record<string, any>): Promise<ResolutionMetrics> =>
      api.get('/analytics/resolution', { params: filters }).then(res => res.data),
    
    getAutomation: (filters?: Record<string, any>): Promise<AutomationAnalysis[]> =>
      api.get('/analytics/automation', { params: filters }).then(res => res.data),
    
    getDashboard: (filters?: Record<string, any>): Promise<DashboardData> =>
      api.get('/analytics/dashboard', { params: filters }).then(res => res.data),
  },

  // Export endpoints
  export: {
    request: (options: {
      format: 'csv' | 'pdf'
      dataType: string
      filters?: Record<string, any>
    }): Promise<{ download_url: string; job_id: string }> =>
      api.post('/export', options).then(res => res.data),
    
    getStatus: (jobId: string): Promise<{
      id: string
      status: 'pending' | 'processing' | 'completed' | 'failed'
      progress: number
      download_url?: string
      error?: string
    }> =>
      api.get(`/export/${jobId}/status`).then(res => res.data),
    
    download: (downloadUrl: string): Promise<Blob> =>
      api.get(downloadUrl, { responseType: 'blob' }).then(res => res.data),
  },
}