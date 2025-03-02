package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// AsyncTaskController handles async task-related HTTP requests
type AsyncTaskController struct {
	asyncTaskService *services.AsyncTaskService
}

// NewAsyncTaskController creates a new AsyncTaskController
func NewAsyncTaskController() *AsyncTaskController {
	return &AsyncTaskController{
		asyncTaskService: services.NewAsyncTaskService(),
	}
}

// GetTask returns a specific task by ID
func (c *AsyncTaskController) GetTask(ctx *gin.Context) {
	// Get authenticated user ID from context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Get task ID from URL parameter
	taskID := ctx.Param("id")
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	// Get the task
	task, err := c.asyncTaskService.GetTaskByID(taskID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if the authenticated user owns this task
	if task.UserID != authenticatedUserID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only view your own tasks"})
		return
	}

	ctx.JSON(http.StatusOK, task)
}

// GetUserTasks returns all tasks for the authenticated user
func (c *AsyncTaskController) GetUserTasks(ctx *gin.Context) {
	// Get authenticated user ID from context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Get the tasks
	tasks, err := c.asyncTaskService.GetTasksByUser(authenticatedUserID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	ctx.JSON(http.StatusOK, tasks)
}

// GetResourceTasks returns all tasks for a specific resource
func (c *AsyncTaskController) GetResourceTasks(ctx *gin.Context) {
	// Get authenticated user ID from context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Get resource ID from URL parameter
	resourceID := ctx.Param("resourceId")
	if resourceID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Resource ID is required"})
		return
	}

	// Get the tasks
	tasks, err := c.asyncTaskService.GetTasksByResource(resourceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	// Filter tasks to only include those owned by the authenticated user
	userTasks := make([]interface{}, 0)
	for _, task := range tasks {
		if task.UserID == authenticatedUserID.(string) {
			userTasks = append(userTasks, task)
		}
	}

	ctx.JSON(http.StatusOK, userTasks)
}

// RegisterRoutes registers the routes for the AsyncTaskController
func (c *AsyncTaskController) RegisterRoutes(router *gin.RouterGroup) {
	tasks := router.Group("/tasks")
	//tasks.Use(middleware.RequireAuth())
	{
		tasks.GET("", c.GetUserTasks)
		tasks.GET("/:id", c.GetTask)
		tasks.GET("/resource/:resourceId", c.GetResourceTasks)
	}
}
