package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/controllers"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/middleware"
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
	}

	// Health check
	r.GET("/health", controllers.HealthCheck)
}
