import { useState, useEffect, useRef, ReactNode } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Maximize2, Minimize2 } from 'lucide-react'

interface OptimizedChartContainerProps {
  title: string
  description: string
  children: ReactNode
  defaultHeight?: number
  expandedHeight?: number
  className?: string
  enableOptimization?: boolean
}

export function OptimizedChartContainer({
  title,
  description,
  children,
  defaultHeight = 300,
  expandedHeight = 500,
  className = "",
  enableOptimization = true
}: OptimizedChartContainerProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [isVisible, setIsVisible] = useState(false)
  const [shouldRender, setShouldRender] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const observerRef = useRef<IntersectionObserver | null>(null)

  // Intersection Observer for lazy loading
  useEffect(() => {
    if (!enableOptimization || !containerRef.current) return

    observerRef.current = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true)
          setShouldRender(true)
          // Stop observing once visible
          if (observerRef.current && containerRef.current) {
            observerRef.current.unobserve(containerRef.current)
          }
        }
      },
      {
        threshold: 0.1, // Trigger when 10% of the element is visible
        rootMargin: '50px' // Start loading 50px before entering viewport
      }
    )

    if (containerRef.current) {
      observerRef.current.observe(containerRef.current)
    }

    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect()
      }
    }
  }, [enableOptimization])

  // Render optimization based on visibility
  const renderContent = () => {
    if (!enableOptimization) {
      return children
    }

    // Only render when visible or already rendered
    if (shouldRender || isVisible) {
      return children
    }

    // Show placeholder while not rendered
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center text-muted-foreground">
          <div className="animate-pulse mb-2">
            <div className="h-6 bg-muted rounded w-32 mx-auto"></div>
          </div>
          <p className="text-sm">Loading chart...</p>
        </div>
      </div>
    )
  }

  const currentHeight = isExpanded ? expandedHeight : defaultHeight

  return (
    <Card 
      ref={containerRef}
      className={`overflow-hidden ${className}`}
    >
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div>
          <CardTitle className="text-lg">{title}</CardTitle>
          <p className="text-sm text-muted-foreground">{description}</p>
        </div>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setIsExpanded(!isExpanded)}
          aria-label={isExpanded ? "Minimize chart" : "Expand chart"}
        >
          {isExpanded ? (
            <Minimize2 className="h-4 w-4" />
          ) : (
            <Maximize2 className="h-4 w-4" />
          )}
        </Button>
      </CardHeader>
      <CardContent 
        className="p-0"
        style={{ height: `${currentHeight}px` }}
      >
        {renderContent()}
      </CardContent>
    </Card>
  )
}