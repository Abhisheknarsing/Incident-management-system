import { Button } from '@/components/ui/button'
import { ExportDialog } from './ExportDialog'
import { FilterState } from '@/types'
import { Download } from 'lucide-react'

interface ExportButtonProps {
  dataType?: string
  filters?: Partial<FilterState>
  variant?: 'default' | 'outline' | 'ghost'
  size?: 'sm' | 'default' | 'lg'
  className?: string
  children?: React.ReactNode
}

export function ExportButton({
  dataType = 'all',
  filters = {},
  variant = 'outline',
  size = 'sm',
  className = "",
  children
}: ExportButtonProps) {
  return (
    <ExportDialog
      dataType={dataType}
      filters={filters}
      trigger={
        <Button variant={variant} size={size} className={className}>
          <Download className="h-4 w-4 mr-2" />
          {children || 'Export'}
        </Button>
      }
    />
  )
}