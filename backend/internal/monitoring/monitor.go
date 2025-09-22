package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"incident-management-system/internal/errors"
	"incident-management-system/internal/logging"
)

// ErrorTracker tracks and monitors errors across the application
type ErrorTracker struct {
	mu           sync.RWMutex
	errors       []ErrorEvent
	metrics      *ErrorMetrics
	logger       *logging.Logger
	maxEvents    int
	alertThresholds *AlertThresholds
}

// ErrorEvent represents a tracked error event
type ErrorEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Error       *errors.APIError       `json:"error"`
	Context     map[string]interface{} `json:"context"`
	Severity    string                 `json:"severity"`
	Component   string                 `json:"component"`
	Operation   string                 `json:"operation"`
	UserID      string                 `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Count       int                    `json:"count"` // For duplicate errors
}

// ErrorMetrics holds error statistics
type ErrorMetrics struct {
	TotalErrors      int64                    `json:"total_errors"`
	ErrorsByCode     map[errors.ErrorCode]int `json:"errors_by_code"`
	ErrorsBySeverity map[string]int           `json:"errors_by_severity"`
	ErrorsByComponent map[string]int          `json:"errors_by_component"`
	LastHourErrors   int                      `json:"last_hour_errors"`
	LastDayErrors    int                      `json:"last_day_errors"`
	ErrorRate        float64                  `json:"error_rate"` // errors per minute
	AvgResolutionTime time.Duration          `json:"avg_resolution_time"`
}

// AlertThresholds defines when to trigger alerts
type AlertThresholds struct {
	ErrorRatePerMinute    float64       `json:"error_rate_per_minute"`
	CriticalErrorsPerHour int           `json:"critical_errors_per_hour"`
	MaxUnresolvedErrors   int           `json:"max_unresolved_errors"`
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"`
}

// PerformanceMetrics tracks system performance
type PerformanceMetrics struct {
	mu                sync.RWMutex
	RequestCount      int64         `json:"request_count"`
	AvgResponseTime   time.Duration `json:"avg_response_time"`
	SlowRequests      int           `json:"slow_requests"`
	DatabaseQueryTime time.Duration `json:"database_query_time"`
	MemoryUsage       uint64        `json:"memory_usage"`
	GoroutineCount    int           `json:"goroutine_count"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// HealthStatus represents the overall system health
type HealthStatus struct {
	Status           string             `json:"status"` // healthy, degraded, unhealthy
	Timestamp        time.Time          `json:"timestamp"`
	ErrorMetrics     *ErrorMetrics      `json:"error_metrics"`
	Performance      *PerformanceMetrics `json:"performance"`
	DatabaseHealth   string             `json:"database_health"`
	ServiceHealth    map[string]string  `json:"service_health"`
	Alerts           []Alert            `json:"alerts"`
	Uptime           time.Duration      `json:"uptime"`
}

// Alert represents a system alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context"`
	Acknowledged bool                  `json:"acknowledged"`
	Resolved    bool                   `json:"resolved"`
}

// DefaultAlertThresholds returns default alert thresholds
func DefaultAlertThresholds() *AlertThresholds {
	return &AlertThresholds{
		ErrorRatePerMinute:    10.0,
		CriticalErrorsPerHour: 5,
		MaxUnresolvedErrors:   50,
		ResponseTimeThreshold: 3 * time.Second,
	}
}

// NewErrorTracker creates a new error tracker
func NewErrorTracker(logger *logging.Logger, maxEvents int) *ErrorTracker {
	if maxEvents <= 0 {
		maxEvents = 1000
	}
	
	return &ErrorTracker{
		errors:          make([]ErrorEvent, 0, maxEvents),
		metrics:         &ErrorMetrics{
			ErrorsByCode:      make(map[errors.ErrorCode]int),
			ErrorsBySeverity:  make(map[string]int),
			ErrorsByComponent: make(map[string]int),
		},
		logger:          logger,
		maxEvents:       maxEvents,
		alertThresholds: DefaultAlertThresholds(),
	}
}

// TrackError tracks a new error event
func (et *ErrorTracker) TrackError(ctx context.Context, err *errors.APIError, component, operation string) {
	et.mu.Lock()
	defer et.mu.Unlock()
	
	// Create error event
	event := ErrorEvent{
		ID:        fmt.Sprintf("err_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Error:     err,
		Severity:  errors.GetErrorSeverity(err),
		Component: component,
		Operation: operation,
		RequestID: logging.GetRequestID(ctx),
		UserID:    logging.GetUserID(ctx),
		Count:     1,
		Context:   make(map[string]interface{}),
	}
	
	// Add stack trace for critical errors
	if event.Severity == "critical" {
		stack := make([]byte, 4096)
		length := runtime.Stack(stack, false)
		event.StackTrace = string(stack[:length])
	}
	
	// Check for duplicate errors (same code, component, operation)
	duplicate := et.findDuplicateError(err.Code, component, operation)
	if duplicate != nil {
		duplicate.Count++
		duplicate.Timestamp = time.Now()
	} else {
		// Add new error event
		et.errors = append(et.errors, event)
		
		// Maintain max events limit
		if len(et.errors) > et.maxEvents {
			et.errors = et.errors[1:]
		}
	}
	
	// Update metrics
	et.updateMetrics(err, event.Severity, component)
	
	// Log the error
	et.logger.WithContext(ctx).WithComponent(component).WithOperation(operation).
		Error("Error tracked", err, 
			logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
				"error_code": err.Code,
				"severity":   event.Severity,
				"event_id":   event.ID,
			}))
	
	// Check for alerts
	et.checkAlerts()
}

// findDuplicateError finds an existing error with the same characteristics
func (et *ErrorTracker) findDuplicateError(code errors.ErrorCode, component, operation string) *ErrorEvent {
	for i := len(et.errors) - 1; i >= 0; i-- {
		event := &et.errors[i]
		if event.Error.Code == code && 
		   event.Component == component && 
		   event.Operation == operation &&
		   !event.Resolved &&
		   time.Since(event.Timestamp) < time.Hour {
			return event
		}
	}
	return nil
}

// updateMetrics updates error metrics
func (et *ErrorTracker) updateMetrics(err *errors.APIError, severity, component string) {
	et.metrics.TotalErrors++
	et.metrics.ErrorsByCode[err.Code]++
	et.metrics.ErrorsBySeverity[severity]++
	et.metrics.ErrorsByComponent[component]++
	
	// Update time-based metrics
	now := time.Now()
	hourAgo := now.Add(-time.Hour)
	dayAgo := now.Add(-24 * time.Hour)
	
	et.metrics.LastHourErrors = et.countErrorsSince(hourAgo)
	et.metrics.LastDayErrors = et.countErrorsSince(dayAgo)
	
	// Calculate error rate (errors per minute in last hour)
	if et.metrics.LastHourErrors > 0 {
		et.metrics.ErrorRate = float64(et.metrics.LastHourErrors) / 60.0
	}
}

// countErrorsSince counts errors since a given time
func (et *ErrorTracker) countErrorsSince(since time.Time) int {
	count := 0
	for _, event := range et.errors {
		if event.Timestamp.After(since) {
			count += event.Count
		}
	}
	return count
}

// checkAlerts checks if any alert thresholds are exceeded
func (et *ErrorTracker) checkAlerts() {
	// Check error rate
	if et.metrics.ErrorRate > et.alertThresholds.ErrorRatePerMinute {
		et.triggerAlert("HIGH_ERROR_RATE", "critical", 
			fmt.Sprintf("Error rate exceeded threshold: %.2f errors/min", et.metrics.ErrorRate),
			map[string]interface{}{
				"current_rate": et.metrics.ErrorRate,
				"threshold":    et.alertThresholds.ErrorRatePerMinute,
			})
	}
	
	// Check critical errors
	criticalErrors := et.metrics.ErrorsBySeverity["critical"]
	if criticalErrors > et.alertThresholds.CriticalErrorsPerHour {
		et.triggerAlert("HIGH_CRITICAL_ERRORS", "critical",
			fmt.Sprintf("Critical errors exceeded threshold: %d errors/hour", criticalErrors),
			map[string]interface{}{
				"current_count": criticalErrors,
				"threshold":     et.alertThresholds.CriticalErrorsPerHour,
			})
	}
	
	// Check unresolved errors
	unresolvedCount := et.countUnresolvedErrors()
	if unresolvedCount > et.alertThresholds.MaxUnresolvedErrors {
		et.triggerAlert("HIGH_UNRESOLVED_ERRORS", "high",
			fmt.Sprintf("Unresolved errors exceeded threshold: %d errors", unresolvedCount),
			map[string]interface{}{
				"current_count": unresolvedCount,
				"threshold":     et.alertThresholds.MaxUnresolvedErrors,
			})
	}
}

// countUnresolvedErrors counts unresolved errors
func (et *ErrorTracker) countUnresolvedErrors() int {
	count := 0
	for _, event := range et.errors {
		if !event.Resolved {
			count++
		}
	}
	return count
}

// triggerAlert triggers a system alert
func (et *ErrorTracker) triggerAlert(alertType, severity, message string, context map[string]interface{}) {
	alert := Alert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Type:      alertType,
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
		Context:   context,
	}
	
	// Log the alert
	et.logger.WithComponent("monitoring").Error("Alert triggered", nil,
		et.logger.WithMetadata(map[string]interface{}{
			"alert_id":   alert.ID,
			"alert_type": alert.Type,
			"severity":   alert.Severity,
			"context":    alert.Context,
		}))
}

// GetMetrics returns current error metrics
func (et *ErrorTracker) GetMetrics() *ErrorMetrics {
	et.mu.RLock()
	defer et.mu.RUnlock()
	
	// Create a copy to avoid race conditions
	metrics := &ErrorMetrics{
		TotalErrors:       et.metrics.TotalErrors,
		ErrorsByCode:      make(map[errors.ErrorCode]int),
		ErrorsBySeverity:  make(map[string]int),
		ErrorsByComponent: make(map[string]int),
		LastHourErrors:    et.metrics.LastHourErrors,
		LastDayErrors:     et.metrics.LastDayErrors,
		ErrorRate:         et.metrics.ErrorRate,
		AvgResolutionTime: et.metrics.AvgResolutionTime,
	}
	
	for k, v := range et.metrics.ErrorsByCode {
		metrics.ErrorsByCode[k] = v
	}
	for k, v := range et.metrics.ErrorsBySeverity {
		metrics.ErrorsBySeverity[k] = v
	}
	for k, v := range et.metrics.ErrorsByComponent {
		metrics.ErrorsByComponent[k] = v
	}
	
	return metrics
}

// GetRecentErrors returns recent error events
func (et *ErrorTracker) GetRecentErrors(limit int) []ErrorEvent {
	et.mu.RLock()
	defer et.mu.RUnlock()
	
	if limit <= 0 || limit > len(et.errors) {
		limit = len(et.errors)
	}
	
	// Return most recent errors
	start := len(et.errors) - limit
	if start < 0 {
		start = 0
	}
	
	events := make([]ErrorEvent, limit)
	copy(events, et.errors[start:])
	
	return events
}

// ResolveError marks an error as resolved
func (et *ErrorTracker) ResolveError(errorID string) bool {
	et.mu.Lock()
	defer et.mu.Unlock()
	
	for i := range et.errors {
		if et.errors[i].ID == errorID {
			et.errors[i].Resolved = true
			now := time.Now()
			et.errors[i].ResolvedAt = &now
			return true
		}
	}
	return false
}

// NewPerformanceMetrics creates a new performance metrics tracker
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		LastUpdated: time.Now(),
	}
}

// UpdatePerformanceMetrics updates performance metrics
func (pm *PerformanceMetrics) UpdatePerformanceMetrics(responseTime time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.RequestCount++
	
	// Update average response time
	if pm.AvgResponseTime == 0 {
		pm.AvgResponseTime = responseTime
	} else {
		pm.AvgResponseTime = (pm.AvgResponseTime + responseTime) / 2
	}
	
	// Track slow requests (> 3 seconds)
	if responseTime > 3*time.Second {
		pm.SlowRequests++
	}
	
	// Update system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	pm.MemoryUsage = m.Alloc
	pm.GoroutineCount = runtime.NumGoroutine()
	pm.LastUpdated = time.Now()
}

// GetPerformanceMetrics returns current performance metrics
func (pm *PerformanceMetrics) GetPerformanceMetrics() *PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	return &PerformanceMetrics{
		RequestCount:      pm.RequestCount,
		AvgResponseTime:   pm.AvgResponseTime,
		SlowRequests:      pm.SlowRequests,
		DatabaseQueryTime: pm.DatabaseQueryTime,
		MemoryUsage:       pm.MemoryUsage,
		GoroutineCount:    pm.GoroutineCount,
		LastUpdated:       pm.LastUpdated,
	}
}

// Global monitoring instances
var (
	globalErrorTracker      *ErrorTracker
	globalPerformanceMetrics *PerformanceMetrics
	startTime               time.Time
)

// InitMonitoring initializes global monitoring
func InitMonitoring(logger *logging.Logger) {
	globalErrorTracker = NewErrorTracker(logger, 1000)
	globalPerformanceMetrics = NewPerformanceMetrics()
	startTime = time.Now()
}

// TrackError tracks an error globally
func TrackError(ctx context.Context, err *errors.APIError, component, operation string) {
	if globalErrorTracker != nil {
		globalErrorTracker.TrackError(ctx, err, component, operation)
	}
}

// UpdatePerformance updates global performance metrics
func UpdatePerformance(responseTime time.Duration) {
	if globalPerformanceMetrics != nil {
		globalPerformanceMetrics.UpdatePerformanceMetrics(responseTime)
	}
}

// GetHealthStatus returns the overall system health status
func GetHealthStatus() *HealthStatus {
	status := &HealthStatus{
		Timestamp:      time.Now(),
		DatabaseHealth: "healthy", // This would be determined by actual health checks
		ServiceHealth:  make(map[string]string),
		Alerts:         []Alert{},
		Uptime:         time.Since(startTime),
	}
	
	if globalErrorTracker != nil {
		status.ErrorMetrics = globalErrorTracker.GetMetrics()
	}
	
	if globalPerformanceMetrics != nil {
		status.Performance = globalPerformanceMetrics.GetPerformanceMetrics()
	}
	
	// Determine overall status
	if status.ErrorMetrics != nil {
		if status.ErrorMetrics.ErrorRate > 5.0 || status.ErrorMetrics.ErrorsBySeverity["critical"] > 0 {
			status.Status = "unhealthy"
		} else if status.ErrorMetrics.ErrorRate > 2.0 || status.ErrorMetrics.LastHourErrors > 10 {
			status.Status = "degraded"
		} else {
			status.Status = "healthy"
		}
	} else {
		status.Status = "healthy"
	}
	
	return status
}

// ExportMetrics exports metrics in JSON format
func ExportMetrics() ([]byte, error) {
	health := GetHealthStatus()
	return json.MarshalIndent(health, "", "  ")
}