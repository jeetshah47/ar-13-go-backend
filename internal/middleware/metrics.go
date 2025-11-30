package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// EndpointMetrics tracks metrics for a specific endpoint
type EndpointMetrics struct {
	Endpoint        string    `json:"endpoint"`
	Service         string    `json:"service"`
	RequestCount    int64     `json:"request_count"`
	TotalDuration   int64     `json:"total_duration_ms"` // in milliseconds
	AvgDuration     float64   `json:"avg_duration_ms"`
	MinDuration     int64     `json:"min_duration_ms"`
	MaxDuration     int64     `json:"max_duration_ms"`
	TotalBytesIn    int64     `json:"total_bytes_in"`
	TotalBytesOut   int64     `json:"total_bytes_out"`
	AvgBytesIn      float64   `json:"avg_bytes_in"`
	AvgBytesOut     float64   `json:"avg_bytes_out"`
	LastRequestTime time.Time `json:"last_request_time"`
	mu              sync.RWMutex
}

// MetricsCollector collects metrics for all endpoints
type MetricsCollector struct {
	metrics map[string]*EndpointMetrics
	mu      sync.RWMutex
}

var globalMetricsCollector = &MetricsCollector{
	metrics: make(map[string]*EndpointMetrics),
}

// GetMetricsCollector returns the global metrics collector
func GetMetricsCollector() *MetricsCollector {
	return globalMetricsCollector
}

// getServiceName extracts service name from endpoint path
func getServiceName(path string) string {
	// Extract service from path like /api/project/all -> project
	// /api/tasks/all/:projectId -> tasks
	// /api/users/all -> users
	parts := splitPath(path)
	if len(parts) >= 2 && parts[0] == "api" {
		return parts[1]
	}
	if len(parts) >= 1 {
		return parts[0]
	}
	return "unknown"
}

// splitPath splits a path into parts
func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, char := range path {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// normalizeEndpoint normalizes endpoint path (removes IDs, etc.)
func normalizeEndpoint(path string) string {
	parts := splitPath(path)
	var normalized []string
	for _, part := range parts {
		// Replace IDs with :id placeholder
		if len(part) == 24 && isHexString(part) {
			// MongoDB ObjectID format
			normalized = append(normalized, ":id")
		} else if isNumeric(part) {
			normalized = append(normalized, ":id")
		} else {
			normalized = append(normalized, part)
		}
	}
	result := ""
	for i, part := range normalized {
		if i > 0 {
			result += "/"
		}
		result += part
	}
	return result
}

// isHexString checks if string is a hex string (MongoDB ObjectID)
func isHexString(s string) bool {
	if len(s) != 24 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// isNumeric checks if string is numeric
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// MetricsMiddleware tracks request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Get request size
		requestSize := c.Request.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}

		// Create a custom response writer to track response size
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
		}
		c.Writer = w

		// Process request
		c.Next()

		// Calculate metrics
		duration := time.Since(start)
		durationMs := duration.Milliseconds()
		
		responseSize := int64(len(w.body))
		if responseSize == 0 {
			// Fallback to Content-Length header if available
			if cl := c.Writer.Header().Get("Content-Length"); cl != "" {
				// Try to parse, but don't fail if it doesn't work
				_ = responseSize // placeholder
			}
		}

		// Normalize endpoint path
		endpoint := normalizeEndpoint(c.Request.URL.Path)
		service := getServiceName(c.Request.URL.Path)

		// Record metrics
		globalMetricsCollector.RecordMetrics(endpoint, service, durationMs, requestSize, responseSize)
	}
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body = append(w.body, []byte(s)...)
	return w.ResponseWriter.WriteString(s)
}

// RecordMetrics records metrics for an endpoint
func (mc *MetricsCollector) RecordMetrics(endpoint, service string, durationMs int64, bytesIn, bytesOut int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := endpoint
	metric, exists := mc.metrics[key]
	if !exists {
		metric = &EndpointMetrics{
			Endpoint:        endpoint,
			Service:         service,
			MinDuration:     durationMs,
			MaxDuration:     durationMs,
			LastRequestTime: time.Now(),
		}
		mc.metrics[key] = metric
	}

	metric.mu.Lock()
	defer metric.mu.Unlock()

	metric.RequestCount++
	metric.TotalDuration += durationMs
	metric.AvgDuration = float64(metric.TotalDuration) / float64(metric.RequestCount)
	
	if durationMs < metric.MinDuration {
		metric.MinDuration = durationMs
	}
	if durationMs > metric.MaxDuration {
		metric.MaxDuration = durationMs
	}

	metric.TotalBytesIn += bytesIn
	metric.TotalBytesOut += bytesOut
	metric.AvgBytesIn = float64(metric.TotalBytesIn) / float64(metric.RequestCount)
	metric.AvgBytesOut = float64(metric.TotalBytesOut) / float64(metric.RequestCount)
	
	metric.LastRequestTime = time.Now()
}

// GetMetrics returns all metrics
func (mc *MetricsCollector) GetMetrics() map[string]*EndpointMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]*EndpointMetrics)
	for k, v := range mc.metrics {
		v.mu.RLock()
		result[k] = &EndpointMetrics{
			Endpoint:        v.Endpoint,
			Service:         v.Service,
			RequestCount:    v.RequestCount,
			TotalDuration:   v.TotalDuration,
			AvgDuration:     v.AvgDuration,
			MinDuration:     v.MinDuration,
			MaxDuration:     v.MaxDuration,
			TotalBytesIn:    v.TotalBytesIn,
			TotalBytesOut:   v.TotalBytesOut,
			AvgBytesIn:      v.AvgBytesIn,
			AvgBytesOut:     v.AvgBytesOut,
			LastRequestTime: v.LastRequestTime,
		}
		v.mu.RUnlock()
	}
	return result
}

// GetMetricsByService returns metrics grouped by service
func (mc *MetricsCollector) GetMetricsByService() map[string]*ServiceMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	serviceMap := make(map[string]*ServiceMetrics)

	for _, metric := range mc.metrics {
		metric.mu.RLock()
		service := metric.Service
		if service == "" {
			service = "unknown"
		}

		serviceMetric, exists := serviceMap[service]
		if !exists {
			serviceMetric = &ServiceMetrics{
				Service:         service,
				RequestCount:    0,
				TotalDuration:   0,
				TotalBytesIn:    0,
				TotalBytesOut:   0,
				Endpoints:       make([]*EndpointMetrics, 0),
			}
			serviceMap[service] = serviceMetric
		}

		serviceMetric.RequestCount += metric.RequestCount
		serviceMetric.TotalDuration += metric.TotalDuration
		serviceMetric.TotalBytesIn += metric.TotalBytesIn
		serviceMetric.TotalBytesOut += metric.TotalBytesOut

		serviceMetric.Endpoints = append(serviceMetric.Endpoints, &EndpointMetrics{
			Endpoint:        metric.Endpoint,
			Service:         metric.Service,
			RequestCount:    metric.RequestCount,
			TotalDuration:   metric.TotalDuration,
			AvgDuration:     metric.AvgDuration,
			MinDuration:     metric.MinDuration,
			MaxDuration:     metric.MaxDuration,
			TotalBytesIn:    metric.TotalBytesIn,
			TotalBytesOut:   metric.TotalBytesOut,
			AvgBytesIn:      metric.AvgBytesIn,
			AvgBytesOut:     metric.AvgBytesOut,
			LastRequestTime: metric.LastRequestTime,
		})
		metric.mu.RUnlock()
	}

	// Calculate averages for services
	for _, serviceMetric := range serviceMap {
		if serviceMetric.RequestCount > 0 {
			serviceMetric.AvgDuration = float64(serviceMetric.TotalDuration) / float64(serviceMetric.RequestCount)
			serviceMetric.AvgBytesIn = float64(serviceMetric.TotalBytesIn) / float64(serviceMetric.RequestCount)
			serviceMetric.AvgBytesOut = float64(serviceMetric.TotalBytesOut) / float64(serviceMetric.RequestCount)
		}
	}

	return serviceMap
}

// ServiceMetrics aggregates metrics by service
type ServiceMetrics struct {
	Service         string             `json:"service"`
	RequestCount    int64              `json:"request_count"`
	TotalDuration   int64              `json:"total_duration_ms"`
	AvgDuration     float64            `json:"avg_duration_ms"`
	TotalBytesIn    int64              `json:"total_bytes_in"`
	TotalBytesOut   int64              `json:"total_bytes_out"`
	AvgBytesIn      float64            `json:"avg_bytes_in"`
	AvgBytesOut     float64            `json:"avg_bytes_out"`
	Endpoints       []*EndpointMetrics `json:"endpoints"`
}

// ResetMetrics clears all metrics (useful for testing or periodic resets)
func (mc *MetricsCollector) ResetMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.metrics = make(map[string]*EndpointMetrics)
}

