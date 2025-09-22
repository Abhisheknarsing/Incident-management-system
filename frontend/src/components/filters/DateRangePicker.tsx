import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Calendar } from 'lucide-react'
import { format, subDays, subMonths } from 'date-fns'

interface DateRangePickerProps {
  value?: { start: string; end: string }
  onChange: (value: { start: string; end: string } | undefined) => void
  className?: string
}

interface DateRangePreset {
  label: string
  getValue: () => { start: string; end: string }
}

const dateRangePresets: DateRangePreset[] = [
  {
    label: 'Last 7 days',
    getValue: () => ({
      start: format(subDays(new Date(), 7), 'yyyy-MM-dd'),
      end: format(new Date(), 'yyyy-MM-dd')
    })
  },
  {
    label: 'Last 30 days',
    getValue: () => ({
      start: format(subDays(new Date(), 30), 'yyyy-MM-dd'),
      end: format(new Date(), 'yyyy-MM-dd')
    })
  },
  {
    label: 'Last 3 months',
    getValue: () => ({
      start: format(subMonths(new Date(), 3), 'yyyy-MM-dd'),
      end: format(new Date(), 'yyyy-MM-dd')
    })
  },
  {
    label: 'Last 6 months',
    getValue: () => ({
      start: format(subMonths(new Date(), 6), 'yyyy-MM-dd'),
      end: format(new Date(), 'yyyy-MM-dd')
    })
  },
  {
    label: 'This year',
    getValue: () => ({
      start: format(new Date(new Date().getFullYear(), 0, 1), 'yyyy-MM-dd'),
      end: format(new Date(), 'yyyy-MM-dd')
    })
  }
]

export function DateRangePicker({ value, onChange, className = "" }: DateRangePickerProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [startDate, setStartDate] = useState(value?.start || '')
  const [endDate, setEndDate] = useState(value?.end || '')

  const handleApply = () => {
    if (startDate && endDate) {
      onChange({ start: startDate, end: endDate })
    } else {
      onChange(undefined)
    }
    setIsOpen(false)
  }

  const handleClear = () => {
    setStartDate('')
    setEndDate('')
    onChange(undefined)
    setIsOpen(false)
  }

  const handlePresetSelect = (preset: DateRangePreset) => {
    const range = preset.getValue()
    setStartDate(range.start)
    setEndDate(range.end)
    onChange(range)
    setIsOpen(false)
  }

  const formatDisplayValue = () => {
    if (value?.start && value?.end) {
      return `${value.start} to ${value.end}`
    }
    return 'Select date range...'
  }

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className={`w-full justify-start text-left font-normal ${className}`}
        >
          <Calendar className="mr-2 h-4 w-4" />
          {formatDisplayValue()}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <div className="p-4 space-y-4">
          <div className="space-y-2">
            <h4 className="font-medium text-sm">Quick Select</h4>
            <div className="grid grid-cols-1 gap-2">
              {dateRangePresets.map((preset) => (
                <Button
                  key={preset.label}
                  variant="ghost"
                  size="sm"
                  className="justify-start"
                  onClick={() => handlePresetSelect(preset)}
                >
                  {preset.label}
                </Button>
              ))}
            </div>
          </div>
          
          <div className="border-t pt-4 space-y-3">
            <h4 className="font-medium text-sm">Custom Range</h4>
            <div className="space-y-2">
              <div>
                <label className="text-xs text-muted-foreground">Start Date</label>
                <Input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="w-full"
                />
              </div>
              <div>
                <label className="text-xs text-muted-foreground">End Date</label>
                <Input
                  type="date"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  min={startDate}
                  className="w-full"
                />
              </div>
            </div>
            
            <div className="flex gap-2 pt-2">
              <Button
                size="sm"
                onClick={handleApply}
                disabled={!startDate || !endDate}
                className="flex-1"
              >
                Apply
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={handleClear}
                className="flex-1"
              >
                Clear
              </Button>
            </div>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}