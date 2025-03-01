package services

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"gorm.io/gorm"
)

// PostService handles business logic related to posts
type PostService struct {
	collectionService *UserGitRepoCollectionService
	db                *gorm.DB
}

// NewPostService creates a new post service
func NewPostService() *PostService {
	return &PostService{
		collectionService: NewUserGitRepoCollectionService(),
		db:                database.DB,
	}
}

// GetAllPosts retrieves all posts from the database
func (s *PostService) GetAllPosts() ([]models.PostResponse, error) {
	var posts []models.Post
	result := s.db.Preload("User").Preload("Collection").Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}

	// Convert to response objects
	var responses []models.PostResponse
	for _, post := range posts {
		response := post.ToResponse(true, true)

		// Load content from file if needed
		content, err := s.getPostContent(&post)
		if err == nil {
			response.Content = string(content)
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetPostsByCollection retrieves all posts from a specific collection
func (s *PostService) GetPostsByCollection(collectionID uint) ([]models.PostResponse, error) {
	var posts []models.Post
	result := s.db.Preload("User").Preload("Collection").Where("collection_id = ?", collectionID).Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}

	// Convert to response objects
	var responses []models.PostResponse
	for _, post := range posts {
		response := post.ToResponse(true, true)

		// Load content from file if needed
		content, err := s.getPostContent(&post)
		if err == nil {
			response.Content = string(content)
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetPostByID retrieves a post by ID
func (s *PostService) GetPostByID(id uint) (models.PostResponse, error) {
	var post models.Post
	result := s.db.Preload("User").Preload("Collection").First(&post, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.PostResponse{}, errors.New("post not found")
		}
		return models.PostResponse{}, result.Error
	}

	response := post.ToResponse(true, true)

	// Load content from file
	content, err := s.getPostContent(&post)
	if err == nil {
		response.Content = string(content)
	}

	return response, nil
}

// GetPostByPath retrieves a post by its file path within a collection
func (s *PostService) GetPostByPath(collectionID uint, filePath string) (models.PostResponse, error) {
	var post models.Post
	result := s.db.Preload("User").Preload("Collection").
		Where("collection_id = ? AND file_path = ?", collectionID, filePath).
		First(&post)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.PostResponse{}, errors.New("post not found")
		}
		return models.PostResponse{}, result.Error
	}

	response := post.ToResponse(true, true)

	// Load content from file
	content, err := s.getPostContent(&post)
	if err == nil {
		response.Content = string(content)
	}

	return response, nil
}

// getPostContent reads the content of a post from its file
func (s *PostService) getPostContent(post *models.Post) ([]byte, error) {
	// Get repository to get the local path
	var repo models.UserGitRepo
	if err := s.db.First(&repo, post.Collection.RepoID).Error; err != nil {
		return nil, errors.New("repository not found")
	}

	// Create the full path to the file
	fullPath := filepath.Join(repo.LocalPath, post.Collection.Path, post.FilePath)

	// Read the file content
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, errors.New("failed to read file: " + err.Error())
	}

	return content, nil
}

// CreatePost creates a new post
func (s *PostService) CreatePost(req models.CreatePostRequest) (models.PostResponse, error) {
	// Get collection to verify it exists and get the repo path
	collection, err := s.collectionService.GetCollectionByID(req.CollectionID)
	if err != nil {
		return models.PostResponse{}, errors.New("collection not found")
	}

	// Get repository to get the local path
	var repo models.UserGitRepo
	if err := s.db.First(&repo, collection.RepoID).Error; err != nil {
		return models.PostResponse{}, errors.New("repository not found")
	}

	// Create the full path to the file
	fullPath := filepath.Join(repo.LocalPath, collection.Path, req.FilePath)

	// Ensure the directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return models.PostResponse{}, errors.New("failed to create directory: " + err.Error())
	}

	// Write the content to the file
	if err := ioutil.WriteFile(fullPath, []byte(req.Content), 0644); err != nil {
		return models.PostResponse{}, errors.New("failed to write file: " + err.Error())
	}

	// Create post record in database (without content)
	post := models.Post{
		Title:        req.Title,
		FilePath:     req.FilePath,
		UserID:       req.UserID,
		CollectionID: req.CollectionID,
	}

	result := s.db.Create(&post)
	if result.Error != nil {
		// Try to delete the file if database insert fails
		os.Remove(fullPath)
		return models.PostResponse{}, result.Error
	}

	// Reload the post with associations
	s.db.Preload("User").Preload("Collection").First(&post, post.ID)

	// Create response with content
	response := post.ToResponse(true, true)
	response.Content = req.Content

	return response, nil
}

// CreatePostFromFile creates a new post from an existing file
func (s *PostService) CreatePostFromFile(req models.FilePostRequest) (models.PostResponse, error) {
	// Get collection to verify it exists and get the repo path
	collection, err := s.collectionService.GetCollectionByID(req.CollectionID)
	if err != nil {
		return models.PostResponse{}, errors.New("collection not found")
	}

	// Get repository to get the local path
	var repo models.UserGitRepo
	if err := s.db.First(&repo, collection.RepoID).Error; err != nil {
		return models.PostResponse{}, errors.New("repository not found")
	}

	// Create the full path to the file
	fullPath := filepath.Join(repo.LocalPath, collection.Path, req.FilePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return models.PostResponse{}, errors.New("file does not exist")
	}

	// Read the file content
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return models.PostResponse{}, errors.New("failed to read file: " + err.Error())
	}

	// Extract title from filename (remove extension)
	filename := filepath.Base(req.FilePath)
	title := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Create post record in database (without content)
	post := models.Post{
		Title:        title,
		FilePath:     req.FilePath,
		UserID:       req.UserID,
		CollectionID: req.CollectionID,
	}

	result := s.db.Create(&post)
	if result.Error != nil {
		return models.PostResponse{}, result.Error
	}

	// Reload the post with associations
	s.db.Preload("User").Preload("Collection").First(&post, post.ID)

	// Create response with content
	response := post.ToResponse(true, true)
	response.Content = string(content)

	return response, nil
}

// UpdatePost updates an existing post
func (s *PostService) UpdatePost(id uint, req models.UpdatePostRequest) (models.PostResponse, error) {
	var post models.Post
	result := s.db.Preload("Collection").First(&post, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.PostResponse{}, errors.New("post not found")
		}
		return models.PostResponse{}, result.Error
	}

	// Get repository to get the local path
	var repo models.UserGitRepo
	if err := s.db.First(&repo, post.Collection.RepoID).Error; err != nil {
		return models.PostResponse{}, errors.New("repository not found")
	}

	// Handle file path change if needed
	oldFullPath := filepath.Join(repo.LocalPath, post.Collection.Path, post.FilePath)
	newFullPath := oldFullPath

	if req.FilePath != "" && req.FilePath != post.FilePath {
		newFullPath = filepath.Join(repo.LocalPath, post.Collection.Path, req.FilePath)

		// Ensure the directory exists for the new path
		dir := filepath.Dir(newFullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return models.PostResponse{}, errors.New("failed to create directory: " + err.Error())
		}
	}

	// Update post fields
	if req.Title != "" {
		post.Title = req.Title
	}

	pathChanged := req.FilePath != "" && req.FilePath != post.FilePath
	if pathChanged {
		post.FilePath = req.FilePath
	}

	// Update the file content if provided
	if req.Content != "" {
		if err := ioutil.WriteFile(newFullPath, []byte(req.Content), 0644); err != nil {
			return models.PostResponse{}, errors.New("failed to write file: " + err.Error())
		}
	} else if pathChanged {
		// Move the file if path changed but content didn't change
		// Read the old file
		content, err := ioutil.ReadFile(oldFullPath)
		if err != nil {
			return models.PostResponse{}, errors.New("failed to read file: " + err.Error())
		}

		// Write to the new location
		if err := ioutil.WriteFile(newFullPath, content, 0644); err != nil {
			return models.PostResponse{}, errors.New("failed to write file: " + err.Error())
		}

		// Delete the old file
		if err := os.Remove(oldFullPath); err != nil {
			return models.PostResponse{}, errors.New("failed to delete old file: " + err.Error())
		}
	}

	// Save the updated post to the database
	result = s.db.Save(&post)
	if result.Error != nil {
		return models.PostResponse{}, result.Error
	}

	// Reload the post with associations
	s.db.Preload("User").Preload("Collection").First(&post, post.ID)

	// Create response with content
	response := post.ToResponse(true, true)

	// Load content from file
	content, err := s.getPostContent(&post)
	if err == nil {
		response.Content = string(content)
	}

	return response, nil
}

// DeletePost deletes a post by ID
func (s *PostService) DeletePost(id uint) error {
	var post models.Post
	result := s.db.Preload("Collection").First(&post, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("post not found")
		}
		return result.Error
	}

	// Get repository to get the local path
	var repo models.UserGitRepo
	if err := s.db.First(&repo, post.Collection.RepoID).Error; err != nil {
		return errors.New("repository not found")
	}

	// Create the full path to the file
	fullPath := filepath.Join(repo.LocalPath, post.Collection.Path, post.FilePath)

	// Delete the file
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return errors.New("failed to delete file: " + err.Error())
	}

	// Delete the post from the database
	result = s.db.Delete(&post)
	return result.Error
}

// SyncPostsFromCollection synchronizes posts from files in a collection
func (s *PostService) SyncPostsFromCollection(collectionID uint, userID string) (int, error) {
	// Get collection
	collection, err := s.collectionService.GetCollectionByID(collectionID)
	if err != nil {
		return 0, errors.New("collection not found")
	}

	// Get repository
	var repo models.UserGitRepo
	if err := s.db.First(&repo, collection.RepoID).Error; err != nil {
		return 0, errors.New("repository not found")
	}

	// Get all files in the collection
	files, err := s.collectionService.ListFilesInCollection(collectionID)
	if err != nil {
		return 0, err
	}

	// Track how many posts were created
	count := 0

	// Process each file
	for _, file := range files {
		// Skip directories
		if file.IsDir {
			continue
		}

		// Skip non-markdown files if collection format is markdown
		if collection.Format == models.FormatMarkdown && filepath.Ext(file.Path) != ".md" {
			continue
		}

		// Check if post already exists for this file
		var existingPost models.Post
		result := s.db.Where("collection_id = ? AND file_path = ?", collectionID, file.Path).First(&existingPost)

		// Skip if post already exists
		if result.Error == nil {
			continue
		}

		// Create a new post from this file
		_, err := s.CreatePostFromFile(models.FilePostRequest{
			CollectionID: collectionID,
			FilePath:     file.Path,
			UserID:       userID,
		})

		if err == nil {
			count++
		}
	}

	return count, nil
}

// SyncCollectionPosts synchronizes posts from a collection's repository files
func (s *PostService) SyncCollectionPosts(collectionID uint, userID string) ([]*models.Post, error) {
	// Get the collection
	var collection models.UserGitRepoCollection
	if err := s.db.Preload("Repo").First(&collection, collectionID).Error; err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	// Check if the user has access to the collection
	if collection.Repo.UserID != userID {
		return nil, fmt.Errorf("user does not have access to this collection")
	}

	// Get the repository path
	repoPath := collection.Repo.LocalPath
	if repoPath == "" {
		return nil, fmt.Errorf("repository local path is not set")
	}

	// Get the collection path
	collectionPath := filepath.Join(repoPath, collection.Path)
	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("collection path does not exist: %s", collectionPath)
	}

	// Find all markdown files in the collection path
	var markdownFiles []string
	err := filepath.Walk(collectionPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".md") || strings.HasSuffix(info.Name(), ".markdown")) {
			markdownFiles = append(markdownFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk collection path: %w", err)
	}

	// Process each markdown file
	var syncedPosts []*models.Post
	for _, filePath := range markdownFiles {
		// Get the relative path from the collection path
		relPath, err := filepath.Rel(collectionPath, filePath)
		if err != nil {
			log.Printf("Failed to get relative path for %s: %v", filePath, err)
			continue
		}

		// Read the file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read file %s: %v", filePath, err)
			continue
		}

		// Parse the markdown content to extract title
		title := extractTitleFromMarkdown(string(content))
		if title == "" {
			// Use the filename as title if no title is found
			title = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		}

		// Check if a post with this file path already exists
		var existingPost models.Post
		result := s.db.Where("file_path = ? AND collection_id = ?", relPath, collectionID).First(&existingPost)

		if result.Error == nil {
			// Update the existing post
			existingPost.Title = title
			existingPost.UpdatedAt = time.Now()

			if err := s.db.Save(&existingPost).Error; err != nil {
				log.Printf("Failed to update post for %s: %v", filePath, err)
				continue
			}

			// Add content to the response
			syncedPosts = append(syncedPosts, &existingPost)
		} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create a new post
			newPost := &models.Post{
				Title:        title,
				FilePath:     relPath,
				UserID:       userID,
				CollectionID: collectionID,
			}

			if err := s.db.Create(newPost).Error; err != nil {
				log.Printf("Failed to create post for %s: %v", filePath, err)
				continue
			}

			syncedPosts = append(syncedPosts, newPost)
		} else {
			log.Printf("Error checking for existing post: %v", result.Error)
			continue
		}
	}

	return syncedPosts, nil
}

// extractTitleFromMarkdown extracts the title from markdown content
// It looks for a # heading or YAML frontmatter title
func extractTitleFromMarkdown(content string) string {
	// Check for YAML frontmatter
	if strings.HasPrefix(content, "---") {
		// Find the end of the frontmatter
		endIndex := strings.Index(content[3:], "---")
		if endIndex != -1 {
			frontmatter := content[3 : endIndex+3]
			// Look for title: in the frontmatter
			titleRegex := regexp.MustCompile(`(?m)^title:\s*(.+)$`)
			matches := titleRegex.FindStringSubmatch(frontmatter)
			if len(matches) > 1 {
				return strings.TrimSpace(matches[1])
			}
		}
	}

	// Look for the first # heading
	headingRegex := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	matches := headingRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}
