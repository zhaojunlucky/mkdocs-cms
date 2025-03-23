package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/config"
	"github.com/zhaojunlucky/mkdocs-cms/controllers"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/middleware"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
	"github.com/zhaojunlucky/mkdocs-cms/utils"
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
		return
	}

	ctx := createContext(appConfig)
	// Initialize Gin router
	router := gin.New()

	// Apply middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORSWithConfig(appConfig))
	router.Use(middleware.NewAuthMiddleware(ctx)) // Update to use new auth middleware
	router.Use(gin.Recovery())

	services.InitServices(ctx)

	api := router.Group("/api")
	controllers.InitAPIControllers(ctx, api)
	v1 := api.Group("/v1")
	controllers.InitV1Controllers(ctx, v1)

	// Setup API routes
	setupRoutes(router)

	// Start the server
	log.Println("Server starting on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createContext(appConfig *config.Config) *core.APPContext {
	githubAppSettings := &models.GitHubAppSettings{
		AppID:          appConfig.GitHub.App.AppID,
		AppName:        appConfig.GitHub.App.Name,
		Description:    appConfig.GitHub.App.Description,
		HomepageURL:    appConfig.GitHub.App.HomepageURL,
		WebhookURL:     appConfig.GitHub.App.WebhookURL,
		WebhookSecret:  appConfig.GitHub.App.WebhookSecret,
		PrivateKeyPath: appConfig.GitHub.App.PrivateKeyPath,
	}
	ctx := &core.APPContext{
		GithubAppSettings: githubAppSettings,
		Config:            appConfig,
		RepoBasePath:      appConfig.RepoBasePath,
		ServiceMap:        make(map[string]interface{}),
	}
	ctx.GithubAppClient = utils.CreateGitHubAppClient(ctx)
	return ctx
}

// setupRoutes configures all the routes for our application
func setupRoutes(r *gin.Engine) {

	// Serve Angular app for any other routes
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/web/browser/index.html")
	})
}
