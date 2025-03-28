package controllers

import (
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// AsyncTaskController handles async task-related HTTP requests
type AsyncTaskController struct {
	BaseController
	asyncTaskService *services.AsyncTaskService
}

func (c *AsyncTaskController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	c.ctx = ctx
	c.asyncTaskService = ctx.MustGetService("asyncTaskService").(*services.AsyncTaskService)
	tasks := router.Group("/tasks")
	{
		tasks.GET("", c.GetUserTasks)
		tasks.GET("/:id", c.GetTask)
		tasks.GET("/resource/:resourceId", c.GetResourceTasks)
	}
}

// GetTask returns a specific task by ID
func (c *AsyncTaskController) GetTask(ctx *gin.Context) {
	// Get authenticated user ID from context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		log.Errorf("Failed to get authenticated user ID from context")
		core.ResponseErrStr(ctx, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get task ID from URL parameter
	taskID := ctx.Param("id")
	if taskID == "" {
		log.Errorf("Task ID is required")
		core.ResponseErrStr(ctx, http.StatusBadRequest, "Task ID is required")
		return
	}

	// Get the task
	task, err := c.asyncTaskService.GetTaskByID(taskID)
	if err != nil {
		log.Errorf("Failed to retrieve task: %v", err)
		core.ResponseErrStr(ctx, http.StatusNotFound, "Task not found")
		return
	}

	// Check if the authenticated user owns this task
	if task.UserID != authenticatedUserID.(string) {
		log.Errorf("User %s does not own task %s", authenticatedUserID, taskID)
		core.ResponseErrStr(ctx, http.StatusForbidden, "You can only view your own tasks")
		return
	}

	ctx.JSON(http.StatusOK, task)
}

// GetUserTasks returns all tasks for the authenticated user
func (c *AsyncTaskController) GetUserTasks(ctx *gin.Context) {
	// Get authenticated user ID from context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		log.Errorf("Failed to get authenticated user ID from context")
		core.ResponseErrStr(ctx, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get the tasks
	tasks, err := c.asyncTaskService.GetTasksByUser(authenticatedUserID.(string))
	if err != nil {
		log.Errorf("Failed to retrieve tasks: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to retrieve tasks")
		return
	}

	core.ResponseOKArr(ctx, tasks)
}

// GetResourceTasks returns all tasks for a specific resource
func (c *AsyncTaskController) GetResourceTasks(ctx *gin.Context) {
	// Get authenticated user ID from context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		log.Errorf("Failed to get authenticated user ID from context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Get resource ID from URL parameter
	resourceID := ctx.Param("resourceId")
	if resourceID == "" {
		log.Errorf("Resource ID is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Resource ID is required"})
		return
	}

	// Get the tasks
	tasks, err := c.asyncTaskService.GetTasksByResource(resourceID)
	if err != nil {
		log.Errorf("Failed to retrieve tasks: %v", err)
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

	core.ResponseOKArr(ctx, userTasks)
}
