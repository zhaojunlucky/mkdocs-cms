package main

import (
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/controllers"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/middleware"
	"github.com/zhaojunlucky/mkdocs-cms/models"
)

func main() {
	// Initialize database
	database.Initialize()

	// Initialize Gin router
	router := gin.New()

	// Apply middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())

	// Setup API routes
	setupRoutes(router)

	// Start the server
	log.Println("Server starting on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRoutes configures all the routes for our application
func setupRoutes(r *gin.Engine) {
	// Initialize controllers
	siteConfigController := controllers.NewSiteConfigController()
	
	// Initialize GitHub App controllers
	appID := int64(0)
	if appIDStr := os.Getenv("GITHUB_APP_ID"); appIDStr != "" {
		if id, err := strconv.ParseInt(appIDStr, 10, 64); err == nil {
			appID = id
		}
	}
	
	// Load private key for GitHub App
	var privateKey []byte
	if privateKeyPath := os.Getenv("GITHUB_APP_PRIVATE_KEY_PATH"); privateKeyPath != "" {
		var err error
		privateKey, err = os.ReadFile(privateKeyPath)
		if err != nil {
			log.Printf("Warning: Failed to read GitHub App private key: %v", err)
		}
	}
	
	// Create GitHub App settings
	githubAppSettings := &models.GitHubAppSettings{
		AppID:         appID,
		AppName:       os.Getenv("GITHUB_APP_NAME"),
		Description:   os.Getenv("GITHUB_APP_DESCRIPTION"),
		HomepageURL:   os.Getenv("GITHUB_APP_HOMEPAGE_URL"),
		WebhookURL:    os.Getenv("GITHUB_APP_WEBHOOK_URL"),
		WebhookSecret: os.Getenv("GITHUB_APP_WEBHOOK_SECRET"),
		PrivateKeyPath: os.Getenv("GITHUB_APP_PRIVATE_KEY_PATH"),
	}
	
	githubAppController := controllers.NewGitHubAppController(appID, privateKey, githubAppSettings)
	githubWebhookController := controllers.NewGitHubWebhookController(githubAppSettings.WebhookSecret)
	
	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// User routes
		v1.GET("/users", controllers.GetUsers)
		v1.GET("/users/:id", controllers.GetUser)
		v1.POST("/users", controllers.CreateUser)
		v1.PUT("/users/:id", controllers.UpdateUser)
		v1.DELETE("/users/:id", controllers.DeleteUser)

		// Post routes
		v1.GET("/posts", controllers.GetPosts)
		v1.GET("/posts/:id", controllers.GetPost)
		v1.POST("/posts", controllers.CreatePost)
		v1.PUT("/posts/:id", controllers.UpdatePost)
		v1.DELETE("/posts/:id", controllers.DeletePost)
		
		// Git Repository routes
		v1.GET("/repos", controllers.GetRepos)
		v1.GET("/users/:user_id/repos", controllers.GetReposByUser)
		v1.GET("/repos/:id", controllers.GetRepo)
		v1.POST("/repos", controllers.CreateRepo)
		v1.PUT("/repos/:id", controllers.UpdateRepo)
		v1.DELETE("/repos/:id", controllers.DeleteRepo)
		v1.POST("/repos/:id/sync", controllers.SyncRepo)
		
		// Event routes
		v1.GET("/events", controllers.GetEvents)
		v1.GET("/events/:id", controllers.GetEvent)
		v1.GET("/events/resources/:resource_type", controllers.GetEventsByResource)
		v1.POST("/events", controllers.CreateEvent)
		v1.PUT("/events/:id", controllers.UpdateEvent)
		v1.DELETE("/events/:id", controllers.DeleteEvent)
		
		// Collection routes
		v1.GET("/collections", controllers.GetCollections)
		v1.GET("/repos/:repo_id/collections", controllers.GetCollectionsByRepo)
		v1.GET("/collections/:id", controllers.GetCollection)
		v1.POST("/collections", controllers.CreateCollection)
		v1.PUT("/collections/:id", controllers.UpdateCollection)
		v1.DELETE("/collections/:id", controllers.DeleteCollection)
		v1.GET("/repos/:repo_id/collections/by-path", controllers.GetCollectionByPath)
		v1.GET("/collections/:collection_id/files", controllers.GetCollectionFiles)
		v1.GET("/collections/:collection_id/browse", controllers.GetCollectionFilesInPath)
		v1.GET("/collections/:collection_id/file", controllers.GetFileContent)
		v1.PUT("/collections/:collection_id/file", controllers.UpdateFileContent)
		v1.DELETE("/collections/:collection_id/file", controllers.DeleteFile)
		v1.GET("/collections/:collection_id/file/json", controllers.GetFileContentJSON)
		v1.POST("/collections/:collection_id/file", controllers.UploadFile)
		
		// Site Configuration routes
		v1.GET("/site-configs", siteConfigController.GetAllSiteConfigs)
		v1.GET("/site-configs/:id", siteConfigController.GetSiteConfigByID)
		v1.POST("/site-configs", siteConfigController.CreateSiteConfig)
		v1.PUT("/site-configs/:id", siteConfigController.UpdateSiteConfig)
		v1.DELETE("/site-configs/:id", siteConfigController.DeleteSiteConfig)
		
		// GitHub App routes
		github := v1.Group("/github")
		{
			github.GET("/app", githubAppController.GetAppInfo)
			github.GET("/installations", githubAppController.GetInstallations)
			github.GET("/installations/:installation_id/repositories", githubAppController.GetInstallationRepositories)
			github.POST("/installations/:installation_id/import", githubAppController.ImportRepositories)
			github.POST("/installations/:installation_id/webhooks", githubAppController.CreateWebhook)
		}
	}
	
	// GitHub webhook endpoint
	r.POST("/api/webhooks/github", githubWebhookController.HandleWebhook)

	// Health check
	r.GET("/health", controllers.HealthCheck)
}
