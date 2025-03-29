package models

import (
	"time"
)

// TaskStatus represents the status of an async task
type TaskStatus string

const (
	// TaskStatusPending indicates the task is waiting to be processed
	TaskStatusPending TaskStatus = "pending"
	// TaskStatusRunning indicates the task is currently running
	TaskStatusRunning TaskStatus = "running"
	// TaskStatusCompleted indicates the task has completed successfully
	TaskStatusCompleted TaskStatus = "completed"
	// TaskStatusFailed indicates the task has failed
	TaskStatusFailed TaskStatus = "failed"
)

// TaskType represents the type of async task
type TaskType string

const (
	// TaskTypeSync indicates a repository sync task
	TaskTypeSync TaskType = "sync"
	// Add more task types as needed
)

// AsyncTask represents an asynchronous task in the system
type AsyncTask struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Type        TaskType   `json:"type" gorm:"type:varchar(50);not null"`
	Status      TaskStatus `json:"status" gorm:"type:varchar(50);not null"`
	ResourceID  string     `json:"resource_id" gorm:"type:varchar(255);not null"` // ID of the resource being processed (e.g., repo ID)
	UserID      string     `json:"user_id" gorm:"type:varchar(255);not null"`     // ID of the user who initiated the task
	Message     string     `json:"message" gorm:"type:text"`                      // Status message or error message
	Progress    int        `json:"progress" gorm:"default:0"`                     // Progress percentage (0-100)
	StartedAt   *time.Time `json:"started_at"`                                    // When the task started running
	CompletedAt *time.Time `json:"completed_at"`                                  // When the task completed or failed
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
