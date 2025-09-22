import { useState, useEffect, useCallback, useRef } from 'react'

interface PerformanceMetrics {
  memoryUsage: number
  cpuUsage: number
  fps: number
  renderTime: number
  bundleSize: number
}

interface OptimizationConfig {
  enableVirtualScrolling: boolean
  enableCodeSplitting: boolean
  enableImageOptimization: boolean
  enableCaching: boolean
  maxDataPoints: number
}

interface MemoryInfo {
  usedJSHeapSize: number
  totalJSHeapSize: number
  jsHeapSizeLimit: number
}

export function usePerformanceOptimization(config: Partial<OptimizationConfig> = {}) {
  const [metrics, setMetrics] = useState<PerformanceMetrics>({
    memoryUsage: 0,
    cpuUsage: 0,
    fps: 60,
    renderTime: 0,
    bundleSize: 0
  })
  
  const [isOptimized, setIsOptimized] = useState(false)
  const [optimizationSuggestions, setOptimizationSuggestions] = useState<string[]>([])
  const animationFrameRef = useRef<number>()
  const lastFrameTimeRef = useRef<number>(0)
  const frameCountRef = useRef<number>(0)
  const lastFpsUpdateRef = useRef<number>(0)

  const defaultConfig: OptimizationConfig = {
    enableVirtualScrolling: true,
    enableCodeSplitting: true,
    enableImageOptimization: true,
    enableCaching: true,
    maxDataPoints: 1000
  }

  const finalConfig = { ...defaultConfig, ...config }

  // Calculate FPS
  const calculateFps = useCallback(() => {
    const now = performance.now()
    frameCountRef.current++
    
    if (now >= lastFpsUpdateRef.current + 1000) {
      const fps = Math.round((frameCountRef.current * 1000) / (now - lastFpsUpdateRef.current))
      setMetrics(prev => ({ ...prev, fps }))
      frameCountRef.current = 0
      lastFpsUpdateRef.current = now
    }
    
    animationFrameRef.current = requestAnimationFrame(calculateFps)
  }, [])

  // Monitor performance metrics
  useEffect(() => {
    // Start FPS monitoring
    lastFpsUpdateRef.current = performance.now()
    animationFrameRef.current = requestAnimationFrame(calculateFps)
    
    // Memory usage monitoring (if available)
    const monitorMemory = () => {
      // @ts-ignore - performance.memory is non-standard
      if (performance.memory) {
        // @ts-ignore - performance.memory is non-standard
        const memory: MemoryInfo = performance.memory
        setMetrics(prev => ({
          ...prev,
          memoryUsage: Math.round(memory.usedJSHeapSize / 1048576), // MB
          cpuUsage: Math.min(100, Math.round((memory.totalJSHeapSize / memory.jsHeapSizeLimit) * 100))
        }))
      }
    }
    
    const memoryInterval = setInterval(monitorMemory, 2000)
    monitorMemory()
    
    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current)
      }
      clearInterval(memoryInterval)
    }
  }, [calculateFps])

  // Generate optimization suggestions
  useEffect(() => {
    const suggestions: string[] = []
    
    if (metrics.fps < 30) {
      suggestions.push('Frame rate is below 30 FPS. Consider reducing DOM complexity or using virtual scrolling.')
    }
    
    if (metrics.memoryUsage > 100) {
      suggestions.push('High memory usage detected. Consider implementing data pagination or virtualization.')
    }
    
    if (!finalConfig.enableVirtualScrolling && metrics.renderTime > 16) {
      suggestions.push('Enable virtual scrolling for large data sets to improve render performance.')
    }
    
    if (!finalConfig.enableCodeSplitting) {
      suggestions.push('Enable code splitting to reduce initial bundle size and improve load times.')
    }
    
    if (!finalConfig.enableCaching) {
      suggestions.push('Enable caching for API responses to reduce network requests.')
    }
    
    setOptimizationSuggestions(suggestions)
    setIsOptimized(suggestions.length === 0)
  }, [metrics, finalConfig])

  // Performance optimization utilities
  const optimizeDataRendering = useCallback(<T,>(data: T[], maxItems: number = finalConfig.maxDataPoints): T[] => {
    if (data.length <= maxItems) {
      return data
    }
    
    // Sample data instead of truncating
    const step = Math.ceil(data.length / maxItems)
    const sampledData: T[] = []
    
    for (let i = 0; i < data.length; i += step) {
      sampledData.push(data[i])
    }
    
    return sampledData
  }, [finalConfig.maxDataPoints])

  const debouncedFunction = useCallback(<T extends (...args: any[]) => any>(func: T, delay: number): T => {
    let timeoutId: ReturnType<typeof setTimeout>
    
    return function (...args: Parameters<T>) {
      clearTimeout(timeoutId)
      timeoutId = setTimeout(() => func(...args), delay)
    } as T
  }, [])

  const throttleFunction = useCallback(<T extends (...args: any[]) => any>(func: T, limit: number): T => {
    let inThrottle: boolean
    
    return function (...args: Parameters<T>) {
      if (!inThrottle) {
        func(...args)
        inThrottle = true
        setTimeout(() => inThrottle = false, limit)
      }
    } as T
  }, [])

  // Memoization utility
  const createMemoizedSelector = useCallback(<T, R>(selector: (state: T) => R): ((state: T) => R) => {
    let lastState: T | null = null
    let lastResult: R | null = null
    
    return (state: T): R => {
      if (lastState === state) {
        return lastResult as R
      }
      
      const result = selector(state)
      lastState = state
      lastResult = result
      return result
    }
  }, [])

  return {
    metrics,
    isOptimized,
    optimizationSuggestions,
    optimizeDataRendering,
    debouncedFunction,
    throttleFunction,
    createMemoizedSelector,
    config: finalConfig
  }
}