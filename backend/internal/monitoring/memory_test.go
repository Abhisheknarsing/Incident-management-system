package monitoring

import (
	"testing"
	"time"

	"incident-management-system/internal/logging"
)

func TestMemoryMonitor_NewMemoryMonitor(t *testing.T) {
	logger, err := logging.NewLogger(&logging.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test creating monitor with nil config (should use defaults)
	monitor := NewMemoryMonitor(logger, nil)
	if monitor == nil {
		t.Fatal("Expected non-nil MemoryMonitor")
	}

	if monitor.logger == nil {
		t.Error("Expected logger to be set")
	}

	if monitor.collectionInterval != 30*time.Second {
		t.Errorf("Expected default collection interval 30s, got %v", monitor.collectionInterval)
	}

	// Test creating monitor with custom config
	config := &MemoryConfig{
		CollectionInterval: 10 * time.Second,
		EnableProfiling:    true,
	}

	monitor = NewMemoryMonitor(logger, config)
	if monitor == nil {
		t.Fatal("Expected non-nil MemoryMonitor")
	}

	if monitor.collectionInterval != 10*time.Second {
		t.Errorf("Expected custom collection interval 10s, got %v", monitor.collectionInterval)
	}
}

func TestMemoryMonitor_GetStats(t *testing.T) {
	logger, err := logging.NewLogger(&logging.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	monitor := NewMemoryMonitor(logger, nil)
	if monitor == nil {
		t.Fatal("Expected non-nil MemoryMonitor")
	}

	stats := monitor.GetStats()
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	// Stats should be initialized (timestamp may or may not be set depending on when collection happens)
	// Just check that we got a valid stats object
	t.Logf("Got stats with timestamp: %v", stats.Timestamp)
}

func TestMemoryMonitor_GetMemoryUsage(t *testing.T) {
	logger, err := logging.NewLogger(&logging.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	monitor := NewMemoryMonitor(logger, nil)
	if monitor == nil {
		t.Fatal("Expected non-nil MemoryMonitor")
	}

	usage := monitor.GetMemoryUsage()
	if usage == nil {
		t.Fatal("Expected non-nil usage map")
	}

	// Check that expected keys exist
	expectedKeys := []string{
		"alloc_mb",
		"heap_alloc_mb",
		"sys_mb",
		"heap_objects",
		"gc_cpu_fraction",
		"num_gc",
		"timestamp",
	}

	for _, key := range expectedKeys {
		if _, exists := usage[key]; !exists {
			t.Errorf("Expected key %s in memory usage map", key)
		}
	}
}

func TestMemoryMonitor_IsMemoryUsageHigh(t *testing.T) {
	logger, err := logging.NewLogger(&logging.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	monitor := NewMemoryMonitor(logger, nil)
	if monitor == nil {
		t.Fatal("Expected non-nil MemoryMonitor")
	}

	// Test with low threshold (should return false)
	if monitor.IsMemoryUsageHigh(1000.0) {
		t.Error("Expected memory usage to not be high with 1000MB threshold")
	}

	// Test with negative threshold (should return true)
	if !monitor.IsMemoryUsageHigh(-1.0) {
		t.Error("Expected memory usage to be high with negative threshold")
	}
}

func TestMemoryMonitor_SuggestOptimizations(t *testing.T) {
	logger, err := logging.NewLogger(&logging.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	monitor := NewMemoryMonitor(logger, nil)
	if monitor == nil {
		t.Fatal("Expected non-nil MemoryMonitor")
	}

	suggestions := monitor.SuggestOptimizations()
	if len(suggestions) == 0 {
		t.Error("Expected at least one suggestion")
	}

	// Should have default "normal range" suggestion
	found := false
	for _, suggestion := range suggestions {
		if suggestion == "Memory usage is within normal range." {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected 'normal range' suggestion")
	}
}

func TestMemoryMonitor_StartStop(t *testing.T) {
	logger, err := logging.NewLogger(&logging.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	config := &MemoryConfig{
		CollectionInterval: 100 * time.Millisecond, // Fast interval for testing
	}

	monitor := NewMemoryMonitor(logger, config)
	if monitor == nil {
		t.Fatal("Expected non-nil MemoryMonitor")
	}

	// Start monitoring
	monitor.Start()

	// Give it time to collect stats
	time.Sleep(200 * time.Millisecond)

	// Get stats after collection
	stats := monitor.GetStats()
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	// Stats should have been updated
	if stats.Timestamp.IsZero() {
		t.Error("Expected timestamp to be updated")
	}

	// Stop monitoring
	monitor.Stop()

	// Start again (should not panic)
	monitor.Start()
	monitor.Stop()
}
