package controllers

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// UserGitRepoController handles git repository-related HTTP requests
type UserGitRepoController struct {
	BaseController
	userGitRepoService           *services.UserGitRepoService
	userGitRepoCollectionService *services.UserGitRepoCollectionService
	asyncTaskService             *services.AsyncTaskService
	userGitRepoLockService       *services.UserGitRepoLockService
}

func (c *UserGitRepoController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	c.ctx = ctx
	c.userGitRepoService = ctx.MustGetService("userGitRepoService").(*services.UserGitRepoService)
	c.asyncTaskService = ctx.MustGetService("asyncTaskService").(*services.AsyncTaskService)
	c.userGitRepoLockService = ctx.MustGetService("userGitRepoLockService").(*services.UserGitRepoLockService)
	c.userGitRepoCollectionService = ctx.MustGetService("userGitRepoCollectionService").(*services.UserGitRepoCollectionService)
	repos := router.Group("/repos")
	{
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

// GetReposByUser returns all git repositories for a specific user
func (c *UserGitRepoController) GetReposByUser(ctx *gin.Context) {
	// Get the authenticated user ID from the context
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", true, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")

	if err := reqParam.Handle(ctx); err != nil {
		core.HandleError(ctx, err)
		return
	}

	repos, err := c.userGitRepoService.GetReposByUser(userId.String())
	if err != nil {
		log.Errorf("Failed to get repositories: %v", err)
		core.ResponseErr(ctx, http.StatusInternalServerError, err)
		return
	}

	var response []models.UserGitRepoResponse = make([]models.UserGitRepoResponse, 0)
	for _, repo := range repos {
		response = append(response, repo.ToResponse(false))
	}

	core.ResponseOKArr(ctx, response)
}

// GetRepo returns a specific git repository
func (c *UserGitRepoController) GetRepo(ctx *gin.Context) {
	// Get the authenticated user ID from the context
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", true, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("id", true, regexp.MustCompile("\\d+"))

	if err := reqParam.Handle(ctx); err != nil {
		core.HandleError(ctx, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(ctx, http.StatusBadRequest, "Invalid repository ID")
		return
	}

	// Verify repository ownership
	repo, err := c.userGitRepoCollectionService.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, repo.ToResponse(true))
}

// UpdateRepo updates an existing git repository
func (c *UserGitRepoController) UpdateRepo(ctx *gin.Context) {
	// Get the authenticated user ID from the context
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", true, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("id", true, regexp.MustCompile("\\d+"))
	var request models.UpdateUserGitRepoRequest

	if err := reqParam.HandleWithBody(ctx, &request); err != nil {
		core.HandleError(ctx, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(ctx, http.StatusBadRequest, "Invalid repository ID")
		return
	}

	// Verify repository ownership
	repo, err := c.userGitRepoCollectionService.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(ctx, err)
		return
	}

	repo, err = c.userGitRepoService.UpdateRepo(repo, request)
	if err != nil {
		log.Errorf("Failed to update repository: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		core.ResponseErr(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, repo.ToResponse(false))
}

// DeleteRepo deletes a git repository
func (c *UserGitRepoController) DeleteRepo(ctx *gin.Context) {
	// Get the authenticated user ID from the context
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", true, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("id", true, regexp.MustCompile("\\d+"))
	var request models.UpdateUserGitRepoRequest

	if err := reqParam.HandleWithBody(ctx, &request); err != nil {
		core.HandleError(ctx, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(ctx, http.StatusBadRequest, "Invalid repository ID")
		return
	}

	// Verify repository ownership
	repo, err := c.userGitRepoCollectionService.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(ctx, err)
		return
	}

	if err := c.userGitRepoService.DeleteRepo(repo); err != nil {
		log.Errorf("Failed to delete repository: %v", err)
		core.ResponseErr(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Repository deleted successfully"})
}

// SyncRepo synchronizes a git repository with its remote
func (c *UserGitRepoController) SyncRepo(ctx *gin.Context) {
	// Get the authenticated user ID from the context
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", true, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("id", true, regexp.MustCompile("\\d+"))
	var request models.UpdateUserGitRepoRequest

	if err := reqParam.HandleWithBody(ctx, &request); err != nil {
		core.HandleError(ctx, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(ctx, http.StatusBadRequest, "Invalid repository ID")
		return
	}

	// Verify repository ownership
	repo, err := c.userGitRepoCollectionService.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(ctx, err)
		return
	}

	// Create an async task for the sync operation
	task, err := c.asyncTaskService.CreateTask(models.TaskTypeSync, repoIDParam.String(), userId.String())
	if err != nil {
		log.Errorf("Failed to create sync task: %v", err)
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to create sync task")
		return
	}

	lock := c.userGitRepoLockService.Acquire(repoIDParam.String())
	lock.Lock()

	defer lock.Unlock()

	// Update status to syncing
	if err := c.userGitRepoService.UpdateRepoStatus(repo, models.StatusSyncing, ""); err != nil {
		log.Errorf("Failed to update repository status: %v", err)
		core.ResponseErr(ctx, http.StatusInternalServerError, err)
		return
	}

	// Start sync in a goroutine
	go func() {
		// Update task status to running
		c.asyncTaskService.UpdateTaskStatus(task.ID, models.TaskStatusRunning, "Repository sync in progress")

		if err := c.userGitRepoService.SyncRepo(repo, ""); err != nil {
			// Update task status to failed
			log.Errorf("Failed to sync repository: %v", err)
			c.userGitRepoService.UpdateRepoStatus(repo, models.StatusFailed, err.Error())
			c.asyncTaskService.UpdateTaskStatus(task.ID, models.TaskStatusFailed, err.Error())
			return
		}
		if err := c.userGitRepoService.CheckWebHooks(repo); err != nil {
			// Update task status to failed
			log.Errorf("Failed to check webhooks: %v", err)
			c.userGitRepoService.UpdateRepoStatus(repo, models.StatusFailed, err.Error())
			c.asyncTaskService.UpdateTaskStatus(task.ID, models.TaskStatusFailed, err.Error())
			return
		}

		// The SyncRepo method now handles setting the status (synced or warning)
		// so we don't need to set it here again

		// Get the updated repository to check its status
		updatedRepo, _ := c.userGitRepoService.GetRepoByID(repoIDParam.String())

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
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", true, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("id", true, regexp.MustCompile("\\d+"))

	if err := reqParam.Handle(ctx); err != nil {
		core.HandleError(ctx, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(ctx, http.StatusBadRequest, "Invalid repository ID")
		return
	}

	// Verify repository ownership
	repo, err := c.userGitRepoCollectionService.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(ctx, err)
		return
	}

	// Get the branches
	branches, err := c.userGitRepoService.GetRepoBranches(repo)
	if err != nil {
		log.Errorf("Failed to get repository branches: %v", err)
		core.ResponseErr(ctx, http.StatusInternalServerError, err)
		return
	}

	core.ResponseOKArr(ctx, branches)
}
