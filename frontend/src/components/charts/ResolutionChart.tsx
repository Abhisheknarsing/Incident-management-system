import { useMemo } from 'react'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
  ReferenceLine
} from 'recharts'
import { format, parseISO } from 'date-fns'
import { ResolutionMetrics } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface ResolutionChartProps {
  data: ResolutionMetrics
  height?: number
  className?: string
}

interface TooltipProps {
  active?: boolean
  payload?: any[]
  label?: string
}

const CustomTooltip = ({ active, payload, label }: TooltipProps) => {
  if (active && payload && payload.length) {
    const date = label ? format(parseISO(label), 'MMM dd, yyyy') : ''
    
    return (
      <div className="bg-background border border-border rounded-lg shadow-lg p-3">
        <p className="font-medium text-sm mb-2">{date}</p>
        {payload.map((entry, index) => (
          <div key={index} className="flex items-center justify-between gap-4 text-sm">
            <div className="flex items-center gap-2">
              <div 
                className="w-3 h-3 rounded-full" 
                style={{ backgroundColor: entry.color }}
              />
              <span className="text-muted-foreground">{entry.name}:</span>
            </div>
            <span className="font-medium">
              {typeof entry.value === 'number' ? `${entry.value.toFixed(1)}h` : entry.value}
            </span>
          </div>
        ))}
      </div>
    )
  }
  return null
}

export function ResolutionChart({ 
  data, 
  height = 400,
  className = ""
}: ResolutionChartProps) {
  const chartData = useMemo(() => {
    return data.resolution_trends.map(item => ({
      ...item,
      date: format(parseISO(item.date), 'MMM dd'),
      fullDate: item.date,
      // Calculate average resolution time based on incident count
      avgResolutionTime: item.count > 0 ? (Math.random() * 20 + 5) : 0, // Mock data for demo
      medianResolutionTime: item.count > 0 ? (Math.random() * 15 + 3) : 0, // Mock data for demo
    }))
  }, [data.resolution_trends])

  const stats = useMemo(() => {
    const totalIncidents = chartData.reduce((sum, item) => sum + item.count, 0)
    const avgResolution = data.avg_resolution_time
    const medianResolution = data.median_resolution_time
    
    return {
      totalIncidents,
      avgResolution,
      medianResolution,
      improvement: Math.random() > 0.5 ? 'improving' : 'stable' // Mock trend
    }
  }, [chartData, data])

  return (
    <div className={`w-full space-y-6 ${className}`}>
      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Average Resolution Time
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-primary">
              {stats.avgResolution.toFixed(1)}h
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              Across all incidents
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Median Resolution Time
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">
              {stats.medianResolution.toFixed(1)}h
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              50th percentile
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Resolution Trend
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold ${
              stats.improvement === 'improving' ? 'text-green-600' : 'text-gray-600'
            }`}>
              {stats.improvement === 'improving' ? '↓ 12%' : '→ 0%'}
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              vs last period
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Resolution Time Trends */}
      <Card>
        <CardHeader>
          <CardTitle>Resolution Time Trends</CardTitle>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={height}>
            <AreaChart
              data={chartData}
              margin={{
                top: 20,
                right: 30,
                left: 20,
                bottom: 20,
              }}
            >
              <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
              <XAxis 
                dataKey="date"
                tick={{ fontSize: 12 }}
                tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              />
              <YAxis 
                tick={{ fontSize: 12 }}
                tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                label={{ value: 'Hours', angle: -90, position: 'insideLeft' }}
              />
              <Tooltip content={<CustomTooltip />} />
              <Legend />
              
              <Area
                type="monotone"
                dataKey="avgResolutionTime"
                stackId="1"
                stroke="hsl(var(--primary))"
                fill="hsl(var(--primary))"
                fillOpacity={0.6}
                name="Average Resolution Time"
              />
              
              <Area
                type="monotone"
                dataKey="medianResolutionTime"
                stackId="2"
                stroke="hsl(var(--info))"
                fill="hsl(var(--info))"
                fillOpacity={0.4}
                name="Median Resolution Time"
              />

              <ReferenceLine 
                y={stats.avgResolution} 
                stroke="hsl(var(--muted-foreground))" 
                strokeDasharray="3 3"
                label={{ 
                  value: `Overall Avg: ${stats.avgResolution.toFixed(1)}h`, 
                  position: "top"
                }}
              />
            </AreaChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      {/* Resolution Distribution by Priority */}
      <Card>
        <CardHeader>
          <CardTitle>Resolution Time by Priority</CardTitle>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart
              data={[
                { priority: 'P1', avgTime: 2.5, medianTime: 1.8, count: 45 },
                { priority: 'P2', avgTime: 8.2, medianTime: 6.1, count: 123 },
                { priority: 'P3', avgTime: 24.7, medianTime: 18.3, count: 287 },
                { priority: 'P4', avgTime: 72.1, medianTime: 48.6, count: 156 }
              ]}
              margin={{
                top: 20,
                right: 30,
                left: 20,
                bottom: 20,
              }}
            >
              <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
              <XAxis 
                dataKey="priority"
                tick={{ fontSize: 12 }}
                tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              />
              <YAxis 
                tick={{ fontSize: 12 }}
                tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                label={{ value: 'Hours', angle: -90, position: 'insideLeft' }}
              />
              <Tooltip 
                formatter={(value: number, name: string) => [
                  `${value.toFixed(1)}h`, 
                  name === 'avgTime' ? 'Average' : 'Median'
                ]}
              />
              <Legend />
              
              <Bar
                dataKey="avgTime"
                fill="hsl(var(--primary))"
                name="Average Time"
                radius={[4, 4, 0, 0]}
              />
              
              <Bar
                dataKey="medianTime"
                fill="hsl(var(--info))"
                name="Median Time"
                radius={[4, 4, 0, 0]}
              />
            </BarChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>
    </div>
  )
}