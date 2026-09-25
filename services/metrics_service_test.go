package services

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/core"
)

func TestMetricsServiceExposesSpringCompatibleMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &MetricsService{}
	svc.Init(&core.APPContext{})

	r := gin.New()
	r.Use(svc.Middleware())
	r.GET("/api/metrics", func(c *gin.Context) {
		svc.Handler().ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/api/v1/sites/:id", func(c *gin.Context) {
		c.Status(200)
	})

	req := httptest.NewRequest("GET", "/api/v1/sites/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	metricsReq := httptest.NewRequest("GET", "/api/metrics", nil)
	metricsW := httptest.NewRecorder()
	r.ServeHTTP(metricsW, metricsReq)
	if metricsW.Code != 200 {
		t.Fatalf("metrics status = %d, want 200", metricsW.Code)
	}

	body := metricsW.Body.String()
	for _, want := range []string{
		`application="markdown-writer"`,
		`http_server_requests_seconds_count{`,
		`uri="/api/v1/sites/:id"`,
		`status="200"`,
		`outcome="SUCCESS"`,
		`exception="None"`,
		`app_info{application="markdown-writer",env="development"} 1`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("expected metrics output to contain %q, got:\n%s", want, body)
		}
	}
}
