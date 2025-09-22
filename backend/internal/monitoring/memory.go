package monitoring

import (
	"runtime"
	"sync"
	"time"

	"incident-management-system/internal/logging"
)

// MemoryMonitor tracks memory usage and provides optimization utilities
type MemoryMonitor struct {
	mu                 sync.RWMutex
	stats              *MemStats
	logger             *logging.Logger
	monitoring         bool
	stopChan           chan struct{}
	collectionInterval time.Duration
}

// MemStats holds memory statistics
type MemStats struct {
	Alloc         uint64      `json:"alloc"`           // bytes allocated and not yet freed
	TotalAlloc    uint64      `json:"total_alloc"`     // bytes allocated (even if freed)
	Sys           uint64      `json:"sys"`             // bytes obtained from system
	Lookups       uint64      `json:"lookups"`         // number of pointer lookups
	Mallocs       uint64      `json:"mallocs"`         // number of mallocs
	Frees         uint64      `json:"frees"`           // number of frees
	HeapAlloc     uint64      `json:"heap_alloc"`      // bytes allocated and not yet freed (same as Alloc)
	HeapSys       uint64      `json:"heap_sys"`        // bytes obtained from system
	HeapIdle      uint64      `json:"heap_idle"`       // bytes in idle spans
	HeapInuse     uint64      `json:"heap_inuse"`      // bytes in non-idle spans
	HeapReleased  uint64      `json:"heap_released"`   // bytes released to the OS
	HeapObjects   uint64      `json:"heap_objects"`    // total number of allocated objects
	StackInuse    uint64      `json:"stack_inuse"`     // bytes used by stack allocator
	StackSys      uint64      `json:"stack_sys"`       // bytes obtained from system for stack allocator
	MSpanInuse    uint64      `json:"mspan_inuse"`     // bytes used for mspan structures
	MSpanSys      uint64      `json:"mspan_sys"`       // bytes used for mspan structures obtained from system
	MCacheInuse   uint64      `json:"mcache_inuse"`    // bytes used for mcache structures
	MCacheSys     uint64      `json:"mcache_sys"`      // bytes used for mcache structures obtained from system
	BuckHashSys   uint64      `json:"buck_hash_sys"`   // bytes used by profiling bucket hash table
	GCSys         uint64      `json:"gc_sys"`          // bytes used for garbage collection system metadata
	OtherSys      uint64      `json:"other_sys"`       // bytes used for other system allocations
	NextGC        uint64      `json:"next_gc"`         // next collection will happen when HeapAlloc > this amount
	LastGC        uint64      `json:"last_gc"`         // end time of last collection (nanoseconds since 1970)
	PauseTotalNs  uint64      `json:"pause_total_ns"`  // total nanoseconds in GC stop-the-world pauses
	PauseNs       [256]uint64 `json:"pause_ns"`        // circular buffer of recent GC pause durations
	NumGC         uint32      `json:"num_gc"`          // number of garbage collections
	NumForcedGC   uint32      `json:"num_forced_gc"`   // number of forced collections
	GCCPUFraction float64     `json:"gc_cpu_fraction"` // fraction of CPU time used by GC
	EnableGC      bool        `json:"enable_gc"`       // whether GC is enabled
	DebugGC       bool        `json:"debug_gc"`        // whether to print debug info about GC
	Timestamp     time.Time   `json:"timestamp"`       // when these stats were collected
}

// MemoryConfig holds memory monitoring configuration
type MemoryConfig struct {
	CollectionInterval time.Duration // How often to collect stats
	EnableProfiling    bool          // Whether to enable memory profiling
}

// DefaultMemoryConfig returns default memory monitoring configuration
func DefaultMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		CollectionInterval: 30 * time.Second,
		EnableProfiling:    false,
	}
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(logger *logging.Logger, config *MemoryConfig) *MemoryMonitor {
	if config == nil {
		config = DefaultMemoryConfig()
	}

	return &MemoryMonitor{
		logger:             logger.WithComponent("memory_monitor"),
		collectionInterval: config.CollectionInterval,
		stats:              &MemStats{},
	}
}

// Start begins memory monitoring
func (m *MemoryMonitor) Start() {
	m.mu.Lock()
	if m.monitoring {
		m.mu.Unlock()
		return
	}
	m.monitoring = true
	m.stopChan = make(chan struct{})
	m.mu.Unlock()

	go m.collectStats()
	m.logger.Info("Memory monitoring started")
}

// Stop stops memory monitoring
func (m *MemoryMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.monitoring {
		return
	}

	m.monitoring = false
	close(m.stopChan)
	m.logger.Info("Memory monitoring stopped")
}

// collectStats periodically collects memory statistics
func (m *MemoryMonitor) collectStats() {
	ticker := time.NewTicker(m.collectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.updateStats()
		case <-m.stopChan:
			return
		}
	}
}

// updateStats updates the current memory statistics
func (m *MemoryMonitor) updateStats() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	m.mu.Lock()
	m.stats = &MemStats{
		Alloc:         ms.Alloc,
		TotalAlloc:    ms.TotalAlloc,
		Sys:           ms.Sys,
		Lookups:       ms.Lookups,
		Mallocs:       ms.Mallocs,
		Frees:         ms.Frees,
		HeapAlloc:     ms.HeapAlloc,
		HeapSys:       ms.HeapSys,
		HeapIdle:      ms.HeapIdle,
		HeapInuse:     ms.HeapInuse,
		HeapReleased:  ms.HeapReleased,
		HeapObjects:   ms.HeapObjects,
		StackInuse:    ms.StackInuse,
		StackSys:      ms.StackSys,
		MSpanInuse:    ms.MSpanInuse,
		MSpanSys:      ms.MSpanSys,
		MCacheInuse:   ms.MCacheInuse,
		MCacheSys:     ms.MCacheSys,
		BuckHashSys:   ms.BuckHashSys,
		GCSys:         ms.GCSys,
		OtherSys:      ms.OtherSys,
		NextGC:        ms.NextGC,
		LastGC:        ms.LastGC,
		PauseTotalNs:  ms.PauseTotalNs,
		PauseNs:       ms.PauseNs,
		NumGC:         ms.NumGC,
		NumForcedGC:   ms.NumForcedGC,
		GCCPUFraction: ms.GCCPUFraction,
		EnableGC:      ms.EnableGC,
		DebugGC:       ms.DebugGC,
		Timestamp:     time.Now(),
	}
	m.mu.Unlock()

	// Log if memory usage is high
	if ms.Alloc > 100*1024*1024 { // 100MB
		m.logger.Warn("High memory usage detected",
			nil,
			m.logger.WithMetadata(map[string]interface{}{
				"alloc_mb":      float64(ms.Alloc) / 1024 / 1024,
				"heap_alloc_mb": float64(ms.HeapAlloc) / 1024 / 1024,
				"sys_mb":        float64(ms.Sys) / 1024 / 1024,
			}))
	}
}

// GetStats returns current memory statistics
func (m *MemoryMonitor) GetStats() *MemStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	stats := *m.stats
	return &stats
}

// ForceGC forces garbage collection
func (m *MemoryMonitor) ForceGC() {
	m.logger.Info("Forcing garbage collection")
	runtime.GC()
	m.updateStats()
}

// GetMemoryUsage returns current memory usage in a simplified format
func (m *MemoryMonitor) GetMemoryUsage() map[string]interface{} {
	m.mu.RLock()
	stats := m.stats
	m.mu.RUnlock()

	return map[string]interface{}{
		"alloc_mb":        float64(stats.Alloc) / 1024 / 1024,
		"heap_alloc_mb":   float64(stats.HeapAlloc) / 1024 / 1024,
		"sys_mb":          float64(stats.Sys) / 1024 / 1024,
		"heap_objects":    stats.HeapObjects,
		"gc_cpu_fraction": stats.GCCPUFraction,
		"num_gc":          stats.NumGC,
		"timestamp":       stats.Timestamp,
	}
}

// IsMemoryUsageHigh checks if memory usage is above threshold
func (m *MemoryMonitor) IsMemoryUsageHigh(thresholdMB float64) bool {
	m.mu.RLock()
	allocMB := float64(m.stats.Alloc) / 1024 / 1024
	m.mu.RUnlock()

	return allocMB > thresholdMB
}

// SuggestOptimizations provides memory optimization suggestions based on current usage
func (m *MemoryMonitor) SuggestOptimizations() []string {
	var suggestions []string

	m.mu.RLock()
	stats := m.stats
	m.mu.RUnlock()

	allocMB := float64(stats.Alloc) / 1024 / 1024
	heapAllocMB := float64(stats.HeapAlloc) / 1024 / 1024
	gcFraction := stats.GCCPUFraction

	// High memory usage suggestions
	if allocMB > 200 {
		suggestions = append(suggestions, "Memory usage is high. Consider optimizing data structures or implementing pagination.")
	}

	if heapAllocMB > 150 {
		suggestions = append(suggestions, "Heap allocation is high. Check for memory leaks or consider more efficient data processing.")
	}

	// GC pressure suggestions
	if gcFraction > 0.25 {
		suggestions = append(suggestions, "GC is consuming significant CPU. Consider reducing allocations or tuning GC settings.")
	}

	// Low memory suggestions
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Memory usage is within normal range.")
	}

	return suggestions
}
