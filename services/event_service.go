package services

import (
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"time"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"gorm.io/gorm"
)

// EventService handles business logic for events
type EventService struct {
	BaseService
}

func (s *EventService) Init(ctx *core.APPContext) {
	s.InitService("eventService", ctx, s)
}

// CreateEvent creates a new event
func (s *EventService) CreateEvent(request models.CreateEventRequest) (models.Event, error) {
	event := models.Event{
		Level:        request.Level,
		Source:       request.Source,
		Message:      request.Message,
		Details:      request.Details,
		StackTrace:   request.StackTrace,
		UserID:       request.UserID,
		ResourceID:   request.ResourceID,
		ResourceType: request.ResourceType,
		IPAddress:    request.IPAddress,
		UserAgent:    request.UserAgent,
		Resolved:     false,
		CreatedAt:    time.Now(),
	}

	// If user ID is provided, check if user exists
	if request.UserID != nil {
		var user models.User
		if err := database.DB.First(&user, *request.UserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// If user doesn't exist, set UserID to nil
				event.UserID = nil
			}
		}
	}

	result := database.DB.Create(&event)
	return event, result.Error
}
