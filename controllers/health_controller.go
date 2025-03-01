package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

var healthService = services.NewHealthService()

// HealthCheck returns the health status of the API
func HealthCheck(c *gin.Context) {
	healthy, err := healthService.CheckHealth()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if !healthy {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "System is not healthy",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "API is healthy",
	})
}
