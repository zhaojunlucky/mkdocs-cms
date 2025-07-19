package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID         string    `json:"id" gorm:"primaryKey;type:text"`
	Username   string    `json:"username" gorm:"unique;not null"`
	Email      string    `json:"email" gorm:"unique;not null"`
	Password   string    `json:"-"` // Password is not exposed in JSON
	Name       string    `json:"name"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	AvatarURL  string    `json:"avatar_url"`
	Provider   string    `json:"provider"`    // OAuth provider (github, google, etc.)
	ProviderID string    `json:"provider_id"` // ID from the OAuth provider
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsActive   bool      `json:"is_active"`
	Roles      []*Role   `gorm:"many2many:user_roles;"` // Use pointer slice []*Role
}

func (u *User) GetQuotaRole() *Role {
	for _, role := range u.Roles {
		if role.Quota {
			return role
		}
	}
	return nil
}

// UserResponse is the structure returned to clients
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL string    `json:"avatar_url"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt int64     `json:"expires_at"`
}

// ToResponse converts a User to a UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Name:      u.Name,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		AvatarURL: u.AvatarURL,
		Provider:  u.Provider,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (u *User) ToResponseWithExpires(expireInt int64) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Name:      u.Name,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		AvatarURL: u.AvatarURL,
		Provider:  u.Provider,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		ExpiresAt: expireInt,
	}
}

// CreateUserRequest is the structure for user creation requests
type CreateUserRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UpdateUserRequest is the structure for user update requests
type UpdateUserRequest struct {
	Email     string `json:"email" binding:"omitempty,email"`
	Password  string `json:"password" binding:"omitempty,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
