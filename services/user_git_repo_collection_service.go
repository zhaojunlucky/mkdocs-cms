package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"gopkg.in/yaml.v3"
)

// UserGitRepoCollectionService handles business logic for git repository collections
type UserGitRepoCollectionService struct{}

// NewUserGitRepoCollectionService creates a new UserGitRepoCollectionService
func NewUserGitRepoCollectionService() *UserGitRepoCollectionService {
	return &UserGitRepoCollectionService{}
}

// VedaConfig represents the structure of veda/config.yml
type VedaConfig struct {
	Collections []Collection `yaml:"collections"`
}

// Collection represents a collection in veda/config.yml
type Collection struct {
	Name   string  `yaml:"name"`
	Label  string  `yaml:"label"`
	Path   string  `yaml:"path"`
	Format string  `yaml:"format"`
	Fields []Field `yaml:"fields,omitempty"`
}

// Field represents a field in a collection
type Field struct {
	Type     string `yaml:"type"`
	Name     string `yaml:"name"`
	Label    string `yaml:"label"`
	Required bool   `yaml:"required,omitempty"`
	Format   string `yaml:"format,omitempty"`
	List     bool   `yaml:"list,omitempty"`
}

// GetAllCollections returns all collections
func (s *UserGitRepoCollectionService) GetAllCollections() ([]models.UserGitRepoCollection, error) {
	// This method is not applicable anymore as collections are read from veda/config.yml
	return nil, errors.New("collections are now stored in veda/config.yml, use GetCollectionsByRepo instead")
}

// GetCollectionsByRepo returns all collections for a specific repository
func (s *UserGitRepoCollectionService) GetCollectionsByRepo(repoID uint) ([]models.UserGitRepoCollection, error) {
	// Get the repository to find its local path
	var repo models.UserGitRepo
	if err := database.DB.First(&repo, repoID).Error; err != nil {
		return nil, errors.New("repository not found")
	}

	// Read collections from veda/config.yml
	collections, err := s.readCollectionsFromConfig(repo)
	if err != nil {
		return nil, err
	}

	return collections, nil
}

// readCollectionsFromConfig reads collections from veda/config.yml
func (s *UserGitRepoCollectionService) readCollectionsFromConfig(repo models.UserGitRepo) ([]models.UserGitRepoCollection, error) {
	configPath := filepath.Join(repo.LocalPath, "veda", "config.yml")

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("veda/config.yml not found in repository %s", repo.Name)
	}

	// Read the config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read veda/config.yml: %v", err)
	}

	// Parse the YAML
	var config VedaConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("invalid YAML format in veda/config.yml: %v", err)
	}

	// Convert to UserGitRepoCollection models
	var collections []models.UserGitRepoCollection
	for _, col := range config.Collections {
		// Resolve the path relative to the repository
		fullPath := filepath.Join(repo.LocalPath, col.Path)
		
		collection := models.UserGitRepoCollection{
			Name:        col.Name,
			Label:       col.Label,
			Path:        fullPath,
			Format:      models.ContentFormat(col.Format),
			Description: "", // No description in veda/config.yml
			RepoID:      repo.ID,
		}
		collections = append(collections, collection)
	}

	return collections, nil
}

// GetCollectionByID returns a specific collection by ID
func (s *UserGitRepoCollectionService) GetCollectionByID(id uint) (models.UserGitRepoCollection, error) {
	// Since collections are now read from veda/config.yml, we need to find the repository first
	var collection models.UserGitRepoCollection
	if err := database.DB.Preload("Repo").First(&collection, id).Error; err != nil {
		// If we can't find it in the database, try to find it by name in the repository
		var repo models.UserGitRepo
		if err := database.DB.First(&repo, "id = ?", id).Error; err != nil {
			return models.UserGitRepoCollection{}, errors.New("repository not found")
		}
		
		// Read collections from veda/config.yml
		collections, err := s.readCollectionsFromConfig(repo)
		if err != nil {
			return models.UserGitRepoCollection{}, err
		}
		
		// Find the collection by ID (which is now just an index)
		if int(id) < len(collections) {
			return collections[id], nil
		}
		
		return models.UserGitRepoCollection{}, errors.New("collection not found")
	}
	
	return collection, nil
}

// GetCollectionByName returns a collection by its name within a repository
func (s *UserGitRepoCollectionService) GetCollectionByName(repoID uint, name string) (models.UserGitRepoCollection, error) {
	collections, err := s.GetCollectionsByRepo(repoID)
	if err != nil {
		return models.UserGitRepoCollection{}, err
	}

	for _, collection := range collections {
		if collection.Name == name {
			return collection, nil
		}
	}

	return models.UserGitRepoCollection{}, errors.New("collection not found")
}

// GetCollectionByPath returns a collection by its path within a repository
func (s *UserGitRepoCollectionService) GetCollectionByPath(repoID uint, path string) (models.UserGitRepoCollection, error) {
	collections, err := s.GetCollectionsByRepo(repoID)
	if err != nil {
		return models.UserGitRepoCollection{}, err
	}

	for _, collection := range collections {
		if strings.HasSuffix(collection.Path, path) {
			return collection, nil
		}
	}

	return models.UserGitRepoCollection{}, errors.New("collection not found")
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
