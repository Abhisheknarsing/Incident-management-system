import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { 
  TimelineChart, 
  TrendAnalysisChart, 
  ResponsiveChartContainer,
  PriorityChart,
  ApplicationChart,
  ResolutionChart,
  SentimentChart,
  AutomationChart,
  MetricsDashboard,
  IncidentVolumeCard,
  ResolutionTimeCard,
  AutomationPotentialCard,
  SentimentScoreCard,
  ActiveApplicationsCard
} from '@/components/charts'
import { FilterPanel } from '@/components/filters'
import { ExportButton } from '@/components/export'
import { 
  useTimelineData, 
  usePriorityAnalysis, 
  useApplicationAnalysis,
  useResolutionMetrics,
  useSentimentAnalysis,
  useAutomationAnalysis
} from '@/hooks/useAnalytics'
import { useFilterState } from '@/hooks/useFilters'
import { useFilterOptions, useAvailableApplications } from '@/hooks/useFilterOptions'
import { LoadingSpinner } from '@/components/ui/loading-spinner'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertCircle } from 'lucide-react'

export function DashboardPage() {
  const { filters, updateFilters, hasActiveFilters, activeFilterCount } = useFilterState()
  const { data: filterOptions } = useFilterOptions()
  const { data: availableApplications } = useAvailableApplications()
  
  const { 
    data: timelineData, 
    isLoading: timelineLoading, 
    error: timelineError 
  } = useTimelineData(filters)

  const { 
    data: priorityData, 
    isLoading: priorityLoading, 
    error: priorityError 
  } = usePriorityAnalysis(filters)

  const { 
    data: applicationData, 
    isLoading: applicationLoading, 
    error: applicationError 
  } = useApplicationAnalysis(filters)

  const { 
    data: resolutionData, 
    isLoading: resolutionLoading, 
    error: resolutionError 
  } = useResolutionMetrics(filters)

  const { 
    data: sentimentData, 
    isLoading: sentimentLoading, 
    error: sentimentError 
  } = useSentimentAnalysis(filters)

  const { 
    data: automationData, 
    isLoading: automationLoading, 
    error: automationError 
  } = useAutomationAnalysis(filters)





  // Calculate summary metrics for the dashboard
  const summaryMetrics = {
    totalIncidents: timelineData?.reduce((sum, item) => sum + item.count, 0) || 0,
    avgResolutionTime: resolutionData?.avg_resolution_time || 0,
    automationPotential: automationData?.filter(item => item.automation_feasible).length || 0,
    sentimentScore: sentimentData ? 
      ((sentimentData.positive - sentimentData.negative) / 
       (sentimentData.positive + sentimentData.negative + sentimentData.neutral)) : 0,
    activeApplications: applicationData?.length || 0,
    criticalApplications: applicationData?.filter(app => app.trend === 'up').length || 0
  }

  return (
    <MetricsDashboard className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Analytics Dashboard</h1>
          <p className="text-muted-foreground">
            View comprehensive incident analytics and reports
            {hasActiveFilters && (
              <span className="ml-2 text-primary">
                ({activeFilterCount} filter{activeFilterCount !== 1 ? 's' : ''} applied)
              </span>
            )}
          </p>
        </div>
        <ExportButton 
          dataType="all" 
          filters={filters}
          variant="default"
          size="default"
        >
          Export Dashboard
        </ExportButton>
      </div>

      {/* Filter Panel */}
      <FilterPanel
        filters={filters}
        onFiltersChange={updateFilters}
        availableOptions={{
          priorities: filterOptions?.priorities || [],
          applications: availableApplications || [],
          statuses: filterOptions?.statuses || []
        }}
      />

      {/* Key Metrics Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <IncidentVolumeCard 
          total={summaryMetrics.totalIncidents}
          trend="stable"
          trendValue="2.1%"
        />
        <ResolutionTimeCard 
          avgTime={summaryMetrics.avgResolutionTime}
          trend="down"
          trendValue="5.2%"
        />
        <AutomationPotentialCard 
          percentage={(summaryMetrics.automationPotential / Math.max(automationData?.length || 1, 1)) * 100}
          processCount={summaryMetrics.automationPotential}
        />
        <SentimentScoreCard 
          score={summaryMetrics.sentimentScore}
          trend="up"
          trendValue="0.15"
        />
        <ActiveApplicationsCard 
          count={summaryMetrics.activeApplications}
          criticalCount={summaryMetrics.criticalApplications}
        />
      </div>

      {/* Timeline Analysis Section */}
      <div className="grid grid-cols-1 gap-6">
        {timelineError ? (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Failed to load timeline data. Please try again later.
            </AlertDescription>
          </Alert>
        ) : timelineLoading ? (
          <Card>
            <CardHeader>
              <CardTitle>Timeline Analysis</CardTitle>
              <CardDescription>Loading incident trends...</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <LoadingSpinner size="lg" />
            </CardContent>
          </Card>
        ) : timelineData && timelineData.length > 0 ? (
          <>
            {/* Main trend analysis with interactive features */}
            <TrendAnalysisChart 
              data={timelineData}
              height={400}
              className="col-span-full"
            />
            
            {/* Simple timeline chart */}
            <ResponsiveChartContainer
              title="Incident Timeline"
              description="Daily incident counts with priority breakdown"
              exportDataType="timeline"
              filters={filters}
              defaultHeight={300}
              expandedHeight={500}
            >
              <TimelineChart 
                data={timelineData}
                height={300}
                showBrush={true}
                showLegend={true}
              />
            </ResponsiveChartContainer>
          </>
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Timeline Analysis</CardTitle>
              <CardDescription>No incident data available</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <p className="text-muted-foreground">
                Upload and process incident data to view timeline analysis
              </p>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Priority and Application Analysis */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Priority Distribution */}
        {priorityError ? (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Failed to load priority data. Please try again later.
            </AlertDescription>
          </Alert>
        ) : priorityLoading ? (
          <Card>
            <CardHeader>
              <CardTitle>Priority Distribution</CardTitle>
              <CardDescription>Loading priority breakdown...</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <LoadingSpinner size="lg" />
            </CardContent>
          </Card>
        ) : priorityData && priorityData.length > 0 ? (
          <ResponsiveChartContainer
            title="Priority Distribution"
            description="Breakdown by priority levels"
            exportDataType="priority"
            filters={filters}
            defaultHeight={350}
            expandedHeight={500}
          >
            <div className="relative">
              <PriorityChart 
                data={priorityData}
                chartType="donut"
                height={350}
                showPercentages={true}
              />
            </div>
          </ResponsiveChartContainer>
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Priority Distribution</CardTitle>
              <CardDescription>No priority data available</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <p className="text-muted-foreground">
                Upload and process incident data to view priority analysis
              </p>
            </CardContent>
          </Card>
        )}

        {/* Application Analysis */}
        {applicationError ? (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Failed to load application data. Please try again later.
            </AlertDescription>
          </Alert>
        ) : applicationLoading ? (
          <Card>
            <CardHeader>
              <CardTitle>Application Analysis</CardTitle>
              <CardDescription>Loading application breakdown...</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <LoadingSpinner size="lg" />
            </CardContent>
          </Card>
        ) : applicationData && applicationData.length > 0 ? (
          <ResponsiveChartContainer
            title="Application Analysis"
            description="Incidents by application with resolution times"
            exportDataType="application"
            filters={filters}
            defaultHeight={350}
            expandedHeight={600}
          >
            <ApplicationChart 
              data={applicationData}
              height={350}
              showResolutionTime={true}
              maxApplications={8}
            />
          </ResponsiveChartContainer>
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Application Analysis</CardTitle>
              <CardDescription>No application data available</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <p className="text-muted-foreground">
                Upload and process incident data to view application analysis
              </p>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Resolution Analysis */}
      <div className="grid grid-cols-1 gap-6">
        {resolutionError ? (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Failed to load resolution data. Please try again later.
            </AlertDescription>
          </Alert>
        ) : resolutionLoading ? (
          <Card>
            <CardHeader>
              <CardTitle>Resolution Analysis</CardTitle>
              <CardDescription>Loading resolution metrics...</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <LoadingSpinner size="lg" />
            </CardContent>
          </Card>
        ) : resolutionData ? (
          <ResolutionChart 
            data={resolutionData}
            height={300}
            className="col-span-full"
          />
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Resolution Analysis</CardTitle>
              <CardDescription>No resolution data available</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <p className="text-muted-foreground">
                Upload and process incident data to view resolution analysis
              </p>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Sentiment and Automation Analysis */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Sentiment Analysis */}
        {sentimentError ? (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Failed to load sentiment data. Please try again later.
            </AlertDescription>
          </Alert>
        ) : sentimentLoading ? (
          <Card>
            <CardHeader>
              <CardTitle>Sentiment Analysis</CardTitle>
              <CardDescription>Loading sentiment breakdown...</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <LoadingSpinner size="lg" />
            </CardContent>
          </Card>
        ) : sentimentData ? (
          <ResponsiveChartContainer
            title="Sentiment Analysis"
            description="Incident sentiment breakdown"
            exportDataType="sentiment"
            filters={filters}
            defaultHeight={350}
            expandedHeight={500}
          >
            <SentimentChart 
              data={sentimentData}
              chartType="pie"
              height={350}
              showIcons={true}
            />
          </ResponsiveChartContainer>
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Sentiment Analysis</CardTitle>
              <CardDescription>No sentiment data available</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <p className="text-muted-foreground">
                Upload and process incident data to view sentiment analysis
              </p>
            </CardContent>
          </Card>
        )}

        {/* Automation Opportunities Preview */}
        {automationError ? (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Failed to load automation data. Please try again later.
            </AlertDescription>
          </Alert>
        ) : automationLoading ? (
          <Card>
            <CardHeader>
              <CardTitle>Automation Opportunities</CardTitle>
              <CardDescription>Loading automation analysis...</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <LoadingSpinner size="lg" />
            </CardContent>
          </Card>
        ) : automationData && automationData.length > 0 ? (
          <ResponsiveChartContainer
            title="Automation Opportunities"
            description="Process automation potential"
            exportDataType="automation"
            filters={filters}
            defaultHeight={350}
            expandedHeight={600}
          >
            <AutomationChart 
              data={automationData}
              height={350}
              showScatterPlot={false}
            />
          </ResponsiveChartContainer>
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Automation Opportunities</CardTitle>
              <CardDescription>No automation data available</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-64">
              <p className="text-muted-foreground">
                Upload and process incident data to view automation analysis
              </p>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Detailed Automation Analysis */}
      {automationData && automationData.length > 0 && (
        <div className="grid grid-cols-1 gap-6">
          <AutomationChart 
            data={automationData}
            height={400}
            showScatterPlot={true}
            className="col-span-full"
          />
        </div>
      )}
    </MetricsDashboard>
  )
}