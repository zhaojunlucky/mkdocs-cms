package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/config"
	"github.com/zhaojunlucky/mkdocs-cms/controllers"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/middleware"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
	"log"
	"os"
)

func main() {
	// Parse command line flags
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/oauth_config.yaml"
	}

	migrate := flag.Bool("migrate", false, "Run database migrations")
	createTestUser := flag.Bool("create-test-user", false, "Create a test user")
	flag.Parse()

	// Load configuration
	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Debug: Print loaded configuration
	log.Printf("Loaded configuration:")
	log.Printf("OAuth Redirect URL: %s", appConfig.OAuth.RedirectURL)
	log.Printf("GitHub OAuth Client ID: %s", appConfig.GitHub.OAuth.ClientID)
	log.Printf("GitHub OAuth Client Secret: %s (length: %d)",
		appConfig.GitHub.OAuth.ClientSecret[:4]+"...",
		len(appConfig.GitHub.OAuth.ClientSecret))
	log.Printf("GitHub App ID: %d", appConfig.GitHub.App.AppID)

	// Initialize database
	database.Initialize()

	// Run migrations if requested
	if *migrate {
		log.Println("Running database migrations...")
		database.MigrateDB()
		if *createTestUser {
			database.CreateTestUser()
		}
		return
	}

	// Initialize Gin router
	router := gin.New()

	// Apply middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORSWithConfig(appConfig))
	router.Use(middleware.NewAuthMiddleware()) // Update to use new auth middleware
	router.Use(gin.Recovery())

	// Setup API routes
	setupRoutes(router, appConfig)

	// Start the server
	log.Println("Server starting on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRoutes configures all the routes for our application
func setupRoutes(r *gin.Engine, appConfig *config.Config) {
	// Create GitHub App settings
	githubAppSettings := &models.GitHubAppSettings{
		AppID:          appConfig.GitHub.App.AppID,
		AppName:        appConfig.GitHub.App.Name,
		Description:    appConfig.GitHub.App.Description,
		HomepageURL:    appConfig.GitHub.App.HomepageURL,
		WebhookURL:     appConfig.GitHub.App.WebhookURL,
		WebhookSecret:  appConfig.GitHub.App.WebhookSecret,
		PrivateKeyPath: appConfig.GitHub.App.PrivateKeyPath,
	}

	controllers.InitUserGitRepoCollectionController(githubAppSettings)

	// Read private key
	bytes, err := os.ReadFile(appConfig.GitHub.App.PrivateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key file: %v", err)
		panic(err)
	}

	// Initialize services
	userService := services.NewUserService()
	userGitRepoService := services.NewUserGitRepoService(githubAppSettings)

	// Initialize controllers
	authController := controllers.NewAuthController(userService, appConfig)
	postController := controllers.NewPostController(githubAppSettings)
	asyncTaskController := controllers.NewAsyncTaskController()
	userGitRepoController := controllers.NewUserGitRepoController(userGitRepoService)
	collectionController := controllers.NewUserGitRepoCollectionController()

	// Initialize GitHub App controllers
	githubAppController := controllers.NewGitHubAppController(
		appConfig.GitHub.App.AppID,
		bytes,
		githubAppSettings,
		userGitRepoService,
	)
	githubWebhookController := controllers.NewGitHubWebhookController(appConfig.GitHub.App.WebhookSecret, githubAppSettings)

	// API routes
	api := r.Group("/api")

	// Register auth routes
	authController.RegisterRoutes(api)

	// API v1 routes
	v1 := api.Group("/v1")
	{
		// Post routes - Register the post controller
		postController.RegisterRoutes(v1)

		// Git Repository routes
		repos := v1.Group("/repos")
		repos.Use(middleware.RequireAuth())
		{
			//repos.GET("", userGitRepoController.GetRepos)
			repos.GET("/:id", userGitRepoController.GetRepo)
			repos.PUT("/:id", userGitRepoController.UpdateRepo)
			repos.DELETE("/:id", userGitRepoController.DeleteRepo)
			repos.POST("/:id/sync", userGitRepoController.SyncRepo)
			repos.GET("/:id/branches", userGitRepoController.GetRepoBranches)
		}

		// User repositories route
		userRepos := v1.Group("/users/repos")
		userRepos.Use(middleware.RequireAuth())
		{
			userRepos.GET("/:user_id", userGitRepoController.GetReposByUser)
		}

		// Collection routes
		collections := v1.Group("/collections")
		{
			collections.GET("/repo/:repoId", collectionController.GetCollectionsByRepo)

			// Collection file routes
			collections.GET("/repo/:repoId/:collectionName/files", collectionController.GetCollectionFiles)
			collections.POST("/repo/:repoId/:collectionName/files/folder", collectionController.CreateFolder)
			collections.GET("/repo/:repoId/:collectionName/files/path", collectionController.GetCollectionFilesInPath)
			collections.GET("/repo/:repoId/:collectionName/files/content", collectionController.GetFileContent)
			collections.PUT("/repo/:repoId/:collectionName/files/content", collectionController.UpdateFileContent)
			collections.DELETE("/repo/:repoId/:collectionName/files", collectionController.DeleteFile)
			collections.POST("/repo/:repoId/:collectionName/files/upload", collectionController.UploadFile)
			collections.PUT("/repo/:repoId/:collectionName/files/rename", collectionController.RenameFile)
		}

		// GitHub App routes
		github := v1.Group("/github")
		github.Use(middleware.RequireAuth())
		{
			github.GET("/app", githubAppController.GetAppInfo)
			github.GET("/installations", githubAppController.GetInstallations)
			github.GET("/installations/:installation_id/repositories", githubAppController.GetInstallationRepositories)
			github.POST("/installations/:installation_id/import", githubAppController.ImportRepositories)
		}

		// AsyncTask routes
		asyncTaskController.RegisterRoutes(v1)
	}

	// GitHub webhook endpoint
	r.POST("/api/webhooks/github", githubWebhookController.HandleWebhook)

	// Health check
	r.GET("/health", controllers.HealthCheck)

	// Serve Angular app for any other routes
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/web/browser/index.html")
	})
}
