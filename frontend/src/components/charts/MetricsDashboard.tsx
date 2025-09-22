import { ReactNode } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

import { 
  TrendingUp, 
  TrendingDown, 
  Minus, 
  AlertTriangle, 
  CheckCircle, 
  Clock,
  Users,
  Zap
} from 'lucide-react'

interface MetricCardProps {
  title: string
  value: string | number
  subtitle?: string
  trend?: 'up' | 'down' | 'stable'
  trendValue?: string
  icon?: ReactNode
  variant?: 'default' | 'success' | 'warning' | 'destructive'
  className?: string
}

interface MetricsDashboardProps {
  children: ReactNode
  className?: string
}

const getTrendIcon = (trend: 'up' | 'down' | 'stable') => {
  switch (trend) {
    case 'up':
      return <TrendingUp className="h-4 w-4 text-green-500" />
    case 'down':
      return <TrendingDown className="h-4 w-4 text-red-500" />
    default:
      return <Minus className="h-4 w-4 text-gray-500" />
  }
}

const getVariantStyles = (variant: MetricCardProps['variant']) => {
  switch (variant) {
    case 'success':
      return 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950'
    case 'warning':
      return 'border-yellow-200 bg-yellow-50 dark:border-yellow-800 dark:bg-yellow-950'
    case 'destructive':
      return 'border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-950'
    default:
      return ''
  }
}

export function MetricCard({
  title,
  value,
  subtitle,
  trend,
  trendValue,
  icon,
  variant = 'default',
  className = ""
}: MetricCardProps) {
  return (
    <Card className={`${getVariantStyles(variant)} ${className}`}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
        {icon && <div className="text-muted-foreground">{icon}</div>}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {subtitle && (
          <p className="text-xs text-muted-foreground mt-1">
            {subtitle}
          </p>
        )}
        {trend && trendValue && (
          <div className="flex items-center gap-1 mt-2">
            {getTrendIcon(trend)}
            <span className="text-xs text-muted-foreground">
              {trendValue} from last period
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export function MetricsDashboard({ children, className = "" }: MetricsDashboardProps) {
  return (
    <div className={`space-y-6 ${className}`}>
      {children}
    </div>
  )
}

// Pre-built metric card components for common use cases
export function IncidentVolumeCard({ 
  total, 
  trend, 
  trendValue 
}: { 
  total: number
  trend?: 'up' | 'down' | 'stable'
  trendValue?: string 
}) {
  return (
    <MetricCard
      title="Total Incidents"
      value={total.toLocaleString()}
      subtitle="All time"
      trend={trend}
      trendValue={trendValue}
      icon={<AlertTriangle className="h-4 w-4" />}
      variant={trend === 'up' ? 'warning' : trend === 'down' ? 'success' : 'default'}
    />
  )
}

export function ResolutionTimeCard({ 
  avgTime, 
  trend, 
  trendValue 
}: { 
  avgTime: number
  trend?: 'up' | 'down' | 'stable'
  trendValue?: string 
}) {
  return (
    <MetricCard
      title="Avg Resolution Time"
      value={`${avgTime.toFixed(1)}h`}
      subtitle="Across all priorities"
      trend={trend}
      trendValue={trendValue}
      icon={<Clock className="h-4 w-4" />}
      variant={trend === 'up' ? 'destructive' : trend === 'down' ? 'success' : 'default'}
    />
  )
}

export function AutomationPotentialCard({ 
  percentage, 
  processCount 
}: { 
  percentage: number
  processCount: number 
}) {
  return (
    <MetricCard
      title="Automation Potential"
      value={`${percentage.toFixed(1)}%`}
      subtitle={`${processCount} processes identified`}
      icon={<Zap className="h-4 w-4" />}
      variant={percentage > 70 ? 'success' : percentage > 40 ? 'warning' : 'default'}
    />
  )
}

export function SentimentScoreCard({ 
  score, 
  trend, 
  trendValue 
}: { 
  score: number
  trend?: 'up' | 'down' | 'stable'
  trendValue?: string 
}) {
  const getScoreVariant = (score: number) => {
    if (score > 0.1) return 'success'
    if (score < -0.1) return 'destructive'
    return 'default'
  }

  return (
    <MetricCard
      title="Sentiment Score"
      value={score.toFixed(2)}
      subtitle="Overall incident sentiment"
      trend={trend}
      trendValue={trendValue}
      icon={score > 0 ? <CheckCircle className="h-4 w-4" /> : <AlertTriangle className="h-4 w-4" />}
      variant={getScoreVariant(score)}
    />
  )
}

export function ActiveApplicationsCard({ 
  count, 
  criticalCount 
}: { 
  count: number
  criticalCount: number 
}) {
  return (
    <MetricCard
      title="Active Applications"
      value={count}
      subtitle={`${criticalCount} with critical incidents`}
      icon={<Users className="h-4 w-4" />}
      variant={criticalCount > count * 0.3 ? 'warning' : 'default'}
    />
  )
}

// Status indicator component
export function StatusIndicator({ 
  status, 
  label 
}: { 
  status: 'healthy' | 'warning' | 'critical'
  label: string 
}) {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'bg-green-500'
      case 'warning':
        return 'bg-yellow-500'
      case 'critical':
        return 'bg-red-500'
      default:
        return 'bg-gray-500'
    }
  }

  return (
    <div className="flex items-center gap-2">
      <div className={`w-2 h-2 rounded-full ${getStatusColor(status)}`} />
      <span className="text-sm text-muted-foreground">{label}</span>
    </div>
  )
}

// Quick stats grid component
export function QuickStatsGrid({ 
  stats 
}: { 
  stats: Array<{
    label: string
    value: string | number
    change?: string
    trend?: 'up' | 'down' | 'stable'
  }> 
}) {
  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
      {stats.map((stat, index) => (
        <div key={index} className="text-center">
          <div className="text-2xl font-bold text-primary">{stat.value}</div>
          <div className="text-sm text-muted-foreground">{stat.label}</div>
          {stat.change && stat.trend && (
            <div className="flex items-center justify-center gap-1 mt-1">
              {getTrendIcon(stat.trend)}
              <span className="text-xs text-muted-foreground">{stat.change}</span>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}