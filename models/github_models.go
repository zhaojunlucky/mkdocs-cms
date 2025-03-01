package models

import "time"

// GitHubAppSettings contains configuration for the GitHub App
type GitHubAppSettings struct {
	AppID       int64  `json:"app_id"`
	AppName     string `json:"app_name"`
	Description string `json:"description"`
	HomepageURL string `json:"homepage_url"`
	WebhookURL  string `json:"webhook_url"`
	WebhookSecret string `json:"webhook_secret"`
	PrivateKeyPath string `json:"private_key_path"`
}

// ImportRepositoriesRequest represents a request to import repositories from GitHub
type ImportRepositoriesRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	RepositoryIDs []string `json:"repository_ids" binding:"required"`
}

// CreateWebhookRequest represents a request to create a webhook for a GitHub repository
type CreateWebhookRequest struct {
	RepositoryFullName []string `json:"repository_full_name" binding:"required"`
	WebhookURL         string   `json:"webhook_url" binding:"required"`
	Secret             string   `json:"secret" binding:"required"`
	Events             []string `json:"events" binding:"required"`
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Private       bool      `json:"private"`
	HTMLURL       string    `json:"html_url"`
	CloneURL      string    `json:"clone_url"`
	DefaultBranch string    `json:"default_branch"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GitHubUser represents a GitHub user or organization
type GitHubUser struct {
	ID        string `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	Type      string `json:"type"`
}

// GitHubPushEvent represents a GitHub push event
type GitHubPushEvent struct {
	Ref        string           `json:"ref"`
	Before     string           `json:"before"`
	After      string           `json:"after"`
	Repository GitHubRepository `json:"repository"`
	Pusher     struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
	Sender GitHubUser `json:"sender"`
	Commits []struct {
		ID        string    `json:"id"`
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commits"`
}

// GitHubPullRequestEvent represents a GitHub pull request event
type GitHubPullRequestEvent struct {
	Action      string           `json:"action"`
	Number      int              `json:"number"`
	PullRequest struct {
		URL       string    `json:"url"`
		ID        string    `json:"id"`
		Number    int       `json:"number"`
		State     string    `json:"state"`
		Locked    bool      `json:"locked"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		ClosedAt  time.Time `json:"closed_at"`
		MergedAt  time.Time `json:"merged_at"`
		Merged    bool      `json:"merged"`
		Base      struct {
			Ref  string           `json:"ref"`
			Repo GitHubRepository `json:"repo"`
		} `json:"base"`
		Head struct {
			Ref  string           `json:"ref"`
			Repo GitHubRepository `json:"repo"`
		} `json:"head"`
		User GitHubUser `json:"user"`
	} `json:"pull_request"`
	Repository GitHubRepository `json:"repository"`
	Sender     GitHubUser       `json:"sender"`
}

// GitHubInstallationEvent represents a GitHub App installation event
type GitHubInstallationEvent struct {
	Action       string `json:"action"`
	Installation struct {
		ID      string     `json:"id"`
		Account GitHubUser `json:"account"`
	} `json:"installation"`
	Repositories []GitHubRepository `json:"repositories"`
	Sender       GitHubUser         `json:"sender"`
}

// GitHubInstallationRepositoriesEvent represents a GitHub App installation repositories event
type GitHubInstallationRepositoriesEvent struct {
	Action              string `json:"action"`
	RepositoriesAdded   []GitHubRepository `json:"repositories_added"`
	RepositoriesRemoved []GitHubRepository `json:"repositories_removed"`
	Installation        struct {
		ID      string     `json:"id"`
		Account GitHubUser `json:"account"`
	} `json:"installation"`
	Sender GitHubUser `json:"sender"`
}
