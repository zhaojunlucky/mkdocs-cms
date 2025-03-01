package services

import (
	"errors"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"gorm.io/gorm"
)

// PostService handles business logic related to posts
type PostService struct{}

// NewPostService creates a new post service
func NewPostService() *PostService {
	return &PostService{}
}

// GetAllPosts retrieves all posts from the database
func (s *PostService) GetAllPosts() ([]models.PostResponse, error) {
	var posts []models.Post
	result := database.DB.Preload("User").Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}

	// Convert to response objects
	var responses []models.PostResponse
	for _, post := range posts {
		responses = append(responses, post.ToResponse(true))
	}

	return responses, nil
}

// GetPostByID retrieves a post by ID
func (s *PostService) GetPostByID(id uint) (models.PostResponse, error) {
	var post models.Post
	result := database.DB.Preload("User").First(&post, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.PostResponse{}, errors.New("post not found")
		}
		return models.PostResponse{}, result.Error
	}

	return post.ToResponse(true), nil
}

// CreatePost creates a new post
func (s *PostService) CreatePost(req models.CreatePostRequest) (models.PostResponse, error) {
	// Verify that the user exists
	var user models.User
	result := database.DB.First(&user, req.UserID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.PostResponse{}, errors.New("user not found")
		}
		return models.PostResponse{}, result.Error
	}

	post := models.Post{
		Title:   req.Title,
		Content: req.Content,
		UserID:  req.UserID,
	}

	result = database.DB.Create(&post)
	if result.Error != nil {
		return models.PostResponse{}, errors.New("failed to create post")
	}

	// Fetch the complete post with user data
	database.DB.Preload("User").First(&post, post.ID)

	return post.ToResponse(true), nil
}

// UpdatePost updates an existing post
func (s *PostService) UpdatePost(id uint, req models.UpdatePostRequest) (models.PostResponse, error) {
	var post models.Post
	result := database.DB.First(&post, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.PostResponse{}, errors.New("post not found")
		}
		return models.PostResponse{}, result.Error
	}

	// Update fields if provided
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}

	result = database.DB.Save(&post)
	if result.Error != nil {
		return models.PostResponse{}, errors.New("failed to update post")
	}

	// Fetch the complete post with user data
	database.DB.Preload("User").First(&post, post.ID)

	return post.ToResponse(true), nil
}

// DeletePost deletes a post by ID
func (s *PostService) DeletePost(id uint) error {
	result := database.DB.Delete(&models.Post{}, id)
	if result.Error != nil {
		return errors.New("failed to delete post")
	}

	if result.RowsAffected == 0 {
		return errors.New("post not found")
	}

	return nil
}
