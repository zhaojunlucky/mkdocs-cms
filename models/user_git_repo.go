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
	// StatusWarning indicates the repository has a warning (e.g., missing config)
	StatusWarning GitRepoStatus = "warning"
)

// UserGitRepo represents a git repository associated with a user
type UserGitRepo struct {
	ID             uint          `json:"id" gorm:"primaryKey"`
	GitRepoID      int64         `json:"git_repo_id" gorm:"not null"`
	Name           string        `json:"name" gorm:"not null"`
	Description    string        `json:"description"`
	LocalPath      string        `json:"local_path" gorm:"not null"`
	RemoteURL      string        `json:"remote_url" gorm:"not null"`
	CloneURL       string        `json:"clone_url"`
	Branch         string        `json:"branch" gorm:"default:main"`
	UserID         string        `json:"user_id" gorm:"not null"`
	User           User          `json:"user" gorm:"foreignKey:UserID"`
	AuthType       string        `json:"auth_type" gorm:"default:'none'"`
	AuthData       string        `json:"auth_data"`
	Provider       string        `json:"provider" gorm:"default:'git'"`
	LastSyncAt     time.Time     `json:"last_sync_at"`
	Status         GitRepoStatus `json:"status" gorm:"type:string;default:'pending'"`
	ErrorMsg       string        `json:"error_msg"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	InstallationID int64         `json:"installation_id"`
}

// UserGitRepoResponse is the structure returned to clients
type UserGitRepoResponse struct {
	ID             uint          `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	LocalPath      string        `json:"local_path"`
	RemoteURL      string        `json:"remote_url"`
	CloneURL       string        `json:"clone_url"`
	Branch         string        `json:"branch"`
	UserID         string        `json:"user_id"`
	User           UserResponse  `json:"user,omitempty"`
	AuthType       string        `json:"auth_type"`
	LastSyncAt     time.Time     `json:"last_sync_at"`
	Status         GitRepoStatus `json:"status"`
	ErrorMsg       string        `json:"error_msg,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	InstallationID int64         `json:"installation_id"`
}

// ToResponse converts a UserGitRepo to a UserGitRepoResponse
func (r *UserGitRepo) ToResponse(includeUser bool) UserGitRepoResponse {
	response := UserGitRepoResponse{
		ID:             r.ID,
		Name:           r.Name,
		Description:    r.Description,
		LocalPath:      r.LocalPath,
		RemoteURL:      r.RemoteURL,
		CloneURL:       r.CloneURL,
		Branch:         r.Branch,
		UserID:         r.UserID,
		AuthType:       r.AuthType,
		LastSyncAt:     r.LastSyncAt,
		Status:         r.Status,
		ErrorMsg:       r.ErrorMsg,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		InstallationID: r.InstallationID,
	}

	if includeUser {
		response.User = r.User.ToResponse()
	}

	return response
}

// UpdateUserGitRepoRequest is the structure for repository update requests
type UpdateUserGitRepoRequest struct {
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	RemoteURL      string        `json:"remote_url"`
	CloneURL       string        `json:"clone_url"`
	Branch         string        `json:"branch"`
	AuthType       string        `json:"auth_type"`
	AuthData       string        `json:"auth_data"`
	Status         GitRepoStatus `json:"status"`
	ErrorMsg       string        `json:"error_msg"`
	InstallationID int64         `json:"installation_id"`
}
