import * as React from "react"
import { LucideIcon } from "lucide-react"
import { cn } from "@/lib/utils"

interface EmptyStateProps {
  icon?: LucideIcon | React.ReactNode
  title: string
  description?: string
  action?: React.ReactNode
  className?: string
}

export function EmptyState({ icon, title, description, action, className }: EmptyStateProps) {
  const IconComponent = icon as LucideIcon
  
  return (
    <div className={cn("flex flex-col items-center justify-center text-center p-8", className)}>
      {icon && (
        <div className="mb-4 text-muted-foreground">
          {typeof icon === 'function' ? (
            <IconComponent className="h-12 w-12" />
          ) : (
            icon
          )}
        </div>
      )}
      <h3 className="text-lg font-semibold mb-2">{title}</h3>
      {description && (
        <p className="text-muted-foreground mb-4 max-w-md">{description}</p>
      )}
      <div>
        {action}
      </div>
    </div>
  )
}