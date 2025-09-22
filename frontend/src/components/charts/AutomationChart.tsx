import { useState, useMemo } from 'react'
import {
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ScatterChart,
  Scatter,
  ReferenceLine,
  ComposedChart,
  Line
} from 'recharts'
import { AutomationAnalysis } from '@/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Bot, Zap, Clock, TrendingUp } from 'lucide-react'

interface AutomationChartProps {
  data: AutomationAnalysis[]
  height?: number
  showScatterPlot?: boolean
  className?: string
}

interface TooltipProps {
  active?: boolean
  payload?: any[]
  label?: string
}

type ViewMode = 'opportunities' | 'feasibility' | 'impact'

const CustomTooltip = ({ active, payload, label }: TooltipProps) => {
  if (active && payload && payload.length) {
    const data = payload[0].payload
    
    return (
      <div className="bg-background border border-border rounded-lg shadow-lg p-3 min-w-[250px]">
        <p className="font-medium mb-2">{label}</p>
        <div className="space-y-1 text-sm">
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">Incidents:</span>
            <span className="font-medium">{data.incident_count}</span>
          </div>
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">Automation Score:</span>
            <span className="font-medium">{(data.automation_score * 100).toFixed(1)}%</span>
          </div>
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">Feasible:</span>
            <Badge variant={data.automation_feasible ? "default" : "secondary"}>
              {data.automation_feasible ? "Yes" : "No"}
            </Badge>
          </div>
        </div>
      </div>
    )
  }
  return null
}

export function AutomationChart({ 
  data, 
  height = 400,
  showScatterPlot = false,
  className = ""
}: AutomationChartProps) {
  const [viewMode, setViewMode] = useState<ViewMode>('opportunities')

  const chartData = useMemo(() => {
    return data.map(item => ({
      ...item,
      name: item.it_process_group.length > 20 
        ? `${item.it_process_group.substring(0, 20)}...` 
        : item.it_process_group,
      fullName: item.it_process_group,
      automationPercentage: Math.round(item.automation_score * 100),
      potentialSavings: item.incident_count * (item.automation_score * 2), // Mock calculation
      priority: item.automation_feasible && item.automation_score > 0.7 ? 'High' :
                item.automation_feasible && item.automation_score > 0.4 ? 'Medium' : 'Low'
    }))
  }, [data])

  const stats = useMemo(() => {
    const totalIncidents = data.reduce((sum, item) => sum + item.incident_count, 0)
    const automatable = data.filter(item => item.automation_feasible).length
    const highPotential = data.filter(item => item.automation_score > 0.7).length
    const avgScore = data.reduce((sum, item) => sum + item.automation_score, 0) / data.length

    return {
      totalIncidents,
      totalProcesses: data.length,
      automatable,
      highPotential,
      avgScore: avgScore * 100
    }
  }, [data])

  const viewModeButtons: { key: ViewMode; label: string; icon: any }[] = [
    { key: 'opportunities', label: 'Opportunities', icon: Bot },
    { key: 'feasibility', label: 'Feasibility', icon: Zap },
    { key: 'impact', label: 'Impact', icon: TrendingUp }
  ]

  const getChartDataForMode = () => {
    switch (viewMode) {
      case 'feasibility':
        return chartData.filter(item => item.automation_feasible)
      case 'impact':
        return chartData.sort((a, b) => b.potentialSavings - a.potentialSavings).slice(0, 10)
      default:
        return chartData.sort((a, b) => b.automationPercentage - a.automationPercentage)
    }
  }

  const currentData = getChartDataForMode()

  if (showScatterPlot) {
    return (
      <Card className={className}>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bot className="h-5 w-5" />
            Automation Opportunity Matrix
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={height}>
            <ScatterChart
              margin={{
                top: 20,
                right: 30,
                left: 20,
                bottom: 20,
              }}
            >
              <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
              <XAxis 
                type="number"
                dataKey="incident_count"
                name="Incident Count"
                tick={{ fontSize: 12 }}
                tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                label={{ value: 'Incident Volume', position: 'insideBottom', offset: -10 }}
              />
              <YAxis 
                type="number"
                dataKey="automationPercentage"
                name="Automation Score"
                tick={{ fontSize: 12 }}
                tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                label={{ value: 'Automation Score (%)', angle: -90, position: 'insideLeft' }}
              />
              <Tooltip content={<CustomTooltip />} />
              
              <Scatter 
                name="IT Process Groups" 
                data={chartData} 
                fill="hsl(var(--primary))"
              />

              {/* Reference lines for high impact quadrant */}
              <ReferenceLine 
                x={stats.totalIncidents / stats.totalProcesses} 
                stroke="hsl(var(--muted-foreground))" 
                strokeDasharray="3 3"
              />
              <ReferenceLine 
                y={70} 
                stroke="hsl(var(--muted-foreground))" 
                strokeDasharray="3 3"
              />
            </ScatterChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className={`w-full space-y-6 ${className}`}>
      {/* Key Metrics */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2">
              <Bot className="h-4 w-4 text-primary" />
              <div className="text-2xl font-bold">{stats.automatable}</div>
            </div>
            <p className="text-xs text-muted-foreground">Automatable Processes</p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2">
              <Zap className="h-4 w-4 text-yellow-500" />
              <div className="text-2xl font-bold">{stats.highPotential}</div>
            </div>
            <p className="text-xs text-muted-foreground">High Potential</p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-green-500" />
              <div className="text-2xl font-bold">{stats.avgScore.toFixed(1)}%</div>
            </div>
            <p className="text-xs text-muted-foreground">Avg Score</p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-blue-500" />
              <div className="text-2xl font-bold">{stats.totalIncidents}</div>
            </div>
            <p className="text-xs text-muted-foreground">Total Incidents</p>
          </CardContent>
        </Card>
      </div>

      {/* View Mode Controls */}
      <div className="flex items-center gap-2">
        <span className="text-sm text-muted-foreground">View:</span>
        <div className="flex border rounded-md">
          {viewModeButtons.map((button) => {
            const IconComponent = button.icon
            return (
              <Button
                key={button.key}
                variant={viewMode === button.key ? "default" : "ghost"}
                size="sm"
                className="rounded-none first:rounded-l-md last:rounded-r-md"
                onClick={() => setViewMode(button.key)}
              >
                <IconComponent className="h-4 w-4 mr-1" />
                {button.label}
              </Button>
            )
          })}
        </div>
      </div>

      {/* Chart */}
      <Card>
        <CardHeader>
          <CardTitle>
            {viewMode === 'opportunities' && 'Automation Opportunities by Process Group'}
            {viewMode === 'feasibility' && 'Feasible Automation Candidates'}
            {viewMode === 'impact' && 'High Impact Automation Targets'}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={height}>
            <ComposedChart
              data={currentData}
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
                yAxisId="left"
                orientation="left"
                tick={{ fontSize: 12 }}
                tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              />
              <YAxis 
                yAxisId="right"
                orientation="right"
                tick={{ fontSize: 12 }}
                tickLine={{ stroke: 'currentColor', strokeWidth: 1 }}
                axisLine={{ stroke: 'currentColor', strokeWidth: 1 }}
              />
              <Tooltip content={<CustomTooltip />} />
              <Legend />
              
              <Bar
                yAxisId="left"
                dataKey="incident_count"
                fill="hsl(var(--primary))"
                name="Incident Count"
                radius={[4, 4, 0, 0]}
              />
              
              <Line
                yAxisId="right"
                type="monotone"
                dataKey="automationPercentage"
                stroke="hsl(var(--warning))"
                strokeWidth={3}
                dot={{ fill: 'hsl(var(--warning))', strokeWidth: 2, r: 4 }}
                name="Automation Score (%)"
              />

              <ReferenceLine 
                yAxisId="right"
                y={70} 
                stroke="hsl(var(--success))" 
                strokeDasharray="3 3"
                label={{ 
                  value: "High Potential (70%)", 
                  position: "top"
                }}
              />
            </ComposedChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      {/* Process Group Details */}
      <Card>
        <CardHeader>
          <CardTitle>Process Group Priorities</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {chartData
              .filter(item => item.automation_feasible)
              .sort((a, b) => b.automationPercentage - a.automationPercentage)
              .slice(0, 5)
              .map((item, index) => (
                <div key={item.it_process_group} className="flex items-center justify-between p-3 border rounded-lg">
                  <div className="flex items-center gap-3">
                    <div className="text-sm font-medium text-muted-foreground">
                      #{index + 1}
                    </div>
                    <div>
                      <div className="font-medium">{item.fullName}</div>
                      <div className="text-sm text-muted-foreground">
                        {item.incident_count} incidents
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant={
                      item.priority === 'High' ? 'default' :
                      item.priority === 'Medium' ? 'secondary' : 'outline'
                    }>
                      {item.priority} Priority
                    </Badge>
                    <div className="text-right">
                      <div className="font-bold text-primary">
                        {item.automationPercentage}%
                      </div>
                      <div className="text-xs text-muted-foreground">
                        Score
                      </div>
                    </div>
                  </div>
                </div>
              ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}