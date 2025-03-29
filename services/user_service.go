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
}

func (s *UserService) Init(ctx *core.APPContext) {
	s.InitService("userService", ctx, s)
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	var user models.User
	result := database.DB.First(&user, "id = ?", id)
	if result.Error != nil {
		log.Errorf("Failed to get user by ID: %s, %v", id, result.Error)
		return nil, result.Error
	}
	return &user, nil
}

// CreateOrUpdateUser creates a new user or updates an existing one based on email
func (s *UserService) CreateOrUpdateUser(user *models.User) (*models.User, error) {
	// Check if user exists
	var existingUser models.User
	result := database.DB.Where("email = ?", user.Email).First(&existingUser)

	if result.RowsAffected > 0 {
		// User exists, update fields
		existingUser.Username = user.Username
		existingUser.Name = user.Name
		existingUser.AvatarURL = user.AvatarURL
		existingUser.Provider = user.Provider
		existingUser.ProviderID = user.ProviderID

		// Update user
		result = database.DB.Save(&existingUser)
		if result.Error != nil {
			log.Errorf("Failed to update user: %v", result.Error)
			return nil, errors.New("failed to update user")
		}

		return &existingUser, nil
	} else {
		log.Infof("User not found %s", user.Email)
		return nil, core.NewHTTPError(http.StatusUnprocessableEntity, "register is temporary disabled")
	}

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

	// Create the user
	result = database.DB.Create(user)
	if result.Error != nil {
		log.Errorf("Failed to create user: %v", result.Error)
		return nil, errors.New("failed to create user")
	}

	return user, nil
}
