import { useState, useRef, useEffect, ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface VirtualTableProps {
  data: any[]
  columns: Array<{
    key: string
    title: string
    render?: (value: any, row: any, index: number) => ReactNode
    width?: string | number
  }>
  rowHeight?: number
  height?: number
  className?: string
}

export function VirtualTable({
  data,
  columns,
  rowHeight = 50,
  height = 400,
  className
}: VirtualTableProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [scrollTop, setScrollTop] = useState(0)
  const [containerHeight, setContainerHeight] = useState(height)

  // Update container height on resize
  useEffect(() => {
    const updateHeight = () => {
      if (containerRef.current) {
        setContainerHeight(containerRef.current.clientHeight)
      }
    }

    updateHeight()
    window.addEventListener('resize', updateHeight)
    return () => window.removeEventListener('resize', updateHeight)
  }, [])

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    setScrollTop(e.currentTarget.scrollTop)
  }

  // Calculate visible rows
  const visibleRowCount = Math.ceil(containerHeight / rowHeight) + 5 // Add buffer
  const startRowIndex = Math.floor(scrollTop / rowHeight)
  const endRowIndex = Math.min(startRowIndex + visibleRowCount, data.length)
  
  const visibleRows = data.slice(startRowIndex, endRowIndex)
  const totalHeight = data.length * rowHeight
  const offsetY = startRowIndex * rowHeight

  return (
    <div 
      ref={containerRef}
      className={cn("w-full overflow-auto border rounded-md", className)}
      style={{ height: `${height}px` }}
      onScroll={handleScroll}
    >
      <div style={{ height: `${totalHeight}px`, position: 'relative' }}>
        <div style={{ 
          transform: `translateY(${offsetY}px)`,
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0
        }}>
          <table className="w-full">
            <thead className="bg-muted sticky top-0 z-10">
              <tr>
                {columns.map((column) => (
                  <th 
                    key={column.key}
                    className="px-4 py-3 text-left text-sm font-medium text-muted-foreground border-b"
                    style={{ width: column.width }}
                  >
                    {column.title}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {visibleRows.map((row, rowIndex) => (
                <tr 
                  key={startRowIndex + rowIndex}
                  className="border-b hover:bg-muted/50 transition-colors"
                >
                  {columns.map((column) => (
                    <td 
                      key={column.key}
                      className="px-4 py-3 text-sm"
                    >
                      {column.render 
                        ? column.render(row[column.key], row, startRowIndex + rowIndex)
                        : String(row[column.key])
                      }
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

// Memoized row component for better performance
interface VirtualTableRowProps {
  columns: VirtualTableProps['columns']
  row: any
  rowIndex: number
}

export function VirtualTableRow({ columns, row, rowIndex }: VirtualTableRowProps) {
  return (
    <tr className="border-b hover:bg-muted/50 transition-colors">
      {columns.map((column) => (
        <td 
          key={column.key}
          className="px-4 py-3 text-sm"
        >
          {column.render 
            ? column.render(row[column.key], row, rowIndex)
            : String(row[column.key])
          }
        </td>
      ))}
    </tr>
  )
}