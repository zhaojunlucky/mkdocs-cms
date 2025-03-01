package models

import (
	"time"
)

// Post represents a blog post or article
type Post struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PostResponse is the structure returned to clients
type PostResponse struct {
	ID        uint         `json:"id"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	UserID    uint         `json:"user_id"`
	User      UserResponse `json:"user,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// ToResponse converts a Post to a PostResponse
func (p *Post) ToResponse(includeUser bool) PostResponse {
	response := PostResponse{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		UserID:    p.UserID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}

	if includeUser {
		response.User = p.User.ToResponse()
	}

	return response
}

// CreatePostRequest is the structure for post creation requests
type CreatePostRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	UserID  uint   `json:"user_id" binding:"required"`
}

// UpdatePostRequest is the structure for post update requests
type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
