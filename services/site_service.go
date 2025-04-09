package services

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
)

type SiteService struct {
	BaseService
}

func (s *SiteService) Init(ctx *core.APPContext) {
	s.InitService("siteService", ctx, s)
}

func (s *SiteService) GetSiteConfig(key string) (*models.SiteSetting, error) {
	var site models.SiteSetting
	if err := database.DB.Where("key = ?", key).First(&site).Error; err != nil {
		return nil, err
	}
	return &site, nil
}

func (s *SiteService) AllowUserRegistration() bool {
	config, err := s.GetSiteConfig("allow_user_registration")
	if err != nil {
		log.Warnf("Failed to get site config: %v", err)
		return false
	}
	return config.Value == "true"
}
