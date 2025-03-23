package database

import (
	log "github.com/sirupsen/logrus"

	"github.com/zhaojunlucky/mkdocs-cms/models"
)

// MigrateDB runs database migrations
func MigrateDB() {
	log.Println("Running database migrations...")

	// Auto migrate all models
	err := DB.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.UserGitRepo{},
		&models.Event{},
		&models.UserGitRepoCollection{},
		&models.SiteConfig{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Add FilePath and CollectionID columns to posts table if they don't exist
	if !DB.Migrator().HasColumn(&models.Post{}, "file_path") {
		err := DB.Migrator().AddColumn(&models.Post{}, "file_path")
		if err != nil {
			log.Infof("Error adding file_path column: %v", err)
		}
	}

	if !DB.Migrator().HasColumn(&models.Post{}, "collection_id") {
		err := DB.Migrator().AddColumn(&models.Post{}, "collection_id")
		if err != nil {
			log.Infof("Error adding collection_id column: %v", err)
		}
	}

	// Drop Content column from posts table if it exists
	if DB.Migrator().HasColumn(&models.Post{}, "content") {
		err := DB.Migrator().DropColumn(&models.Post{}, "content")
		if err != nil {
			log.Infof("Error dropping content column: %v", err)
		} else {
			log.Info("Dropped content column from posts table")
		}
	}

	log.Info("Database migrations completed successfully")
}
