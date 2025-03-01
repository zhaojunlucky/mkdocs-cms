package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// SiteConfigController handles HTTP requests for site configurations
type SiteConfigController struct {
	siteConfigService *services.SiteConfigService
}

// NewSiteConfigController creates a new SiteConfigController
func NewSiteConfigController() *SiteConfigController {
	return &SiteConfigController{
		siteConfigService: services.NewSiteConfigService(),
	}
}

// GetAllSiteConfigs handles GET /site-configs
func (c *SiteConfigController) GetAllSiteConfigs(ctx *gin.Context) {
	configs, err := c.siteConfigService.GetAllSiteConfigs()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []models.SiteConfigResponse
	for _, config := range configs {
		response = append(response, config.ToResponse())
	}

	ctx.JSON(http.StatusOK, response)
}

// GetSiteConfigByID handles GET /site-configs/:id
func (c *SiteConfigController) GetSiteConfigByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	config, err := c.siteConfigService.GetSiteConfigByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Site configuration not found"})
		return
	}

	ctx.JSON(http.StatusOK, config.ToResponse())
}

// CreateSiteConfig handles POST /site-configs
func (c *SiteConfigController) CreateSiteConfig(ctx *gin.Context) {
	var request models.CreateSiteConfigRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := c.siteConfigService.CreateSiteConfig(request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, config.ToResponse())
}

// UpdateSiteConfig handles PUT /site-configs/:id
func (c *SiteConfigController) UpdateSiteConfig(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var request models.UpdateSiteConfigRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := c.siteConfigService.UpdateSiteConfig(uint(id), request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, config.ToResponse())
}

// DeleteSiteConfig handles DELETE /site-configs/:id
func (c *SiteConfigController) DeleteSiteConfig(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := c.siteConfigService.DeleteSiteConfig(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Site configuration deleted successfully"})
}
