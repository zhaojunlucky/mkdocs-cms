package models

import (
	"time"
)

// SiteConfig represents configuration for a site
type SiteConfig struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Name            string    `json:"name" gorm:"not null"`
	Description     string    `json:"description"`
	RepoWorkingDir  string    `json:"repo_working_dir" gorm:"not null"`
	SiteDomain      string    `json:"site_domain" gorm:"not null"`
	SiteTitle       string    `json:"site_title"`
	SiteDescription string    `json:"site_description"`
	ExtraCSS        string    `json:"extra_css" gorm:"type:text"`
	ExtraJS         string    `json:"extra_js" gorm:"type:text"`
	GoogleAnalytics string    `json:"google_analytics"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// SiteConfigResponse is the structure returned to clients
type SiteConfigResponse struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	RepoWorkingDir  string    `json:"repo_working_dir"`
	SiteDomain      string    `json:"site_domain"`
	SiteTitle       string    `json:"site_title,omitempty"`
	SiteDescription string    `json:"site_description,omitempty"`
	ExtraCSS        string    `json:"extra_css,omitempty"`
	ExtraJS         string    `json:"extra_js,omitempty"`
	GoogleAnalytics string    `json:"google_analytics,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ToResponse converts a SiteConfig to a SiteConfigResponse
func (s *SiteConfig) ToResponse(includeRepo bool) SiteConfigResponse {
	response := SiteConfigResponse{
		ID:              s.ID,
		Name:            s.Name,
		Description:     s.Description,
		RepoWorkingDir:  s.RepoWorkingDir,
		SiteDomain:      s.SiteDomain,
		SiteTitle:       s.SiteTitle,
		SiteDescription: s.SiteDescription,
		ExtraCSS:        s.ExtraCSS,
		ExtraJS:         s.ExtraJS,
		GoogleAnalytics: s.GoogleAnalytics,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}

	return response
}

// CreateSiteConfigRequest is the structure for site configuration creation requests
type CreateSiteConfigRequest struct {
	Name            string `json:"name" binding:"required"`
	Description     string `json:"description"`
	RepoWorkingDir  string `json:"repo_working_dir" binding:"required"`
	SiteDomain      string `json:"site_domain" binding:"required"`
	SiteTitle       string `json:"site_title"`
	SiteDescription string `json:"site_description"`
	ExtraCSS        string `json:"extra_css"`
	ExtraJS         string `json:"extra_js"`
	GoogleAnalytics string `json:"google_analytics"`
}

// UpdateSiteConfigRequest is the structure for site configuration update requests
type UpdateSiteConfigRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	RepoWorkingDir  string `json:"repo_working_dir"`
	SiteDomain      string `json:"site_domain"`
	SiteTitle       string `json:"site_title"`
	SiteDescription string `json:"site_description"`
	ExtraCSS        string `json:"extra_css"`
	ExtraJS         string `json:"extra_js"`
	GoogleAnalytics string `json:"google_analytics"`
}
