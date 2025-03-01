package database

import (
	"log"

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
			log.Printf("Error adding file_path column: %v", err)
		}
	}

	if !DB.Migrator().HasColumn(&models.Post{}, "collection_id") {
		err := DB.Migrator().AddColumn(&models.Post{}, "collection_id")
		if err != nil {
			log.Printf("Error adding collection_id column: %v", err)
		}
	}

	// Drop Content column from posts table if it exists
	if DB.Migrator().HasColumn(&models.Post{}, "content") {
		err := DB.Migrator().DropColumn(&models.Post{}, "content")
		if err != nil {
			log.Printf("Error dropping content column: %v", err)
		} else {
			log.Println("Dropped content column from posts table")
		}
	}

	log.Println("Database migrations completed successfully")
}

// CreateTestUser creates a test user for development
func CreateTestUser() {
	log.Println("Creating test user...")

	// Check if test user already exists
	var count int64
	DB.Model(&models.User{}).Where("email = ?", "test@example.com").Count(&count)
	if count > 0 {
		log.Println("Test user already exists, skipping creation")
		return
	}

	// Create test user
	testUser := models.User{
		ID:        "test-user-id",
		Name:      "Test User",
		Email:     "test@example.com",
		AvatarURL: "https://via.placeholder.com/150",
		Provider:  "test",
	}

	result := DB.Create(&testUser)
	if result.Error != nil {
		log.Fatalf("Failed to create test user: %v", result.Error)
	}

	log.Println("Test user created successfully")
}
