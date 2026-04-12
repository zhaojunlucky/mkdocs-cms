package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

type MetricsController struct {
	ctx *core.APPContext
}

// GetMetrics handles the /metrics endpoint for Prometheus scraping
func (mc *MetricsController) GetMetrics(c *gin.Context) {
	metricsService := mc.ctx.MustGetService("metrics").(*services.MetricsService)
	metricsService.Handler().ServeHTTP(c.Writer, c.Request)
}

func (mc *MetricsController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	mc.ctx = ctx

	router.GET("/metrics", mc.GetMetrics)
}
