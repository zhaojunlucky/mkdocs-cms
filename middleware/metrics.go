package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// MetricsMiddleware creates a middleware that collects HTTP request metrics
func MetricsMiddleware(ctx *core.APPContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get metrics service from context
		metricsService := ctx.ServiceMap["metrics"].(*services.MetricsService)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get response information
		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}
		statusCode := strconv.Itoa(c.Writer.Status())
		responseSize := float64(c.Writer.Size())

		// Record metrics
		metricsService.IncrementHTTPRequests(method, endpoint, statusCode)
		metricsService.ObserveHTTPRequestDuration(method, endpoint, duration)
		
		if responseSize > 0 {
			metricsService.ObserveResponseSize(method, endpoint, responseSize)
		}

		// Record processing time
		metricsService.ObserveProcessingTime("http", "request", duration)
	}
}
