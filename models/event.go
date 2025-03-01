package models

import (
	"time"
)

// EventLevel represents the severity level of an event
type EventLevel string

const (
	// EventLevelInfo represents an informational event
	EventLevelInfo EventLevel = "info"
	// EventLevelWarning represents a warning event
	EventLevelWarning EventLevel = "warning"
	// EventLevelError represents an error event
	EventLevelError EventLevel = "error"
	// EventLevelCritical represents a critical error event
	EventLevelCritical EventLevel = "critical"
)

// EventSource represents the source of an event
type EventSource string

const (
	// EventSourceSystem represents a system-generated event
	EventSourceSystem EventSource = "system"
	// EventSourceUser represents a user-generated event
	EventSourceUser EventSource = "user"
	// EventSourceGitRepo represents a git repository-related event
	EventSourceGitRepo EventSource = "git_repo"
	// EventSourceAPI represents an API-related event
	EventSourceAPI EventSource = "api"
	// EventSourceDatabase represents a database-related event
	EventSourceDatabase EventSource = "database"
)

// Event represents a system event or error
type Event struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	Level       EventLevel  `json:"level" gorm:"type:string;not null;index"`
	Source      EventSource `json:"source" gorm:"type:string;not null;index"`
	Message     string      `json:"message" gorm:"not null"`
	Details     string      `json:"details" gorm:"type:text"`
	StackTrace  string      `json:"stack_trace" gorm:"type:text"`
	UserID      *string     `json:"user_id" gorm:"index"`
	User        *User       `json:"user" gorm:"foreignKey:UserID"`
	ResourceID  *uint       `json:"resource_id"`
	ResourceType string     `json:"resource_type" gorm:"index"`
	IPAddress   string      `json:"ip_address"`
	UserAgent   string      `json:"user_agent"`
	Resolved    bool        `json:"resolved" gorm:"default:false;index"`
	ResolvedAt  *time.Time  `json:"resolved_at"`
	ResolvedBy  *string     `json:"resolved_by"`
	CreatedAt   time.Time   `json:"created_at" gorm:"index"`
}

// EventResponse is the structure returned to clients
type EventResponse struct {
	ID           uint        `json:"id"`
	Level        EventLevel  `json:"level"`
	Source       EventSource `json:"source"`
	Message      string      `json:"message"`
	Details      string      `json:"details,omitempty"`
	StackTrace   string      `json:"stack_trace,omitempty"`
	UserID       *string     `json:"user_id,omitempty"`
	User         *UserResponse `json:"user,omitempty"`
	ResourceID   *uint       `json:"resource_id,omitempty"`
	ResourceType string      `json:"resource_type,omitempty"`
	IPAddress    string      `json:"ip_address,omitempty"`
	UserAgent    string      `json:"user_agent,omitempty"`
	Resolved     bool        `json:"resolved"`
	ResolvedAt   *time.Time  `json:"resolved_at,omitempty"`
	ResolvedBy   *string     `json:"resolved_by,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

// ToResponse converts an Event to an EventResponse
func (e *Event) ToResponse(includeDetails bool, includeUser bool) EventResponse {
	response := EventResponse{
		ID:           e.ID,
		Level:        e.Level,
		Source:       e.Source,
		Message:      e.Message,
		UserID:       e.UserID,
		ResourceID:   e.ResourceID,
		ResourceType: e.ResourceType,
		IPAddress:    e.IPAddress,
		UserAgent:    e.UserAgent,
		Resolved:     e.Resolved,
		ResolvedAt:   e.ResolvedAt,
		ResolvedBy:   e.ResolvedBy,
		CreatedAt:    e.CreatedAt,
	}

	if includeDetails {
		response.Details = e.Details
		response.StackTrace = e.StackTrace
	}

	if includeUser && e.User != nil {
		userResponse := e.User.ToResponse()
		response.User = &userResponse
	}

	return response
}

// CreateEventRequest is the structure for event creation requests
type CreateEventRequest struct {
	Level        EventLevel  `json:"level" binding:"required"`
	Source       EventSource `json:"source" binding:"required"`
	Message      string      `json:"message" binding:"required"`
	Details      string      `json:"details"`
	StackTrace   string      `json:"stack_trace"`
	UserID       *string     `json:"user_id"`
	ResourceID   *uint       `json:"resource_id"`
	ResourceType string      `json:"resource_type"`
	IPAddress    string      `json:"ip_address"`
	UserAgent    string      `json:"user_agent"`
}

// UpdateEventRequest is the structure for event update requests
type UpdateEventRequest struct {
	Resolved   *bool      `json:"resolved"`
	ResolvedBy *string    `json:"resolved_by"`
}
