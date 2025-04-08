package database

import (
	"fmt"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize sets up the database connection
func Initialize(ctx *core.APPContext) {
	var err error

	// Get database path from environment variable or use default
	dbPath := filepath.Join(ctx.Config.WorkingDir, "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		err := os.MkdirAll(dbPath, 0755) // Creates parent directories recursively with permissions 0755
		if err != nil {
			log.Fatalf("error creating directory %s: %v", dbPath, err)
		}
		fmt.Printf("Successfully created directory: %s\n", dbPath)
	}
	dbPath = filepath.Join(dbPath, "cms.db")

	// Configure GORM logger
	newLogger := logger.New(
		log.New(),
		logger.Config{
			LogLevel: logger.Info,
		},
	)

	// Open database connection
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Info("Connected to database")

	// Auto migrate the schema
	err = DB.AutoMigrate(&models.User{}, &models.UserGitRepo{}, &models.Event{}, &models.AsyncTask{}, &models.SiteSetting{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Info("Database migration completed")
}
