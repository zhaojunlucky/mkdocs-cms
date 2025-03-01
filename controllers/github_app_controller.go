package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// GitHubAppController handles GitHub App API endpoints
type GitHubAppController struct {
	appID             int64
	privateKey        []byte
	gitRepoService    *services.UserGitRepoService
	eventService      *services.EventService
	githubAppSettings *models.GitHubAppSettings
}

// NewGitHubAppController creates a new GitHubAppController
func NewGitHubAppController(appID int64, privateKey []byte, settings *models.GitHubAppSettings) *GitHubAppController {
	return &GitHubAppController{
		appID:             appID,
		privateKey:        privateKey,
		gitRepoService:    services.NewUserGitRepoService(),
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
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
		},
	}
	client := github.NewClient(httpClient)

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
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
		},
	}
	client := github.NewClient(httpClient)

	// Get an installation token
	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installation token: " + err.Error()})
		return
	}

	// Create a new client with the installation token
	httpClient = &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: installationToken.GetToken()},
			),
		},
	}
	client = github.NewClient(httpClient)

	// Get repositories for the installation
	repos, _, err := client.Apps.ListRepos(ctx, nil)
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
	repo, err := c.gitRepoService.GetRepoByID(uint(repoID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Return the repository
	ctx.JSON(http.StatusOK, repo.ToResponse(true))
}

// CreateRepositoryFromGitHub creates a repository from GitHub
func (c *GitHubAppController) CreateRepositoryFromGitHub(ctx *gin.Context) {
	// Parse the installation ID from the URL
	installationIDStr := ctx.Param("id")
	installationID, err := strconv.ParseInt(installationIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid installation ID"})
		return
	}

	// Parse request body
	var request struct {
		RepositoryID int64 `json:"repository_id"`
		UserID       uint  `json:"user_id"`
	}
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
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
		},
	}
	client := github.NewClient(httpClient)

	// Get an installation token
	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installation token: " + err.Error()})
		return
	}

	// Create a new client with the installation token
	httpClient = &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: installationToken.GetToken()},
			),
		},
	}
	client = github.NewClient(httpClient)

	// Get repositories for the installation
	repos, _, err := client.Apps.ListRepos(ctx, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repositories: " + err.Error()})
		return
	}

	// Find the selected repository
	var selectedRepo *github.Repository
	for _, repo := range repos.Repositories {
		if repo.GetID() == request.RepositoryID {
			selectedRepo = repo
			break
		}
	}

	if selectedRepo == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Repository not found in installation"})
		return
	}

	// Create the repository in our system
	authData := models.GitHubAuthData{
		InstallationID: installationID,
	}
	authDataJSON, err := json.Marshal(authData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize auth data"})
		return
	}

	newRepo, err := c.gitRepoService.CreateRepo(models.CreateUserGitRepoRequest{
		UserID:      request.UserID,
		Name:        selectedRepo.GetName(),
		Description: selectedRepo.GetDescription(),
		RemoteURL:   selectedRepo.GetCloneURL(),
		Branch:      selectedRepo.GetDefaultBranch(),
		Provider:    "github",
		AuthType:    "github_app",
		AuthData:    string(authDataJSON),
		LocalPath:   "", // Will be set by the service
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create repository: " + err.Error()})
		return
	}

	// Log the event
	c.eventService.CreateEvent(models.CreateEventRequest{
		Level:        models.EventLevelInfo,
		Source:       models.EventSourceGitRepo,
		Message:      "GitHub repository created",
		ResourceID:   &newRepo.ID,
		ResourceType: "repository",
		Details:      fmt.Sprintf("GitHub repository %s created", selectedRepo.GetFullName()),
	})

	// Return the repository
	ctx.JSON(http.StatusCreated, newRepo.ToResponse(true))
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
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
		},
	}
	client := github.NewClient(httpClient)

	// Get an installation token
	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installation token: " + err.Error()})
		return
	}

	// Create a new client with the installation token
	httpClient = &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: installationToken.GetToken()},
			),
		},
	}
	client = github.NewClient(httpClient)

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
		newRepo, err := c.gitRepoService.CreateRepo(models.CreateUserGitRepoRequest{
			UserID:      request.UserID,
			Name:        selectedRepo.GetName(),
			Description: selectedRepo.GetDescription(),
			RemoteURL:   selectedRepo.GetCloneURL(),
			Branch:      selectedRepo.GetDefaultBranch(),
			Provider:    "github",
			AuthType:    "github_app",
			AuthData:    fmt.Sprintf(`{"installation_id": %d}`, installationID),
			LocalPath:   "", // Will be set by the service
		})

		if err != nil {
			continue
		}

		importedRepos = append(importedRepos, newRepo)

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

	// Convert to response format
	var response []models.UserGitRepoResponse
	for _, repo := range importedRepos {
		response = append(response, repo.ToResponse(true))
	}

	ctx.JSON(http.StatusOK, response)
}

// CreateWebhook creates a webhook for a repository
func (c *GitHubAppController) CreateWebhook(ctx *gin.Context) {
	installationID, err := strconv.ParseInt(ctx.Param("installation_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid installation ID"})
		return
	}

	// Parse request body
	var request struct {
		RepositoryFullName string `json:"repository_full_name"`
		WebhookURL         string `json:"webhook_url"`
		UserID             uint   `json:"user_id"`
		Branch             string `json:"branch"`
		LocalPath          string `json:"local_path"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Generate a JWT for GitHub API authentication
	token, err := c.generateJWT()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// Create a GitHub client with the JWT
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
		},
	}
	client := github.NewClient(httpClient)

	// Get an installation token
	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get installation token: " + err.Error()})
		return
	}

	// Create a new client with the installation token
	httpClient = &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: installationToken.GetToken()},
			),
		},
	}
	client = github.NewClient(httpClient)

	// Parse owner and repo from the full name
	parts := strings.Split(request.RepositoryFullName, "/")
	if len(parts) != 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository full name format"})
		return
	}
	owner := parts[0]
	repo := parts[1]

	// Create the webhook
	hook := &github.Hook{
		Config: map[string]interface{}{
			"url":          request.WebhookURL,
			"content_type": "json",
			"secret":       c.githubAppSettings.WebhookSecret,
		},
		Events: []string{"push"},
		Active: github.Bool(true),
	}

	createdHook, _, err := client.Repositories.CreateHook(ctx, owner, repo, hook)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook: " + err.Error()})
		return
	}

	// Create the git repository in our system
	authData := models.GitHubAuthData{
		InstallationID: installationID,
	}
	authDataJSON, err := json.Marshal(authData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize auth data"})
		return
	}

	newRepo, err := c.gitRepoService.CreateRepo(models.CreateUserGitRepoRequest{
		UserID:      request.UserID,
		Name:        repo,
		Description: "GitHub repository " + request.RepositoryFullName,
		RemoteURL:   "https://github.com/" + request.RepositoryFullName + ".git",
		Branch:      request.Branch,
		Provider:    "github",
		AuthType:    "github_app",
		AuthData:    string(authDataJSON),
		LocalPath:   request.LocalPath,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create repository: " + err.Error()})
		return
	}

	// Log the event
	c.eventService.CreateEvent(models.CreateEventRequest{
		Level:        models.EventLevelInfo,
		Source:       models.EventSourceGitRepo,
		Message:      "GitHub repository created",
		ResourceID:   &newRepo.ID,
		ResourceType: "repository",
		Details:      fmt.Sprintf("GitHub repository %s created", request.RepositoryFullName),
	})

	response := map[string]interface{}{
		"hook_id":   createdHook.GetID(),
		"repo_id":   newRepo.ID,
		"repo_name": newRepo.Name,
	}

	ctx.JSON(http.StatusCreated, response)
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
