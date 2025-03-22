package controllers

import (
	"encoding/base64"
	"fmt"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

var userGitRepoCollectionService *services.UserGitRepoCollectionService

// InitUserGitRepoCollectionController initializes the controller with GitHub app settings
func InitUserGitRepoCollectionController(settings *models.GitHubAppSettings) {
	userGitRepoCollectionService = services.NewUserGitRepoCollectionService(services.NewUserGitRepoService(settings))
}

type UserGitRepoCollectionController struct {
	service *services.UserGitRepoCollectionService
}

func NewUserGitRepoCollectionController() *UserGitRepoCollectionController {
	return &UserGitRepoCollectionController{
		service: userGitRepoCollectionService,
	}
}

// GetCollections returns all collections
func (ctrl *UserGitRepoCollectionController) GetCollections(c *gin.Context) {
	collections, err := ctrl.service.GetAllCollections()
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
func (ctrl *UserGitRepoCollectionController) GetCollectionsByRepo(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repoId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collections, err := ctrl.service.GetCollectionsByRepo(uint(repoID))
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
func (ctrl *UserGitRepoCollectionController) GetCollection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("collectionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection, err := ctrl.service.GetCollectionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	c.JSON(http.StatusOK, collection.ToResponse(true))
}

// CreateCollection creates a new collection
// This method is kept for backward compatibility but returns an error message
// as collections are now read-only from veda/config.yml
func (ctrl *UserGitRepoCollectionController) CreateCollection(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Collections are now defined in veda/config.yml file in the repository. Please edit that file directly.",
	})
}

// UpdateCollection updates an existing collection
// This method is kept for backward compatibility but returns an error message
// as collections are now read-only from veda/config.yml
func (ctrl *UserGitRepoCollectionController) UpdateCollection(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Collections are now defined in veda/config.yml file in the repository. Please edit that file directly.",
	})
}

// DeleteCollection deletes a collection
// This method is kept for backward compatibility but returns an error message
// as collections are now read-only from veda/config.yml
func (ctrl *UserGitRepoCollectionController) DeleteCollection(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Collections are now defined in veda/config.yml file in the repository. Please edit that file directly.",
	})
}

// GetCollectionByPath returns a collection by its path
func (ctrl *UserGitRepoCollectionController) GetCollectionByPath(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repoId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path parameter is required"})
		return
	}

	collection, err := ctrl.service.GetCollectionByPath(uint(repoID), path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	c.JSON(http.StatusOK, collection.ToResponse(true))
}

// GetCollectionFiles returns all files in a collection
func (ctrl *UserGitRepoCollectionController) GetCollectionFiles(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repoId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collectionName")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	files, err := ctrl.service.ListFilesInCollection(uint(repoID), collectionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, files)
}

// GetCollectionFilesInPath returns all files in a specific path within a collection
func (ctrl *UserGitRepoCollectionController) GetCollectionFilesInPath(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repoId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collectionName")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	path := c.Query("path")
	if path == "" {
		// If no path is provided, return files at the root of the collection
		files, err := ctrl.service.ListFilesInCollection(uint(repoID), collectionName)
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

	files, err := ctrl.service.ListFilesInPath(uint(repoID), collectionName, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, files)
}

// GetFileContent returns the content of a file within a collection
func (ctrl *UserGitRepoCollectionController) GetFileContent(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repoId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collectionName")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
		return
	}

	content, contentType, err := ctrl.service.GetFileContent(uint(repoID), collectionName, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set the content type and return the file content
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, content)
}

type UpdateFileRequest struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// UpdateFileContent updates the content of a file within a collection
func (ctrl *UserGitRepoCollectionController) UpdateFileContent(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repoId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collectionName")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	// Get file path and content from request body
	var req UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.service.UpdateFileContent(uint(repoID), collectionName, req.Path, []byte(req.Content)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get repository info for GitHub commit
	repo, err := ctrl.service.GetRepo(uint(repoID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get repository: %v", err)})
		return
	}

	// Commit changes using GitHub app
	commitMsg := fmt.Sprintf("Update file: %s", req.Path)
	if err := ctrl.service.CommitWithGithubApp(repo, commitMsg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to commit changes: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File updated and changes committed successfully"})
}

// DeleteFile deletes a file or directory within a collection
func (ctrl *UserGitRepoCollectionController) DeleteFile(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repoId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collectionName")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
		return
	}

	if err := ctrl.service.DeleteFile(uint(repoID), collectionName, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// UploadFile uploads a file to a collection using JSON request
func (ctrl *UserGitRepoCollectionController) UploadFile(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repoId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collectionName")
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

	if err := ctrl.service.UpdateFileContent(uint(repoID), collectionName, request.Path, content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

// GetFileContentJSON returns the content of a file within a collection as JSON
func (ctrl *UserGitRepoCollectionController) GetFileContentJSON(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repoId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	collectionName := c.Param("collectionName")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
		return
	}

	content, contentType, err := ctrl.service.GetFileContent(uint(repoID), collectionName, filePath)
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
	collection, err := ctrl.service.GetCollectionByName(uint(repoID), collectionName)
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

// RenameFileRequest represents the request body for renaming a file
type RenameFileRequest struct {
	OldPath string `json:"oldPath" binding:"required"`
	NewPath string `json:"newPath" binding:"required"`
}

type CreateFolderRequest struct {
	Path   string `json:"path"`
	Folder string `json:"folder" binding:"required"`
}

// RenameFile renames a file in a collection
func (ctrl *UserGitRepoCollectionController) RenameFile(c *gin.Context) {
	// Get repository ID and collection name from path
	repoID, err := strconv.ParseUint(c.Param("repoId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repository ID"})
		return
	}
	collectionName := c.Param("collectionName")

	// Parse request body
	var req RenameFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify repository ownership
	if !ctrl.VerifyRepoOwnership(c, uint(repoID)) {
		return
	}

	// Call service to rename file
	err = ctrl.service.RenameFile(uint(repoID), collectionName, req.OldPath, req.NewPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (ctrl *UserGitRepoCollectionController) VerifyRepoOwnership(c *gin.Context, repoID uint) bool {
	repo, err := ctrl.service.GetRepo(repoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get repository: %v", err)})
		return false
	}
	if repo.UserID != c.MustGet("userId") {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to rename files in this repository"})
		return false
	}
	return true
}

func (ctrl *UserGitRepoCollectionController) CreateFolder(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repoId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repository ID"})
		return
	}
	collectionName := c.Param("collectionName")

	// Parse request body
	var req CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify repository ownership
	if !ctrl.VerifyRepoOwnership(c, uint(repoID)) {
		return
	}

	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	if invalidChars.MatchString(req.Folder) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Folder name contains invalid characters"})
		return
	}

	// Call service to create folder
	err = ctrl.service.CreateFolder(uint(repoID), collectionName, req.Path, req.Folder)
	if err != nil {
		core.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder created successfully"})
}
