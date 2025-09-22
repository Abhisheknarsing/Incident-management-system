import { ReactNode, useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ExportButton } from '@/components/export'
import { FilterState } from '@/types'
import { Maximize2, Minimize2 } from 'lucide-react'

interface ResponsiveChartContainerProps {
  title: string
  description?: string
  children: ReactNode
  onExport?: () => void
  exportDataType?: string
  filters?: Partial<FilterState>
  className?: string
  defaultHeight?: number
  expandedHeight?: number
}

export function ResponsiveChartContainer({
  title,
  description,
  children,
  onExport,
  exportDataType,
  filters = {},
  className = "",
  defaultHeight = 300,
  expandedHeight = 500
}: ResponsiveChartContainerProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [isMobile, setIsMobile] = useState(false)

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768)
    }
    
    checkMobile()
    window.addEventListener('resize', checkMobile)
    
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  const height = isExpanded ? expandedHeight : (isMobile ? 250 : defaultHeight)

  return (
    <Card className={`${className} ${isExpanded ? 'col-span-full' : ''} transition-all duration-300`}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className={`${isMobile ? 'text-lg' : 'text-xl'}`}>
              {title}
            </CardTitle>
            {description && (
              <p className="text-sm text-muted-foreground mt-1">
                {description}
              </p>
            )}
          </div>
          
          <div className="flex items-center gap-2">
            {(onExport || exportDataType) && (
              exportDataType ? (
                <ExportButton
                  dataType={exportDataType}
                  filters={filters}
                  variant="outline"
                  size="sm"
                  className="h-8 px-2"
                />
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onExport}
                  className="h-8 px-2"
                >
                  Export
                </Button>
              )
            )}
            
            <Button
              variant="outline"
              size="sm"
              onClick={() => setIsExpanded(!isExpanded)}
              className="h-8 w-8 p-0"
            >
              {isExpanded ? (
                <Minimize2 className="h-4 w-4" />
              ) : (
                <Maximize2 className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="pt-2">
        <div style={{ height: `${height}px` }}>
          {children}
        </div>
      </CardContent>
    </Card>
  )
}

// Hook for responsive chart dimensions
export function useResponsiveChart() {
  const [dimensions, setDimensions] = useState({
    width: 0,
    height: 0,
    isMobile: false,
    isTablet: false
  })

  useEffect(() => {
    const updateDimensions = () => {
      const width = window.innerWidth
      const height = window.innerHeight
      
      setDimensions({
        width,
        height,
        isMobile: width < 768,
        isTablet: width >= 768 && width < 1024
      })
    }

    updateDimensions()
    window.addEventListener('resize', updateDimensions)
    
    return () => window.removeEventListener('resize', updateDimensions)
  }, [])

  return dimensions
}