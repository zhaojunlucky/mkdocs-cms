package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"net/http"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles business logic related to users
type UserService struct {
	BaseService
	siteService *SiteService
}

func (s *UserService) Init(ctx *core.APPContext) {
	s.InitService("userService", ctx, s)
	s.siteService = ctx.MustGetService("siteService").(*SiteService)
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	var user models.User
	result := database.DB.Preload("Roles").First(&user, "id = ?", id)
	if result.Error != nil {
		log.Errorf("Failed to get user by ID: %s, %v", id, result.Error)
		return nil, core.NewGormHTTPError(result.Error)
	}
	return &user, nil
}

// CreateOrUpdateUser creates a new user or updates an existing one based on email
func (s *UserService) CreateOrUpdateUser(user *models.User) (*models.User, error) {
	// Check if user exists
	var existingUser models.User
	result := database.DB.Where("provider = ? AND provider_id = ? AND username = ?", user.Provider, user.ProviderID, user.Username).First(&existingUser)

	if result.RowsAffected > 0 {
		if existingUser.Username != user.Username ||
			existingUser.Name != user.Name ||
			existingUser.AvatarURL != user.AvatarURL ||
			existingUser.Email != user.Email {
			// User exists, update fields
			existingUser.Username = user.Username
			existingUser.Name = user.Name
			existingUser.AvatarURL = user.AvatarURL
			existingUser.Email = user.Email

			// Update user
			result = database.DB.Save(&existingUser)
			if result.Error != nil {
				log.Errorf("Failed to update user: %v", result.Error)
				return nil, errors.New("failed to update user")
			}
		}

		return &existingUser, nil
	} else {
		log.Infof("User not found %s", user.Email)
		if !s.siteService.AllowUserRegistration() {
			return nil, core.NewHTTPErrorStr(http.StatusUnprocessableEntity, "register is disabled")
		}
	}

	user.IsActive = false
	// User doesn't exist, create new one
	// Set a default password for OAuth users (this will be a random string that can't be used to log in)
	if user.Password == "" {
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err == nil {
			hashedPassword, err := bcrypt.GenerateFromPassword(randomBytes, bcrypt.DefaultCost)
			if err == nil {
				user.Password = string(hashedPassword)
			}
		}
	}

	// Generate a unique ID for the user if not provided
	if user.ID == "" {
		randomBytes := make([]byte, 16)
		if _, err := rand.Read(randomBytes); err == nil {
			user.ID = fmt.Sprintf("%x", randomBytes)
		}
	}
	var role models.Role
	err := database.DB.First(&role, "name = 'user'")
	if err.Error != nil {
		log.Errorf("Failed to get user role: %v", err.Error)
		return nil, errors.New("failed to get user role")
	}
	// Assign the user to the "user" role
	user.Roles = append(user.Roles, &role)
	// Create the user
	result = database.DB.Create(user)
	if result.Error != nil {
		log.Errorf("Failed to create user: %v", result.Error)
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

func (s *UserService) GetUserStorage(userId string) (*models.UserStorage, error) {
	var userStorage models.UserStorage
	result := database.DB.Preload("User").First(&userStorage, "user_id = ?", userId)
	if result.Error != nil {

		log.Errorf("Failed to get user storage by ID: %s, %v", userId, result.Error)
		return nil, core.NewGormHTTPError(result.Error)
	}
	return &userStorage, nil
}

func (s *UserService) CreateUserStorage(storge *models.UserStorage) (*models.UserStorage, error) {
	result := database.DB.Create(storge)
	if result.Error != nil {
		log.Errorf("Failed to create user storage: %v", result.Error)
		return nil, core.NewGormHTTPError(result.Error)
	}
	return storge, nil
}

func (s *UserService) CreateUserStorageFile(file *models.UserStorageFile) (*models.UserStorageFile, error) {
	result := database.DB.Create(file)
	if result.Error != nil {
		log.Errorf("Failed to create user storage file: %v", result.Error)
		return nil, core.NewGormHTTPError(result.Error)
	}
	return file, nil
}

func (s *UserService) GetUserStorageFile(id string, path string) (*models.UserStorageFile, error) {
	var userStorage models.UserStorage
	var userStorageFile models.UserStorageFile
	result := database.DB.Preload("User").First(&userStorage, "user_id = ?", id)
	if result.Error != nil {
		log.Errorf("Failed to get user storage by ID: %s, %v", id, result.Error)
		return nil, core.NewGormHTTPError(result.Error)
	}
	result = database.DB.Preload("UserStorage").First(&userStorageFile, "user_storage_id = ? and url = ?", userStorage.ID, path)
	if result.Error != nil {
		log.Errorf("Failed to get user storage file by path: %s, %v", path, result.Error)
		return nil, core.NewGormHTTPError(result.Error)
	}
	return &userStorageFile, nil
}
