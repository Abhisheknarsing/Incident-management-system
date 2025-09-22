package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
)

// CacheService provides caching functionality for analytics data
type CacheService struct {
	cache *ristretto.Cache
	mu    sync.RWMutex
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	MaxCost     int64
	NumCounters int64
	BufferItems int64
	TTL         time.Duration
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxCost:     100000,     // Max cost of cache
		NumCounters: 1000000,    // Number of keys to track frequency of
		BufferItems: 64,         // Number of keys per Get buffer
		TTL:         5 * time.Minute, // Default TTL of 5 minutes
	}
}

// NewCacheService creates a new cache service
func NewCacheService(config *CacheConfig) (*CacheService, error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: config.NumCounters,
		MaxCost:     config.MaxCost,
		BufferItems: config.BufferItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return &CacheService{
		cache: cache,
	}, nil
}

// Get retrieves a value from cache
func (c *CacheService) Get(key string) (interface{}, bool) {
	value, found := c.cache.Get(key)
	if !found {
		return nil, false
	}
	return value, true
}

// Set stores a value in cache with TTL
func (c *CacheService) Set(key string, value interface{}, cost int64, ttl time.Duration) bool {
	return c.cache.SetWithTTL(key, value, cost, ttl)
}

// Delete removes a value from cache
func (c *CacheService) Delete(key string) {
	c.cache.Del(key)
}

// Clear clears the entire cache
func (c *CacheService) Clear() {
	c.cache.Clear()
}

// Stats returns cache statistics
func (c *CacheService) Stats() *ristretto.Metrics {
	return c.cache.Metrics
}

// Close closes the cache
func (c *CacheService) Close() {
	c.cache.Close()
}

// CachedAnalyticsService wraps AnalyticsService with caching functionality
type CachedAnalyticsService struct {
	*AnalyticsService
	cache *CacheService
}

// NewCachedAnalyticsService creates a new cached analytics service
func NewCachedAnalyticsService(analyticsService *AnalyticsService, cacheConfig *CacheConfig) (*CachedAnalyticsService, error) {
	cache, err := NewCacheService(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache service: %w", err)
	}

	return &CachedAnalyticsService{
		AnalyticsService: analyticsService,
		cache:           cache,
	}, nil
}

// buildCacheKey creates a cache key from filters
func buildCacheKey(prefix string, filters *TimelineFilters) string {
	if filters == nil {
		return prefix
	}

	key := prefix
	if filters.StartDate != nil {
		key += fmt.Sprintf("_start:%s", filters.StartDate.Format("2006-01-02"))
	}
	if filters.EndDate != nil {
		key += fmt.Sprintf("_end:%s", filters.EndDate.Format("2006-01-02"))
	}
	if len(filters.Priorities) > 0 {
		key += fmt.Sprintf("_prios:%v", filters.Priorities)
	}
	if len(filters.Applications) > 0 {
		key += fmt.Sprintf("_apps:%v", filters.Applications)
	}
	if len(filters.Statuses) > 0 {
		key += fmt.Sprintf("_statuses:%v", filters.Statuses)
	}

	return key
}

// getCachedOrFetch retrieves data from cache or fetches it
func (s *CachedAnalyticsService) getCachedOrFetch(ctx context.Context, key string, fetchFunc func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if cached, found := s.cache.Get(key); found {
		return cached, nil
	}

	// Fetch from source
	data, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	// Store in cache
	jsonData, _ := json.Marshal(data)
	s.cache.Set(key, data, int64(len(jsonData)), 5*time.Minute)

	return data, nil
}

// GetDailyTimeline returns cached daily incident timeline data
func (s *CachedAnalyticsService) GetDailyTimeline(ctx context.Context, filters *TimelineFilters) ([]TimelineData, error) {
	key := buildCacheKey("daily_timeline", filters)
	
	result, err := s.getCachedOrFetch(ctx, key, func() (interface{}, error) {
		return s.AnalyticsService.GetDailyTimeline(ctx, filters)
	})
	if err != nil {
		return nil, err
	}
	
	return result.([]TimelineData), nil
}

// GetWeeklyTimeline returns cached weekly incident timeline data
func (s *CachedAnalyticsService) GetWeeklyTimeline(ctx context.Context, filters *TimelineFilters) ([]TimelineData, error) {
	key := buildCacheKey("weekly_timeline", filters)
	
	result, err := s.getCachedOrFetch(ctx, key, func() (interface{}, error) {
		return s.AnalyticsService.GetWeeklyTimeline(ctx, filters)
	})
	if err != nil {
		return nil, err
	}
	
	return result.([]TimelineData), nil
}

// GetTrendAnalysis returns cached trend analysis data
func (s *CachedAnalyticsService) GetTrendAnalysis(ctx context.Context, period string, filters *TimelineFilters) ([]TrendAnalysis, error) {
	key := buildCacheKey(fmt.Sprintf("trend_analysis_%s", period), filters)
	
	result, err := s.getCachedOrFetch(ctx, key, func() (interface{}, error) {
		return s.AnalyticsService.GetTrendAnalysis(ctx, period, filters)
	})
	if err != nil {
		return nil, err
	}
	
	return result.([]TrendAnalysis), nil
}

// GetPriorityAnalysis returns cached priority analysis data
func (s *CachedAnalyticsService) GetPriorityAnalysis(ctx context.Context, filters *TimelineFilters) ([]PriorityAnalysis, error) {
	key := buildCacheKey("priority_analysis", filters)
	
	result, err := s.getCachedOrFetch(ctx, key, func() (interface{}, error) {
		return s.AnalyticsService.GetPriorityAnalysis(ctx, filters)
	})
	if err != nil {
		return nil, err
	}
	
	return result.([]PriorityAnalysis), nil
}

// GetApplicationAnalysis returns cached application analysis data
func (s *CachedAnalyticsService) GetApplicationAnalysis(ctx context.Context, filters *TimelineFilters) ([]ApplicationAnalysis, error) {
	key := buildCacheKey("application_analysis", filters)
	
	result, err := s.getCachedOrFetch(ctx, key, func() (interface{}, error) {
		return s.AnalyticsService.GetApplicationAnalysis(ctx, filters)
	})
	if err != nil {
		return nil, err
	}
	
	return result.([]ApplicationAnalysis), nil
}

// GetSentimentAnalysis returns cached sentiment analysis data
func (s *CachedAnalyticsService) GetSentimentAnalysis(ctx context.Context, filters *TimelineFilters) ([]SentimentAnalysis, error) {
	key := buildCacheKey("sentiment_analysis", filters)
	
	result, err := s.getCachedOrFetch(ctx, key, func() (interface{}, error) {
		return s.AnalyticsService.GetSentimentAnalysis(ctx, filters)
	})
	if err != nil {
		return nil, err
	}
	
	return result.([]SentimentAnalysis), nil
}

// GetAutomationAnalysis returns cached automation analysis data
func (s *CachedAnalyticsService) GetAutomationAnalysis(ctx context.Context, filters *TimelineFilters) ([]AutomationAnalysis, error) {
	key := buildCacheKey("automation_analysis", filters)
	
	result, err := s.getCachedOrFetch(ctx, key, func() (interface{}, error) {
		return s.AnalyticsService.GetAutomationAnalysis(ctx, filters)
	})
	if err != nil {
		return nil, err
	}
	
	return result.([]AutomationAnalysis), nil
}

// GetAnalyticsSummary returns cached analytics summary
func (s *CachedAnalyticsService) GetAnalyticsSummary(ctx context.Context, filters *TimelineFilters) (*AnalyticsSummary, error) {
	key := buildCacheKey("analytics_summary", filters)
	
	result, err := s.getCachedOrFetch(ctx, key, func() (interface{}, error) {
		return s.AnalyticsService.GetAnalyticsSummary(ctx, filters)
	})
	if err != nil {
		return nil, err
	}
	
	return result.(*AnalyticsSummary), nil
}

// InvalidateCache invalidates cache entries for a specific filter set
func (s *CachedAnalyticsService) InvalidateCache(filters *TimelineFilters) {
	// Invalidate all cache entries related to these filters
	keys := []string{
		buildCacheKey("daily_timeline", filters),
		buildCacheKey("weekly_timeline", filters),
		buildCacheKey("trend_analysis_daily", filters),
		buildCacheKey("trend_analysis_weekly", filters),
		buildCacheKey("priority_analysis", filters),
		buildCacheKey("application_analysis", filters),
		buildCacheKey("sentiment_analysis", filters),
		buildCacheKey("automation_analysis", filters),
		buildCacheKey("analytics_summary", filters),
	}
	
	for _, key := range keys {
		s.cache.Delete(key)
	}
}

// ClearCache clears the entire cache
func (s *CachedAnalyticsService) ClearCache() {
	s.cache.Clear()
}