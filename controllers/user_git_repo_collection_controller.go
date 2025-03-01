package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

var userGitRepoCollectionService = services.NewUserGitRepoCollectionService()

// GetCollections returns all collections
func GetCollections(c *gin.Context) {
	collections, err := userGitRepoCollectionService.GetAllCollections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []models.UserGitRepoCollectionResponse
	for _, collection := range collections {
		response = append(response, collection.ToResponse(false))
	}

	c.JSON(http.StatusOK, response)
}

// GetCollectionsByRepo returns all collections for a specific repository
func GetCollectionsByRepo(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collections, err := userGitRepoCollectionService.GetCollectionsByRepo(uint(repoID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []models.UserGitRepoCollectionResponse
	for _, collection := range collections {
		response = append(response, collection.ToResponse(false))
	}

	c.JSON(http.StatusOK, response)
}

// GetCollection returns a specific collection
func GetCollection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection, err := userGitRepoCollectionService.GetCollectionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	c.JSON(http.StatusOK, collection.ToResponse(true))
}

// CreateCollection creates a new collection
func CreateCollection(c *gin.Context) {
	var request models.CreateUserGitRepoCollectionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection, err := userGitRepoCollectionService.CreateCollection(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, collection.ToResponse(false))
}

// UpdateCollection updates an existing collection
func UpdateCollection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var request models.UpdateUserGitRepoCollectionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection, err := userGitRepoCollectionService.UpdateCollection(uint(id), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, collection.ToResponse(false))
}

// DeleteCollection deletes a collection
func DeleteCollection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := userGitRepoCollectionService.DeleteCollection(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collection deleted successfully"})
}

// GetCollectionByPath returns a collection by its path
func GetCollectionByPath(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path parameter is required"})
		return
	}

	collection, err := userGitRepoCollectionService.GetCollectionByPath(uint(repoID), path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	c.JSON(http.StatusOK, collection.ToResponse(true))
}
