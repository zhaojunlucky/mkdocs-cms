package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// MetricsMiddleware creates a middleware that collects HTTP request metrics.
// The service is resolved lazily on first request to avoid initialization order issues.
func MetricsMiddleware(ctx *core.APPContext) gin.HandlerFunc {
	var svc *services.MetricsService
	return func(c *gin.Context) {
		if svc == nil {
			svc = ctx.MustGetService("metrics").(*services.MetricsService)
		}
		svc.Middleware()(c)
	}
}
