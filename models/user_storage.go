package models

import "gorm.io/gorm"

type UserStorage struct {
	gorm.Model
	UserID     string `gorm:"not null uniqueIndex"`
	User       User   `gorm:"foreignKey:UserID"`
	BucketName string
}

type UserStorageFile struct {
	gorm.Model
	UserStorageID uint
	UserStorage   UserStorage `gorm:"foreignKey:UserStorageID"`
	FileName      string      // origin file name
	Url           string      // /repo/:repoId/:collectionName/storage/2025-04-18-{collectionName}-uuid.xxx
	Path          string      `gorm:"index"` // 2025-04/{collectionName}-uuid.xxx minio path
	ContentType   string
}
