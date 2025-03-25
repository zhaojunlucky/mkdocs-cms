package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"io"
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
	gitRepoService         *services.UserGitRepoService
	eventService           *services.EventService
	webhookSecret          string
	userGitRepoLockService *services.UserGitRepoLockService
}

func (c *GitHubWebhookController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	c.ctx = ctx
	c.gitRepoService = ctx.MustGetService("userGitRepoService").(*services.UserGitRepoService)
	c.eventService = ctx.MustGetService("eventService").(*services.EventService)
	c.webhookSecret = ctx.GithubAppSettings.WebhookSecret
	c.userGitRepoLockService = ctx.MustGetService("userGitRepoLockService").(*services.UserGitRepoLockService)

	router.POST("/github/webhook", c.HandleWebhook)

}

// HandleWebhook processes incoming GitHub webhook events
func (c *GitHubWebhookController) HandleWebhook(ctx *gin.Context) {
	// Verify the webhook signature
	signature := ctx.GetHeader("X-Hub-Signature-256")
	if !c.verifySignature(ctx.Request, signature) {
		log.Error("Invalid webhook signature")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// Get the event type
	eventType := ctx.GetHeader("X-GitHub-Event")
	if eventType == "" {
		log.Error("Missing X-GitHub-Event header")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-GitHub-Event header"})
		return
	}

	// Read the request body
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Errorf("Failed to read request body: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	// Process the webhook event based on its type
	switch eventType {
	case "push":
		var event github.PushEvent
		if err := json.Unmarshal(body, &event); err != nil {
			log.Errorf("Invalid push event payload: %v", err)
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
		log.Warnf("Received unhandled GitHub event: %s", eventType)
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
		log.Errorf("Failed to find repositories for URL %s: %v", remoteURL, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find repositories"})
		return
	}

	if len(repos) == 0 {
		log.Warnf("No repositories found for URL %s", remoteURL)
		ctx.JSON(http.StatusOK, gin.H{"message": "No matching repositories found"})
		return
	}

	// Sync each matching repository
	for _, repo := range repos {
		if repo.Branch == branch {
			c.syncRepo(repo, body, *event.After)
		}
	}
	log.Infof("Push event processed for repository %s", repoFullName)
	ctx.JSON(http.StatusOK, gin.H{"message": "Push event processed"})
}

func (c *GitHubWebhookController) syncRepo(repo models.UserGitRepo, body []byte, commitID string) {
	id := fmt.Sprintf("%d", repo.ID)
	lock := c.userGitRepoLockService.Acquire(id)
	lock.Lock()

	defer lock.Unlock()
	if err := c.gitRepoService.SyncRepository(id, commitID); err != nil {
		log.Errorf("Failed to sync repository %s: %v", id, err)
		return
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
		log.Errorf("Failed to find repositories for URL %s: %v", remoteURL, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find repositories"})
		return
	}

	if len(repos) == 0 {
		log.Errorf("No repositories found for URL %s", remoteURL)
		ctx.JSON(http.StatusOK, gin.H{"message": "No matching repositories found"})
		return
	}

	// Sync each matching repository
	for _, repo := range repos {
		if repo.Branch == targetBranch {
			if err := c.gitRepoService.SyncRepository(fmt.Sprintf("%d", repo.ID), ""); err != nil {
				log.Errorf("Failed to sync repository %d: %v", repo.ID, err)
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
