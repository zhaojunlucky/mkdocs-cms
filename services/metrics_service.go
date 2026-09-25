package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/golib/pkg/ginmetrics"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/env"
)

// appName is the "application" label applied to every metric, mirroring
// Spring Boot's management.metrics.tags.application so this service's
// metrics share naming/dashboards with the Spring Boot services.
const appName = "markdown-writer"

type MetricsService struct {
	BaseService

	metrics *ginmetrics.Metrics
}

func (m *MetricsService) Init(ctx *core.APPContext) {
	m.InitService("metrics", ctx, m)
	m.metrics = ginmetrics.New(ginmetrics.Options{
		Application: appName,
		Env:         m.envLabel(),
	})
}

func (m *MetricsService) Handler() http.Handler {
	return m.metrics.Handler()
}

func (m *MetricsService) Middleware() gin.HandlerFunc {
	return m.metrics.Middleware()
}

func (m *MetricsService) envLabel() string {
	if env.IsProduction {
		return "production"
	}
	return "development"
}
