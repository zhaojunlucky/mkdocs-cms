package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger is a middleware that logs the request method, path, status code, and latency
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Get client IP
		clientIP := c.ClientIP()

		// Get request method
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		// Log the request
		fmt.Printf("[GIN] %v | %3d | %13v | %15s | %-7s %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}
