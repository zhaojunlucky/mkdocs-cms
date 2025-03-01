package models

import (
	"time"
)

// Post represents a blog post or article
type Post struct {
	ID           uint                 `json:"id" gorm:"primaryKey"`
	Title        string               `json:"title" gorm:"not null"`
	FilePath     string               `json:"file_path" gorm:"not null"`
	UserID       string               `json:"user_id" gorm:"not null"`
	User         User                 `json:"user" gorm:"foreignKey:UserID"`
	CollectionID uint                 `json:"collection_id" gorm:"not null"`
	Collection   UserGitRepoCollection `json:"collection" gorm:"foreignKey:CollectionID"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

// PostResponse is the structure returned to clients
type PostResponse struct {
	ID           uint                        `json:"id"`
	Title        string                      `json:"title"`
	Content      string                      `json:"content,omitempty"` // Content is loaded from file when needed
	FilePath     string                      `json:"file_path"`
	UserID       string                      `json:"user_id"`
	User         UserResponse                `json:"user,omitempty"`
	CollectionID uint                        `json:"collection_id"`
	Collection   UserGitRepoCollectionResponse `json:"collection,omitempty"`
	CreatedAt    time.Time                   `json:"created_at"`
	UpdatedAt    time.Time                   `json:"updated_at"`
}

// ToResponse converts a Post to a PostResponse
func (p *Post) ToResponse(includeUser bool, includeCollection bool) PostResponse {
	response := PostResponse{
		ID:           p.ID,
		Title:        p.Title,
		FilePath:     p.FilePath,
		UserID:       p.UserID,
		CollectionID: p.CollectionID,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}

	if includeUser {
		response.User = p.User.ToResponse()
	}

	if includeCollection {
		response.Collection = p.Collection.ToResponse(false)
	}

	return response
}

// CreatePostRequest is the structure for post creation requests
type CreatePostRequest struct {
	Title        string `json:"title" binding:"required"`
	Content      string `json:"content" binding:"required"` // Content to write to file
	FilePath     string `json:"file_path" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	CollectionID uint   `json:"collection_id" binding:"required"`
}

// UpdatePostRequest is the structure for post update requests
type UpdatePostRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"` // Content to write to file
	FilePath string `json:"file_path"`
}

// FilePostRequest is the structure for creating or updating a post from a file
type FilePostRequest struct {
	CollectionID uint   `json:"collection_id" binding:"required"`
	FilePath     string `json:"file_path" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
}
