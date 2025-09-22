import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Input } from '@/components/ui/input'
import { Check, ChevronDown, X, Search } from 'lucide-react'
import { cn } from '@/lib/utils'

interface MultiSelectFilterProps {
  options: string[]
  selectedValues: string[]
  onChange: (values: string[]) => void
  placeholder?: string
  className?: string
  maxDisplayItems?: number
}

export function MultiSelectFilter({
  options,
  selectedValues,
  onChange,
  placeholder = "Select options...",
  className = "",
  maxDisplayItems = 3
}: MultiSelectFilterProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')

  const filteredOptions = options.filter(option =>
    option.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const handleToggleOption = (option: string) => {
    const isSelected = selectedValues.includes(option)
    if (isSelected) {
      onChange(selectedValues.filter(value => value !== option))
    } else {
      onChange([...selectedValues, option])
    }
  }

  const handleSelectAll = () => {
    if (selectedValues.length === filteredOptions.length) {
      // Deselect all filtered options
      onChange(selectedValues.filter(value => !filteredOptions.includes(value)))
    } else {
      // Select all filtered options
      const newSelected = [...selectedValues]
      filteredOptions.forEach(option => {
        if (!newSelected.includes(option)) {
          newSelected.push(option)
        }
      })
      onChange(newSelected)
    }
  }

  const handleClearAll = () => {
    onChange([])
    setIsOpen(false)
  }

  const getDisplayText = () => {
    if (selectedValues.length === 0) {
      return placeholder
    }
    
    if (selectedValues.length <= maxDisplayItems) {
      return selectedValues.join(', ')
    }
    
    return `${selectedValues.slice(0, maxDisplayItems).join(', ')} +${selectedValues.length - maxDisplayItems} more`
  }

  const allFilteredSelected = filteredOptions.length > 0 && 
    filteredOptions.every(option => selectedValues.includes(option))

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className={cn(
            "w-full justify-between text-left font-normal",
            selectedValues.length === 0 && "text-muted-foreground",
            className
          )}
        >
          <span className="truncate">{getDisplayText()}</span>
          <div className="flex items-center gap-1">
            {selectedValues.length > 0 && (
              <Badge variant="secondary" className="ml-1">
                {selectedValues.length}
              </Badge>
            )}
            <ChevronDown className="h-4 w-4 opacity-50" />
          </div>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="start">
        <div className="p-3 space-y-3">
          {/* Search */}
          <div className="relative">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search options..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-8"
            />
          </div>

          {/* Actions */}
          <div className="flex items-center justify-between">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleSelectAll}
              disabled={filteredOptions.length === 0}
            >
              {allFilteredSelected ? 'Deselect All' : 'Select All'}
            </Button>
            {selectedValues.length > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleClearAll}
                className="text-muted-foreground hover:text-foreground"
              >
                Clear All
              </Button>
            )}
          </div>

          {/* Options List */}
          <div className="max-h-60 overflow-y-auto space-y-1">
            {filteredOptions.length === 0 ? (
              <div className="text-center text-muted-foreground py-4">
                {searchTerm ? 'No options found' : 'No options available'}
              </div>
            ) : (
              filteredOptions.map((option) => {
                const isSelected = selectedValues.includes(option)
                return (
                  <div
                    key={option}
                    className={cn(
                      "flex items-center space-x-2 rounded-sm px-2 py-1.5 cursor-pointer hover:bg-accent hover:text-accent-foreground",
                      isSelected && "bg-accent text-accent-foreground"
                    )}
                    onClick={() => handleToggleOption(option)}
                  >
                    <div className={cn(
                      "flex h-4 w-4 items-center justify-center rounded-sm border border-primary",
                      isSelected ? "bg-primary text-primary-foreground" : "opacity-50"
                    )}>
                      {isSelected && <Check className="h-3 w-3" />}
                    </div>
                    <span className="flex-1 text-sm">{option}</span>
                  </div>
                )
              })
            )}
          </div>

          {/* Selected Items Display */}
          {selectedValues.length > 0 && (
            <div className="border-t pt-3">
              <div className="text-xs text-muted-foreground mb-2">
                Selected ({selectedValues.length}):
              </div>
              <div className="flex flex-wrap gap-1">
                {selectedValues.map((value) => (
                  <Badge
                    key={value}
                    variant="secondary"
                    className="text-xs gap-1"
                  >
                    {value}
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-auto p-0 ml-1"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleToggleOption(value)
                      }}
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </Badge>
                ))}
              </div>
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
}