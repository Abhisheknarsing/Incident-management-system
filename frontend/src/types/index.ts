export interface Upload {
  id: string
  filename: string
  original_filename: string
  status: 'uploaded' | 'processing' | 'completed' | 'failed'
  record_count: number
  processed_count: number
  error_count: number
  errors?: string[]
  created_at: string
  processed_at?: string
}

export interface Incident {
  id: string
  upload_id: string
  incident_id: string
  report_date: string
  resolve_date?: string
  last_resolve_date?: string
  brief_description: string
  description: string
  application_name: string
  resolution_group: string
  resolved_person: string
  priority: string
  
  // Additional fields
  category?: string
  subcategory?: string
  impact?: string
  urgency?: string
  status?: string
  customer_affected?: string
  business_service?: string
  root_cause?: string
  resolution_notes?: string
  
  // Derived fields
  sentiment_score?: number
  sentiment_label?: 'positive' | 'negative' | 'neutral'
  resolution_time_hours?: number
  automation_score?: number
  automation_feasible?: boolean
  it_process_group?: string
  
  created_at: string
  updated_at: string
}

export interface TimelineData {
  date: string
  count: number
  p1_count: number
  p2_count: number
  p3_count: number
  p4_count: number
}

export interface PriorityAnalysis {
  priority: string
  count: number
  percentage: number
}

export interface ApplicationAnalysis {
  application_name: string
  incident_count: number
  avg_resolution_time: number
  trend: string
}

export interface SentimentAnalysis {
  positive: number
  negative: number
  neutral: number
}

export interface ResolutionMetrics {
  avg_resolution_time: number
  median_resolution_time: number
  resolution_trends: TimelineData[]
}

export interface AutomationAnalysis {
  it_process_group: string
  automation_score: number
  incident_count: number
  automation_feasible: boolean
}

export interface DashboardData {
  timeline: TimelineData[]
  priorities: PriorityAnalysis[]
  applications: ApplicationAnalysis[]
  sentiment: SentimentAnalysis
  resolutionMetrics: ResolutionMetrics
  automationOpportunities: AutomationAnalysis[]
}

export interface FilterState {
  dateRange: { start: string; end: string }
  priorities: string[]
  applications: string[]
  statuses: string[]
}

export interface ExportOptions {
  format: 'csv' | 'pdf'
  dataType: string
  includeFilters: boolean
}