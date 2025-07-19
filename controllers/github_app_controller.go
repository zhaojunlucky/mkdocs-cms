package controllers

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v45/github"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
	"golang.org/x/oauth2"
)

// GitHubAppController handles GitHub App related operations
type GitHubAppController struct {
	BaseController
	appID              int64
	privateKey         []byte
	eventService       *services.EventService
	githubAppSettings  *models.GitHubAppSettings
	userGitRepoService *services.UserGitRepoService
	userService        *services.UserService
}

func (c *GitHubAppController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	c.ctx = ctx
	c.userGitRepoService = ctx.MustGetService("userGitRepoService").(*services.UserGitRepoService)
	c.eventService = ctx.MustGetService("eventService").(*services.EventService)
	c.githubAppSettings = ctx.GithubAppSettings
	c.userService = ctx.MustGetService("userService").(*services.UserService)

	c.initConfig()
	gh := router.Group("/github")
	{
		gh.GET("/app", c.GetAppInfo)
		gh.GET("/installations", c.GetInstallations)
		gh.GET("/installations/:installation_id/repositories", c.GetInstallationRepositories)
		gh.POST("/installations/:installation_id/import", c.ImportRepositories)
	}

}

// NewGitHubAppController creates a new GitHubAppController
func (c *GitHubAppController) initConfig() {
	c.appID = c.ctx.GithubAppSettings.AppID
	bytes, err := os.ReadFile(c.ctx.Config.GitHub.App.PrivateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key file: %v", err)
	}
	c.privateKey = bytes
}

// GetAppInfo returns information about the GitHub App
func (c *GitHubAppController) GetAppInfo(ctx *gin.Context) {
	reqParam := core.NewRequestParam()
	_ = reqParam.AddContextParam("userId", true, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	if err := reqParam.Handle(ctx); err != nil {
		core.HandleError(ctx, err)
		return
	}

	appInfo := map[string]interface{}{
		"app_id":      c.appID,
		"name":        c.githubAppSettings.AppName,
		"description": c.githubAppSettings.Description,
		"homepage":    c.githubAppSettings.HomepageURL,
		"html_url":    fmt.Sprintf("https://github.com/apps/%s", c.githubAppSettings.AppName),
	}

	ctx.JSON(http.StatusOK, appInfo)
}

// GetInstallations returns installations of the GitHub App for the current user
func (c *GitHubAppController) GetInstallations(ctx *gin.Context) {
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	if err := reqParam.Handle(ctx); err != nil {
		core.HandleError(ctx, err)
		return
	}

	// Get all installations
	installations, _, err := c.ctx.GithubAppClient.Apps.ListInstallations(ctx, nil)
	if err != nil {
		log.Errorf("Failed to get installations: %v", err)
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to get installations: "+err.Error())
		return
	}

	// Get user's GitHub username from database
	user, err := c.userService.GetUserByID(fmt.Sprintf("%v", userId.String()))
	if err != nil {
		log.Errorf("Failed to get user information: %v", err)
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to get user information: "+err.Error())
		return
	}

	// Convert to a simpler format for the response and filter by user
	response := make([]map[string]interface{}, 0)
	for _, inst := range installations {
		// Only include installations for the current user's GitHub account
		// For GitHub provider, the username is typically the GitHub username
		if user.Provider == "github" && strings.EqualFold(inst.GetAccount().GetLogin(), user.Username) {
			installation := map[string]interface{}{
				"id": inst.GetID(),
				"account": map[string]interface{}{
					"login":      inst.GetAccount().GetLogin(),
					"avatar_url": inst.GetAccount().GetAvatarURL(),
				},
				"created_at": inst.GetCreatedAt().Format(time.RFC3339),
				"updated_at": inst.GetUpdatedAt().Format(time.RFC3339),
			}
			response = append(response, installation)
		}
	}

	core.ResponseOKArr(ctx, response)
}

// GetInstallationRepositories returns repositories for a specific installation
func (c *GitHubAppController) GetInstallationRepositories(ctx *gin.Context) {
	// Get current user from context
	reqParam := core.NewRequestParam()
	userID := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	if err := reqParam.Handle(ctx); err != nil {
		core.HandleError(ctx, err)
		return
	}

	installationID, err := strconv.ParseInt(ctx.Param("installation_id"), 10, 64)
	if err != nil {
		log.Errorf("Invalid installation ID: %v", err)
		core.ResponseErrStr(ctx, http.StatusBadRequest, "Invalid installation ID")
		return
	}

	// Verify that the installation belongs to the user
	installations, _, err := c.ctx.GithubAppClient.Apps.ListInstallations(ctx, nil)
	if err != nil {
		log.Errorf("Failed to get installations: %v", err)
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to get installations: "+err.Error())
		return
	}

	// Get user's GitHub username from database
	user, err := c.userService.GetUserByID(fmt.Sprintf("%v", userID.String()))
	if err != nil {
		log.Errorf("Failed to get user information: %v", err)
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to get user information: "+err.Error())
		return
	}

	// Check if the requested installation belongs to the user
	installationBelongsToUser := false
	for _, inst := range installations {
		if inst.GetID() == installationID &&
			user.Provider == "github" &&
			strings.EqualFold(inst.GetAccount().GetLogin(), user.Username) {
			installationBelongsToUser = true
			break
		}
	}

	if !installationBelongsToUser {
		log.Errorf("Installation does not belong to the authenticated user")
		core.ResponseErrStr(ctx, http.StatusForbidden, "Installation does not belong to the authenticated user")
		return
	}

	// Get an installation token
	installationToken, _, err := c.ctx.GithubAppClient.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		log.Errorf("Failed to get installation token: %v", err)
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to get installation token: "+err.Error())
		return
	}

	// Create a new client with the installation token
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: installationToken.GetToken()},
			),
		},
	}
	client := github.NewClient(httpClient)

	// Get repositories for the installation
	repos, _, err := client.Apps.ListRepos(ctx, nil)
	if err != nil {
		log.Errorf("Failed to get repositories: %v", err)
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to get repositories: "+err.Error())
		return
	}

	userExistingRepos, err := c.userGitRepoService.GetReposByUser(userID.String())
	if err != nil {
		log.Errorf("Failed to get user existing repos: %v", err)
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to get user existing repos: "+err.Error())
		return
	}
	var existingRepoIDs map[int64]string = make(map[int64]string)

	for _, repo := range userExistingRepos {
		existingRepoIDs[repo.GitRepoID] = repo.Name
	}

	// Convert to a simpler format for the response
	var response []map[string]interface{}
	for _, repo := range repos.Repositories {
		if existingRepoIDs[repo.GetID()] != "" {
			continue
		}
		repository := map[string]interface{}{
			"id":             repo.GetID(),
			"name":           repo.GetName(),
			"full_name":      repo.GetFullName(),
			"private":        repo.GetPrivate(),
			"html_url":       repo.GetHTMLURL(),
			"clone_url":      repo.GetCloneURL(),
			"default_branch": repo.GetDefaultBranch(),
			"created_at":     repo.GetCreatedAt().Format(time.RFC3339),
			"updated_at":     repo.GetUpdatedAt().Format(time.RFC3339),
		}
		response = append(response, repository)
	}

	core.ResponseOKArr(ctx, response)
}

// ImportRepositories imports repositories from a GitHub App installation
func (c *GitHubAppController) ImportRepositories(ctx *gin.Context) {
	reqParam := core.NewRequestParam()
	userID := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	if err := reqParam.Handle(ctx); err != nil {
		core.HandleError(ctx, err)
		return
	}

	installationID, err := strconv.ParseInt(ctx.Param("installation_id"), 10, 64)
	if err != nil {
		log.Errorf("Invalid installation ID: %v", err)
		core.ResponseErrStr(ctx, http.StatusBadRequest, "Invalid installation ID")
		return
	}

	var request models.ImportRepositoriesRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Errorf("Invalid request: %v", err)
		core.ResponseErr(ctx, http.StatusBadRequest, err)
		return
	}
	request.UserID = userID.String()
	user, err := c.userService.GetUserByID(userID.String())
	if err != nil {
		log.Errorf("Failed to get user information: %v", err)
		core.ResponseErrStr(ctx, http.StatusBadRequest, "Failed to get user information: "+err.Error())
		return
	}
	role := user.GetQuotaRole()
	if role == nil {
		log.Errorf("Failed to get user quota role: %v", err)
		core.ResponseErrStr(ctx, http.StatusForbidden, "user quota role not found")
		return
	}
	var roleQuota models.UserRoleQuota
	err = database.DB.Preload("Role").Where("role_id = ?", role.ID).First(&roleQuota).Error
	if err != nil {
		log.Errorf("Failed to get user quota role: %v", err)
		core.ResponseErrStr(ctx, http.StatusForbidden, "user quota role not found")
		return
	}
	var userGitRepos []models.UserGitRepo
	err = database.DB.Where("user_id = ?", userID.String()).Find(&userGitRepos).Error
	if err != nil {
		log.Errorf("Failed to list user git repo: %v", err)
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to list user git repo: "+err.Error())
		return
	}
	if roleQuota.RepoCount != 0 && len(userGitRepos) >= roleQuota.RepoCount {
		log.Errorf("User quota exceeded: %v", err)
		core.ResponseErrStr(ctx, http.StatusForbidden, "user quota exceeded, contact admin")
		return
	}

	// Get an installation token
	installationToken, _, err := c.ctx.GithubAppClient.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		log.Errorf("Failed to get installation token: %v", err)
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to get installation token: "+err.Error())
		return
	}

	// Create a new client with the installation token
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: installationToken.GetToken()},
			),
		},
	}
	client := github.NewClient(httpClient)

	// Get repositories for the installation
	repos, _, err := client.Apps.ListRepos(ctx, nil)
	if err != nil {
		log.Errorf("Failed to get repositories: %v", err)
		core.ResponseErrStr(ctx, http.StatusInternalServerError, "Failed to get repositories: "+err.Error())
		return
	}

	// Import the selected repositories
	var importedRepos []models.UserGitRepo
	for _, repoID := range request.RepositoryIDs {
		// Find the repository in the list
		var selectedRepo *github.Repository
		for _, repo := range repos.Repositories {
			if repo.GetID() == repoID {
				selectedRepo = repo
				break
			}
		}

		if selectedRepo == nil {
			continue
		}

		// Create the repository in our system
		newRepo := &models.UserGitRepo{
			UserID:         request.UserID,
			Name:           selectedRepo.GetName(),
			Description:    selectedRepo.GetDescription(),
			RemoteURL:      selectedRepo.GetCloneURL(),
			Branch:         selectedRepo.GetDefaultBranch(),
			Provider:       "github",
			AuthType:       "github_app",
			InstallationID: installationID,
			GitRepoID:      *selectedRepo.ID,
			AuthData:       fmt.Sprintf(`{"installation_id": %d}`, installationID),
			LocalPath:      "", // Will be set by the service
		}

		err = c.userGitRepoService.CreateRepo(newRepo)
		if err != nil {
			log.Warnf("Failed to create repository: %v", err)
			continue
		}

		importedRepos = append(importedRepos, *newRepo)

		// Log the event
		c.eventService.CreateEvent(models.CreateEventRequest{
			Level:        models.EventLevelInfo,
			Source:       models.EventSourceGitRepo,
			Message:      "GitHub repository imported",
			ResourceID:   &newRepo.ID,
			ResourceType: "repository",
			Details:      fmt.Sprintf("GitHub repository %s imported", selectedRepo.GetFullName()),
		})
	}
	go func() {
		// Sync the imported repositories
		for _, repo := range importedRepos {
			log.Infof("Syncing repository %d", repo.ID)
			err := c.userGitRepoService.SyncRepo(&repo, "")
			if err != nil {
				log.Errorf("Failed to sync repository %d: %v", repo.ID, err)
				c.userGitRepoService.UpdateRepoStatus(&repo, models.StatusFailed, err.Error())
			}

			if err := c.userGitRepoService.CheckWebHooks(&repo); err != nil {
				// Update task status to failed
				log.Errorf("Failed to check webhooks: %v", err)
				c.userGitRepoService.UpdateRepoStatus(&repo, models.StatusFailed, err.Error())
				return
			}
		}
	}()
	// Convert to response format
	var response []models.UserGitRepoResponse
	for _, repo := range importedRepos {
		response = append(response, repo.ToResponse(true))
	}

	core.ResponseOKArr(ctx, response)
}
