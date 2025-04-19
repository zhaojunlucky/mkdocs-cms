package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/services"
	"io"
	"net/http"
)

type StorageController struct {
	BaseController
	storageService *services.StorageService
}

func (s *StorageController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	s.ctx = ctx
	s.storageService = ctx.MustGetService("storageService").(*services.StorageService)
	collections := router.Group("/storage")
	{
		collections.POST("", s.AttachFile)
		collections.GET("/:fileName", s.GetAttachedFile)
	}
}

func (s *StorageController) AttachFile(c *gin.Context) {

	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")

	if err := reqParam.Handle(c); err != nil {
		core.HandleError(c, err)
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		core.ResponseErrStr(c, http.StatusBadRequest, "Unable to get form: %s", err.Error())
		return
	}

	files := form.File["file[]"]

	if len(files) == 0 {
		core.ResponseErrStr(c, http.StatusBadRequest, "No files uploaded")
		return
	}

	data, err := s.storageService.AttachFile(userId.String(), files)
	if err != nil {
		core.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, data)
}

func (s *StorageController) GetAttachedFile(c *gin.Context) {
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	fileName := reqParam.AddUrlParam("fileName", false, nil)

	if err := reqParam.Handle(c); err != nil {
		core.HandleError(c, err)
		return
	}

	eTag := c.GetHeader("If-None-Match")

	userStorageFile, object, err := s.storageService.GetAttachedFile(userId.String(), fileName.String(), eTag)
	if err != nil {
		log.Errorf("Failed to get attached file: %v", err)
		core.HandleError(c, err)
		return
	} else if object == nil {
		// 304
		c.Status(http.StatusNotModified)
		return
	}
	defer object.Close()
	c.Header("Content-Type", userStorageFile.ContentType)
	c.Header("Age", "86400")
	c.Header("Cache-Control", "public, max-age=31536000")
	c.Header("ETag", fmt.Sprintf("%d", userStorageFile.ID))
	_, err = io.Copy(c.Writer, object)
	if err != nil {
		core.ResponseErr(c, http.StatusInternalServerError, err)
		return
	}
}
