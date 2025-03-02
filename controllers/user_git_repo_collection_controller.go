package controllers

import (
	"encoding/base64"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	id, err := strconv.ParseUint(c.Param("collection_id"), 10, 32)
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
// This method is kept for backward compatibility but returns an error message
// as collections are now read-only from veda/config.yml
func CreateCollection(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Collections are now defined in veda/config.yml file in the repository. Please edit that file directly.",
	})
}

// UpdateCollection updates an existing collection
// This method is kept for backward compatibility but returns an error message
// as collections are now read-only from veda/config.yml
func UpdateCollection(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Collections are now defined in veda/config.yml file in the repository. Please edit that file directly.",
	})
}

// DeleteCollection deletes a collection
// This method is kept for backward compatibility but returns an error message
// as collections are now read-only from veda/config.yml
func DeleteCollection(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Collections are now defined in veda/config.yml file in the repository. Please edit that file directly.",
	})
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

// GetCollectionFiles returns all files in a collection
func GetCollectionFiles(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collection_name")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	files, err := userGitRepoCollectionService.ListFilesInCollection(uint(repoID), collectionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, files)
}

// GetCollectionFilesInPath returns all files in a specific path within a collection
func GetCollectionFilesInPath(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collection_name")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	path := c.Query("path")
	if path == "" {
		// If no path is provided, return files at the root of the collection
		files, err := userGitRepoCollectionService.ListFilesInCollection(uint(repoID), collectionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, files)
		return
	}
	path = strings.ReplaceAll(path, ".", "")
	path = strings.Trim(path, "/")
	path = strings.Trim(path, "\\")

	files, err := userGitRepoCollectionService.ListFilesInPath(uint(repoID), collectionName, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, files)
}

// GetFileContent returns the content of a file within a collection
func GetFileContent(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collection_name")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
		return
	}

	content, contentType, err := userGitRepoCollectionService.GetFileContent(uint(repoID), collectionName, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set the content type and return the file content
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, content)
}

// UpdateFileContent updates the content of a file within a collection
func UpdateFileContent(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collection_name")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
		return
	}

	// Read the request body
	content, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	if err := userGitRepoCollectionService.UpdateFileContent(uint(repoID), collectionName, filePath, content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File updated successfully"})
}

// DeleteFile deletes a file or directory within a collection
func DeleteFile(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collection_name")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
		return
	}

	if err := userGitRepoCollectionService.DeleteFile(uint(repoID), collectionName, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// UploadFile uploads a file to a collection using JSON request
func UploadFile(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collection_name")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	var request models.FileUploadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert content from base64 if needed, or use as-is if it's plain text
	var content []byte
	// Check if content is base64 encoded (simple heuristic)
	if strings.HasPrefix(request.Content, "data:") && strings.Contains(request.Content, ";base64,") {
		// Extract the base64 part
		parts := strings.Split(request.Content, ";base64,")
		if len(parts) != 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 content format"})
			return
		}

		// Decode base64
		var err error
		content, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode base64 content: " + err.Error()})
			return
		}
	} else {
		// Use content as-is (plain text)
		content = []byte(request.Content)
	}

	if err := userGitRepoCollectionService.UpdateFileContent(uint(repoID), collectionName, request.Path, content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

// GetFileContentJSON returns the content of a file within a collection as JSON
func GetFileContentJSON(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collection_name")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
		return
	}

	content, contentType, err := userGitRepoCollectionService.GetFileContent(uint(repoID), collectionName, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get repository to find its local path
	var repo models.UserGitRepo
	if err := database.DB.First(&repo, repoID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Repository not found"})
		return
	}

	// Get collection to find its path
	collection, err := userGitRepoCollectionService.GetCollectionByName(uint(repoID), collectionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fullPath := filepath.Join(collection.Path, filePath)
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Prepare response
	response := models.FileResponse{
		Name:      filepath.Base(filePath),
		Path:      filePath,
		IsDir:     fileInfo.IsDir(),
		Size:      fileInfo.Size(),
		ModTime:   fileInfo.ModTime(),
		Extension: filepath.Ext(filePath),
	}

	// For binary files, encode as base64
	if !strings.HasPrefix(contentType, "text/") && contentType != "application/json" && contentType != "application/xml" {
		response.Content = "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(content)
	} else {
		response.Content = string(content)
	}

	c.JSON(http.StatusOK, response)
}
