package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v45/github"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// GitHubWebhookController handles webhook events from GitHub
type GitHubWebhookController struct {
	BaseController
	gitRepoService *services.UserGitRepoService
	eventService   *services.EventService
	webhookSecret  string
}

func (c *GitHubWebhookController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	c.ctx = ctx
	c.gitRepoService = ctx.MustGetService("userGitRepoService").(*services.UserGitRepoService)
	c.eventService = ctx.MustGetService("eventService").(*services.EventService)
	c.webhookSecret = ctx.GithubAppSettings.WebhookSecret

	router.POST("/webhooks/github", c.HandleWebhook)

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
		var event github.PushEvent
		if err := json.Unmarshal(body, &event); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid push event payload"})
			return
		}
		c.handlePushEvent(ctx, &event, body)
	//case "pull_request":
	//	c.handlePullRequestEvent(ctx, body)
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
func (c *GitHubWebhookController) handlePushEvent(ctx *gin.Context, event *github.PushEvent, body []byte) {
	// Extract repository information
	repoName := event.GetRepo().GetName()
	repoOwner := event.GetRepo().GetOwner().GetName()
	repoFullName := repoOwner + "/" + repoName
	branch := event.GetRef()
	if len(branch) > 11 && branch[:11] == "refs/heads/" {
		branch = branch[11:]
	}

	// Find repositories that match the remote URL
	remoteURL := "https://github.com/" + repoFullName + ".git"
	repos, err := c.gitRepoService.GetReposByURL(remoteURL)
	if err != nil {
		log.Printf("Failed to find repositories for URL %s: %v", remoteURL, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find repositories"})
		return
	}

	if len(repos) == 0 {
		log.Printf("No repositories found for URL %s", remoteURL)
		ctx.JSON(http.StatusOK, gin.H{"message": "No matching repositories found"})
		return
	}

	// Sync each matching repository
	for _, repo := range repos {
		if repo.Branch == branch {
			if err := c.gitRepoService.SyncRepository(fmt.Sprintf("%d", repo.ID)); err != nil {
				log.Printf("Failed to sync repository %d: %v", repo.ID, err)
				continue
			}

			// Log the event
			repoID := repo.ID
			c.eventService.CreateEvent(models.CreateEventRequest{
				Level:        models.EventLevelInfo,
				Source:       models.EventSourceGitRepo,
				Message:      "Repository synced due to GitHub push event",
				ResourceID:   &repoID,
				ResourceType: "repository",
				Details:      string(body),
			})
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Push event processed"})
}

// handlePullRequestEvent processes GitHub pull request events
func (c *GitHubWebhookController) handlePullRequestEvent(ctx *gin.Context, body []byte) {
	var event github.PullRequestEvent
	if err := json.Unmarshal(body, &event); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request event payload"})
		return
	}

	// Only process closed pull requests that were merged
	if event.GetAction() != "closed" || !event.GetPullRequest().GetMerged() {
		ctx.JSON(http.StatusOK, gin.H{"message": "Pull request not merged, no action needed"})
		return
	}

	// Extract repository information
	repoName := event.GetRepo().GetName()
	repoOwner := event.GetRepo().GetOwner().GetLogin()
	repoFullName := repoOwner + "/" + repoName
	targetBranch := event.GetPullRequest().GetBase().GetRef()

	// Find repositories that match the remote URL
	remoteURL := "https://github.com/" + repoFullName + ".git"
	repos, err := c.gitRepoService.GetReposByURL(remoteURL)
	if err != nil {
		log.Printf("Failed to find repositories for URL %s: %v", remoteURL, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find repositories"})
		return
	}

	if len(repos) == 0 {
		log.Printf("No repositories found for URL %s", remoteURL)
		ctx.JSON(http.StatusOK, gin.H{"message": "No matching repositories found"})
		return
	}

	// Sync each matching repository
	for _, repo := range repos {
		if repo.Branch == targetBranch {
			if err := c.gitRepoService.SyncRepository(fmt.Sprintf("%d", repo.ID)); err != nil {
				log.Printf("Failed to sync repository %d: %v", repo.ID, err)
				continue
			}

			// Log the event
			repoID := repo.ID
			c.eventService.CreateEvent(models.CreateEventRequest{
				Level:        models.EventLevelInfo,
				Source:       models.EventSourceGitRepo,
				Message:      "Repository synced due to GitHub pull request merge",
				ResourceID:   &repoID,
				ResourceType: "repository",
				Details:      string(body),
			})
		}

	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Pull request event processed"})
}

// handleInstallationEvent processes GitHub App installation events
func (c *GitHubWebhookController) handleInstallationEvent(ctx *gin.Context, body []byte) {
	var event github.InstallationEvent
	if err := json.Unmarshal(body, &event); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid installation event payload"})
		return
	}

	// Log the installation event
	c.eventService.CreateEvent(models.CreateEventRequest{
		Level:        models.EventLevelInfo,
		Source:       models.EventSourceAPI,
		Message:      "GitHub App " + event.GetAction(),
		ResourceType: "github_app",
		Details:      string(body),
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "Installation event processed"})
}

// handleInstallationRepositoriesEvent processes GitHub App installation repositories events
func (c *GitHubWebhookController) handleInstallationRepositoriesEvent(ctx *gin.Context, body []byte) {
	var event github.InstallationRepositoriesEvent
	if err := json.Unmarshal(body, &event); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid installation repositories event payload"})
		return
	}

	// Log the installation repositories event
	c.eventService.CreateEvent(models.CreateEventRequest{
		Level:        models.EventLevelInfo,
		Source:       models.EventSourceAPI,
		Message:      "GitHub App " + event.GetAction() + " repositories",
		ResourceType: "github_app",
		Details:      string(body),
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "Installation repositories event processed"})
}
