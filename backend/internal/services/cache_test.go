package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheService(t *testing.T) {
	// Test creating cache service with default config
	cache, err := NewCacheService(nil)
	require.NoError(t, err)
	assert.NotNil(t, cache)

	// Test creating cache service with custom config
	config := &CacheConfig{
		MaxCost:     50000,
		NumCounters: 500000,
		BufferItems: 32,
		TTL:         10 * time.Minute,
	}

	cache, err = NewCacheService(config)
	require.NoError(t, err)
	assert.NotNil(t, cache)
}

func TestCacheService_SetGet(t *testing.T) {
	cache, err := NewCacheService(nil)
	require.NoError(t, err)

	// Test setting and getting a value
	key := "test_key"
	value := "test_value"
	cost := int64(100)
	ttl := 5 * time.Minute

	success := cache.Set(key, value, cost, ttl)
	assert.True(t, success)

	// Wait a bit for the cache to process the item
	time.Sleep(10 * time.Millisecond)

	// Test getting the value
	retrieved, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Test getting a non-existent key
	_, found = cache.Get("non_existent_key")
	assert.False(t, found)
}

func TestCacheService_Delete(t *testing.T) {
	cache, err := NewCacheService(nil)
	require.NoError(t, err)

	// Set a value
	key := "test_key"
	value := "test_value"
	cache.Set(key, value, 100, 5*time.Minute)

	// Wait a bit for the cache to process the item
	time.Sleep(10 * time.Millisecond)

	// Verify it exists
	_, found := cache.Get(key)
	assert.True(t, found)

	// Delete the value
	cache.Delete(key)

	// Wait a bit for the cache to process the deletion
	time.Sleep(10 * time.Millisecond)

	// Verify it no longer exists
	_, found = cache.Get(key)
	assert.False(t, found)
}

func TestCacheService_Clear(t *testing.T) {
	cache, err := NewCacheService(nil)
	require.NoError(t, err)

	// Set multiple values
	cache.Set("key1", "value1", 100, 5*time.Minute)
	cache.Set("key2", "value2", 100, 5*time.Minute)
	cache.Set("key3", "value3", 100, 5*time.Minute)

	// Wait a bit for the cache to process the items
	time.Sleep(10 * time.Millisecond)

	// Verify they exist
	_, found := cache.Get("key1")
	assert.True(t, found)
	_, found = cache.Get("key2")
	assert.True(t, found)
	_, found = cache.Get("key3")
	assert.True(t, found)

	// Clear the cache
	cache.Clear()

	// Wait a bit for the cache to process the clear
	time.Sleep(10 * time.Millisecond)

	// Verify they no longer exist
	_, found = cache.Get("key1")
	assert.False(t, found)
	_, found = cache.Get("key2")
	assert.False(t, found)
	_, found = cache.Get("key3")
	assert.False(t, found)
}

func TestCachedAnalyticsService(t *testing.T) {
	// Create a mock analytics service (we'll test with a simple struct)
	mockService := &AnalyticsService{}

	// Test creating cached analytics service with default config
	cachedService, err := NewCachedAnalyticsService(mockService, nil)
	require.NoError(t, err)
	assert.NotNil(t, cachedService)
	assert.Equal(t, mockService, cachedService.AnalyticsService)

	// Test creating cached analytics service with custom config
	config := &CacheConfig{
		MaxCost:     10000,
		NumCounters: 100000,
		BufferItems: 16,
		TTL:         2 * time.Minute,
	}

	cachedService, err = NewCachedAnalyticsService(mockService, config)
	require.NoError(t, err)
	assert.NotNil(t, cachedService)
}

func TestBuildCacheKey(t *testing.T) {
	// Test with nil filters
	key := buildCacheKey("test_prefix", nil)
	assert.Equal(t, "test_prefix", key)

	// Test with empty filters
	filters := &TimelineFilters{}
	key = buildCacheKey("test_prefix", filters)
	assert.Equal(t, "test_prefix", key)

	// Test with date filters
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	filters = &TimelineFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	key = buildCacheKey("test_prefix", filters)
	assert.Contains(t, key, "test_prefix")
	assert.Contains(t, key, "start:2024-01-01")
	assert.Contains(t, key, "end:2024-01-31")

	// Test with priority filters
	filters = &TimelineFilters{
		Priorities: []string{"P1", "P2"},
	}

	key = buildCacheKey("test_prefix", filters)
	assert.Contains(t, key, "test_prefix")
	assert.Contains(t, key, "[P1 P2]")

	// Test with application filters
	filters = &TimelineFilters{
		Applications: []string{"App1", "App2"},
	}

	key = buildCacheKey("test_prefix", filters)
	assert.Contains(t, key, "test_prefix")
	assert.Contains(t, key, "[App1 App2]")

	// Test with status filters
	filters = &TimelineFilters{
		Statuses: []string{"Open", "Closed"},
	}

	key = buildCacheKey("test_prefix", filters)
	assert.Contains(t, key, "test_prefix")
	assert.Contains(t, key, "[Open Closed]")

	// Test with all filters
	filters = &TimelineFilters{
		StartDate:    &startDate,
		EndDate:      &endDate,
		Priorities:   []string{"P1", "P2"},
		Applications: []string{"App1", "App2"},
		Statuses:     []string{"Open", "Closed"},
	}

	key = buildCacheKey("test_prefix", filters)
	assert.Contains(t, key, "test_prefix")
	assert.Contains(t, key, "start:2024-01-01")
	assert.Contains(t, key, "end:2024-01-31")
	assert.Contains(t, key, "[P1 P2]")
	assert.Contains(t, key, "[App1 App2]")
	assert.Contains(t, key, "[Open Closed]")
}
