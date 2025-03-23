package controllers

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"

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
	// Get current user from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get all installations
	installations, _, err := c.ctx.GithubAppClient.Apps.ListInstallations(ctx, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installations: " + err.Error()})
		return
	}

	// Get user's GitHub username from database
	user, err := c.userService.GetUserByID(fmt.Sprintf("%v", userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information: " + err.Error()})
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

	ctx.JSON(http.StatusOK, response)
}

// GetInstallationRepositories returns repositories for a specific installation
func (c *GitHubAppController) GetInstallationRepositories(ctx *gin.Context) {
	// Get current user from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	installationID, err := strconv.ParseInt(ctx.Param("installation_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid installation ID"})
		return
	}

	// Verify that the installation belongs to the user
	installations, _, err := c.ctx.GithubAppClient.Apps.ListInstallations(ctx, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installations: " + err.Error()})
		return
	}

	// Get user's GitHub username from database
	user, err := c.userService.GetUserByID(fmt.Sprintf("%v", userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information: " + err.Error()})
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
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Installation does not belong to the authenticated user"})
		return
	}

	// Get an installation token
	installationToken, _, err := c.ctx.GithubAppClient.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installation token: " + err.Error()})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repositories: " + err.Error()})
		return
	}

	userExistingRepos, err := c.userGitRepoService.GetReposByUser(userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	ctx.JSON(http.StatusOK, core.EnsureNonNilArr(response))
}

// GetRepositoryByID returns a specific repository by ID
func (c *GitHubAppController) GetRepositoryByID(ctx *gin.Context) {
	// Parse the repository ID from the URL
	repoIDStr := ctx.Param("id")
	repoID, err := strconv.ParseUint(repoIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	// Get the repository from the database
	repo, err := c.userGitRepoService.GetRepoByID(strconv.FormatUint(repoID, 10))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Return the repository
	ctx.JSON(http.StatusOK, repo.ToResponse(true))
}

// ImportRepositories imports repositories from a GitHub App installation
func (c *GitHubAppController) ImportRepositories(ctx *gin.Context) {
	authenticatedUserID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	installationID, err := strconv.ParseInt(ctx.Param("installation_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid installation ID"})
		return
	}

	var request models.ImportRepositoriesRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	request.UserID = authenticatedUserID.(string)

	// Get an installation token
	installationToken, _, err := c.ctx.GithubAppClient.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installation token: " + err.Error()})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repositories: " + err.Error()})
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

		hook := &github.Hook{
			Config: map[string]interface{}{
				"url":          c.githubAppSettings.WebhookURL,
				"content_type": "json",
				"secret":       c.githubAppSettings.WebhookSecret,
			},
			Events: []string{"push"},
			Active: github.Bool(true),
		}

		parts := strings.Split(*selectedRepo.FullName, "/")
		createdHook, _, err := client.Repositories.CreateHook(context.Background(), parts[0], parts[1], hook)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook: " + err.Error()})
			return
		}
		c.eventService.CreateEvent(models.CreateEventRequest{
			Level:        models.EventLevelInfo,
			Source:       models.EventSourceGitRepo,
			Message:      fmt.Sprintf("GitHub repository webhook created: %d", createdHook.ID),
			ResourceID:   &newRepo.ID,
			ResourceType: "repository",
			Details:      fmt.Sprintf("GitHub repository webhook %d created", createdHook.ID),
		})
	}

	// Convert to response format
	var response []models.UserGitRepoResponse
	for _, repo := range importedRepos {
		response = append(response, repo.ToResponse(true))
	}

	ctx.JSON(http.StatusOK, response)
}
