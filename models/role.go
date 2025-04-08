package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model         // Includes fields ID, CreatedAt, UpdatedAt, DeletedAt
	Name        string `gorm:"uniqueIndex;not null;size:50"` // Role name (e.g., "admin", "editor")
	Description string `gorm:"size:255"`                     // Optional description

	// --- Relationships ---
	// Many-to-Many relationship with User (inverse of User.Roles)
	Users []*User `gorm:"many2many:user_roles;"` // Use pointer slice []*User
}
