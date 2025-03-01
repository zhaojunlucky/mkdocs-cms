package models

import (
	"time"
)

// GitRepoStatus represents the status of a git repository
type GitRepoStatus string

const (
	// StatusSynced indicates the repository is in sync with remote
	StatusSynced GitRepoStatus = "synced"
	// StatusSyncing indicates the repository is currently syncing
	StatusSyncing GitRepoStatus = "syncing"
	// StatusFailed indicates the repository sync failed
	StatusFailed GitRepoStatus = "failed"
	// StatusPending indicates the repository is pending for sync
	StatusPending GitRepoStatus = "pending"
)

// UserGitRepo represents a git repository associated with a user
type UserGitRepo struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"not null"`
	Description string       `json:"description"`
	LocalPath   string       `json:"local_path" gorm:"not null"`
	RemoteURL   string       `json:"remote_url" gorm:"not null"`
	Branch      string       `json:"branch" gorm:"default:main"`
	UserID      uint         `json:"user_id" gorm:"not null"`
	User        User         `json:"user" gorm:"foreignKey:UserID"`
	LastSyncAt  time.Time    `json:"last_sync_at"`
	Status      GitRepoStatus `json:"status" gorm:"type:string;default:'pending'"`
	ErrorMsg    string       `json:"error_msg"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// UserGitRepoResponse is the structure returned to clients
type UserGitRepoResponse struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	LocalPath   string       `json:"local_path"`
	RemoteURL   string       `json:"remote_url"`
	Branch      string       `json:"branch"`
	UserID      uint         `json:"user_id"`
	User        UserResponse `json:"user,omitempty"`
	LastSyncAt  time.Time    `json:"last_sync_at"`
	Status      GitRepoStatus `json:"status"`
	ErrorMsg    string       `json:"error_msg,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// ToResponse converts a UserGitRepo to a UserGitRepoResponse
func (r *UserGitRepo) ToResponse(includeUser bool) UserGitRepoResponse {
	response := UserGitRepoResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		LocalPath:   r.LocalPath,
		RemoteURL:   r.RemoteURL,
		Branch:      r.Branch,
		UserID:      r.UserID,
		LastSyncAt:  r.LastSyncAt,
		Status:      r.Status,
		ErrorMsg:    r.ErrorMsg,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	if includeUser {
		response.User = r.User.ToResponse()
	}

	return response
}

// CreateUserGitRepoRequest is the structure for repository creation requests
type CreateUserGitRepoRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	RemoteURL   string `json:"remote_url" binding:"required"`
	Branch      string `json:"branch"`
	UserID      uint   `json:"user_id" binding:"required"`
}

// UpdateUserGitRepoRequest is the structure for repository update requests
type UpdateUserGitRepoRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	RemoteURL   string       `json:"remote_url"`
	Branch      string       `json:"branch"`
	Status      GitRepoStatus `json:"status"`
	ErrorMsg    string       `json:"error_msg"`
}
