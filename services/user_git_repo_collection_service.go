package services

import (
	"errors"
	"path/filepath"
	"time"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
)

// UserGitRepoCollectionService handles business logic for git repository collections
type UserGitRepoCollectionService struct{}

// NewUserGitRepoCollectionService creates a new UserGitRepoCollectionService
func NewUserGitRepoCollectionService() *UserGitRepoCollectionService {
	return &UserGitRepoCollectionService{}
}

// GetAllCollections returns all collections
func (s *UserGitRepoCollectionService) GetAllCollections() ([]models.UserGitRepoCollection, error) {
	var collections []models.UserGitRepoCollection
	result := database.DB.Find(&collections)
	return collections, result.Error
}

// GetCollectionsByRepo returns all collections for a specific repository
func (s *UserGitRepoCollectionService) GetCollectionsByRepo(repoID uint) ([]models.UserGitRepoCollection, error) {
	var collections []models.UserGitRepoCollection
	result := database.DB.Where("repo_id = ?", repoID).Find(&collections)
	return collections, result.Error
}

// GetCollectionByID returns a specific collection by ID
func (s *UserGitRepoCollectionService) GetCollectionByID(id uint) (models.UserGitRepoCollection, error) {
	var collection models.UserGitRepoCollection
	result := database.DB.Preload("Repo").First(&collection, id)
	return collection, result.Error
}

// CreateCollection creates a new collection
func (s *UserGitRepoCollectionService) CreateCollection(request models.CreateUserGitRepoCollectionRequest) (models.UserGitRepoCollection, error) {
	// Check if repository exists
	var repo models.UserGitRepo
	if err := database.DB.First(&repo, request.RepoID).Error; err != nil {
		return models.UserGitRepoCollection{}, errors.New("repository not found")
	}

	// Validate path format
	if !filepath.IsAbs(request.Path) {
		// If path is not absolute, make it relative to the repository path
		request.Path = filepath.Join(repo.LocalPath, request.Path)
	}

	// Set default format if not provided
	format := request.Format
	if format == "" {
		format = models.FormatMarkdown
	}

	collection := models.UserGitRepoCollection{
		Name:        request.Name,
		Label:       request.Label,
		Path:        request.Path,
		Format:      format,
		Description: request.Description,
		RepoID:      request.RepoID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result := database.DB.Create(&collection)
	return collection, result.Error
}

// UpdateCollection updates an existing collection
func (s *UserGitRepoCollectionService) UpdateCollection(id uint, request models.UpdateUserGitRepoCollectionRequest) (models.UserGitRepoCollection, error) {
	var collection models.UserGitRepoCollection
	if err := database.DB.First(&collection, id).Error; err != nil {
		return models.UserGitRepoCollection{}, err
	}

	// Update fields if provided
	if request.Name != "" {
		collection.Name = request.Name
	}
	if request.Label != "" {
		collection.Label = request.Label
	}
	if request.Path != "" {
		// If path is not absolute, make it relative to the repository path
		if !filepath.IsAbs(request.Path) {
			var repo models.UserGitRepo
			if err := database.DB.First(&repo, collection.RepoID).Error; err != nil {
				return models.UserGitRepoCollection{}, errors.New("repository not found")
			}
			collection.Path = filepath.Join(repo.LocalPath, request.Path)
		} else {
			collection.Path = request.Path
		}
	}
	if request.Format != "" {
		collection.Format = request.Format
	}
	if request.Description != "" {
		collection.Description = request.Description
	}

	collection.UpdatedAt = time.Now()
	result := database.DB.Save(&collection)
	return collection, result.Error
}

// DeleteCollection deletes a collection
func (s *UserGitRepoCollectionService) DeleteCollection(id uint) error {
	var collection models.UserGitRepoCollection
	if err := database.DB.First(&collection, id).Error; err != nil {
		return err
	}

	return database.DB.Delete(&collection).Error
}

// GetCollectionByPath returns a collection by its path within a repository
func (s *UserGitRepoCollectionService) GetCollectionByPath(repoID uint, path string) (models.UserGitRepoCollection, error) {
	var collection models.UserGitRepoCollection
	result := database.DB.Where("repo_id = ? AND path = ?", repoID, path).First(&collection)
	return collection, result.Error
}

// GetCollectionByName returns a collection by its name within a repository
func (s *UserGitRepoCollectionService) GetCollectionByName(repoID uint, name string) (models.UserGitRepoCollection, error) {
	var collection models.UserGitRepoCollection
	result := database.DB.Where("repo_id = ? AND name = ?", repoID, name).First(&collection)
	return collection, result.Error
}
