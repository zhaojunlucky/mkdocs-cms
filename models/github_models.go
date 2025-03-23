package models

import "time"

// GitHubAppSettings contains configuration for the GitHub App
type GitHubAppSettings struct {
	AppID          int64  `json:"app_id"`
	AppName        string `json:"app_name"`
	Description    string `json:"description"`
	HomepageURL    string `json:"homepage_url"`
	WebhookURL     string `json:"webhook_url"`
	WebhookSecret  string `json:"webhook_secret"`
	PrivateKeyPath string `json:"private_key_path"`
}

// ImportRepositoriesRequest represents a request to import repositories from GitHub
type ImportRepositoriesRequest struct {
	UserID        string  `json:"user_id"`
	RepositoryIDs []int64 `json:"repositories" binding:"required"`
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
