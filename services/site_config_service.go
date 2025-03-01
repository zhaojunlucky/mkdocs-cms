package services

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
)

// SiteConfigService handles business logic for site configurations
type SiteConfigService struct{}

// NewSiteConfigService creates a new SiteConfigService
func NewSiteConfigService() *SiteConfigService {
	return &SiteConfigService{}
}

// GetAllSiteConfigs returns all site configurations
func (s *SiteConfigService) GetAllSiteConfigs() ([]models.SiteConfig, error) {
	var configs []models.SiteConfig
	result := database.DB.Find(&configs)
	return configs, result.Error
}

// GetSiteConfigByID returns a specific site configuration by ID
func (s *SiteConfigService) GetSiteConfigByID(id uint) (models.SiteConfig, error) {
	var config models.SiteConfig
	result := database.DB.First(&config, id)
	return config, result.Error
}

// GetSiteConfigByDomain returns a site configuration by its domain
func (s *SiteConfigService) GetSiteConfigByDomain(domain string) (models.SiteConfig, error) {
	var config models.SiteConfig
	result := database.DB.Where("site_domain = ?", domain).First(&config)
	return config, result.Error
}

// CreateSiteConfig creates a new site configuration
func (s *SiteConfigService) CreateSiteConfig(request models.CreateSiteConfigRequest) (models.SiteConfig, error) {
	// Validate working directory
	workingDir := request.RepoWorkingDir
	if !filepath.IsAbs(workingDir) {
		return models.SiteConfig{}, errors.New("working directory must be an absolute path")
	}

	// Check if working directory exists
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		return models.SiteConfig{}, errors.New("working directory does not exist")
	}

	// Check if domain is already in use
	var existingConfig models.SiteConfig
	if err := database.DB.Where("site_domain = ?", request.SiteDomain).First(&existingConfig).Error; err == nil {
		return models.SiteConfig{}, errors.New("domain is already in use by another site")
	}

	config := models.SiteConfig{
		Name:            request.Name,
		Description:     request.Description,
		RepoWorkingDir:  workingDir,
		SiteDomain:      request.SiteDomain,
		SiteTitle:       request.SiteTitle,
		SiteDescription: request.SiteDescription,
		ExtraCSS:        request.ExtraCSS,
		ExtraJS:         request.ExtraJS,
		GoogleAnalytics: request.GoogleAnalytics,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	result := database.DB.Create(&config)
	return config, result.Error
}

// UpdateSiteConfig updates an existing site configuration
func (s *SiteConfigService) UpdateSiteConfig(id uint, request models.UpdateSiteConfigRequest) (models.SiteConfig, error) {
	var config models.SiteConfig
	if err := database.DB.First(&config, id).Error; err != nil {
		return models.SiteConfig{}, err
	}

	// Update fields if provided
	if request.Name != "" {
		config.Name = request.Name
	}
	if request.Description != "" {
		config.Description = request.Description
	}
	if request.RepoWorkingDir != "" {
		// Validate working directory
		workingDir := request.RepoWorkingDir
		if !filepath.IsAbs(workingDir) {
			return models.SiteConfig{}, errors.New("working directory must be an absolute path")
		}

		// Check if working directory exists
		if _, err := os.Stat(workingDir); os.IsNotExist(err) {
			return models.SiteConfig{}, errors.New("working directory does not exist")
		}

		config.RepoWorkingDir = workingDir
	}
	if request.SiteDomain != "" {
		// Check if domain is already in use by another site
		var existingConfig models.SiteConfig
		if err := database.DB.Where("site_domain = ? AND id != ?", request.SiteDomain, id).First(&existingConfig).Error; err == nil {
			return models.SiteConfig{}, errors.New("domain is already in use by another site")
		}
		config.SiteDomain = request.SiteDomain
	}
	if request.SiteTitle != "" {
		config.SiteTitle = request.SiteTitle
	}
	if request.SiteDescription != "" {
		config.SiteDescription = request.SiteDescription
	}
	if request.ExtraCSS != "" {
		config.ExtraCSS = request.ExtraCSS
	}
	if request.ExtraJS != "" {
		config.ExtraJS = request.ExtraJS
	}
	if request.GoogleAnalytics != "" {
		config.GoogleAnalytics = request.GoogleAnalytics
	}

	config.UpdatedAt = time.Now()
	result := database.DB.Save(&config)
	return config, result.Error
}

// DeleteSiteConfig deletes a site configuration
func (s *SiteConfigService) DeleteSiteConfig(id uint) error {
	var config models.SiteConfig
	if err := database.DB.First(&config, id).Error; err != nil {
		return err
	}

	return database.DB.Delete(&config).Error
}
