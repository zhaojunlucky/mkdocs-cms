package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// MetricsMiddleware creates a middleware that collects HTTP request metrics
func MetricsMiddleware(ctx *core.APPContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		metricsService := ctx.MustGetService("metrics").(*services.MetricsService)

		if c.Request != nil && c.Request.URL != nil && metricsService.IsMetricsPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()
		method := c.Request.Method
		route := metricsService.InitialRouteLabel(c)
		metricsService.IncInFlight(method, route)
		defer metricsService.ObserveGinRequest(c, start, method, route)

		// Process request
		c.Next()
	}
}
