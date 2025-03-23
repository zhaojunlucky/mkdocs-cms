package controllers

import (
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// UserGitRepoController handles git repository-related HTTP requests
type UserGitRepoController struct {
	BaseController
	userGitRepoService *services.UserGitRepoService
	asyncTaskService   *services.AsyncTaskService
}

func (c *UserGitRepoController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	c.ctx = ctx
	c.userGitRepoService = ctx.MustGetService("userGitRepoService").(*services.UserGitRepoService)
	c.asyncTaskService = ctx.MustGetService("asyncTaskService").(*services.AsyncTaskService)
	repos := router.Group("/repos")
	{
		//repos.GET("", userGitRepoController.GetRepos)
		repos.GET("/:id", c.GetRepo)
		repos.PUT("/:id", c.UpdateRepo)
		repos.DELETE("/:id", c.DeleteRepo)
		repos.POST("/:id/sync", c.SyncRepo)
		repos.GET("/:id/branches", c.GetRepoBranches)
	}

	// User repositories route
	userRepos := router.Group("/users/repos")
	{
		userRepos.GET("/:user_id", c.GetReposByUser)
	}
}

// GetRepos returns all git repositories for the authenticated user
func (c *UserGitRepoController) GetRepos(ctx *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get only the repositories owned by the authenticated user
	repos, err := c.userGitRepoService.GetReposByUser(authenticatedUserID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []models.UserGitRepoResponse
	for _, repo := range repos {
		response = append(response, repo.ToResponse(false))
	}

	ctx.JSON(http.StatusOK, response)
}

// GetReposByUser returns all git repositories for a specific user
func (c *UserGitRepoController) GetReposByUser(ctx *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get the requested user ID from the URL parameter
	requestedUserID := ctx.Param("user_id")

	// Check if the authenticated user is trying to access their own repositories
	// This ensures users can only see their own repositories
	if authenticatedUserID.(string) != requestedUserID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own repositories"})
		return
	}

	repos, err := c.userGitRepoService.GetReposByUser(requestedUserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []models.UserGitRepoResponse = make([]models.UserGitRepoResponse, 0)
	for _, repo := range repos {
		response = append(response, repo.ToResponse(false))
	}

	ctx.JSON(http.StatusOK, response)
}

// GetRepo returns a specific git repository
func (c *UserGitRepoController) GetRepo(ctx *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := ctx.Param("id")
	repo, err := c.userGitRepoService.GetRepoByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Check if the authenticated user owns this repository
	if repo.UserID != authenticatedUserID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own repositories"})
		return
	}

	ctx.JSON(http.StatusOK, repo.ToResponse(true))
}

// UpdateRepo updates an existing git repository
func (c *UserGitRepoController) UpdateRepo(ctx *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := ctx.Param("id")

	// First, get the existing repository to check ownership
	existingRepo, err := c.userGitRepoService.GetRepoByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Check if the authenticated user owns this repository
	if existingRepo.UserID != authenticatedUserID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own repositories"})
		return
	}

	var request models.UpdateUserGitRepoRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repo, err := c.userGitRepoService.UpdateRepo(id, request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, repo.ToResponse(false))
}

// DeleteRepo deletes a git repository
func (c *UserGitRepoController) DeleteRepo(ctx *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := ctx.Param("id")

	// First, get the existing repository to check ownership
	existingRepo, err := c.userGitRepoService.GetRepoByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Check if the authenticated user owns this repository
	if existingRepo.UserID != authenticatedUserID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own repositories"})
		return
	}

	if err := c.userGitRepoService.DeleteRepo(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Repository deleted successfully"})
}

// SyncRepo synchronizes a git repository with its remote
func (c *UserGitRepoController) SyncRepo(ctx *gin.Context) {
	// Get authenticated user ID from context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	id := ctx.Param("id")

	// First, get the existing repository to check ownership
	repo, err := c.userGitRepoService.GetRepoByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Check if the authenticated user owns this repository
	if repo.UserID != authenticatedUserID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only sync your own repositories"})
		return
	}

	// Create an async task for the sync operation
	task, err := c.asyncTaskService.CreateTask(models.TaskTypeSync, id, authenticatedUserID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sync task"})
		return
	}

	// Update status to syncing
	if err := c.userGitRepoService.UpdateRepoStatus(id, models.StatusSyncing, ""); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Start sync in a goroutine
	go func() {
		// Update task status to running
		c.asyncTaskService.UpdateTaskStatus(task.ID, models.TaskStatusRunning, "Repository sync in progress")

		if err := c.userGitRepoService.SyncRepo(id); err != nil {
			c.userGitRepoService.UpdateRepoStatus(id, models.StatusFailed, err.Error())
			c.asyncTaskService.UpdateTaskStatus(task.ID, models.TaskStatusFailed, err.Error())
			return
		}
		if err := c.userGitRepoService.CheckWebHooks(id); err != nil {
			c.userGitRepoService.UpdateRepoStatus(id, models.StatusFailed, err.Error())
			c.asyncTaskService.UpdateTaskStatus(task.ID, models.TaskStatusFailed, err.Error())
			return
		}

		// The SyncRepo method now handles setting the status (synced or warning)
		// so we don't need to set it here again

		// Get the updated repository to check its status
		updatedRepo, _ := c.userGitRepoService.GetRepoByID(id)

		if updatedRepo.Status == models.StatusWarning {
			c.asyncTaskService.UpdateTaskStatus(task.ID, models.TaskStatusCompleted,
				"Repository sync completed with warnings: "+updatedRepo.ErrorMsg)
		} else {
			c.asyncTaskService.UpdateTaskStatus(task.ID, models.TaskStatusCompleted,
				"Repository sync completed successfully")
		}
	}()

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Repository sync started",
		"task_id": task.ID,
	})
}

// GetRepoBranches returns all branches for a specific git repository
func (c *UserGitRepoController) GetRepoBranches(ctx *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get the repository ID from the URL
	repoID := ctx.Param("id")

	// Get the repository to verify ownership
	repo, err := c.userGitRepoService.GetRepoByID(repoID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Verify that the repository belongs to the authenticated user
	if repo.UserID != authenticatedUserID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this repository"})
		return
	}

	// Get the branches
	branches, err := c.userGitRepoService.GetRepoBranches(repoID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, branches)
}
