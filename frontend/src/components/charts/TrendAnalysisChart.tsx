import { useState, useMemo } from 'react'
import {
  ComposedChart,
  Line,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ReferenceLine
} from 'recharts'
import { format, parseISO, subDays, isAfter } from 'date-fns'
import { TimelineData } from '@/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface TrendAnalysisChartProps {
  data: TimelineData[]
  height?: number
  className?: string
}

interface TooltipProps {
  active?: boolean
  payload?: any[]
  label?: string
}

type TimeRange = '7d' | '30d' | '90d' | 'all'

const CustomTooltip = ({ active, payload, label }: TooltipProps) => {
  if (active && payload && payload.length) {
    const date = label ? format(parseISO(label), 'MMM dd, yyyy') : ''
    
    return (
      <div className="bg-background border border-border rounded-lg shadow-lg p-3 min-w-[200px]">
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
            <span className="font-medium">{entry.value}</span>
          </div>
        ))}
      </div>
    )
  }
  return null
}

export function TrendAnalysisChart({ 
  data, 
  height = 500, 
  className = ""
}: TrendAnalysisChartProps) {
  const [timeRange, setTimeRange] = useState<TimeRange>('30d')
  const [showMovingAverage, setShowMovingAverage] = useState(true)

  const filteredData = useMemo(() => {
    if (timeRange === 'all') return data

    const days = timeRange === '7d' ? 7 : timeRange === '30d' ? 30 : 90
    const cutoffDate = subDays(new Date(), days)
    
    return data.filter(item => isAfter(parseISO(item.date), cutoffDate))
  }, [data, timeRange])

  const chartData = useMemo(() => {
    const processedData = filteredData.map((item, index) => {
      // Calculate 7-day moving average
      let movingAverage = 0
      if (showMovingAverage && index >= 6) {
        const last7Days = filteredData.slice(index - 6, index + 1)
        movingAverage = last7Days.reduce((sum, day) => sum + day.count, 0) / 7
      }

      return {
        ...item,
        date: format(parseISO(item.date), 'MMM dd'),
        fullDate: item.date,
        total: item.count,
        movingAverage: movingAverage > 0 ? Math.round(movingAverage) : null,
        criticalCount: item.p1_count + item.p2_count,
        normalCount: item.p3_count + item.p4_count
      }
    })

    return processedData
  }, [filteredData, showMovingAverage])

  const stats = useMemo(() => {
    if (chartData.length === 0) return null

    const totalIncidents = chartData.reduce((sum, item) => sum + item.total, 0)
    const avgDaily = Math.round(totalIncidents / chartData.length)
    const maxDaily = Math.max(...chartData.map(item => item.total))
    const criticalPercent = Math.round(
      (chartData.reduce((sum, item) => sum + item.criticalCount, 0) / totalIncidents) * 100
    )

    return { totalIncidents, avgDaily, maxDaily, criticalPercent }
  }, [chartData])

  const timeRangeButtons: { key: TimeRange; label: string }[] = [
    { key: '7d', label: '7 Days' },
    { key: '30d', label: '30 Days' },
    { key: '90d', label: '90 Days' },
    { key: 'all', label: 'All Time' }
  ]

  return (
    <Card className={className}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Incident Trend Analysis</CardTitle>
          <div className="flex items-center gap-2">
            <Button
              variant={showMovingAverage ? "default" : "outline"}
              size="sm"
              onClick={() => setShowMovingAverage(!showMovingAverage)}
            >
              7-Day Avg
            </Button>
            <div className="flex border rounded-md">
              {timeRangeButtons.map((button) => (
                <Button
                  key={button.key}
                  variant={timeRange === button.key ? "default" : "ghost"}
                  size="sm"
                  className="rounded-none first:rounded-l-md last:rounded-r-md"
                  onClick={() => setTimeRange(button.key)}
                >
                  {button.label}
                </Button>
              ))}
            </div>
          </div>
        </div>
        
        {stats && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-primary">{stats.totalIncidents}</div>
              <div className="text-sm text-muted-foreground">Total Incidents</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">{stats.avgDaily}</div>
              <div className="text-sm text-muted-foreground">Daily Average</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-orange-600">{stats.maxDaily}</div>
              <div className="text-sm text-muted-foreground">Peak Day</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-red-600">{stats.criticalPercent}%</div>
              <div className="text-sm text-muted-foreground">Critical (P1/P2)</div>
            </div>
          </div>
        )}
      </CardHeader>
      
      <CardContent>
        <ResponsiveContainer width="100%" height={height}>
          <ComposedChart
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
            />
            <Tooltip content={<CustomTooltip />} />
            <Legend 
              wrapperStyle={{ paddingTop: '20px' }}
            />
            
            {/* Stacked bars for critical vs normal incidents */}
            <Bar
              dataKey="criticalCount"
              stackId="priority"
              fill="hsl(var(--destructive))"
              name="Critical (P1/P2)"
              radius={[0, 0, 0, 0]}
            />
            <Bar
              dataKey="normalCount"
              stackId="priority"
              fill="hsl(var(--primary))"
              name="Normal (P3/P4)"
              radius={[4, 4, 0, 0]}
            />
            
            {/* Moving average line */}
            {showMovingAverage && (
              <Line
                type="monotone"
                dataKey="movingAverage"
                stroke="hsl(var(--warning))"
                strokeWidth={3}
                strokeDasharray="5 5"
                dot={false}
                name="7-Day Moving Average"
                connectNulls={false}
              />
            )}

            {/* Reference line for average */}
            {stats && (
              <ReferenceLine 
                y={stats.avgDaily} 
                stroke="hsl(var(--muted-foreground))" 
                strokeDasharray="3 3"
                label={{ 
                  value: `Avg: ${stats.avgDaily}`, 
                  position: "top"
                }}
              />
            )}
          </ComposedChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  )
}