import { useMemo } from 'react'
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip,
  Legend,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid
} from 'recharts'
import { PriorityAnalysis } from '@/types'

interface PriorityChartProps {
  data: PriorityAnalysis[]
  chartType?: 'pie' | 'donut' | 'bar'
  height?: number
  showPercentages?: boolean
  className?: string
}

interface TooltipProps {
  active?: boolean
  payload?: any[]
  label?: string
}

const PRIORITY_COLORS = {
  P1: 'hsl(var(--destructive))',
  P2: 'hsl(var(--warning))', 
  P3: 'hsl(var(--info))',
  P4: 'hsl(var(--success))',
  // Fallback colors for other priorities
  default: 'hsl(var(--muted-foreground))'
}

const CustomTooltip = ({ active, payload }: TooltipProps) => {
  if (active && payload && payload.length) {
    const data = payload[0].payload
    
    return (
      <div className="bg-background border border-border rounded-lg shadow-lg p-3">
        <div className="flex items-center gap-2 mb-2">
          <div 
            className="w-3 h-3 rounded-full" 
            style={{ backgroundColor: payload[0].color }}
          />
          <span className="font-medium">{data.priority}</span>
        </div>
        <div className="space-y-1 text-sm">
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">Count:</span>
            <span className="font-medium">{data.count}</span>
          </div>
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">Percentage:</span>
            <span className="font-medium">{data.percentage}%</span>
          </div>
        </div>
      </div>
    )
  }
  return null
}

const renderCustomLabel = (entry: any) => {
  return `${entry.priority}: ${entry.percentage}%`
}

export function PriorityChart({ 
  data, 
  chartType = 'donut',
  height = 300,
  showPercentages = true,
  className = ""
}: PriorityChartProps) {
  const chartData = useMemo(() => {
    return data.map(item => ({
      ...item,
      fill: PRIORITY_COLORS[item.priority as keyof typeof PRIORITY_COLORS] || PRIORITY_COLORS.default
    }))
  }, [data])

  const totalIncidents = useMemo(() => {
    return data.reduce((sum, item) => sum + item.count, 0)
  }, [data])

  if (chartType === 'bar') {
    return (
      <div className={`w-full ${className}`}>
        <ResponsiveContainer width="100%" height={height}>
          <BarChart
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
              dataKey="priority"
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
            <Bar 
              dataKey="count" 
              radius={[4, 4, 0, 0]}
              fill="hsl(var(--primary))"
            />
          </BarChart>
        </ResponsiveContainer>
      </div>
    )
  }

  return (
    <div className={`w-full ${className}`}>
      <ResponsiveContainer width="100%" height={height}>
        <PieChart>
          <Pie
            data={chartData}
            cx="50%"
            cy="50%"
            labelLine={false}
            label={showPercentages ? renderCustomLabel : false}
            outerRadius={chartType === 'donut' ? 100 : 120}
            innerRadius={chartType === 'donut' ? 60 : 0}
            fill="#8884d8"
            dataKey="count"
            stroke="hsl(var(--background))"
            strokeWidth={2}
          >
            {chartData.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={entry.fill} />
            ))}
          </Pie>
          <Tooltip content={<CustomTooltip />} />
          <Legend 
            verticalAlign="bottom" 
            height={36}
            formatter={(value, entry: any) => (
              <span style={{ color: entry.color }}>
                {value} ({entry.payload?.count || 0})
              </span>
            )}
          />
        </PieChart>
      </ResponsiveContainer>
      
      {/* Center text for donut chart */}
      {chartType === 'donut' && (
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <div className="text-center">
            <div className="text-2xl font-bold text-primary">{totalIncidents}</div>
            <div className="text-sm text-muted-foreground">Total Incidents</div>
          </div>
        </div>
      )}
    </div>
  )
}