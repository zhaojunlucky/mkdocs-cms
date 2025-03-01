package services

import (
	"github.com/zhaojunlucky/mkdocs-cms/database"
)

// HealthService handles business logic related to health checks
type HealthService struct{}

// NewHealthService creates a new health service
func NewHealthService() *HealthService {
	return &HealthService{}
}

// CheckHealth checks the health of the application
func (s *HealthService) CheckHealth() (bool, error) {
	// Check database connection
	sqlDB, err := database.DB.DB()
	if err != nil {
		return false, err
	}

	// Ping the database
	err = sqlDB.Ping()
	if err != nil {
		return false, err
	}

	return true, nil
}
