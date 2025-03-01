package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/middleware"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

// PostController handles HTTP requests related to posts
type PostController struct {
	postService *services.PostService
}

// NewPostController creates a new post controller
func NewPostController() *PostController {
	return &PostController{
		postService: services.NewPostService(),
	}
}

// RegisterRoutes registers routes for the post controller
func (c *PostController) RegisterRoutes(router *gin.RouterGroup) {
	posts := router.Group("/posts")
	{
		posts.GET("", c.GetAllPosts)
		posts.GET("/:id", c.GetPostByID)
		posts.POST("", middleware.RequireAuth(), c.CreatePost)
		posts.PUT("/:id", middleware.RequireAuth(), c.UpdatePost)
		posts.DELETE("/:id", middleware.RequireAuth(), c.DeletePost)
		
		// Collection-specific routes
		posts.GET("/collection/:collectionId", c.GetPostsByCollection)
		posts.GET("/collection/:collectionId/file/*filePath", c.GetPostByPath)
		posts.POST("/collection/:collectionId/sync", middleware.RequireAuth(), c.SyncPostsFromCollection)
		posts.POST("/collection/:collectionId/sync-collection", middleware.RequireAuth(), c.SyncCollectionPosts)
		posts.POST("/file", middleware.RequireAuth(), c.CreatePostFromFile)
	}
}

// GetAllPosts retrieves all posts
func (c *PostController) GetAllPosts(ctx *gin.Context) {
	posts, err := c.postService.GetAllPosts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, posts)
}

// GetPostsByCollection retrieves all posts from a specific collection
func (c *PostController) GetPostsByCollection(ctx *gin.Context) {
	collectionID, err := strconv.ParseUint(ctx.Param("collectionId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection ID"})
		return
	}

	posts, err := c.postService.GetPostsByCollection(uint(collectionID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, posts)
}

// GetPostByID retrieves a post by ID
func (c *PostController) GetPostByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	post, err := c.postService.GetPostByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, post)
}

// GetPostByPath retrieves a post by its file path within a collection
func (c *PostController) GetPostByPath(ctx *gin.Context) {
	collectionID, err := strconv.ParseUint(ctx.Param("collectionId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection ID"})
		return
	}

	// Get the file path parameter and remove the leading slash
	filePath := ctx.Param("filePath")
	if len(filePath) > 0 && filePath[0] == '/' {
		filePath = filePath[1:]
	}

	post, err := c.postService.GetPostByPath(uint(collectionID), filePath)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, post)
}

// CreatePost creates a new post
func (c *PostController) CreatePost(ctx *gin.Context) {
	var req models.CreatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	req.UserID = userID.(string)

	post, err := c.postService.CreatePost(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, post)
}

// CreatePostFromFile creates a new post from an existing file
func (c *PostController) CreatePostFromFile(ctx *gin.Context) {
	var req models.FilePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	req.UserID = userID.(string)

	post, err := c.postService.CreatePostFromFile(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, post)
}

// UpdatePost updates an existing post
func (c *PostController) UpdatePost(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	var req models.UpdatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := c.postService.UpdatePost(uint(id), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, post)
}

// DeletePost deletes a post
func (c *PostController) DeletePost(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	err = c.postService.DeletePost(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
}

// SyncPostsFromCollection synchronizes posts from files in a collection
func (c *PostController) SyncPostsFromCollection(ctx *gin.Context) {
	collectionID, err := strconv.ParseUint(ctx.Param("collectionId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection ID"})
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := c.postService.SyncPostsFromCollection(uint(collectionID), userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "posts synchronized successfully",
		"count":   count,
	})
}

// SyncCollectionPosts synchronizes posts from a collection's repository files
func (c *PostController) SyncCollectionPosts(ctx *gin.Context) {
	// Get collection ID from path parameter
	collectionIDStr := ctx.Param("collectionId")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Sync posts from the collection
	syncedPosts, err := c.postService.SyncCollectionPosts(uint(collectionID), userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync posts: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Posts synchronized successfully", "posts": syncedPosts})
}
