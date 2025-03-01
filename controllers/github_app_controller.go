package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v45/github"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// GitHubAppController handles GitHub App API endpoints
type GitHubAppController struct {
	appID             int64
	privateKey        []byte
	gitRepoService    *services.GitRepoService
	eventService      *services.EventService
	githubAppSettings *models.GitHubAppSettings
}

// NewGitHubAppController creates a new GitHubAppController
func NewGitHubAppController(appID int64, privateKey []byte, settings *models.GitHubAppSettings) *GitHubAppController {
	return &GitHubAppController{
		appID:             appID,
		privateKey:        privateKey,
		gitRepoService:    services.NewGitRepoService(),
		eventService:      services.NewEventService(),
		githubAppSettings: settings,
	}
}

// GetAppInfo returns information about the GitHub App
func (c *GitHubAppController) GetAppInfo(ctx *gin.Context) {
	appInfo := map[string]interface{}{
		"app_id":      c.appID,
		"app_name":    c.githubAppSettings.AppName,
		"description": c.githubAppSettings.Description,
		"homepage":    c.githubAppSettings.HomepageURL,
	}

	ctx.JSON(http.StatusOK, appInfo)
}

// GetInstallations returns all installations of the GitHub App
func (c *GitHubAppController) GetInstallations(ctx *gin.Context) {
	// Generate a JWT for GitHub API authentication
	token, err := c.generateJWT()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// Create a GitHub client with the JWT
	client := github.NewClient(nil).WithAuthToken(token)

	// Get all installations
	installations, _, err := client.Apps.ListInstallations(ctx, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installations: " + err.Error()})
		return
	}

	// Convert to a simpler format for the response
	var response []map[string]interface{}
	for _, inst := range installations {
		installation := map[string]interface{}{
			"id":           inst.GetID(),
			"account":      inst.GetAccount().GetLogin(),
			"account_type": inst.GetAccount().GetType(),
			"created_at":   inst.GetCreatedAt().Format(time.RFC3339),
			"updated_at":   inst.GetUpdatedAt().Format(time.RFC3339),
		}
		response = append(response, installation)
	}

	ctx.JSON(http.StatusOK, response)
}

// GetInstallationRepositories returns repositories for a specific installation
func (c *GitHubAppController) GetInstallationRepositories(ctx *gin.Context) {
	installationID, err := strconv.ParseInt(ctx.Param("installation_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid installation ID"})
		return
	}

	// Generate a JWT for GitHub API authentication
	token, err := c.generateJWT()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// Create a GitHub client with the JWT
	client := github.NewClient(nil).WithAuthToken(token)

	// Get an installation token
	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installation token: " + err.Error()})
		return
	}

	// Create a new client with the installation token
	installationClient := github.NewClient(nil).WithAuthToken(installationToken.GetToken())

	// Get repositories for the installation
	repos, _, err := installationClient.Apps.ListRepos(ctx, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repositories: " + err.Error()})
		return
	}

	// Convert to a simpler format for the response
	var response []map[string]interface{}
	for _, repo := range repos.Repositories {
		repository := map[string]interface{}{
			"id":            repo.GetID(),
			"name":          repo.GetName(),
			"full_name":     repo.GetFullName(),
			"private":       repo.GetPrivate(),
			"html_url":      repo.GetHTMLURL(),
			"clone_url":     repo.GetCloneURL(),
			"default_branch": repo.GetDefaultBranch(),
			"created_at":    repo.GetCreatedAt().Format(time.RFC3339),
			"updated_at":    repo.GetUpdatedAt().Format(time.RFC3339),
		}
		response = append(response, repository)
	}

	ctx.JSON(http.StatusOK, response)
}

// ImportRepositories imports repositories from a GitHub App installation
func (c *GitHubAppController) ImportRepositories(ctx *gin.Context) {
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

	// Generate a JWT for GitHub API authentication
	token, err := c.generateJWT()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// Create a GitHub client with the JWT
	client := github.NewClient(nil).WithAuthToken(token)

	// Get an installation token
	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installation token: " + err.Error()})
		return
	}

	// Create a new client with the installation token
	installationClient := github.NewClient(nil).WithAuthToken(installationToken.GetToken())

	// Get repositories for the installation
	repos, _, err := installationClient.Apps.ListRepos(ctx, nil)
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
		newRepo, err := c.gitRepoService.CreateRepo(models.CreateGitRepoRequest{
			UserID:      request.UserID,
			Name:        selectedRepo.GetName(),
			Description: selectedRepo.GetDescription(),
			URL:         selectedRepo.GetCloneURL(),
			Branch:      selectedRepo.GetDefaultBranch(),
			LocalPath:   "", // Will be set by the service
			AuthType:    "github_app",
			AuthData:    fmt.Sprintf(`{"installation_id": %d}`, installationID),
		})

		if err != nil {
			continue
		}

		importedRepos = append(importedRepos, newRepo)

		// Log the event
		c.eventService.CreateEvent(models.CreateEventRequest{
			ResourceType: "repository",
			ResourceID:   int(newRepo.ID),
			EventType:    "github_repo_imported",
			Message:      "Repository imported from GitHub App installation",
			Data:         fmt.Sprintf(`{"repository_id": %d, "installation_id": %d}`, repoID, installationID),
		})
	}

	// Convert to response format
	var response []models.GitRepoResponse
	for _, repo := range importedRepos {
		response = append(response, repo.ToResponse())
	}

	ctx.JSON(http.StatusOK, response)
}

// generateJWT generates a JWT for GitHub App authentication
func (c *GitHubAppController) generateJWT() (string, error) {
	// Parse the private key
	key, err := jwt.ParseRSAPrivateKeyFromPEM(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	// Create the JWT
	now := time.Now()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    strconv.FormatInt(c.appID, 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %v", err)
	}

	return signedToken, nil
}

// CreateWebhook creates a webhook for a repository
func (c *GitHubAppController) CreateWebhook(ctx *gin.Context) {
	installationID, err := strconv.ParseInt(ctx.Param("installation_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid installation ID"})
		return
	}

	var request models.CreateWebhookRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a JWT for GitHub API authentication
	token, err := c.generateJWT()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// Create a GitHub client with the JWT
	client := github.NewClient(nil).WithAuthToken(token)

	// Get an installation token
	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installation token: " + err.Error()})
		return
	}

	// Create a new client with the installation token
	installationClient := github.NewClient(nil).WithAuthToken(installationToken.GetToken())

	// Parse owner and repo from the full name
	parts := request.RepositoryFullName
	if len(parts) != 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository full name"})
		return
	}
	owner := parts[0]
	repo := parts[1]

	// Create the webhook
	hook := &github.Hook{
		Config: map[string]interface{}{
			"url":          request.WebhookURL,
			"content_type": "json",
			"secret":       request.Secret,
		},
		Events: request.Events,
		Active: github.Bool(true),
	}

	createdHook, _, err := installationClient.Repositories.CreateHook(ctx, owner, repo, hook)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook: " + err.Error()})
		return
	}

	// Log the event
	c.eventService.CreateEvent(models.CreateEventRequest{
		ResourceType: "repository",
		ResourceID:   0, // No specific repository ID in our system
		EventType:    "github_webhook_created",
		Message:      "GitHub webhook created",
		Data:         fmt.Sprintf(`{"repository": "%s/%s", "webhook_id": %d}`, owner, repo, createdHook.GetID()),
	})

	response := map[string]interface{}{
		"id":           createdHook.GetID(),
		"url":          createdHook.GetURL(),
		"events":       createdHook.Events,
		"active":       createdHook.GetActive(),
		"created_at":   createdHook.GetCreatedAt().Format(time.RFC3339),
		"updated_at":   createdHook.GetUpdatedAt().Format(time.RFC3339),
	}

	ctx.JSON(http.StatusCreated, response)
}
