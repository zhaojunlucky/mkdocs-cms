package controllers

import (
	"encoding/base64"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

type UserGitRepoCollectionController struct {
	BaseController
	service                *services.UserGitRepoCollectionService
	userGitRepoLockService *services.UserGitRepoLockService
}

func (ctrl *UserGitRepoCollectionController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	ctrl.ctx = ctx
	ctrl.service = ctx.MustGetService("userGitRepoCollectionService").(*services.UserGitRepoCollectionService)
	ctrl.userGitRepoLockService = ctx.MustGetService("userGitRepoLockService").(*services.UserGitRepoLockService)
	collections := router.Group("/collections")
	{
		collections.GET("/repo/:repoId", ctrl.GetCollectionsByRepo)

		// Collection file routes
		collections.GET("/repo/:repoId/:collectionName/files", ctrl.GetCollectionFilesInPath)
		collections.POST("/repo/:repoId/:collectionName/files/folder", ctrl.CreateFolder)
		collections.GET("/repo/:repoId/:collectionName/files/content", ctrl.GetFileContent)
		collections.PUT("/repo/:repoId/:collectionName/files/content", ctrl.UpdateFileContent)
		collections.DELETE("/repo/:repoId/:collectionName/files", ctrl.DeleteFile)
		collections.POST("/repo/:repoId/:collectionName/files/upload", ctrl.UploadFile)
		collections.PUT("/repo/:repoId/:collectionName/files/rename", ctrl.RenameFile)
	}
}

// GetCollectionsByRepo returns all collections for a specific repository
func (ctrl *UserGitRepoCollectionController) GetCollectionsByRepo(c *gin.Context) {

	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("repoId", false, regexp.MustCompile("\\d+"))

	if err := reqParam.Handle(c); err != nil {
		core.HandleError(c, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(c, http.StatusBadRequest, "Invalid repository ID")
		return
	}

	// Verify repository ownership
	repo, err := ctrl.service.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(c, err)
		return
	}

	collections, err := ctrl.service.GetCollectionsByRepo(repo)
	if err != nil {
		log.Errorf("Failed to get collections: %v", err)
		core.ResponseErr(c, http.StatusInternalServerError, err)
		return
	}

	var response []models.UserGitRepoCollectionResponse
	for _, collection := range collections {
		response = append(response, collection.ToResponse(false))
	}

	core.ResponseOKArr(c, response)
}

// GetCollectionFilesInPath returns all files in a specific path within a collection
func (ctrl *UserGitRepoCollectionController) GetCollectionFilesInPath(c *gin.Context) {
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("repoId", false, regexp.MustCompile("\\d+"))
	collectionName := reqParam.AddUrlParam("collectionName", true, nil)
	pathParam := reqParam.AddQueryParam("path", true, nil)

	if err := reqParam.Handle(c); err != nil {
		core.HandleError(c, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(c, http.StatusBadRequest, "Invalid repository ID")
		return
	}

	// Verify repository ownership
	repo, err := ctrl.service.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(c, err)
		return
	}
	path := pathParam.String()
	if path == "" {
		// If no path is provided, return files at the root of the collection
		files, err := ctrl.service.ListFilesInCollection(repo, collectionName.String())
		if err != nil {
			log.Errorf("Failed to list files in collection: %v", err)
			core.HandleError(c, err)
			return
		}
		core.ResponseOKArr(c, files)
		return
	}
	path = strings.ReplaceAll(path, ".", "")
	path = strings.Trim(path, "/")
	path = strings.Trim(path, "\\")

	files, err := ctrl.service.ListFilesInPath(repo, collectionName.String(), path)
	if err != nil {
		log.Errorf("Failed to list files in collection: %v", err)
		core.ResponseErr(c, http.StatusInternalServerError, err)
		return
	}

	core.ResponseOKArr(c, files)
}

// GetFileContent returns the content of a file within a collection
func (ctrl *UserGitRepoCollectionController) GetFileContent(c *gin.Context) {
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("repoId", false, regexp.MustCompile("\\d+"))
	collectionName := reqParam.AddUrlParam("collectionName", false, nil)
	pathParam := reqParam.AddQueryParam("path", false, nil)

	if err := reqParam.Handle(c); err != nil {
		core.HandleError(c, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(c, http.StatusBadRequest, "Invalid repository ID")
		return
	}
	// Verify repository ownership
	repo, err := ctrl.service.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(c, err)
		return
	}

	filePath := pathParam.String()
	if filePath == "" {
		log.Errorf("File path is required")
		core.ResponseErrStr(c, http.StatusBadRequest, "File path is required")
		return
	}

	content, contentType, err := ctrl.service.GetFileContent(repo, collectionName.String(), filePath)
	if err != nil {
		log.Errorf("Failed to get file content: %v", err)
		core.ResponseErr(c, http.StatusInternalServerError, err)
		return
	}
	log.Infof("File %s content retrieved successfully", filePath)
	// Set the content type and return the file content
	c.Header("Content-paramType", contentType)
	c.Data(http.StatusOK, contentType, content)
}

type UpdateFileRequest struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// UpdateFileContent updates the content of a file within a collection
func (ctrl *UserGitRepoCollectionController) UpdateFileContent(c *gin.Context) {
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("repoId", false, regexp.MustCompile("\\d+"))
	collectionName := reqParam.AddUrlParam("collectionName", false, nil)
	var req UpdateFileRequest
	if err := reqParam.HandleWithBody(c, &req); err != nil {
		core.HandleError(c, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(c, http.StatusBadRequest, "Invalid repository ID")
		return
	}
	// Verify repository ownership
	repo, err := ctrl.service.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(c, err)
		return
	}

	lock := ctrl.userGitRepoLockService.Acquire(c.Param("repoId"))
	lock.Lock()

	defer lock.Unlock()

	hasFrontMatter := strings.EqualFold("true", c.GetHeader("X-File-Front-Matter"))
	var isDraft *bool = nil
	if hasFrontMatter {
		isDraft = new(bool)
		*isDraft = strings.EqualFold("true", c.GetHeader("X-File-Front-Matter-Draft"))
	}

	if err := ctrl.service.UpdateFileContent(repo, collectionName.String(), req.Path, []byte(req.Content), isDraft); err != nil {
		log.Errorf("Failed to update file content: %v", err)
		core.ResponseErr(c, http.StatusInternalServerError, err)
		return
	}

	log.Infof("File %s updated and changes committed successfully", req.Path)
	c.JSON(http.StatusOK, gin.H{"message": "File updated and changes committed successfully"})
}

// DeleteFile deletes a file or directory within a collection
func (ctrl *UserGitRepoCollectionController) DeleteFile(c *gin.Context) {
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("repoId", false, regexp.MustCompile("\\d+"))
	collectionName := reqParam.AddUrlParam("collectionName", true, nil)
	pathParam := reqParam.AddQueryParam("path", false, nil)

	if err := reqParam.Handle(c); err != nil {
		core.HandleError(c, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(c, http.StatusBadRequest, "Invalid repository ID")
		return
	}
	// Verify repository ownership
	repo, err := ctrl.service.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(c, err)
		return
	}
	filePath := pathParam.String()

	lock := ctrl.userGitRepoLockService.Acquire(c.Param("repoId"))
	lock.Lock()

	defer lock.Unlock()

	if err := ctrl.service.DeleteFile(repo, collectionName.String(), filePath); err != nil {
		log.Errorf("Failed to delete file: %v", err)
		core.ResponseErr(c, http.StatusInternalServerError, err)
		return
	}
	log.Infof("File %s deleted successfully", filePath)
	c.Status(http.StatusNoContent)
}

// UploadFile uploads a file to a collection using JSON request
func (ctrl *UserGitRepoCollectionController) UploadFile(c *gin.Context) {
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("repoId", false, regexp.MustCompile("\\d+"))
	collectionName := reqParam.AddUrlParam("collectionName", false, nil)
	var request models.FileUploadRequest

	if err := reqParam.HandleWithBody(c, &request); err != nil {
		core.HandleError(c, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(c, http.StatusBadRequest, "Invalid repository ID")
		return
	}
	// Verify repository ownership
	repo, err := ctrl.service.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(c, err)
		return
	}

	// Convert content from base64 if needed, or use as-is if it's plain text
	var content []byte
	// Check if content is base64 encoded (simple heuristic)
	if strings.HasPrefix(request.Content, "data:") && strings.Contains(request.Content, ";base64,") {
		// Extract the base64 part
		parts := strings.Split(request.Content, ";base64,")
		if len(parts) != 2 {
			log.Errorf("Invalid base64 content format")
			core.ResponseErrStr(c, http.StatusBadRequest, "Invalid base64 content format")
			return
		}

		// Decode base64
		var err error
		content, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			log.Errorf("Failed to decode base64 content: %v", err)
			core.ResponseErrStr(c, http.StatusBadRequest, "Failed to decode base64 content: "+err.Error())
			return
		}
	} else {
		// Use content as-is (plain text)
		content = []byte(request.Content)
	}

	lock := ctrl.userGitRepoLockService.Acquire(c.Param("repoId"))
	lock.Lock()

	defer lock.Unlock()
	hasFrontMatter := strings.EqualFold("true", c.GetHeader("X-File-Front-Matter"))
	var isDraft *bool = nil
	if hasFrontMatter {
		isDraft = new(bool)
		*isDraft = strings.EqualFold("true", c.GetHeader("X-File-Front-Matter-Draft"))
	}

	if err := ctrl.service.UpdateFileContent(repo, collectionName.String(), request.Path, content, isDraft); err != nil {
		log.Errorf("Failed to update file content: %v", err)
		core.ResponseErr(c, http.StatusInternalServerError, err)
		return
	}

	log.Infof("File %s uploaded successfully", request.Path)
	c.JSON(http.StatusCreated, gin.H{"message": "File uploaded successfully"})
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
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("repoId", false, regexp.MustCompile("\\d+"))
	collectionName := reqParam.AddUrlParam("collectionName", false, nil)
	// Parse request body
	var req RenameFileRequest

	if err := reqParam.HandleWithBody(c, &req); err != nil {
		core.HandleError(c, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(c, http.StatusBadRequest, "Invalid repository ID")
		return
	}
	// Verify repository ownership
	repo, err := ctrl.service.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(c, err)
		return
	}

	lock := ctrl.userGitRepoLockService.Acquire(c.Param("repoId"))
	lock.Lock()

	defer lock.Unlock()

	// Call service to rename file
	err = ctrl.service.RenameFile(repo, collectionName.String(), req.OldPath, req.NewPath)
	if err != nil {
		log.Errorf("Failed to rename file: %v", err)
		core.ResponseErr(c, http.StatusInternalServerError, err)
		return
	}

	log.Infof("File %s renamed to %s successfully", req.OldPath, req.NewPath)
	c.Status(http.StatusOK)
}

func (ctrl *UserGitRepoCollectionController) CreateFolder(c *gin.Context) {
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	repoIDParam := reqParam.AddUrlParam("repoId", false, regexp.MustCompile("\\d+"))
	collectionName := reqParam.AddUrlParam("collectionName", false, nil)
	// Parse request body
	var req CreateFolderRequest

	if err := reqParam.HandleWithBody(c, &req); err != nil {
		core.HandleError(c, err)
		return
	}

	repoID, err := repoIDParam.UInt64()
	if err != nil {
		log.Errorf("Failed to parse repository ID: %v", err)
		core.ResponseErrStr(c, http.StatusBadRequest, "Invalid repository ID")
		return
	}
	// Verify repository ownership
	repo, err := ctrl.service.VerifyRepoOwnership(userId.String(), uint(repoID))
	if err != nil {
		log.Errorf("Failed to verify repository ownership: %v", err)
		core.HandleError(c, err)
		return
	}

	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	if invalidChars.MatchString(req.Folder) {
		log.Errorf("Folder name %s contains invalid characters", req.Folder)
		core.ResponseErrStr(c, http.StatusBadRequest, "Folder name contains invalid characters")
		return
	}

	lock := ctrl.userGitRepoLockService.Acquire(c.Param("repoId"))
	lock.Lock()

	defer lock.Unlock()

	// Call service to create folder
	err = ctrl.service.CreateFolder(repo, collectionName.String(), req.Path, req.Folder)
	if err != nil {
		log.Errorf("Failed to create folder: %v", err)
		core.HandleError(c, err)
		return
	}
	log.Infof("Folder %s created successfully", req.Folder)
	c.JSON(http.StatusOK, gin.H{"message": "Folder created successfully"})
}
