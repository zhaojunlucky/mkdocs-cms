package models

import "gorm.io/gorm"

type SiteSetting struct {
	gorm.Model
	Key   string `json:"key"`
	Value string `json:"value"`
}
