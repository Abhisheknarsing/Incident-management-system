import { useMemo } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Brush,
  ReferenceLine
} from 'recharts'
import { format, parseISO } from 'date-fns'
import { TimelineData } from '@/types'

interface TimelineChartProps {
  data: TimelineData[]
  height?: number
  showBrush?: boolean
  showLegend?: boolean
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
          <div key={index} className="flex items-center gap-2 text-sm">
            <div 
              className="w-3 h-3 rounded-full" 
              style={{ backgroundColor: entry.color }}
            />
            <span className="text-muted-foreground">{entry.name}:</span>
            <span className="font-medium">{entry.value}</span>
          </div>
        ))}
      </div>
    )
  }
  return null
}

export function TimelineChart({ 
  data, 
  height = 400, 
  showBrush = true, 
  showLegend = true,
  className = ""
}: TimelineChartProps) {
  const chartData = useMemo(() => {
    return data.map(item => ({
      ...item,
      date: format(parseISO(item.date), 'MMM dd'),
      fullDate: item.date,
      total: item.count
    }))
  }, [data])

  const maxValue = useMemo(() => {
    return Math.max(...data.map(item => item.count))
  }, [data])

  return (
    <div className={`w-full ${className}`}>
      <ResponsiveContainer width="100%" height={height}>
        <LineChart
          data={chartData}
          margin={{
            top: 20,
            right: 30,
            left: 20,
            bottom: showBrush ? 80 : 20,
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
          {showLegend && (
            <Legend 
              wrapperStyle={{ paddingTop: '20px' }}
              iconType="line"
            />
          )}
          
          <Line
            type="monotone"
            dataKey="total"
            stroke="hsl(var(--primary))"
            strokeWidth={3}
            dot={{ fill: 'hsl(var(--primary))', strokeWidth: 2, r: 4 }}
            activeDot={{ r: 6, stroke: 'hsl(var(--primary))', strokeWidth: 2 }}
            name="Total Incidents"
          />
          
          <Line
            type="monotone"
            dataKey="p1_count"
            stroke="hsl(var(--destructive))"
            strokeWidth={2}
            dot={{ fill: 'hsl(var(--destructive))', strokeWidth: 1, r: 3 }}
            name="P1 (Critical)"
          />
          
          <Line
            type="monotone"
            dataKey="p2_count"
            stroke="hsl(var(--warning))"
            strokeWidth={2}
            dot={{ fill: 'hsl(var(--warning))', strokeWidth: 1, r: 3 }}
            name="P2 (High)"
          />
          
          <Line
            type="monotone"
            dataKey="p3_count"
            stroke="hsl(var(--info))"
            strokeWidth={2}
            dot={{ fill: 'hsl(var(--info))', strokeWidth: 1, r: 3 }}
            name="P3 (Medium)"
          />
          
          <Line
            type="monotone"
            dataKey="p4_count"
            stroke="hsl(var(--success))"
            strokeWidth={2}
            dot={{ fill: 'hsl(var(--success))', strokeWidth: 1, r: 3 }}
            name="P4 (Low)"
          />

          {/* Reference line for average */}
          <ReferenceLine 
            y={maxValue * 0.7} 
            stroke="hsl(var(--muted-foreground))" 
            strokeDasharray="5 5"
            label={{ value: "High Activity", position: "top" }}
          />
          
          {showBrush && (
            <Brush 
              dataKey="date" 
              height={30}
              stroke="hsl(var(--primary))"
              fill="hsl(var(--muted))"
            />
          )}
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}