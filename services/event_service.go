package services

import (
	"time"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"gorm.io/gorm"
)

// EventService handles business logic for events
type EventService struct{}

// NewEventService creates a new EventService
func NewEventService() *EventService {
	return &EventService{}
}

// GetAllEvents returns all events with pagination
func (s *EventService) GetAllEvents(page, pageSize int, level models.EventLevel, source models.EventSource, resolved *bool) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64
	
	query := database.DB.Model(&models.Event{})
	
	// Apply filters
	if level != "" {
		query = query.Where("level = ?", level)
	}
	
	if source != "" {
		query = query.Where("source = ?", source)
	}
	
	if resolved != nil {
		query = query.Where("resolved = ?", *resolved)
	}
	
	// Count total matching records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}
	
	return events, total, nil
}

// GetEventByID returns a specific event by ID
func (s *EventService) GetEventByID(id uint) (models.Event, error) {
	var event models.Event
	result := database.DB.Preload("User").First(&event, id)
	return event, result.Error
}

// GetEventsByResourceType returns events for a specific resource type
func (s *EventService) GetEventsByResourceType(resourceType string, resourceID *uint) ([]models.Event, error) {
	var events []models.Event
	query := database.DB.Where("resource_type = ?", resourceType)
	
	if resourceID != nil {
		query = query.Where("resource_id = ?", *resourceID)
	}
	
	result := query.Order("created_at DESC").Find(&events)
	return events, result.Error
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

// UpdateEvent updates an existing event
func (s *EventService) UpdateEvent(id uint, request models.UpdateEventRequest) (models.Event, error) {
	var event models.Event
	if err := database.DB.First(&event, id).Error; err != nil {
		return models.Event{}, err
	}
	
	// Update fields if provided
	if request.Resolved != nil {
		event.Resolved = *request.Resolved
		
		if *request.Resolved {
			now := time.Now()
			event.ResolvedAt = &now
			event.ResolvedBy = request.ResolvedBy
		} else {
			event.ResolvedAt = nil
			event.ResolvedBy = nil
		}
	}
	
	result := database.DB.Save(&event)
	return event, result.Error
}

// DeleteEvent deletes an event
func (s *EventService) DeleteEvent(id uint) error {
	var event models.Event
	if err := database.DB.First(&event, id).Error; err != nil {
		return err
	}
	
	return database.DB.Delete(&event).Error
}

// LogSystemEvent is a helper function to log system events
func (s *EventService) LogSystemEvent(level models.EventLevel, message string, details string) (models.Event, error) {
	request := models.CreateEventRequest{
		Level:   level,
		Source:  models.EventSourceSystem,
		Message: message,
		Details: details,
	}
	
	return s.CreateEvent(request)
}

// LogErrorEvent is a helper function to log error events
func (s *EventService) LogErrorEvent(source models.EventSource, message string, details string, stackTrace string) (models.Event, error) {
	request := models.CreateEventRequest{
		Level:      models.EventLevelError,
		Source:     source,
		Message:    message,
		Details:    details,
		StackTrace: stackTrace,
	}
	
	return s.CreateEvent(request)
}
