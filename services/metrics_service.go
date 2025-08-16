package services

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/zhaojunlucky/mkdocs-cms/core"
)

type MetricsService struct {
	BaseService

	// Counter metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	DatabaseQueriesTotal *prometheus.CounterVec
	ErrorsTotal          *prometheus.CounterVec

	// Gauge metrics
	ActiveConnections prometheus.Gauge
	MemoryUsage       prometheus.Gauge
	ActiveUsers       prometheus.Gauge

	// Histogram metrics
	HTTPRequestDuration *prometheus.HistogramVec
	DatabaseQueryTime   *prometheus.HistogramVec

	// Summary metrics
	ResponseSize   *prometheus.SummaryVec
	ProcessingTime *prometheus.SummaryVec
}

func (m *MetricsService) Init(ctx *core.APPContext) {
	m.InitService("metrics", ctx, m)
	m.initializeMetrics()
}

func (m *MetricsService) initializeMetrics() {
	// Initialize Counter metrics
	m.HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mkdocs_cms",
			Name:      "mkdocs_cms_http_requests_total",
			Help:      "The total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	m.DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mkdocs_cms",
			Name:      "mkdocs_cms_database_queries_total",
			Help:      "The total number of database queries",
		},
		[]string{"operation", "table"},
	)

	m.ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mkdocs_cms",
			Name:      "mkdocs_cms_errors_total",
			Help:      "The total number of errors",
		},
		[]string{"type", "component"},
	)

	// Initialize Gauge metrics
	m.ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "mkdocs_cms",
			Name:      "mkdocs_cms_active_connections",
			Help:      "The number of active connections",
		},
	)

	m.MemoryUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "mkdocs_cms",
			Name:      "mkdocs_cms_memory_usage_bytes",
			Help:      "Current memory usage in bytes",
		},
	)

	m.ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "mkdocs_cms",
			Name:      "mkdocs_cms_active_users",
			Help:      "The number of active users",
		},
	)

	// Initialize Histogram metrics
	m.HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "mkdocs_cms",
			Name:      "mkdocs_cms_http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	m.DatabaseQueryTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "mkdocs_cms",
			Name:      "mkdocs_cms_database_query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "table"},
	)

	// Initialize Summary metrics
	m.ResponseSize = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "mkdocs_cms",
			Name:       "mkdocs_cms_response_size_bytes",
			Help:       "HTTP response size in bytes",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method", "endpoint"},
	)

	m.ProcessingTime = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "mkdocs_cms",
			Name:       "mkdocs_cms_processing_time_seconds",
			Help:       "Request processing time in seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"component", "operation"},
	)
}

// Counter methods
func (m *MetricsService) IncrementHTTPRequests(method, endpoint, statusCode string) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

func (m *MetricsService) IncrementDatabaseQueries(operation, table string) {
	m.DatabaseQueriesTotal.WithLabelValues(operation, table).Inc()
}

func (m *MetricsService) IncrementErrors(errorType, component string) {
	m.ErrorsTotal.WithLabelValues(errorType, component).Inc()
}

// Gauge methods
func (m *MetricsService) SetActiveConnections(count float64) {
	m.ActiveConnections.Set(count)
}

func (m *MetricsService) SetMemoryUsage(bytes float64) {
	m.MemoryUsage.Set(bytes)
}

func (m *MetricsService) SetActiveUsers(count float64) {
	m.ActiveUsers.Set(count)
}

func (m *MetricsService) IncActiveUsers() {
	m.ActiveUsers.Inc()
}

func (m *MetricsService) DecActiveUsers() {
	m.ActiveUsers.Dec()
}

// Histogram methods
func (m *MetricsService) ObserveHTTPRequestDuration(method, endpoint string, duration time.Duration) {
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

func (m *MetricsService) ObserveDatabaseQueryTime(operation, table string, duration time.Duration) {
	m.DatabaseQueryTime.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// Summary methods
func (m *MetricsService) ObserveResponseSize(method, endpoint string, size float64) {
	m.ResponseSize.WithLabelValues(method, endpoint).Observe(size)
}

func (m *MetricsService) ObserveProcessingTime(component, operation string, duration time.Duration) {
	m.ProcessingTime.WithLabelValues(component, operation).Observe(duration.Seconds())
}

// Helper method to get a timer for measuring duration
func (m *MetricsService) NewTimer() *prometheus.Timer {
	return prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		// This is a generic timer, specific metrics should use their own observers
	}))
}
