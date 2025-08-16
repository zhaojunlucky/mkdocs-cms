package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zhaojunlucky/mkdocs-cms/core"
)

type MetricsController struct {
	ctx *core.APPContext
}

// GetMetrics handles the /metrics endpoint for Prometheus scraping
func (mc *MetricsController) GetMetrics(c *gin.Context) {
	// Use the Prometheus HTTP handler to serve metrics
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

func (mc *MetricsController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	mc.ctx = ctx

	router.GET("/metrics", mc.GetMetrics)
}
