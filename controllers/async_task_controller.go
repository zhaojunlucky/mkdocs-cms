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
	reqParam := core.NewRequestParam()

	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "You must be logged in to view tasks")
	taskId := reqParam.AddUrlParam("id", false, nil)

	if err := reqParam.Handle(ctx); err != nil {
		core.HandleError(ctx, err)
		return
	}

	// Get the task
	task, err := c.asyncTaskService.GetTaskByID(taskId.String())
	if err != nil {
		log.Errorf("Failed to retrieve task: %v", err)
		core.ResponseErrStr(ctx, http.StatusNotFound, "Task not found")
		return
	}

	// Check if the authenticated user owns this task
	if task.UserID != userId.String() {
		log.Errorf("User %s does not own task %s", userId.String(), task.ID)
		core.ResponseErrStr(ctx, http.StatusForbidden, "You can only view your own tasks")
		return
	}

	ctx.JSON(http.StatusOK, task)
}

// GetUserTasks returns all tasks for the authenticated user
func (c *AsyncTaskController) GetUserTasks(ctx *gin.Context) {

	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")

	if err := reqParam.Handle(ctx); err != nil {
		core.HandleError(ctx, err)
		return
	}

	// Get the tasks
	tasks, err := c.asyncTaskService.GetTasksByUser(userId.String())
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
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	resourceId := reqParam.AddUrlParam("resourceId", false, nil)

	if err := reqParam.Handle(ctx); err != nil {
		core.HandleError(ctx, err)
		return
	}

	// Get the tasks
	tasks, err := c.asyncTaskService.GetTasksByResource(resourceId.String())
	if err != nil {
		log.Errorf("Failed to retrieve tasks: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	// Filter tasks to only include those owned by the authenticated user
	userTasks := make([]interface{}, 0)
	for _, task := range tasks {
		if task.UserID == userId.String() {
			userTasks = append(userTasks, task)
		}
	}

	core.ResponseOKArr(ctx, userTasks)
}
