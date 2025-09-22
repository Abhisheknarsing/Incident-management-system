import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Upload, BarChart3, Settings } from 'lucide-react'

const navigation = [
  { name: 'Upload', href: '/', icon: Upload },
  { name: 'Processing', href: '/processing', icon: Settings },
  { name: 'Dashboard', href: '/dashboard', icon: BarChart3 },
]

export function Navigation() {
  const location = useLocation()

  return (
    <nav className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container mx-auto px-4">
        <div className="flex h-16 items-center justify-between">
          <div className="flex items-center space-x-8">
            <Link to="/" className="text-xl font-bold">
              Incident Management System
            </Link>
            <div className="flex space-x-4">
              {navigation.map((item) => {
                const Icon = item.icon
                return (
                  <Link
                    key={item.name}
                    to={item.href}
                    className={cn(
                      'flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors',
                      location.pathname === item.href
                        ? 'bg-primary text-primary-foreground'
                        : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                    )}
                  >
                    <Icon className="h-4 w-4" />
                    <span>{item.name}</span>
                  </Link>
                )
              })}
            </div>
          </div>
        </div>
      </div>
    </nav>
  )
}