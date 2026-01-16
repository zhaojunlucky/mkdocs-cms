package services

import (
	"context"
	"mime/multipart"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
)

type MinIOService struct {
	BaseService
	MinIO *minio.Client
}

func (s *MinIOService) Init(ctx *core.APPContext) {
	s.InitService("minioService", ctx, s)
	var err error
	s.MinIO, err = minio.New(ctx.Config.MinIOConfig.APIURL, &minio.Options{
		Creds:      credentials.NewStaticV4(ctx.Config.MinIOConfig.AccessKey, ctx.Config.MinIOConfig.SecretKey, ""),
		Secure:     strings.HasPrefix(ctx.Config.FrontendURL, "https"),
		MaxRetries: 3,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (s *MinIOService) BucketExists(bucketName string) (bool, error) {
	return s.MinIO.BucketExists(context.Background(), bucketName)
}

func (s *MinIOService) UploadFile(name string, objectName string, file *multipart.FileHeader) (minio.UploadInfo, error) {
	src, err := file.Open()
	if err != nil {
		return minio.UploadInfo{}, err
	}
	defer src.Close()
	return s.MinIO.PutObject(context.Background(), name, objectName, src, file.Size,
		minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})
}

func (s *MinIOService) DeleteFile(name string, path string) error {
	return s.MinIO.RemoveObject(context.Background(), name, path, minio.RemoveObjectOptions{ForceDelete: true})
}

func (s *MinIOService) DownloadFile(name string, path string) (*minio.Object, error) {
	return s.MinIO.GetObject(context.Background(), name, path, minio.GetObjectOptions{})
}
