package services

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zhaojunlucky/mkdocs-cms/core"
)

type MetricsService struct {
	BaseService

	registry        *prometheus.Registry
	requestDuration *prometheus.HistogramVec
	requestTotal    *prometheus.CounterVec
	inFlight        *prometheus.GaugeVec
}

func (m *MetricsService) Init(ctx *core.APPContext) {
	m.InitService("metrics", ctx, m)
	m.initializeMetrics()
}

func (m *MetricsService) initializeMetrics() {
	m.registry = prometheus.NewRegistry()

	m.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mkdocs_cms_http_requests_seconds",
			Help:    "Duration of HTTP server requests in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "route", "status"},
	)

	m.requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mkdocs_cms_http_requests_total",
			Help: "Total number of HTTP server requests.",
		},
		[]string{"method", "route", "status"},
	)

	m.inFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mkdocs_cms_http_active_requests",
			Help: "Current number of in-flight HTTP server requests.",
		},
		[]string{"method", "route"},
	)

	m.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewBuildInfoCollector(),
		m.requestDuration,
		m.requestTotal,
		m.inFlight,
	)
}

func (m *MetricsService) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *MetricsService) IsMetricsPath(path string) bool {
	return path == "/metrics" || path == "/api/metrics"
}

func (m *MetricsService) InitialRouteLabel(c *gin.Context) string {
	if route := c.FullPath(); route != "" {
		return route
	}
	return "UNKNOWN"
}

func (m *MetricsService) FinalRouteLabel(c *gin.Context, fallback string) string {
	if route := c.FullPath(); route != "" {
		return route
	}
	if c.Writer.Status() == http.StatusNotFound {
		return "NOT_FOUND"
	}
	return fallback
}

func (m *MetricsService) IncInFlight(method, route string) {
	m.inFlight.WithLabelValues(method, route).Inc()
}

func (m *MetricsService) DecInFlight(method, route string) {
	m.inFlight.WithLabelValues(method, route).Dec()
}

func (m *MetricsService) ObserveRequest(method, route string, statusCode int, duration time.Duration) {
	status := strconv.Itoa(statusCode)
	m.requestTotal.WithLabelValues(method, route, status).Inc()
	m.requestDuration.WithLabelValues(method, route, status).Observe(duration.Seconds())
}

func (m *MetricsService) ObserveGinRequest(c *gin.Context, startedAt time.Time, method, route string) {
	finalRoute := m.FinalRouteLabel(c, route)
	m.DecInFlight(method, route)
	m.ObserveRequest(method, finalRoute, c.Writer.Status(), time.Since(startedAt))
}
