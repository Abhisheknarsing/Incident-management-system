import { lazy } from 'react'

export const LazyTimelineChart = lazy(() =>
  import('@/components/charts/TimelineChart').then(module => ({
    default: module.TimelineChart
  }))
)

export const LazyTrendAnalysisChart = lazy(() =>
  import('@/components/charts/TrendAnalysisChart').then(module => ({
    default: module.TrendAnalysisChart
  }))
)

export const LazyPriorityChart = lazy(() =>
  import('@/components/charts/PriorityChart').then(module => ({
    default: module.PriorityChart
  }))
)

export const LazyApplicationChart = lazy(() =>
  import('@/components/charts/ApplicationChart').then(module => ({
    default: module.ApplicationChart
  }))
)

export const LazyResolutionChart = lazy(() =>
  import('@/components/charts/ResolutionChart').then(module => ({
    default: module.ResolutionChart
  }))
)

export const LazySentimentChart = lazy(() =>
  import('@/components/charts/SentimentChart').then(module => ({
    default: module.SentimentChart
  }))
)

export const LazyAutomationChart = lazy(() =>
  import('@/components/charts/AutomationChart').then(module => ({
    default: module.AutomationChart
  }))
)