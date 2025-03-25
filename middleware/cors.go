package middleware

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/config"
	"strings"
)

// CORS is a middleware that adds CORS headers to responses
func CORS() gin.HandlerFunc {
	return CORSWithConfig(nil)
}

// CORSWithConfig is a middleware that adds CORS headers to responses with configuration
func CORSWithConfig(appConfig *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := "*"
		allowCredentials := "true"
		allowMethods := "POST, OPTIONS, GET, PUT, DELETE"
		allowHeaders := "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"

		// Use configuration if provided
		if appConfig != nil && appConfig.Security.CORS.AllowedOrigins != nil && len(appConfig.Security.CORS.AllowedOrigins) > 0 {
			origin = strings.Join(appConfig.Security.CORS.AllowedOrigins, ", ")
		}

		if appConfig != nil && appConfig.Security.CORS.AllowedMethods != nil && len(appConfig.Security.CORS.AllowedMethods) > 0 {
			allowMethods = strings.Join(appConfig.Security.CORS.AllowedMethods, ", ")
		}

		if appConfig != nil && appConfig.Security.CORS.AllowedHeaders != nil && len(appConfig.Security.CORS.AllowedHeaders) > 0 {
			allowHeaders = strings.Join(appConfig.Security.CORS.AllowedHeaders, ", ")
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", allowCredentials)
		c.Writer.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		c.Writer.Header().Set("Access-Control-Allow-Methods", allowMethods)

		if c.Request.Method == "OPTIONS" {
			log.Infof("OPTIONS request for %s", c.Request.URL.Path)
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
