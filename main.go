package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/config"
	"github.com/zhaojunlucky/mkdocs-cms/controllers"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/middleware"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
	"github.com/zhaojunlucky/mkdocs-cms/utils"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	// Parse command line flags
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/config-dev.yaml"
	}

	// Load configuration
	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	ctx := createContext(appConfig)

	setupLog(ctx)

	// Debug: Print loaded configuration
	log.Infof("Loaded configuration:")
	log.Infof("OAuth Redirect URL: %s", appConfig.OAuth.RedirectURL)
	log.Infof("GitHub OAuth Client ID: %s", appConfig.GitHub.OAuth.ClientID)
	log.Debugf("GitHub OAuth Client Secret: %s (length: %d)",
		appConfig.GitHub.OAuth.ClientSecret[:4]+"...",
		len(appConfig.GitHub.OAuth.ClientSecret))
	log.Infof("GitHub App ID: %d", appConfig.GitHub.App.AppID)

	// Initialize database
	database.Initialize()

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
	log.Infof("Server starting on http://localhost:8080")
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
		RepoBasePath:      filepath.Join(appConfig.WorkingDir, "repos"),
		LogDirPath:        filepath.Join(appConfig.WorkingDir, "log"),
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

func setupLog(ctx *core.APPContext) {

	switch strings.ToLower(ctx.Config.LogLevel) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Fatalf("invalid log level: %s", ctx.Config.LogLevel)
	}
	logPath := ctx.LogDirPath
	log.Infof("log path: %s", logPath)

	fiInfo, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(logPath, 0755)
		if err != nil {
			panic(err)
		}
	} else if !fiInfo.IsDir() {
		log.Fatalf("%s must be a directory", logPath)
	}

	logFilePath := path.Join(logPath, "mkdocs-cms.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to log to file, using default stderr")
	}
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
			//return frame.Function, fileName
			return "", fileName
		},
	})

	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

}
