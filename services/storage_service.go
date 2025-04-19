package services

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/env"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type StorageService struct {
	BaseService
	userService          *UserService
	minioService         *MinIOService
	invalidFilenameChars *regexp.Regexp
}

func (s *StorageService) Init(ctx *core.APPContext) {
	s.InitService("storageService", ctx, s)
	s.userService = ctx.MustGetService("userService").(*UserService)
	s.minioService = ctx.MustGetService("minioService").(*MinIOService)
	s.invalidFilenameChars = regexp.MustCompile(`\W`)
}

func (s *StorageService) AttachFile(userId string, files []*multipart.FileHeader) (map[string]interface{}, error) {
	userStorge, err := s.userService.GetUserStorage(userId)
	if err != nil {
		var httpErr *core.HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			// create user storage
		} else {
			return nil, err
		}
	}
	if userStorge == nil {
		var bucketPrexix = "mkdocs-cms"
		if !env.IsProduction {
			bucketPrexix = "mkdocs-cms-test"
		}
		user, err := s.userService.GetUserByID(userId)
		if err != nil {
			return nil, err
		}

		bucketName := fmt.Sprintf("%s-%s", bucketPrexix, user.Username)

		userStorge = &models.UserStorage{
			UserID:     user.ID,
			BucketName: bucketName,
		}
		if userStorge, err = s.userService.CreateUserStorage(userStorge); err != nil {
			return nil, err
		}

	}

	found, err := s.minioService.BucketExists(userStorge.BucketName)
	if err != nil {
		return nil, core.NewHTTPErrorStr(http.StatusBadGateway, err.Error())
	}
	if !found {
		return nil, core.NewHTTPErrorStr(http.StatusUnprocessableEntity, fmt.Sprintf("bucket %s not found, please contact admin.", userStorge.BucketName))
	}
	uploadedFiles := map[string]string{}
	errorFiles := map[string]string{}

	currentTime := time.Now()
	path := currentTime.Format("2006-01")
	for _, file := range files {
		log.Infof("uploading file: %s", file.Filename)
		ext := filepath.Ext(file.Filename)
		sanitizeFileName := file.Filename[:strings.LastIndex(file.Filename, ".")]
		sanitizeFileName = s.invalidFilenameChars.ReplaceAllString(sanitizeFileName, "")
		fileName := fmt.Sprintf("%s-%s%s", sanitizeFileName, uuid.New().String(), ext)
		filePath := filepath.Join(path, fileName)

		uploadInfo, err := s.minioService.UploadFile(userStorge.BucketName, filePath, file)
		if err != nil {
			errorFiles[file.Filename] = err.Error()
			continue
		}
		log.Info(uploadInfo)

		noDirPath := strings.ReplaceAll(filePath, "/", "-")

		fileUrl := fmt.Sprintf("/api/v1/storage/%s", noDirPath)
		log.Info(fileUrl)

		uploadedFiles[file.Filename] = fileUrl

		userStorgeFile := &models.UserStorageFile{
			UserStorageID: userStorge.ID,
			FileName:      file.Filename,
			Path:          filePath,
			Url:           noDirPath,
			ContentType:   file.Header.Get("Content-Type"),
		}
		if _, err := s.userService.CreateUserStorageFile(userStorgeFile); err != nil {
			s.minioService.DeleteFile(userStorge.BucketName, filePath)
			return nil, err
		}

	}
	return map[string]interface{}{
		"uploadedFiles": uploadedFiles,
		"errorFiles":    errorFiles,
	}, nil
}

func (s *StorageService) GetAttachedFile(userId string, filePath string, eTag string) (*models.UserStorageFile, *minio.Object, error) {
	userStorageFile, err := s.userService.GetUserStorageFile(userId, filePath)
	if err != nil {
		return nil, nil, err
	}
	if eTag == fmt.Sprintf("%d", userStorageFile.ID) {
		return userStorageFile, nil, nil
	}
	reader, err := s.minioService.DownloadFile(userStorageFile.UserStorage.BucketName, userStorageFile.Path)
	return userStorageFile, reader, err
}
