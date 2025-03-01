package services

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

// FileInfo represents information about a file or directory
type FileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"is_dir"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"mod_time"`
	Extension string    `json:"extension,omitempty"`
}

// ListFilesInCollection lists all files under a collection path
func (s *UserGitRepoCollectionService) ListFilesInCollection(collectionID uint) ([]FileInfo, error) {
	// Get the collection
	collection, err := s.GetCollectionByID(collectionID)
	if err != nil {
		return nil, err
	}

	// Check if the path exists
	if _, err := os.Stat(collection.Path); os.IsNotExist(err) {
		return nil, errors.New("collection path does not exist")
	}

	// Read the directory
	entries, err := os.ReadDir(collection.Path)
	if err != nil {
		return nil, err
	}

	// Convert to FileInfo
	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(collection.Path, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		// Add extension for files
		if !entry.IsDir() {
			fileInfo.Extension = filepath.Ext(entry.Name())
		}

		files = append(files, fileInfo)
	}

	return files, nil
}

// ListFilesInPath lists all files under a specific path within a collection
func (s *UserGitRepoCollectionService) ListFilesInPath(collectionID uint, subPath string) ([]FileInfo, error) {
	// Get the collection
	collection, err := s.GetCollectionByID(collectionID)
	if err != nil {
		return nil, err
	}

	// Ensure the subPath doesn't try to escape the collection directory
	cleanSubPath := filepath.Clean(subPath)
	if cleanSubPath == ".." || filepath.IsAbs(cleanSubPath) || strings.HasPrefix(cleanSubPath, "../") {
		return nil, errors.New("invalid path")
	}

	// Construct the full path
	fullPath := filepath.Join(collection.Path, cleanSubPath)

	// Check if the path exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, errors.New("path does not exist")
	}

	// Read the directory
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	// Convert to FileInfo
	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		relativePath := filepath.Join(cleanSubPath, entry.Name())
		fileInfo := FileInfo{
			Name:    entry.Name(),
			Path:    relativePath,
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		// Add extension for files
		if !entry.IsDir() {
			fileInfo.Extension = filepath.Ext(entry.Name())
		}

		files = append(files, fileInfo)
	}

	return files, nil
}

// GetFileContent retrieves the content of a file within a collection
func (s *UserGitRepoCollectionService) GetFileContent(collectionID uint, filePath string) ([]byte, string, error) {
	// Get the collection
	collection, err := s.GetCollectionByID(collectionID)
	if err != nil {
		return nil, "", err
	}

	// Ensure the filePath doesn't try to escape the collection directory
	cleanFilePath := filepath.Clean(filePath)
	if cleanFilePath == ".." || filepath.IsAbs(cleanFilePath) || strings.HasPrefix(cleanFilePath, "../") {
		return nil, "", errors.New("invalid path")
	}

	// Construct the full path
	fullPath := filepath.Join(collection.Path, cleanFilePath)

	// Check if the path exists and is a file
	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, "", errors.New("file does not exist")
	}
	if err != nil {
		return nil, "", err
	}
	if fileInfo.IsDir() {
		return nil, "", errors.New("path is a directory, not a file")
	}

	// Read the file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, "", err
	}

	// Determine content type based on file extension
	contentType := "text/plain"
	ext := filepath.Ext(fullPath)
	switch strings.ToLower(ext) {
	case ".html", ".htm":
		contentType = "text/html"
	case ".css":
		contentType = "text/css"
	case ".js":
		contentType = "application/javascript"
	case ".json":
		contentType = "application/json"
	case ".xml":
		contentType = "application/xml"
	case ".md":
		contentType = "text/markdown"
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	case ".pdf":
		contentType = "application/pdf"
	}

	return content, contentType, nil
}

// UpdateFileContent updates the content of a file within a collection
func (s *UserGitRepoCollectionService) UpdateFileContent(collectionID uint, filePath string, content []byte) error {
	// Get the collection
	collection, err := s.GetCollectionByID(collectionID)
	if err != nil {
		return err
	}

	// Ensure the filePath doesn't try to escape the collection directory
	cleanFilePath := filepath.Clean(filePath)
	if cleanFilePath == ".." || filepath.IsAbs(cleanFilePath) || strings.HasPrefix(cleanFilePath, "../") {
		return errors.New("invalid path")
	}

	// Construct the full path
	fullPath := filepath.Join(collection.Path, cleanFilePath)

	// Check if the file exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If the file exists, check if it's a directory
	if err == nil && fileInfo.IsDir() {
		return errors.New("cannot update a directory")
	}

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}

	// Write the content to the file
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return err
	}

	return nil
}

// DeleteFile deletes a file or directory within a collection
func (s *UserGitRepoCollectionService) DeleteFile(collectionID uint, filePath string) error {
	// Get the collection
	collection, err := s.GetCollectionByID(collectionID)
	if err != nil {
		return err
	}

	// Ensure the filePath doesn't try to escape the collection directory
	cleanFilePath := filepath.Clean(filePath)
	if cleanFilePath == ".." || filepath.IsAbs(cleanFilePath) || strings.HasPrefix(cleanFilePath, "../") {
		return errors.New("invalid path")
	}

	// Construct the full path
	fullPath := filepath.Join(collection.Path, cleanFilePath)

	// Check if the path exists
	_, err = os.Stat(fullPath)
	if os.IsNotExist(err) {
		return errors.New("file or directory does not exist")
	}
	if err != nil {
		return err
	}

	// Delete the file or directory
	if err := os.RemoveAll(fullPath); err != nil {
		return err
	}

	return nil
}
