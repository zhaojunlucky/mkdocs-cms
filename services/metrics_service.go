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
	"github.com/zhaojunlucky/mkdocs-cms/env"
)

const metricsPath = "/metrics"
const apiMetricsPath = "/api/metrics"

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
			Namespace: "go_gin",
			Subsystem: "http_server",
			Name:      "requests_seconds",
			Help:      "Duration of HTTP server requests in seconds.",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "route", "status"},
	)

	m.requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "go_gin",
			Subsystem: "http_server",
			Name:      "requests_total",
			Help:      "Total number of HTTP server requests.",
		},
		[]string{"method", "route", "status"},
	)

	m.inFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "go_gin",
			Subsystem: "http_server",
			Name:      "active_requests",
			Help:      "Current number of in-flight HTTP server requests.",
		},
		[]string{"method", "route"},
	)

	appInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "go_gin",
			Name:      "app_info",
			Help:      "Static information about the application instance.",
		},
		[]string{"env"},
	)

	m.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewBuildInfoCollector(),
		m.requestDuration,
		m.requestTotal,
		m.inFlight,
		appInfo,
	)
	appInfo.WithLabelValues(m.envLabel()).Set(1)
}

func (m *MetricsService) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *MetricsService) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request != nil && c.Request.URL != nil && isMetricsPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		method := c.Request.Method
		route := initialMetricsRouteLabel(c)
		m.inFlight.WithLabelValues(method, route).Inc()
		start := time.Now()

		defer func() {
			m.inFlight.WithLabelValues(method, route).Dec()

			status := strconv.Itoa(c.Writer.Status())
			finalRoute := finalMetricsRouteLabel(c, route)
			m.requestTotal.WithLabelValues(method, finalRoute, status).Inc()
			m.requestDuration.WithLabelValues(method, finalRoute, status).Observe(time.Since(start).Seconds())
		}()

		c.Next()
	}
}

func (m *MetricsService) envLabel() string {
	if env.IsProduction {
		return "production"
	}
	return "development"
}

func isMetricsPath(path string) bool {
	return path == metricsPath || path == apiMetricsPath
}

func initialMetricsRouteLabel(c *gin.Context) string {
	if route := c.FullPath(); route != "" {
		return route
	}
	return "UNKNOWN"
}

func finalMetricsRouteLabel(c *gin.Context, fallback string) string {
	if route := c.FullPath(); route != "" {
		return route
	}
	if c.Writer.Status() == http.StatusNotFound {
		return "NOT_FOUND"
	}
	return fallback
}
