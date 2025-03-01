package database

import (
	"log"

	"github.com/zhaojunlucky/mkdocs-cms/models"
)

// MigrateDB performs database migrations
func MigrateDB() {
	if DB == nil {
		log.Fatal("Database not initialized")
	}

	// Drop tables if they exist
	for _, model := range []interface{}{
		&models.User{},
		&models.Post{},
		&models.UserGitRepo{},
		&models.Event{},
		&models.UserGitRepoCollection{},
		&models.SiteConfig{},
	} {
		if DB.Migrator().HasTable(model) {
			if err := DB.Migrator().DropTable(model); err != nil {
				log.Printf("Warning: Failed to drop table for %T: %v", model, err)
			}
		}
	}

	// Create tables with new schema
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

	log.Println("Database migration completed successfully")
}

// CreateTestUser creates a test user for development
func CreateTestUser() {
	user := models.User{
		ID:       "test-user-id-123456",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "$2a$10$zJBMxLd8H4ywXUzWJJA.wOmFj1V/Jv/q8W8X0ICDYFxVdwrKKK1Uy", // password is "password"
		Name:     "Test User",
	}

	result := DB.Create(&user)
	if result.Error != nil {
		log.Printf("Failed to create test user: %v", result.Error)
	} else {
		log.Println("Test user created successfully")
	}
}
