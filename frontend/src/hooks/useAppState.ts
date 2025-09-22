import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { FilterState, Upload } from '@/types'

interface AppState {
  // Filter state
  filters: FilterState
  setFilters: (filters: Partial<FilterState>) => void
  clearFilters: () => void
  
  // Current upload selection
  selectedUpload: Upload | null
  setSelectedUpload: (upload: Upload | null) => void
  
  // UI state
  sidebarOpen: boolean
  setSidebarOpen: (open: boolean) => void
  
  // Notification state
  notifications: Array<{
    id: string
    type: 'success' | 'error' | 'warning' | 'info'
    title: string
    message: string
    timestamp: number
  }>
  addNotification: (notification: Omit<AppState['notifications'][0], 'id' | 'timestamp'>) => void
  removeNotification: (id: string) => void
  clearNotifications: () => void
}

const defaultFilters: FilterState = {
  dateRange: { start: '', end: '' },
  priorities: [],
  applications: [],
  statuses: [],
}

export const useAppState = create<AppState>()(
  persist(
    (set) => ({
      // Filter state
      filters: defaultFilters,
      setFilters: (newFilters) =>
        set((state) => ({
          filters: { ...state.filters, ...newFilters },
        })),
      clearFilters: () => set({ filters: defaultFilters }),
      
      // Current upload selection
      selectedUpload: null,
      setSelectedUpload: (upload) => set({ selectedUpload: upload }),
      
      // UI state
      sidebarOpen: true,
      setSidebarOpen: (open) => set({ sidebarOpen: open }),
      
      // Notification state
      notifications: [],
      addNotification: (notification) => {
        const id = crypto.randomUUID()
        const timestamp = Date.now()
        set((state) => ({
          notifications: [
            ...state.notifications,
            { ...notification, id, timestamp },
          ],
        }))
        
        // Auto-remove notification after 5 seconds
        setTimeout(() => {
          set((state) => ({
            notifications: state.notifications.filter((n) => n.id !== id),
          }))
        }, 5000)
      },
      removeNotification: (id) =>
        set((state) => ({
          notifications: state.notifications.filter((n) => n.id !== id),
        })),
      clearNotifications: () => set({ notifications: [] }),
    }),
    {
      name: 'incident-management-app-state',
      partialize: (state) => ({
        filters: state.filters,
        selectedUpload: state.selectedUpload,
        sidebarOpen: state.sidebarOpen,
      }),
    }
  )
)

// Convenience hooks for specific state slices
export const useFilters = () => {
  const filters = useAppState((state) => state.filters)
  const setFilters = useAppState((state) => state.setFilters)
  const clearFilters = useAppState((state) => state.clearFilters)
  
  return { filters, setFilters, clearFilters }
}

export const useSelectedUpload = () => {
  const selectedUpload = useAppState((state) => state.selectedUpload)
  const setSelectedUpload = useAppState((state) => state.setSelectedUpload)
  
  return { selectedUpload, setSelectedUpload }
}

export const useNotifications = () => {
  const notifications = useAppState((state) => state.notifications)
  const addNotification = useAppState((state) => state.addNotification)
  const removeNotification = useAppState((state) => state.removeNotification)
  const clearNotifications = useAppState((state) => state.clearNotifications)
  
  return { notifications, addNotification, removeNotification, clearNotifications }
}