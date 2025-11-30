package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v45/github"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/core/md"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"gopkg.in/yaml.v3"
)

// UserGitRepoCollectionService handles business logic for git repository collections
type UserGitRepoCollectionService struct {
	BaseService
	userGitRepoService         *UserGitRepoService
	userFileDraftStatusService *UserFileDraftStatusService
	mdHandler                  *md.MDHandler
}

func (s *UserGitRepoCollectionService) Init(ctx *core.APPContext) {
	s.InitService("userGitRepoCollectionService", ctx, s)
	s.userGitRepoService = ctx.MustGetService("userGitRepoService").(*UserGitRepoService)
	s.userFileDraftStatusService = ctx.MustGetService("userFileDraftStatusService").(*UserFileDraftStatusService)
	s.mdHandler = md.NewMDHandler()
}

// VedaConfig represents the structure of veda/config.yml
type VedaConfig struct {
	Collections []Collection `yaml:"collections"`
	MDConfig    *md.MDConfig `yaml:"md_config"`
}

// Collection represents a collection in veda/config.yml
type Collection struct {
	Name              string             `yaml:"name"`
	Label             string             `yaml:"label"`
	Path              string             `yaml:"path"`
	Format            string             `yaml:"format"`
	FileNameGenerator *FileNameGenerator `yaml:"file_name_generator"`
	Fields            []Field            `yaml:"fields,omitempty"`
}

type FileNameGenerator struct {
	Type  string `yaml:"type" json:"type"`
	First string `yaml:"first" json:"first"`
}

// Field represents a field in a collection
type Field struct {
	Type     string `yaml:"type" json:"type"`
	Name     string `yaml:"name" json:"name"`
	Label    string `yaml:"label" json:"label"`
	Required bool   `yaml:"required,omitempty" json:"required"`
	Format   string `yaml:"format,omitempty" json:"format"`
	List     bool   `yaml:"list,omitempty" json:"list"`
	Default  string `yaml:"default,omitempty" json:"default"`
}

// GetAllCollections returns all collections
func (s *UserGitRepoCollectionService) GetAllCollections() ([]models.UserGitRepoCollection, error) {
	// This method is not applicable anymore as collections are read from veda/config.yml
	return nil, errors.New("collections are now stored in veda/config.yml, use GetCollectionsByRepoID instead")
}

// GetCollectionsByRepoID returns all collections for a specific repository
func (s *UserGitRepoCollectionService) GetCollectionsByRepoID(repoID uint) ([]models.UserGitRepoCollection, error) {
	// Get the repository to find its local path
	var repo models.UserGitRepo
	if err := database.DB.First(&repo, repoID).Error; err != nil {
		return nil, fmt.Errorf("repository %d not found", repoID)
	}

	// Read collections from veda/config.yml
	collections, err := s.readCollectionsFromConfig(repo)
	if err != nil {
		log.Errorf("Failed to read collections from veda/config.yml: %v", err)
		return nil, err
	}

	return collections, nil
}

// GetCollectionsByRepo returns all collections for a specific repository
func (s *UserGitRepoCollectionService) GetCollectionsByRepo(repo *models.UserGitRepo) ([]models.UserGitRepoCollection, error) {
	// Get the repository to find its local path

	// Read collections from veda/config.yml
	collections, err := s.readCollectionsFromConfig(*repo)
	if err != nil {
		log.Errorf("Failed to read collections from veda/config.yml: %v", err)
		return nil, core.NewHTTPErrorStr(http.StatusInternalServerError, err.Error())
	}

	return collections, nil
}

func (s *UserGitRepoCollectionService) readRepoConfig(repo models.UserGitRepo) (*VedaConfig, error) {
	configPath := filepath.Join(repo.LocalPath, "veda", "config.yml")

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Errorf("veda/config.yml not found in repository %s", repo.Name)
		return nil, fmt.Errorf("veda/config.yml not found in repository %s", repo.Name)
	}

	// Read the config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		log.Errorf("Failed to read veda/config.yml: %v", err)
		return nil, fmt.Errorf("failed to read veda/config.yml: %v", err)
	}

	// Parse the YAML
	var config *VedaConfig = &VedaConfig{}
	if err := yaml.Unmarshal(configData, config); err != nil {
		log.Errorf("Invalid YAML format in veda/config.yml: %v", err)
		return nil, fmt.Errorf("invalid YAML format in veda/config.yml: %v", err)
	}
	return config, nil
}

// readCollectionsFromConfig reads collections from veda/config.yml
func (s *UserGitRepoCollectionService) readCollectionsFromConfig(repo models.UserGitRepo) ([]models.UserGitRepoCollection, error) {
	var collections []models.UserGitRepoCollection

	config, err := s.readRepoConfig(repo)
	if err != nil {
		return collections, nil
	}
	// Convert to UserGitRepoCollection models
	for _, col := range config.Collections {
		// Resolve the path relative to the repository
		fullPath := filepath.Join(repo.LocalPath, col.Path)

		var modelFields []models.Field
		for _, f := range col.Fields {
			modelFields = append(modelFields, models.Field{
				Type:     f.Type,
				Name:     f.Name,
				Label:    f.Label,
				Required: f.Required,
				Format:   f.Format,
				List:     f.List,
				Default:  f.Default,
			})
		}

		collection := models.UserGitRepoCollection{
			Name:        col.Name,
			Label:       col.Label,
			Path:        fullPath,
			Format:      models.ContentFormat(col.Format),
			Description: "", // No description in veda/config.yml
			RepoID:      repo.ID,
			Fields:      modelFields,
		}

		if col.FileNameGenerator != nil {
			collection.FileNameGenerator = &models.FileNameGenerator{
				Type:  col.FileNameGenerator.Type,
				First: col.FileNameGenerator.First,
			}
		}
		collections = append(collections, collection)
	}

	return collections, nil
}

// GetCollectionByID returns a specific collection by ID
func (s *UserGitRepoCollectionService) GetCollectionByID(id uint) (models.UserGitRepoCollection, error) {
	// Since collections are now read from veda/config.yml, we need to find the repository first
	var collection models.UserGitRepoCollection
	if err := database.DB.Preload("Repo").First(&collection, id).Error; err != nil {
		// If we can't find it in the database, try to find it by name in the repository
		var repo models.UserGitRepo
		if err := database.DB.First(&repo, "id = ?", id).Error; err != nil {
			return models.UserGitRepoCollection{}, fmt.Errorf("repository %d not found", id)
		}

		// Read collections from veda/config.yml
		collections, err := s.readCollectionsFromConfig(repo)
		if err != nil {
			return models.UserGitRepoCollection{}, err
		}

		// Find the collection by ID (which is now just an index)
		if int(id) < len(collections) {
			return collections[id], nil
		}

		return models.UserGitRepoCollection{}, errors.New("collection not found")
	}

	return collection, nil
}

// GetCollectionByName returns a collection by its name within a repository
func (s *UserGitRepoCollectionService) GetCollectionByName(repo *models.UserGitRepo, name string) (models.UserGitRepoCollection, error) {
	collections, err := s.GetCollectionsByRepo(repo)
	if err != nil {
		return models.UserGitRepoCollection{}, err
	}

	for _, collection := range collections {
		if collection.Name == name {
			return collection, nil
		}
	}

	return models.UserGitRepoCollection{}, core.NewHTTPErrorStr(http.StatusNotFound, "collection not found")
}

// GetCollectionByPath returns a collection by its path within a repository
func (s *UserGitRepoCollectionService) GetCollectionByPath(repoID uint, path string) (models.UserGitRepoCollection, error) {
	collections, err := s.GetCollectionsByRepoID(repoID)
	if err != nil {
		return models.UserGitRepoCollection{}, err
	}

	for _, collection := range collections {
		if strings.HasSuffix(collection.Path, path) {
			return collection, nil
		}
	}

	return models.UserGitRepoCollection{}, errors.New("collection not found")
}

// FileInfo represents information about a file or directory
type FileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"is_dir"`
	IsDraft   bool      `json:"is_draft"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"mod_time"`
	Extension string    `json:"extension,omitempty"`
}

// ListFilesInCollection lists all files under a collection path
func (s *UserGitRepoCollectionService) ListFilesInCollection(repo *models.UserGitRepo, collectionName string) ([]FileInfo, error) {
	// Get the collection
	collection, err := s.GetCollectionByName(repo, collectionName)
	if err != nil {
		return nil, err
	}

	// Check if the path exists
	if _, err := os.Stat(collection.Path); os.IsNotExist(err) {
		return nil, core.NewHTTPErrorStr(http.StatusBadRequest, "collection path does not exist")
	}

	// Read the directory
	entries, err := os.ReadDir(collection.Path)
	if err != nil {
		return nil, err
	}

	statusMap, err := s.userFileDraftStatusService.GetDraftStatus(repo.UserID, repo.ID, collectionName)
	if err != nil {
		log.Errorf("failed to get draft status: %v", err)
	}

	// Convert to FileInfo
	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if !entry.IsDir() && filepath.Ext(entry.Name()) != fmt.Sprintf(".%s", collection.Format) {
			continue
		}

		fileInfo := FileInfo{
			Name:    entry.Name(),
			Path:    entry.Name(),
			IsDraft: statusMap[entry.Name()],
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		// Add extension for files
		if !entry.IsDir() {
			fileInfo.Extension = filepath.Ext(entry.Name())
		}

		files = append(files, fileInfo)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name > files[j].Name
	})

	return files, nil
}

// ListFilesInPath lists all files under a specific path within a collection
func (s *UserGitRepoCollectionService) ListFilesInPath(repo *models.UserGitRepo, collectionName string, subPath string) ([]FileInfo, error) {
	// Get the collection
	collection, err := s.GetCollectionByName(repo, collectionName)
	if err != nil {
		return nil, err
	}

	// Ensure the subPath doesn't try to escape the collection directory
	cleanSubPath := filepath.Clean(subPath)
	if cleanSubPath == ".." || filepath.IsAbs(cleanSubPath) || strings.HasPrefix(cleanSubPath, "../") {
		return nil, errors.New("invalid path")
	}

	// Construct the full path
	fullPath := filepath.Join(collection.Path, cleanSubPath)

	// Check if the path exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, errors.New("path does not exist")
	}

	// Read the directory
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	statusMap, err := s.userFileDraftStatusService.GetDraftStatus(repo.UserID, repo.ID, collectionName)
	if err != nil {
		log.Errorf("failed to get draft status: %v", err)
	}

	// Convert to FileInfo
	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if !entry.IsDir() && filepath.Ext(entry.Name()) != fmt.Sprintf(".%s", collection.Format) {
			continue
		}

		relativePath := filepath.Join(cleanSubPath, entry.Name())
		fileInfo := FileInfo{
			Name:    entry.Name(),
			Path:    relativePath,
			IsDir:   entry.IsDir(),
			IsDraft: statusMap[relativePath],
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		// Add extension for files
		if !entry.IsDir() {
			fileInfo.Extension = filepath.Ext(entry.Name())
		}

		files = append(files, fileInfo)
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir && !files[j].IsDir {
			return true // Directories come before files
		}
		if !files[i].IsDir && files[j].IsDir {
			return false // Files come after directories
		}
		// If both are directories or both are files, sort by name
		return files[i].Name > files[j].Name
	})

	return files, nil
}

// GetFileContent retrieves the content of a file within a collection
func (s *UserGitRepoCollectionService) GetFileContent(repo *models.UserGitRepo, collectionName string, filePath string) ([]byte, string, error) {
	// Get the collection
	collection, err := s.GetCollectionByName(repo, collectionName)
	if err != nil {
		return nil, "", err
	}

	// Ensure the filePath doesn't try to escape the collection directory
	cleanFilePath := filepath.Clean(filePath)
	if cleanFilePath == ".." || filepath.IsAbs(cleanFilePath) || strings.HasPrefix(cleanFilePath, "../") {
		return nil, "", errors.New("invalid path")
	}

	// Construct the full path
	fullPath := filepath.Join(collection.Path, cleanFilePath)

	// Check if the path exists and is a file
	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, "", errors.New("file does not exist")
	}
	if err != nil {
		return nil, "", err
	}
	if fileInfo.IsDir() {
		return nil, "", errors.New("path is a directory, not a file")
	}

	// Read the file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, "", err
	}

	// Determine content type based on file extension
	contentType := "text/plain"
	ext := filepath.Ext(fullPath)
	switch strings.ToLower(ext) {
	case ".html", ".htm":
		contentType = "text/html"
	case ".css":
		contentType = "text/css"
	case ".js":
		contentType = "application/javascript"
	case ".json":
		contentType = "application/json"
	case ".xml":
		contentType = "application/xml"
	case ".md":
		contentType = "text/markdown"
		content = s.handleMarkdown(repo, content, md.DirectionRead)

	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	case ".pdf":
		contentType = "application/pdf"
	}

	return content, contentType, nil
}

func (s *UserGitRepoCollectionService) handleMarkdown(repo *models.UserGitRepo, content []byte, direction string) []byte {

	config, err := s.readRepoConfig(*repo)
	if err != nil {
		log.Errorf("unable to read repo config, skip md handler: %v", err)
		return content
	}
	return s.mdHandler.Handle(config.MDConfig, content, direction)
}

// UpdateFileContent updates the content of a file within a collection
func (s *UserGitRepoCollectionService) UpdateFileContent(repo *models.UserGitRepo, collectionName string, filePath string, content []byte, isDraft *bool) error {
	// Get the collection
	collection, err := s.GetCollectionByName(repo, collectionName)
	if err != nil {
		return err
	}

	// Ensure the filePath doesn't try to escape the collection directory
	cleanFilePath := filepath.Clean(filePath)
	if cleanFilePath == ".." || filepath.IsAbs(cleanFilePath) || strings.HasPrefix(cleanFilePath, "../") {
		return errors.New("invalid path")
	}

	ext := filepath.Ext(filePath)
	if ext == ".md" {
		content = s.handleMarkdown(repo, content, md.DirectionWrite)
	}

	// Construct the full path
	fullPath := filepath.Join(collection.Path, cleanFilePath)

	// Check if the file exists and validate
	fileInfo, err := os.Stat(fullPath)
	isNewFile := err != nil && os.IsNotExist(err)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If the file exists, check if it's a directory
	if err == nil && fileInfo.IsDir() {
		return errors.New("cannot update a directory")
	}

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}

	// Write the content to the file
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return err
	}

	// Commit message based on whether it's a new file or update
	commitMsg := fmt.Sprintf("Update file %s in collection %s", cleanFilePath, collectionName)
	if isNewFile {
		commitMsg = fmt.Sprintf("Create new file %s in collection %s", cleanFilePath, collectionName)
	}

	// Commit the changes
	if err := s.CommitWithGithubApp(*repo, commitMsg); err != nil {
		return fmt.Errorf("failed to commit changes: %v", err)
	}

	if isDraft != nil {
		_ = s.userFileDraftStatusService.SetDraftStatus(repo.UserID, repo.ID, collectionName, cleanFilePath, *isDraft)
	}

	return nil
}

// DeleteFile deletes a file or directory within a collection
func (s *UserGitRepoCollectionService) DeleteFile(repo *models.UserGitRepo, collectionName string, filePath string) error {
	// Get the collection
	collection, err := s.GetCollectionByName(repo, collectionName)
	if err != nil {
		return err
	}

	// Ensure the filePath doesn't try to escape the collection directory
	cleanFilePath := filepath.Clean(filePath)
	if cleanFilePath == ".." || filepath.IsAbs(cleanFilePath) || strings.HasPrefix(cleanFilePath, "../") {
		return errors.New("invalid path")
	}

	// Construct the full path
	fullPath := filepath.Join(collection.Path, cleanFilePath)

	// Check if the path exists
	_, err = os.Stat(fullPath)
	if os.IsNotExist(err) {
		return errors.New("file or directory does not exist")
	}
	if err != nil {
		return err
	}

	// Delete the file or directory
	if err := os.RemoveAll(fullPath); err != nil {
		return err
	}

	// Commit the changes
	if err := s.CommitWithGithubApp(*repo, fmt.Sprintf("Delete %s from collection %s", filePath, collectionName)); err != nil {
		return fmt.Errorf("failed to commit changes: %v", err)
	}

	_ = s.userFileDraftStatusService.SetDraftStatus(repo.UserID, repo.ID, collectionName, cleanFilePath, false)

	return nil
}

func (s *UserGitRepoCollectionService) GetRepo(repoID uint) (models.UserGitRepo, error) {
	var repo models.UserGitRepo
	if err := database.DB.Preload("User").First(&repo, repoID).Error; err != nil {
		return models.UserGitRepo{}, errors.New("repository not found")
	}
	return repo, nil
}

// CommitWithGithubApp commits changes using GitHub app authentication
func (s *UserGitRepoCollectionService) CommitWithGithubApp(repo models.UserGitRepo, message string) error {
	// Get installation token
	opts := &github.InstallationTokenOptions{
		RepositoryIDs: []int64{repo.GitRepoID},
		Permissions: &github.InstallationPermissions{
			Contents: github.String("write"),
			Metadata: github.String("read"),
		},
	}

	// Get an installation token
	token, _, err := s.ctx.GithubAppClient.Apps.CreateInstallationToken(context.Background(), repo.InstallationID, opts)

	if err != nil {
		log.Errorf("Failed to get installation token: %v", err)
		return fmt.Errorf("failed to get installation token: %v", err)
	}

	// Set up git config with token
	remoteURL := fmt.Sprintf("https://x-access-token:%s@%s", token.GetToken(), repo.RemoteURL[8:])
	cmd := exec.Command("git", "-C", repo.LocalPath, "remote", "set-url", "origin", remoteURL)
	if output, err := cmd.CombinedOutput(); err != nil {

		log.Errorf("Failed to configure git with token: %s", string(output))
		return fmt.Errorf("failed to configure git with token: %v", err)
	}

	// Check if there are any changes
	statusCmd := exec.Command("git", "-C", repo.LocalPath, "status", "--porcelain")
	output, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %v", err)
	}

	needPush := len(output) > 0
	// If no changes, return early
	if needPush {
		// Add all changes
		cmd = exec.Command("git", "-C", repo.LocalPath, "add", ".")
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Errorf("Failed to stage changes: %s", string(output))
			return fmt.Errorf("failed to stage changes: %v", err)
		}

		// Commit changes
		cmd = exec.Command("git", "-C", repo.LocalPath, "commit", "-m", message)
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Errorf("Failed to commit changes: %s", string(output))
			return fmt.Errorf("failed to commit changes: %v", err)
		}
	}

	// Push changes
	cmd = exec.Command("git", "-C", repo.LocalPath, "push", "origin", repo.Branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Errorf("Failed to push changes: %s", string(output))
		return fmt.Errorf("failed to push changes: %v", err)
	}

	return nil
}

// RenameFile renames a file in a collection
func (s *UserGitRepoCollectionService) RenameFile(repo *models.UserGitRepo, collectionName string, oldPath string, newPath string) error {
	// Get collection info
	collection, err := s.GetCollectionByName(repo, collectionName)
	if err != nil {
		return err
	}

	// Clean and validate paths
	cleanOldPath := filepath.Clean(oldPath)
	cleanNewPath := filepath.Clean(newPath)

	// Ensure paths are within collection directory
	//if !strings.HasPrefix(cleanOldPath, collectionName+"/") || !strings.HasPrefix(cleanNewPath, collectionName+"/") {
	//	return errors.New("invalid file path")
	//}

	// Construct full paths
	oldFullPath := filepath.Join(collection.Path, cleanOldPath)
	newFullPath := filepath.Join(collection.Path, cleanNewPath)

	// Check if source file exists and is not a directory
	fileInfo, err := os.Stat(oldFullPath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return errors.New("cannot rename a directory")
	}

	// Check if target path doesn't exist
	if _, err := os.Stat(newFullPath); err == nil {
		return errors.New("destination file already exists")
	}

	// Create parent directories for the new path if they don't exist
	parentDir := filepath.Dir(newFullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}

	// Rename the file
	if err := os.Rename(oldFullPath, newFullPath); err != nil {
		return err
	}

	// Commit the changes
	commitMsg := fmt.Sprintf("Rename file from %s to %s in collection %s", cleanOldPath, cleanNewPath, collectionName)
	if err := s.CommitWithGithubApp(*repo, commitMsg); err != nil {
		return fmt.Errorf("failed to commit changes: %v", err)
	}

	_ = s.userFileDraftStatusService.RenameFile(repo.UserID, repo.ID, collectionName, cleanOldPath, cleanNewPath)

	return nil
}

func (s *UserGitRepoCollectionService) CreateFolder(repo *models.UserGitRepo, name string, path string, folder string) error {
	// Get the collection
	collection, err := s.GetCollectionByName(repo, name)
	if err != nil {
		return err
	}

	// Ensure the path doesn't try to escape the collection directory
	cleanPath := filepath.Clean(path)
	if cleanPath == ".." || filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, "../") {
		return core.NewHTTPErrorStr(http.StatusBadRequest, "invalid path")
	}

	// Ensure the folder name is valid
	cleanFolder := filepath.Clean(folder)
	if cleanFolder == ".." || filepath.IsAbs(cleanFolder) || strings.HasPrefix(cleanFolder, "../") || strings.Contains(cleanFolder, "/") {
		return core.NewHTTPErrorStr(http.StatusBadRequest, "invalid folder name")
	}

	// Construct the full path
	fullPath := filepath.Join(collection.Path, cleanPath, cleanFolder)

	// Check if the folder already exists
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		if err == nil {
			return core.NewHTTPErrorStr(http.StatusBadRequest, "folder already exists")
		}
		return err
	}

	// Create the folder
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return fmt.Errorf("failed to create folder: %v", err)
	}

	fi, _ := os.Create(filepath.Join(fullPath, ".gitkeep"))
	if fi != nil {
		_ = fi.Close()
	}

	// Commit the changes
	commitMsg := fmt.Sprintf("Create folder %s in collection %s", filepath.Join(cleanPath, cleanFolder), name)
	if err := s.CommitWithGithubApp(*repo, commitMsg); err != nil {
		log.Errorf("Failed to commit changes: %v", err)
		return fmt.Errorf("failed to commit changes: %v", err)
	}

	return nil
}

func (s *UserGitRepoCollectionService) VerifyRepoOwnership(userID string, repoID uint) (*models.UserGitRepo, error) {
	repo, err := s.GetRepo(repoID)
	if err != nil {
		log.Errorf("Failed to get repository: %v", err)
		return nil, core.NewHTTPErrorStr(http.StatusInternalServerError, err.Error())
	}
	if repo.UserID != userID {
		log.Errorf("You do not have permission to rename files in this repository")
		return nil, core.NewHTTPErrorStr(http.StatusForbidden, "You do not have permission to rename files in this repository")
	}
	return &repo, nil
}
