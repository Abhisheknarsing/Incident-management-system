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
  CartesianGrid,
  RadialBarChart,
  RadialBar
} from 'recharts'
import { SentimentAnalysis } from '@/types'
import { Smile, Frown, Meh } from 'lucide-react'

interface SentimentChartProps {
  data: SentimentAnalysis
  chartType?: 'pie' | 'bar' | 'gauge'
  height?: number
  showIcons?: boolean
  className?: string
}

interface TooltipProps {
  active?: boolean
  payload?: any[]
  label?: string
}

const SENTIMENT_COLORS = {
  positive: 'hsl(var(--success))',
  negative: 'hsl(var(--destructive))',
  neutral: 'hsl(var(--warning))'
}

const SENTIMENT_ICONS = {
  positive: Smile,
  negative: Frown,
  neutral: Meh
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
          <span className="font-medium capitalize">{data.sentiment}</span>
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

export function SentimentChart({ 
  data, 
  chartType = 'pie',
  height = 300,
  showIcons = true,
  className = ""
}: SentimentChartProps) {
  const chartData = useMemo(() => {
    const total = data.positive + data.negative + data.neutral
    
    if (total === 0) return []
    
    return [
      {
        sentiment: 'positive',
        count: data.positive,
        percentage: Math.round((data.positive / total) * 100),
        fill: SENTIMENT_COLORS.positive
      },
      {
        sentiment: 'negative',
        count: data.negative,
        percentage: Math.round((data.negative / total) * 100),
        fill: SENTIMENT_COLORS.negative
      },
      {
        sentiment: 'neutral',
        count: data.neutral,
        percentage: Math.round((data.neutral / total) * 100),
        fill: SENTIMENT_COLORS.neutral
      }
    ].filter(item => item.count > 0)
  }, [data])

  const totalIncidents = useMemo(() => {
    return data.positive + data.negative + data.neutral
  }, [data])

  const sentimentScore = useMemo(() => {
    if (totalIncidents === 0) return 0
    // Calculate weighted sentiment score (-1 to 1)
    return ((data.positive - data.negative) / totalIncidents).toFixed(2)
  }, [data, totalIncidents])

  if (chartType === 'gauge') {
    const gaugeData = [
      {
        name: 'Positive',
        value: data.positive,
        fill: SENTIMENT_COLORS.positive
      },
      {
        name: 'Neutral', 
        value: data.neutral,
        fill: SENTIMENT_COLORS.neutral
      },
      {
        name: 'Negative',
        value: data.negative,
        fill: SENTIMENT_COLORS.negative
      }
    ]

    return (
      <div className={`w-full ${className}`}>
        <ResponsiveContainer width="100%" height={height}>
          <RadialBarChart
            cx="50%"
            cy="50%"
            innerRadius="20%"
            outerRadius="80%"
            data={gaugeData}
            startAngle={180}
            endAngle={0}
          >
            <RadialBar
              dataKey="value"
              cornerRadius={4}
              fill="hsl(var(--primary))"
            />
            <Legend 
              iconSize={18}
              layout="horizontal"
              verticalAlign="bottom"
              align="center"
            />
            <Tooltip content={<CustomTooltip />} />
          </RadialBarChart>
        </ResponsiveContainer>
        
        {/* Center sentiment score */}
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <div className="text-center">
            <div className={`text-3xl font-bold ${
              parseFloat(sentimentScore.toString()) > 0.1 ? 'text-green-600' :
              parseFloat(sentimentScore.toString()) < -0.1 ? 'text-red-600' : 'text-gray-600'
            }`}>
              {sentimentScore}
            </div>
            <div className="text-sm text-muted-foreground">Sentiment Score</div>
          </div>
        </div>
      </div>
    )
  }

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
              dataKey="sentiment"
              tick={{ fontSize: 12 }}
              tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              tickFormatter={(value) => value.charAt(0).toUpperCase() + value.slice(1)}
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
            />
          </BarChart>
        </ResponsiveContainer>
      </div>
    )
  }

  // Default pie chart
  return (
    <div className={`w-full relative ${className}`}>
      <ResponsiveContainer width="100%" height={height}>
        <PieChart>
          <Pie
            data={chartData}
            cx="50%"
            cy="50%"
            labelLine={false}
            label={({ sentiment, percentage }) => `${sentiment}: ${percentage}%`}
            outerRadius={100}
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
              <span style={{ color: entry.color }} className="flex items-center gap-1">
                {showIcons && (() => {
                  const IconComponent = SENTIMENT_ICONS[value as keyof typeof SENTIMENT_ICONS]
                  return IconComponent ? <IconComponent className="h-4 w-4" /> : null
                })()}
                {value.charAt(0).toUpperCase() + value.slice(1)} ({entry.payload?.count || 0})
              </span>
            )}
          />
        </PieChart>
      </ResponsiveContainer>

      {/* Summary stats */}
      <div className="mt-4 grid grid-cols-3 gap-4 text-center">
        <div>
          <div className="text-lg font-bold text-green-600">{data.positive}</div>
          <div className="text-xs text-muted-foreground">Positive</div>
        </div>
        <div>
          <div className="text-lg font-bold text-gray-600">{data.neutral}</div>
          <div className="text-xs text-muted-foreground">Neutral</div>
        </div>
        <div>
          <div className="text-lg font-bold text-red-600">{data.negative}</div>
          <div className="text-xs text-muted-foreground">Negative</div>
        </div>
      </div>
    </div>
  )
}