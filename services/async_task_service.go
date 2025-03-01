package services

import (
	"errors"
	"time"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
)

// AsyncTaskService handles business logic for async tasks
type AsyncTaskService struct{}

// NewAsyncTaskService creates a new AsyncTaskService
func NewAsyncTaskService() *AsyncTaskService {
	return &AsyncTaskService{}
}

// CreateTask creates a new async task
func (s *AsyncTaskService) CreateTask(taskType models.TaskType, resourceID, userID string) (models.AsyncTask, error) {
	task := models.AsyncTask{
		Type:       taskType,
		Status:     models.TaskStatusPending,
		ResourceID: resourceID,
		UserID:     userID,
		Message:    "Task created",
	}

	result := database.DB.Create(&task)
	if result.Error != nil {
		return models.AsyncTask{}, result.Error
	}

	return task, nil
}

// GetTaskByID returns a specific task by ID
func (s *AsyncTaskService) GetTaskByID(id string) (models.AsyncTask, error) {
	var task models.AsyncTask
	result := database.DB.Where("id = ?", id).First(&task)
	if result.Error != nil {
		return models.AsyncTask{}, result.Error
	}
	return task, nil
}

// GetTasksByUser returns all tasks for a specific user
func (s *AsyncTaskService) GetTasksByUser(userID string) ([]models.AsyncTask, error) {
	var tasks []models.AsyncTask
	result := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}
	return tasks, nil
}

// GetTasksByResource returns all tasks for a specific resource
func (s *AsyncTaskService) GetTasksByResource(resourceID string) ([]models.AsyncTask, error) {
	var tasks []models.AsyncTask
	result := database.DB.Where("resource_id = ?", resourceID).Order("created_at DESC").Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}
	return tasks, nil
}

// UpdateTaskStatus updates the status of a task
func (s *AsyncTaskService) UpdateTaskStatus(id string, status models.TaskStatus, message string) error {
	var task models.AsyncTask
	result := database.DB.Where("id = ?", id).First(&task)
	if result.Error != nil {
		return result.Error
	}

	updates := map[string]interface{}{
		"status":  status,
		"message": message,
	}

	// Set started_at if transitioning to running
	if status == models.TaskStatusRunning && task.StartedAt == nil {
		now := time.Now()
		updates["started_at"] = now
	}

	// Set completed_at if transitioning to completed or failed
	if (status == models.TaskStatusCompleted || status == models.TaskStatusFailed) && task.CompletedAt == nil {
		now := time.Now()
		updates["completed_at"] = now
	}

	result = database.DB.Model(&task).Updates(updates)
	return result.Error
}

// UpdateTaskProgress updates the progress of a task
func (s *AsyncTaskService) UpdateTaskProgress(id string, progress int) error {
	if progress < 0 || progress > 100 {
		return errors.New("progress must be between 0 and 100")
	}

	result := database.DB.Model(&models.AsyncTask{}).Where("id = ?", id).Update("progress", progress)
	return result.Error
}

// DeleteTask deletes a task
func (s *AsyncTaskService) DeleteTask(id string) error {
	result := database.DB.Delete(&models.AsyncTask{}, "id = ?", id)
	return result.Error
}

// CleanupOldTasks deletes tasks older than the specified duration
func (s *AsyncTaskService) CleanupOldTasks(age time.Duration) error {
	cutoff := time.Now().Add(-age)
	result := database.DB.Where("created_at < ?", cutoff).Delete(&models.AsyncTask{})
	return result.Error
}
