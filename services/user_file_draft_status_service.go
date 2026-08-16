package services

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"gorm.io/gorm"
	"strings"
)

type UserFileDraftStatusService struct {
	BaseService
}

func (s *UserFileDraftStatusService) Init(ctx *core.APPContext) {
	s.InitService("userFileDraftStatusService", ctx, s)
}

func (s *UserFileDraftStatusService) GetDraftStatus(userId string, repoId uint, collectionName string) (map[string]bool, error) {
	m := make(map[string]bool)

	var status []models.UserFileDraftStatus
	err := database.DB.Where("user_id = ? AND repo_id = ? AND collection_name = ?", userId, repoId, collectionName).Find(&status).Error
	if err != nil {
		log.Errorf("Failed to get draft status: %v", err)
		return m, nil
	}
	for _, v := range status {
		m[v.FilePath] = v.Draft
	}
	return m, nil
}

func (s *UserFileDraftStatusService) SetDraftStatus(userId string, repoId uint, collectionName string, filePath string, draft bool) error {

	var status models.UserFileDraftStatus
	err := database.DB.Where("user_id = ? AND repo_id = ? AND collection_name = ? AND file_path = ?", userId, repoId, collectionName, filePath).First(&status).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if draft && errors.Is(err, gorm.ErrRecordNotFound) {
		status = models.UserFileDraftStatus{
			UserID:         userId,
			RepoID:         repoId,
			CollectionName: collectionName,
			FilePath:       filePath,
			Draft:          true,
		}
		if err := database.DB.Create(&status).Error; err != nil {
			return err
		}
	} else if !draft && err == nil {
		if err := database.DB.Delete(&status).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *UserFileDraftStatusService) RenameFile(userId string, repoId uint, collectionName string, oldPath string, newPath string) error {
	var status models.UserFileDraftStatus
	err := database.DB.Where("user_id = ? AND repo_id = ? AND collection_name = ? AND file_path = ?", userId, repoId, collectionName, oldPath).First(&status).Error
	if err == nil {
		status.FilePath = newPath
		if err := database.DB.Save(&status).Error; err != nil {
			log.Errorf("Failed to save draft status: %v", err)
			return err
		}
	} else {
		log.Errorf("Failed to get draft status: %v", err)
		return err
	}
	return nil

}

func (s *UserFileDraftStatusService) RenameFolder(userId string, repoId uint, collectionName string, oldPath string, newPath string) error {
	oldPrefix := strings.TrimRight(oldPath, "/") + "/"
	newPrefix := strings.TrimRight(newPath, "/") + "/"

	var statuses []models.UserFileDraftStatus
	err := database.DB.Where("user_id = ? AND repo_id = ? AND collection_name = ?", userId, repoId, collectionName).Find(&statuses).Error
	if err != nil {
		log.Errorf("Failed to get folder draft statuses: %v", err)
		return err
	}

	for _, status := range statuses {
		if status.FilePath == oldPath {
			status.FilePath = newPath
		} else if strings.HasPrefix(status.FilePath, oldPrefix) {
			status.FilePath = newPrefix + strings.TrimPrefix(status.FilePath, oldPrefix)
		} else {
			continue
		}
		if err := database.DB.Save(&status).Error; err != nil {
			log.Errorf("Failed to save folder draft status: %v", err)
			return err
		}
	}

	return nil
}

func (s *UserFileDraftStatusService) DeleteFolder(userId string, repoId uint, collectionName string, folderPath string) error {
	folderPrefix := strings.TrimRight(folderPath, "/") + "/"
	var statuses []models.UserFileDraftStatus
	err := database.DB.Where("user_id = ? AND repo_id = ? AND collection_name = ?", userId, repoId, collectionName).Find(&statuses).Error
	if err != nil {
		log.Errorf("Failed to get folder draft statuses: %v", err)
		return err
	}

	for _, status := range statuses {
		if status.FilePath != folderPath && !strings.HasPrefix(status.FilePath, folderPrefix) {
			continue
		}
		if err := database.DB.Delete(&status).Error; err != nil {
			log.Errorf("Failed to delete folder draft status: %v", err)
			return err
		}
	}

	return nil
}
