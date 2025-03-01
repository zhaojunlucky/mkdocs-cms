package services

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"gorm.io/gorm"
)

// UserGitRepoService handles business logic for git repositories
type UserGitRepoService struct{}

// NewUserGitRepoService creates a new UserGitRepoService
func NewUserGitRepoService() *UserGitRepoService {
	return &UserGitRepoService{}
}

// GetAllRepos returns all git repositories
func (s *UserGitRepoService) GetAllRepos() ([]models.UserGitRepo, error) {
	var repos []models.UserGitRepo
	result := database.DB.Find(&repos)
	return repos, result.Error
}

// GetReposByUser returns all git repositories for a specific user
func (s *UserGitRepoService) GetReposByUser(userID uint) ([]models.UserGitRepo, error) {
	var repos []models.UserGitRepo
	result := database.DB.Where("user_id = ?", userID).Find(&repos)
	return repos, result.Error
}

// GetRepoByID returns a specific git repository by ID
func (s *UserGitRepoService) GetRepoByID(id uint) (models.UserGitRepo, error) {
	var repo models.UserGitRepo
	result := database.DB.First(&repo, id)
	return repo, result.Error
}

// CreateRepo creates a new git repository
func (s *UserGitRepoService) CreateRepo(request models.CreateUserGitRepoRequest) (models.UserGitRepo, error) {
	// Check if user exists
	var user models.User
	if err := database.DB.First(&user, request.UserID).Error; err != nil {
		return models.UserGitRepo{}, errors.New("user not found")
	}

	// Generate local path for the repository
	baseRepoPath := os.Getenv("REPO_BASE_PATH")
	if baseRepoPath == "" {
		baseRepoPath = "./repositories"
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseRepoPath, 0755); err != nil {
		return models.UserGitRepo{}, err
	}

	// Create a unique local path for this repository
	localPath := filepath.Join(baseRepoPath, user.Username, request.Name)

	repo := models.UserGitRepo{
		Name:        request.Name,
		Description: request.Description,
		LocalPath:   localPath,
		RemoteURL:   request.RemoteURL,
		UserID:      request.UserID,
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if request.Branch != "" {
		repo.Branch = request.Branch
	} else {
		repo.Branch = "main" // Default branch
	}

	result := database.DB.Create(&repo)
	return repo, result.Error
}

// UpdateRepo updates an existing git repository
func (s *UserGitRepoService) UpdateRepo(id uint, request models.UpdateUserGitRepoRequest) (models.UserGitRepo, error) {
	var repo models.UserGitRepo
	if err := database.DB.First(&repo, id).Error; err != nil {
		return models.UserGitRepo{}, err
	}

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
	if request.Status != "" {
		repo.Status = request.Status
	}
	if request.ErrorMsg != "" {
		repo.ErrorMsg = request.ErrorMsg
	}

	repo.UpdatedAt = time.Now()
	result := database.DB.Save(&repo)
	return repo, result.Error
}

// DeleteRepo deletes a git repository
func (s *UserGitRepoService) DeleteRepo(id uint) error {
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
	// if err := os.RemoveAll(repo.LocalPath); err != nil {
	//     return err
	// }

	return nil
}

// UpdateRepoStatus updates the status of a git repository
func (s *UserGitRepoService) UpdateRepoStatus(id uint, status models.GitRepoStatus, errorMsg string) error {
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
// This is a placeholder for actual git operations
func (s *UserGitRepoService) SyncRepo(id uint) error {
	var repo models.UserGitRepo
	if err := database.DB.First(&repo, id).Error; err != nil {
		return err
	}

	// Update status to syncing
	if err := s.UpdateRepoStatus(id, models.StatusSyncing, ""); err != nil {
		return err
	}

	// TODO: Implement actual git operations here
	// For now, we'll just simulate a successful sync
	time.Sleep(2 * time.Second)

	// Update status to synced
	return s.UpdateRepoStatus(id, models.StatusSynced, "")
}
