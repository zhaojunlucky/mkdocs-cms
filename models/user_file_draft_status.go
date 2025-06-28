package models

import (
	"time"
)

type UserFileDraftStatus struct {
	ID             uint `gorm:"primaryKey"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	UserID         string `gorm:"index:idx_user_repo_collection,not null"`
	RepoID         uint   `gorm:"index:idx_user_repo_collection,not null"`
	CollectionName string `gorm:"index:idx_user_repo_collection,not null"`
	FilePath       string `gorm:"not null"`
	Draft          bool   `gorm:"not null"`
}
