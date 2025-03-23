package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-github/v45/github"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// UserGitRepoService handles business logic for git repositories
type UserGitRepoService struct {
	BaseService
	githubAppSettings *models.GitHubAppSettings
	githubAppClient   *github.Client
}

func (s *UserGitRepoService) Init(ctx *core.APPContext) {
	s.InitService("userGitRepoService", ctx, s)
	s.githubAppSettings = ctx.GithubAppSettings
	s.githubAppClient = ctx.GithubAppClient

}

// GetAllRepos returns all git repositories
func (s *UserGitRepoService) GetAllRepos() ([]models.UserGitRepo, error) {
	var repos []models.UserGitRepo
	result := database.DB.Find(&repos)
	return repos, result.Error
}

// GetReposByUser returns all git repositories for a specific user
func (s *UserGitRepoService) GetReposByUser(userID string) ([]models.UserGitRepo, error) {
	var repos []models.UserGitRepo
	result := database.DB.Where("user_id = ?", userID).Find(&repos)
	return repos, result.Error
}

// GetRepoByID returns a specific git repository by ID
func (s *UserGitRepoService) GetRepoByID(id string) (models.UserGitRepo, error) {
	var repo models.UserGitRepo
	result := database.DB.First(&repo, id)
	return repo, result.Error
}

// GetReposByURL returns repositories that match a specific remote URL
func (s *UserGitRepoService) GetReposByURL(url string) ([]models.UserGitRepo, error) {
	var repos []models.UserGitRepo
	result := database.DB.Where("remote_url = ?", url).Find(&repos)
	return repos, result.Error
}

// CreateRepo creates a new git repository
func (s *UserGitRepoService) CreateRepo(repo *models.UserGitRepo) error {
	// Check if user exists
	var user models.User
	if err := database.DB.First(&user, "id = ?", repo.UserID).Error; err != nil {
		return errors.New("user not found")
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(s.ctx.RepoBasePath, 0755); err != nil {
		return err
	}

	// Create a unique local path for this repository
	repo.LocalPath = filepath.Join(s.ctx.RepoBasePath, user.Username, repo.Name)

	// Set default values if not provided
	if repo.Branch == "" {
		repo.Branch = "main" // Default branch
	}

	repo.Status = models.StatusPending
	repo.CreatedAt = time.Now()
	repo.UpdatedAt = time.Now()

	result := database.DB.Create(repo)
	return result.Error
}

// UpdateRepo updates an existing git repository
func (s *UserGitRepoService) UpdateRepo(id string, request models.UpdateUserGitRepoRequest) (models.UserGitRepo, error) {
	var repo models.UserGitRepo
	if err := database.DB.First(&repo, id).Error; err != nil {
		return models.UserGitRepo{}, err
	}

	// Check if branch is being changed
	branchChanged := request.Branch != "" && request.Branch != repo.Branch

	// Update fields if provided
	if request.Name != "" {
		repo.Name = request.Name
	}
	if request.Description != "" {
		repo.Description = request.Description
	}
	if request.RemoteURL != "" {
		repo.RemoteURL = request.RemoteURL
	}
	if request.Branch != "" {
		repo.Branch = request.Branch
	}
	if request.AuthType != "" {
		repo.AuthType = request.AuthType
	}
	if request.AuthData != "" {
		repo.AuthData = request.AuthData
	}
	if request.Status != "" {
		repo.Status = request.Status
	}
	if request.ErrorMsg != "" {
		repo.ErrorMsg = request.ErrorMsg
	}

	repo.UpdatedAt = time.Now()
	result := database.DB.Save(&repo)
	if result.Error != nil {
		return repo, result.Error
	}

	// If branch was changed, sync the repo and checkout the new branch
	if branchChanged {
		// First sync the repository to ensure we have the latest changes
		if err := s.SyncRepo(id); err != nil {
			return repo, fmt.Errorf("failed to sync repository after branch change: %v", err)
		}

		// Then checkout the new branch
		if err := s.checkoutBranch(repo); err != nil {
			return repo, fmt.Errorf("failed to checkout new branch: %v", err)
		}
	}

	return repo, nil
}

// DeleteRepo deletes a git repository
func (s *UserGitRepoService) DeleteRepo(id string) error {
	var repo models.UserGitRepo
	if err := database.DB.First(&repo, id).Error; err != nil {
		return err
	}

	// Delete the repository from the database
	if err := database.DB.Delete(&repo).Error; err != nil {
		return err
	}

	// Optionally, delete the local repository files
	// This is commented out for safety - uncomment if you want to delete files
	if err := os.RemoveAll(repo.LocalPath); err != nil {
		return err
	}

	return nil
}

// UpdateRepoStatus updates the status of a git repository
func (s *UserGitRepoService) UpdateRepoStatus(id string, status models.GitRepoStatus, errorMsg string) error {
	var repo models.UserGitRepo
	if err := database.DB.First(&repo, id).Error; err != nil {
		return err
	}

	repo.Status = status
	repo.ErrorMsg = errorMsg
	if status == models.StatusSynced {
		repo.LastSyncAt = time.Now()
	}
	repo.UpdatedAt = time.Now()

	return database.DB.Save(&repo).Error
}

// SyncRepo synchronizes a git repository with its remote
func (s *UserGitRepoService) SyncRepo(id string) error {
	var repo models.UserGitRepo
	if err := database.DB.First(&repo, id).Error; err != nil {
		return err
	}

	// Update status to syncing
	if err := s.UpdateRepoStatus(id, models.StatusSyncing, ""); err != nil {
		return err
	}

	var err error
	switch repo.AuthType {
	case "github_app":
		err = s.syncWithGitHubApp(repo)
	default:
		err = s.syncWithGitCommand(repo)
	}

	if err != nil {
		s.UpdateRepoStatus(id, models.StatusFailed, err.Error())
		return err
	}

	// Check if veda/config.yml exists and has valid format
	if err := s.checkVedaConfig(repo); err != nil {
		// Set error message but don't fail the sync
		s.UpdateRepoStatus(id, models.StatusWarning, err.Error())
	} else {
		// Update status to synced
		if err := s.UpdateRepoStatus(id, models.StatusSynced, ""); err != nil {
			return err
		}
	}

	return nil
}

// syncWithGitCommand uses git command line to sync a repository
func (s *UserGitRepoService) syncWithGitCommand(repo models.UserGitRepo) error {
	// Check if repository directory exists
	if _, err := os.Stat(repo.LocalPath); os.IsNotExist(err) {
		// Clone the repository
		cmd := exec.Command("git", "clone", "-b", repo.Branch, repo.RemoteURL, repo.LocalPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %v", err)
		}
	} else {
		// Pull the latest changes
		cmd := exec.Command("git", "-C", repo.LocalPath, "pull", "origin", repo.Branch)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to pull repository: %v", err)
		}
	}

	return nil
}

// syncWithGitHubApp uses GitHub App authentication to sync a repository
func (s *UserGitRepoService) syncWithGitHubApp(repo models.UserGitRepo) error {
	// Parse the auth data
	var authData models.GitHubAuthData
	err := json.Unmarshal([]byte(repo.AuthData), &authData)
	if err != nil {
		return fmt.Errorf("invalid auth data: %v", err)
	}

	// Get an installation token
	ctx := context.Background()
	installationToken, _, err := s.ctx.GithubAppClient.Apps.CreateInstallationToken(ctx, authData.InstallationID, nil)
	if err != nil {
		return fmt.Errorf("failed to get installation token: %v", err)
	}

	// Use the installation token for git operations
	if _, err := os.Stat(repo.LocalPath); os.IsNotExist(err) {
		// Clone the repository
		cloneURL := repo.RemoteURL
		// Insert the token into the URL
		// Example: https://x-access-token:TOKEN@github.com/owner/repo.git
		cloneURL = fmt.Sprintf("https://x-access-token:%s@%s", installationToken.GetToken(), cloneURL[8:])

		cmd := exec.Command("git", "clone", "-b", repo.Branch, cloneURL, repo.LocalPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %v", err)
		}
	} else {
		// Set the remote URL with the token
		remoteURL := fmt.Sprintf("https://x-access-token:%s@%s", installationToken.GetToken(), repo.RemoteURL[8:])
		setRemoteCmd := exec.Command("git", "-C", repo.LocalPath, "remote", "set-url", "origin", remoteURL)
		if err := setRemoteCmd.Run(); err != nil {
			return fmt.Errorf("failed to set remote URL: %v", err)
		}

		// Pull the latest changes
		pullCmd := exec.Command("git", "-C", repo.LocalPath, "pull", "origin") //, repo.Branch
		if err := pullCmd.Run(); err != nil {
			return fmt.Errorf("failed to pull repository: %v", err)
		}

		// Reset the remote URL to the original
		resetRemoteCmd := exec.Command("git", "-C", repo.LocalPath, "remote", "set-url", "origin", repo.RemoteURL)
		if err := resetRemoteCmd.Run(); err != nil {
			return fmt.Errorf("failed to reset remote URL: %v", err)
		}
	}

	return nil
}

// SyncRepository is an alias for SyncRepo for compatibility with the webhook controller
func (s *UserGitRepoService) SyncRepository(id string) error {
	return s.SyncRepo(id)
}

// GetRepoBranches returns all branches for a specific git repository
func (s *UserGitRepoService) GetRepoBranches(id string) ([]string, error) {
	// Get the repository
	repo, err := s.GetRepoByID(id)
	if err != nil {
		return nil, err
	}

	// Check if the repository exists locally
	if _, err := os.Stat(repo.LocalPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository not found locally")
	}

	// Run git branch -r to get remote branches
	cmd := exec.Command("git", "-C", repo.LocalPath, "branch", "-r")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %v", err)
	}

	// Parse the output to extract branch names
	branches := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove "origin/" prefix
		if strings.HasPrefix(line, "origin/") {
			branch := strings.TrimPrefix(line, "origin/")
			// Skip HEAD and other special refs
			if branch != "HEAD" && !strings.Contains(branch, "->") {
				branches = append(branches, branch)
			}
		}
	}

	return branches, nil
}

// checkoutBranch checks out the specified branch in the repository
func (s *UserGitRepoService) checkoutBranch(repo models.UserGitRepo) error {
	// Check if repository directory exists
	if _, err := os.Stat(repo.LocalPath); os.IsNotExist(err) {
		return fmt.Errorf("repository directory does not exist")
	}

	// Fetch all branches to ensure the branch exists locally
	fetchCmd := exec.Command("git", "-C", repo.LocalPath, "fetch", "origin")
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch from remote: %v", err)
	}

	// Check if the branch exists
	checkBranchCmd := exec.Command("git", "-C", repo.LocalPath, "branch", "-r")
	output, err := checkBranchCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list remote branches: %v", err)
	}

	branchExists := false
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "/"+repo.Branch) {
			branchExists = true
			break
		}
	}

	if !branchExists {
		return fmt.Errorf("branch '%s' does not exist in the remote repository", repo.Branch)
	}

	// Checkout the branch
	checkoutCmd := exec.Command("git", "-C", repo.LocalPath, "switch", repo.Branch)
	if err := checkoutCmd.Run(); err != nil {
		// Try to create and checkout the branch if it doesn't exist locally
		createCmd := exec.Command("git", "-C", repo.LocalPath, "checkout", "-b", repo.Branch, "origin/"+repo.Branch)
		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("failed to checkout branch '%s': %v", repo.Branch, err)
		}
	}

	// Pull the latest changes for this branch
	pullCmd := exec.Command("git", "-C", repo.LocalPath, "pull", "origin", repo.Branch)
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull latest changes for branch '%s': %v", repo.Branch, err)
	}

	// Check if veda/config.yml exists and has valid format
	if err := s.checkVedaConfig(repo); err != nil {
		return err
	}

	return nil
}

// checkVedaConfig checks if veda/config.yml exists and has a valid format
func (s *UserGitRepoService) checkVedaConfig(repo models.UserGitRepo) error {
	configPath := filepath.Join(repo.LocalPath, "veda", "config.yml")

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("veda/config.yml not found. Please create this file with proper configuration")
	}

	// Read the config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read veda/config.yml: %v", err)
	}

	// Parse the YAML to validate its structure
	var config map[string]interface{}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("invalid YAML format in veda/config.yml: %v", err)
	}

	// Check for required fields
	collections, ok := config["collections"]
	if !ok {
		return fmt.Errorf("veda/config.yml is missing required 'collections' field")
	}

	// Validate collections structure
	collectionsList, ok := collections.([]interface{})
	if !ok {
		return fmt.Errorf("'collections' field in veda/config.yml must be an array")
	}

	if len(collectionsList) == 0 {
		return fmt.Errorf("veda/config.yml must contain at least one collection")
	}

	// Validate each collection
	for i, col := range collectionsList {
		collection, ok := col.(map[string]interface{})
		if !ok {
			return fmt.Errorf("collection #%d in veda/config.yml has invalid format", i+1)
		}

		// Check for required collection fields
		if _, ok := collection["name"]; !ok {
			return fmt.Errorf("collection #%d in veda/config.yml is missing required 'name' field", i+1)
		}
		if _, ok := collection["label"]; !ok {
			return fmt.Errorf("collection #%d in veda/config.yml is missing required 'label' field", i+1)
		}
		if _, ok := collection["path"]; !ok {
			return fmt.Errorf("collection #%d in veda/config.yml is missing required 'path' field", i+1)
		}
		if _, ok := collection["format"]; !ok {
			return fmt.Errorf("collection #%d in veda/config.yml is missing required 'format' field", i+1)
		}
	}

	return nil
}

// ensureVedaConfig checks if veda/config.yml exists and creates it if it doesn't
func (s *UserGitRepoService) ensureVedaConfig(repo models.UserGitRepo) error {
	configPath := filepath.Join(repo.LocalPath, "veda", "config.yml")

	// Check if the config file already exists
	if _, err := os.Stat(configPath); err == nil {
		// File exists, validate its format
		return s.checkVedaConfig(repo)
	}

	// If file doesn't exist, return an error
	return fmt.Errorf("veda/config.yml not found. Please create this file with proper configuration")
}

// GetInstallationToken gets a GitHub app token for the repository
func (s *UserGitRepoService) GetInstallationToken(installationID int64) (*github.InstallationToken, error) {
	token, _, err := s.githubAppClient.Apps.CreateInstallationToken(context.Background(), installationID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create installation token: %v", err)
	}
	return token, nil
}

func (s *UserGitRepoService) CheckWebHooks(id string) error {
	var repo models.UserGitRepo
	if err := database.DB.First(&repo, id).Error; err != nil {
		return err
	}

	// Update status to syncing
	if err := s.UpdateRepoStatus(id, models.StatusSyncing, ""); err != nil {
		return err
	}

	// Get installation token for GitHub API access
	token, err := s.GetInstallationToken(repo.InstallationID)
	if err != nil {
		s.UpdateRepoStatus(id, models.StatusFailed, fmt.Sprintf("Failed to get installation token: %v", err))
		return err
	}

	// Create GitHub client with installation token
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	// Extract owner and repository name from the remote URL
	// Example: https://github.com/owner/repo.git
	parts := strings.Split(repo.RemoteURL, "/")
	owner := parts[len(parts)-2]
	repoName := strings.TrimSuffix(parts[len(parts)-1], ".git")

	// List webhooks for the repository
	hooks, _, err := client.Repositories.ListHooks(context.Background(), owner, repoName, nil)
	if err != nil {
		s.UpdateRepoStatus(id, models.StatusFailed, fmt.Sprintf("Failed to list webhooks: %v", err))
		return err
	}

	// Log webhook information
	log.Infof("Found %d webhooks for repository %s/%s\n", len(hooks), owner, repoName)
	var repoHook *github.Hook
	for i, hook := range hooks {
		url := hook.Config["url"]
		log.Infof("Webhook %d: ID=%d, URL=%s, Active=%t\n", i+1, hook.GetID(), url, hook.GetActive())

		if url == s.githubAppSettings.WebhookURL {
			repoHook = hook
			break
		}
	}

	if repoHook == nil {
		// create hook
		hook := &github.Hook{
			Config: map[string]interface{}{
				"url":          s.githubAppSettings.WebhookURL,
				"content_type": "json",
				"secret":       s.githubAppSettings.WebhookSecret,
			},
			Events: []string{"push"},
			Active: github.Bool(true),
		}

		createdHook, _, err := client.Repositories.CreateHook(context.Background(), owner, repoName, hook)
		if err != nil {
			s.UpdateRepoStatus(id, models.StatusFailed, fmt.Sprintf("Failed to create webhook: %v", err))
			return err
		} else {
			s.UpdateRepoStatus(id, models.StatusSynced, fmt.Sprintf("created wehbook: %d", createdHook.ID))
		}
	} else if !*repoHook.Active {
		// enable hook
		_, resp, err := client.Repositories.EditHook(context.Background(), owner, repoName, repoHook.GetID(), &github.Hook{
			Active: github.Bool(true),
		})
		if err != nil {
			s.UpdateRepoStatus(id, models.StatusFailed, fmt.Sprintf("Failed to update webhook: %v", err))
			return err
		}
		resp.Body.Close()
	}

	// Update status to synced
	if err := s.UpdateRepoStatus(id, models.StatusSynced, ""); err != nil {
		return err
	}

	return nil
}
