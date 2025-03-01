package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// GitHubWebhookController handles webhook events from GitHub
type GitHubWebhookController struct {
	gitRepoService *services.GitRepoService
	eventService   *services.EventService
	webhookSecret  string
}

// NewGitHubWebhookController creates a new GitHubWebhookController
func NewGitHubWebhookController(webhookSecret string) *GitHubWebhookController {
	return &GitHubWebhookController{
		gitRepoService: services.NewGitRepoService(),
		eventService:   services.NewEventService(),
		webhookSecret:  webhookSecret,
	}
}

// HandleWebhook processes incoming GitHub webhook events
func (c *GitHubWebhookController) HandleWebhook(ctx *gin.Context) {
	// Verify the webhook signature
	signature := ctx.GetHeader("X-Hub-Signature-256")
	if !c.verifySignature(ctx.Request, signature) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// Get the event type
	eventType := ctx.GetHeader("X-GitHub-Event")
	if eventType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-GitHub-Event header"})
		return
	}

	// Read the request body
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	// Process the webhook event based on its type
	switch eventType {
	case "push":
		c.handlePushEvent(ctx, body)
	case "pull_request":
		c.handlePullRequestEvent(ctx, body)
	case "installation":
		c.handleInstallationEvent(ctx, body)
	case "installation_repositories":
		c.handleInstallationRepositoriesEvent(ctx, body)
	default:
		// Log the event but return a success response
		log.Printf("Received unhandled GitHub event: %s", eventType)
		ctx.JSON(http.StatusOK, gin.H{"message": "Event received but not processed"})
	}
}

// verifySignature verifies the GitHub webhook signature
func (c *GitHubWebhookController) verifySignature(r *http.Request, signature string) bool {
	if c.webhookSecret == "" || signature == "" {
		return false
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}

	// Important: Restore the request body for later use
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	// Calculate the HMAC
	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	mac.Write(body)
	expectedMAC := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}

// handlePushEvent processes GitHub push events
func (c *GitHubWebhookController) handlePushEvent(ctx *gin.Context, body []byte) {
	var event models.GitHubPushEvent
	if err := json.Unmarshal(body, &event); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid push event payload"})
		return
	}

	// Extract repository information
	repoFullName := event.Repository.FullName
	repoURL := event.Repository.CloneURL
	branch := strings.TrimPrefix(event.Ref, "refs/heads/")

	// Find repositories in our system that match this GitHub repository
	repos, err := c.gitRepoService.GetReposByURL(repoURL)
	if err != nil || len(repos) == 0 {
		log.Printf("No matching repository found for %s", repoFullName)
		ctx.JSON(http.StatusOK, gin.H{"message": "No matching repository found"})
		return
	}

	// For each matching repository, sync it
	for _, repo := range repos {
		// Only sync if the branch matches
		if repo.Branch == branch {
			if err := c.gitRepoService.SyncRepository(repo.ID); err != nil {
				log.Printf("Failed to sync repository %d: %v", repo.ID, err)
				continue
			}

			// Log the event
			c.eventService.CreateEvent(models.CreateEventRequest{
				ResourceType: "repository",
				ResourceID:   int(repo.ID),
				EventType:    "github_push",
				Message:      "Repository synced due to GitHub push event",
				Data:         string(body),
			})
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Push event processed"})
}

// handlePullRequestEvent processes GitHub pull request events
func (c *GitHubWebhookController) handlePullRequestEvent(ctx *gin.Context, body []byte) {
	var event models.GitHubPullRequestEvent
	if err := json.Unmarshal(body, &event); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request event payload"})
		return
	}

	// Only process closed pull requests that were merged
	if event.Action != "closed" || !event.PullRequest.Merged {
		ctx.JSON(http.StatusOK, gin.H{"message": "Pull request event ignored"})
		return
	}

	// Extract repository information
	repoFullName := event.Repository.FullName
	repoURL := event.Repository.CloneURL
	branch := event.PullRequest.Base.Ref

	// Find repositories in our system that match this GitHub repository
	repos, err := c.gitRepoService.GetReposByURL(repoURL)
	if err != nil || len(repos) == 0 {
		log.Printf("No matching repository found for %s", repoFullName)
		ctx.JSON(http.StatusOK, gin.H{"message": "No matching repository found"})
		return
	}

	// For each matching repository, sync it
	for _, repo := range repos {
		// Only sync if the branch matches
		if repo.Branch == branch {
			if err := c.gitRepoService.SyncRepository(repo.ID); err != nil {
				log.Printf("Failed to sync repository %d: %v", repo.ID, err)
				continue
			}

			// Log the event
			c.eventService.CreateEvent(models.CreateEventRequest{
				ResourceType: "repository",
				ResourceID:   int(repo.ID),
				EventType:    "github_pull_request_merged",
				Message:      "Repository synced due to GitHub pull request merge",
				Data:         string(body),
			})
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Pull request event processed"})
}

// handleInstallationEvent processes GitHub App installation events
func (c *GitHubWebhookController) handleInstallationEvent(ctx *gin.Context, body []byte) {
	var event models.GitHubInstallationEvent
	if err := json.Unmarshal(body, &event); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid installation event payload"})
		return
	}

	// Log the installation event
	c.eventService.CreateEvent(models.CreateEventRequest{
		ResourceType: "github_app",
		ResourceID:   0, // No specific resource ID for app installation
		EventType:    "github_app_" + event.Action,
		Message:      "GitHub App " + event.Action,
		Data:         string(body),
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "Installation event processed"})
}

// handleInstallationRepositoriesEvent processes GitHub App installation repositories events
func (c *GitHubWebhookController) handleInstallationRepositoriesEvent(ctx *gin.Context, body []byte) {
	var event models.GitHubInstallationRepositoriesEvent
	if err := json.Unmarshal(body, &event); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid installation repositories event payload"})
		return
	}

	// Log the installation repositories event
	c.eventService.CreateEvent(models.CreateEventRequest{
		ResourceType: "github_app",
		ResourceID:   0, // No specific resource ID for app installation repositories
		EventType:    "github_app_" + event.Action + "_repositories",
		Message:      "GitHub App " + event.Action + " repositories",
		Data:         string(body),
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "Installation repositories event processed"})
}
