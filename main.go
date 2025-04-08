package main

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/config"
	"github.com/zhaojunlucky/mkdocs-cms/controllers"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/env"
	"github.com/zhaojunlucky/mkdocs-cms/middleware"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
	"github.com/zhaojunlucky/mkdocs-cms/utils"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

//go:embed web/dist/mkdocs-cms-ui/browser/*
var UIFS embed.FS
var Version string = "1.0.1-dev"

func main() {
	// Parse command line flags
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/config-dev.yaml"
	}
	log.Infof("Using config path: %s", configPath)

	port := 8080
	var err error
	portStr := os.Getenv("PORT")
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			log.Fatalf("Failed to parse PORT as an integer: %v", err)
		}
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
	log.Infof("Production: %v", env.IsProduction)
	log.Infof("Frontend URL: %s", appConfig.FrontendURL)
	log.Infof("OAuth Redirect URL: %s", appConfig.OAuth.RedirectURL)
	log.Infof("GitHub OAuth Client ID: %s", appConfig.GitHub.OAuth.ClientID)
	log.Debugf("GitHub OAuth Client Secret: %s (length: %d)",
		appConfig.GitHub.OAuth.ClientSecret[:4]+"...",
		len(appConfig.GitHub.OAuth.ClientSecret))
	log.Infof("GitHub App ID: %d", appConfig.GitHub.App.AppID)

	// Initialize database
	database.Initialize(ctx)

	// Initialize Gin router
	router := gin.New()

	if env.IsProduction {
		log.Infof("Using production mode for gin")
		gin.SetMode(gin.ReleaseMode)
	}
	// Serve static files
	subFS, err := fs.Sub(UIFS, "web/dist/mkdocs-cms-ui/browser")
	if err != nil {
		log.Fatalf("Failed to read UI files: %v", err)
	}
	router.NoRoute(func(c *gin.Context) {
		if middleware.UIPathReg.MatchString(c.Request.URL.Path) {
			c.FileFromFS(c.Request.URL.Path, http.FS(subFS))
			return
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		}
	})

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

	// Start the server
	log.Infof("Server starting on http://localhost:%d", port)
	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
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
		Version:           Version,
	}
	ctx.GithubAppClient = utils.CreateGitHubAppClient(ctx)
	return ctx
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

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	// Configure logrus
	log.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))
	log.SetFormatter(&log.JSONFormatter{})

	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
			//return frame.Function, fileName
			return "", fileName
		},
	})

}
