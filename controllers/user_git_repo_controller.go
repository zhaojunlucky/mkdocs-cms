package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

var userGitRepoService = services.NewUserGitRepoService()

// GetRepos returns all git repositories for the authenticated user
func GetRepos(c *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get only the repositories owned by the authenticated user
	repos, err := userGitRepoService.GetReposByUser(authenticatedUserID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []models.UserGitRepoResponse
	for _, repo := range repos {
		response = append(response, repo.ToResponse(false))
	}

	c.JSON(http.StatusOK, response)
}

// GetReposByUser returns all git repositories for a specific user
func GetReposByUser(c *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get the requested user ID from the URL parameter
	requestedUserID := c.Param("user_id")

	// Check if the authenticated user is trying to access their own repositories
	// This ensures users can only see their own repositories
	if authenticatedUserID.(string) != requestedUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own repositories"})
		return
	}

	repos, err := userGitRepoService.GetReposByUser(requestedUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []models.UserGitRepoResponse = make([]models.UserGitRepoResponse, 0)
	for _, repo := range repos {
		response = append(response, repo.ToResponse(false))
	}

	c.JSON(http.StatusOK, response)
}

// GetRepo returns a specific git repository
func GetRepo(c *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	repo, err := userGitRepoService.GetRepoByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Check if the authenticated user owns this repository
	if repo.UserID != authenticatedUserID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own repositories"})
		return
	}

	c.JSON(http.StatusOK, repo.ToResponse(true))
}

// CreateRepo creates a new git repository
func CreateRepo(c *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var repo models.UserGitRepo
	if err := c.ShouldBindJSON(&repo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the user ID to the authenticated user
	repo.UserID = authenticatedUserID.(string)

	if err := userGitRepoService.CreateRepo(&repo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, repo.ToResponse(false))
}

// UpdateRepo updates an existing git repository
func UpdateRepo(c *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	
	// First, get the existing repository to check ownership
	existingRepo, err := userGitRepoService.GetRepoByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}
	
	// Check if the authenticated user owns this repository
	if existingRepo.UserID != authenticatedUserID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own repositories"})
		return
	}

	var request models.UpdateUserGitRepoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repo, err := userGitRepoService.UpdateRepo(id, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, repo.ToResponse(false))
}

// DeleteRepo deletes a git repository
func DeleteRepo(c *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	
	// First, get the existing repository to check ownership
	existingRepo, err := userGitRepoService.GetRepoByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}
	
	// Check if the authenticated user owns this repository
	if existingRepo.UserID != authenticatedUserID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own repositories"})
		return
	}

	if err := userGitRepoService.DeleteRepo(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Repository deleted successfully"})
}

// SyncRepo synchronizes a git repository with its remote
func SyncRepo(c *gin.Context) {
	// Get the authenticated user ID from the context
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	
	// First, get the existing repository to check ownership
	repo, err := userGitRepoService.GetRepoByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}
	
	// Check if the authenticated user owns this repository
	if repo.UserID != authenticatedUserID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only sync your own repositories"})
		return
	}

	// Update status to syncing
	if err := userGitRepoService.UpdateRepoStatus(id, models.StatusSyncing, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Start sync in a goroutine
	go func() {
		if err := userGitRepoService.SyncRepo(id); err != nil {
			userGitRepoService.UpdateRepoStatus(id, models.StatusFailed, err.Error())
		} else {
			userGitRepoService.UpdateRepoStatus(id, models.StatusSynced, "")
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Repository sync started"})
}
