import { useState, useMemo } from 'react'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ComposedChart,
  Line,
  ReferenceLine
} from 'recharts'
import { ApplicationAnalysis } from '@/types'
import { Button } from '@/components/ui/button'
import { TrendingUp, TrendingDown, Minus } from 'lucide-react'

interface ApplicationChartProps {
  data: ApplicationAnalysis[]
  height?: number
  showResolutionTime?: boolean
  maxApplications?: number
  className?: string
}

interface TooltipProps {
  active?: boolean
  payload?: any[]
  label?: string
}

type SortBy = 'incidents' | 'resolution_time' | 'name'

const CustomTooltip = ({ active, payload, label }: TooltipProps) => {
  if (active && payload && payload.length) {
    const data = payload[0].payload
    
    return (
      <div className="bg-background border border-border rounded-lg shadow-lg p-3 min-w-[200px]">
        <p className="font-medium mb-2">{label}</p>
        <div className="space-y-1 text-sm">
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">Incidents:</span>
            <span className="font-medium">{data.incident_count}</span>
          </div>
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">Avg Resolution:</span>
            <span className="font-medium">{data.avg_resolution_time.toFixed(1)}h</span>
          </div>
          <div className="flex items-center justify-between gap-4">
            <span className="text-muted-foreground">Trend:</span>
            <div className="flex items-center gap-1">
              {data.trend === 'up' && <TrendingUp className="h-3 w-3 text-red-500" />}
              {data.trend === 'down' && <TrendingDown className="h-3 w-3 text-green-500" />}
              {data.trend === 'stable' && <Minus className="h-3 w-3 text-gray-500" />}
              <span className="font-medium capitalize">{data.trend}</span>
            </div>
          </div>
        </div>
      </div>
    )
  }
  return null
}



export function ApplicationChart({ 
  data, 
  height = 400,
  showResolutionTime = true,
  maxApplications = 10,
  className = ""
}: ApplicationChartProps) {
  const [sortBy, setSortBy] = useState<SortBy>('incidents')
  const [showAll, setShowAll] = useState(false)

  const chartData = useMemo(() => {
    let sortedData = [...data]
    
    // Sort data based on selected criteria
    switch (sortBy) {
      case 'incidents':
        sortedData.sort((a, b) => b.incident_count - a.incident_count)
        break
      case 'resolution_time':
        sortedData.sort((a, b) => b.avg_resolution_time - a.avg_resolution_time)
        break
      case 'name':
        sortedData.sort((a, b) => a.application_name.localeCompare(b.application_name))
        break
    }

    // Limit number of applications shown
    const displayData = showAll ? sortedData : sortedData.slice(0, maxApplications)
    
    return displayData.map(item => ({
      ...item,
      name: item.application_name.length > 15 
        ? `${item.application_name.substring(0, 15)}...` 
        : item.application_name,
      fullName: item.application_name,
      resolutionHours: Math.round(item.avg_resolution_time * 10) / 10
    }))
  }, [data, sortBy, showAll, maxApplications])

  const avgResolutionTime = useMemo(() => {
    if (data.length === 0) return 0
    const total = data.reduce((sum, item) => sum + item.avg_resolution_time, 0)
    return total / data.length
  }, [data])

  const sortButtons: { key: SortBy; label: string }[] = [
    { key: 'incidents', label: 'By Incidents' },
    { key: 'resolution_time', label: 'By Resolution Time' },
    { key: 'name', label: 'By Name' }
  ]

  return (
    <div className={`w-full space-y-4 ${className}`}>
      {/* Controls */}
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Sort:</span>
          <div className="flex border rounded-md">
            {sortButtons.map((button) => (
              <Button
                key={button.key}
                variant={sortBy === button.key ? "default" : "ghost"}
                size="sm"
                className="rounded-none first:rounded-l-md last:rounded-r-md text-xs"
                onClick={() => setSortBy(button.key)}
              >
                {button.label}
              </Button>
            ))}
          </div>
        </div>
        
        {data.length > maxApplications && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowAll(!showAll)}
          >
            {showAll ? `Show Top ${maxApplications}` : `Show All (${data.length})`}
          </Button>
        )}
      </div>

      {/* Chart */}
      <ResponsiveContainer width="100%" height={height}>
        {showResolutionTime ? (
          <ComposedChart
            data={chartData}
            margin={{
              top: 20,
              right: 30,
              left: 20,
              bottom: 60,
            }}
          >
            <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
            <XAxis 
              dataKey="name"
              tick={{ fontSize: 11 }}
              tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              angle={-45}
              textAnchor="end"
              height={80}
            />
            <YAxis 
              yAxisId="incidents"
              orientation="left"
              tick={{ fontSize: 12 }}
              tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
            />
            <YAxis 
              yAxisId="time"
              orientation="right"
              tick={{ fontSize: 12 }}
              tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
            />
            <Tooltip content={<CustomTooltip />} />
            <Legend />
            
            <Bar
              yAxisId="incidents"
              dataKey="incident_count"
              fill="hsl(var(--primary))"
              name="Incident Count"
              radius={[4, 4, 0, 0]}
            />
            
            <Line
              yAxisId="time"
              type="monotone"
              dataKey="resolutionHours"
              stroke="hsl(var(--warning))"
              strokeWidth={3}
              dot={{ fill: 'hsl(var(--warning))', strokeWidth: 2, r: 4 }}
              name="Avg Resolution Time (hours)"
            />

            <ReferenceLine 
              yAxisId="time"
              y={avgResolutionTime} 
              stroke="hsl(var(--muted-foreground))" 
              strokeDasharray="3 3"
              label={{ 
                value: `Avg: ${avgResolutionTime.toFixed(1)}h`, 
                position: "top"
              }}
            />
          </ComposedChart>
        ) : (
          <BarChart
            data={chartData}
            margin={{
              top: 20,
              right: 30,
              left: 20,
              bottom: 60,
            }}
          >
            <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
            <XAxis 
              dataKey="name"
              tick={{ fontSize: 11 }}
              tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              angle={-45}
              textAnchor="end"
              height={80}
            />
            <YAxis 
              tick={{ fontSize: 12 }}
              tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
            />
            <Tooltip content={<CustomTooltip />} />
            <Legend />
            
            <Bar
              dataKey="incident_count"
              fill="hsl(var(--primary))"
              name="Incident Count"
              radius={[4, 4, 0, 0]}
            />
          </BarChart>
        )}
      </ResponsiveContainer>

      {/* Application trends summary */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        <div className="text-center">
          <div className="text-lg font-bold text-primary">{data.length}</div>
          <div className="text-muted-foreground">Applications</div>
        </div>
        <div className="text-center">
          <div className="text-lg font-bold text-green-600">
            {data.filter(app => app.trend === 'down').length}
          </div>
          <div className="text-muted-foreground">Improving</div>
        </div>
        <div className="text-center">
          <div className="text-lg font-bold text-red-600">
            {data.filter(app => app.trend === 'up').length}
          </div>
          <div className="text-muted-foreground">Worsening</div>
        </div>
        <div className="text-center">
          <div className="text-lg font-bold text-gray-600">
            {data.filter(app => app.trend === 'stable').length}
          </div>
          <div className="text-muted-foreground">Stable</div>
        </div>
      </div>
    </div>
  )
}