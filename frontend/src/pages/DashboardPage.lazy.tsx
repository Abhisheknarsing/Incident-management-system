import { lazy, Suspense } from 'react'
import { LoadingSpinner } from '@/components/ui/loading-spinner'

// Lazy load heavy chart components
const LazyTimelineChart = lazy(() => 
  import('@/components/charts/TimelineChart').then(module => ({ default: module.TimelineChart }))
)

const LazyTrendAnalysisChart = lazy(() => 
  import('@/components/charts/TrendAnalysisChart').then(module => ({ default: module.TrendAnalysisChart }))
)

const LazyPriorityChart = lazy(() => 
  import('@/components/charts/PriorityChart').then(module => ({ default: module.PriorityChart }))
)

const LazyApplicationChart = lazy(() => 
  import('@/components/charts/ApplicationChart').then(module => ({ default: module.ApplicationChart }))
)

const LazyResolutionChart = lazy(() => 
  import('@/components/charts/ResolutionChart').then(module => ({ default: module.ResolutionChart }))
)

const LazySentimentChart = lazy(() => 
  import('@/components/charts/SentimentChart').then(module => ({ default: module.SentimentChart }))
)

const LazyAutomationChart = lazy(() => 
  import('@/components/charts/AutomationChart').then(module => ({ default: module.AutomationChart }))
)

interface LazyDashboardPageProps {
  className?: string
}

export function LazyDashboardPage({ className }: LazyDashboardPageProps) {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-screen">
        <LoadingSpinner size="lg" />
      </div>
    }>
      <div className={className}>
        <h1>Lazy Loaded Dashboard</h1>
        <p>Dashboard components will load as needed</p>
      </div>
    </Suspense>
  )
}