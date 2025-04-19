package models

import "gorm.io/gorm"

type UserRoleQuota struct {
	gorm.Model
	Role      *Role `gorm:"foreignKey:RoleID"`
	RoleID    string
	RepoCount int
	DocSize   int `gorm:"type:bigint;default:100"` // MB
}
