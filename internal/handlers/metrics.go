package handlers

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// MetricsHandler handles metrics endpoints
type MetricsHandler struct{}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// GetAllMetrics returns all endpoint metrics
func (h *MetricsHandler) GetAllMetrics(c *gin.Context) {
	collector := middleware.GetMetricsCollector()
	metrics := collector.GetMetrics()

	// Convert to slice for sorting
	metricsList := make([]*middleware.EndpointMetrics, 0, len(metrics))
	for _, m := range metrics {
		metricsList = append(metricsList, m)
	}

	// Sort by request count (descending)
	sort.Slice(metricsList, func(i, j int) bool {
		return metricsList[i].RequestCount > metricsList[j].RequestCount
	})

	c.JSON(http.StatusOK, gin.H{
		"endpoints": metricsList,
		"total_endpoints": len(metricsList),
	})
}

// GetMetricsByService returns metrics grouped by service
func (h *MetricsHandler) GetMetricsByService(c *gin.Context) {
	collector := middleware.GetMetricsCollector()
	serviceMetrics := collector.GetMetricsByService()

	// Convert to slice for sorting
	servicesList := make([]*middleware.ServiceMetrics, 0, len(serviceMetrics))
	for _, sm := range serviceMetrics {
		servicesList = append(servicesList, sm)
	}

	// Sort by request count (descending)
	sort.Slice(servicesList, func(i, j int) bool {
		return servicesList[i].RequestCount > servicesList[j].RequestCount
	})

	// Sort endpoints within each service by request count
	for _, sm := range servicesList {
		sort.Slice(sm.Endpoints, func(i, j int) bool {
			return sm.Endpoints[i].RequestCount > sm.Endpoints[j].RequestCount
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"services": servicesList,
		"total_services": len(servicesList),
	})
}

// GetTopServices returns top N services by resource usage
func (h *MetricsHandler) GetTopServices(c *gin.Context) {
	collector := middleware.GetMetricsCollector()
	serviceMetrics := collector.GetMetricsByService()

	// Convert to slice
	servicesList := make([]*middleware.ServiceMetrics, 0, len(serviceMetrics))
	for _, sm := range serviceMetrics {
		servicesList = append(servicesList, sm)
	}

	// Get sort by parameter (default: requests)
	sortBy := c.DefaultQuery("sort_by", "requests")
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		// Try to parse limit, default to 10 if invalid
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Sort based on sort_by parameter
	switch sortBy {
	case "requests":
		sort.Slice(servicesList, func(i, j int) bool {
			return servicesList[i].RequestCount > servicesList[j].RequestCount
		})
	case "duration":
		sort.Slice(servicesList, func(i, j int) bool {
			return servicesList[i].TotalDuration > servicesList[j].TotalDuration
		})
	case "bytes_in":
		sort.Slice(servicesList, func(i, j int) bool {
			return servicesList[i].TotalBytesIn > servicesList[j].TotalBytesIn
		})
	case "bytes_out":
		sort.Slice(servicesList, func(i, j int) bool {
			return servicesList[i].TotalBytesOut > servicesList[j].TotalBytesOut
		})
	case "total_bytes":
		sort.Slice(servicesList, func(i, j int) bool {
			totalI := servicesList[i].TotalBytesIn + servicesList[i].TotalBytesOut
			totalJ := servicesList[j].TotalBytesIn + servicesList[j].TotalBytesOut
			return totalI > totalJ
		})
	default:
		sort.Slice(servicesList, func(i, j int) bool {
			return servicesList[i].RequestCount > servicesList[j].RequestCount
		})
	}

	// Limit results
	if limit > 0 && limit < len(servicesList) {
		servicesList = servicesList[:limit]
	}

	// Calculate summary statistics
	totalRequests := int64(0)
	totalDuration := int64(0)
	totalBytesIn := int64(0)
	totalBytesOut := int64(0)

	for _, sm := range serviceMetrics {
		totalRequests += sm.RequestCount
		totalDuration += sm.TotalDuration
		totalBytesIn += sm.TotalBytesIn
		totalBytesOut += sm.TotalBytesOut
	}

	c.JSON(http.StatusOK, gin.H{
		"top_services": servicesList,
		"summary": gin.H{
			"total_services":     len(serviceMetrics),
			"total_requests":    totalRequests,
			"total_duration_ms": totalDuration,
			"total_bytes_in":    totalBytesIn,
			"total_bytes_out":   totalBytesOut,
			"total_bytes":       totalBytesIn + totalBytesOut,
		},
		"sort_by": sortBy,
		"limit":   limit,
	})
}

// ResetMetrics clears all metrics (admin only - consider adding auth)
func (h *MetricsHandler) ResetMetrics(c *gin.Context) {
	collector := middleware.GetMetricsCollector()
	collector.ResetMetrics()

	c.JSON(http.StatusOK, gin.H{
		"message": "Metrics reset successfully",
	})
}

