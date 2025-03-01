package services

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService handles business logic related to users
type UserService struct{}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{}
}

// GetAllUsers retrieves all users from the database
func (s *UserService) GetAllUsers() ([]models.UserResponse, error) {
	var users []models.User
	result := database.DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	// Convert to response objects
	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	return responses, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	var user models.User
	result := database.DB.First(&user, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := database.DB.First(&user, "email = ?", email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByProviderID retrieves a user by provider and provider ID
func (s *UserService) GetUserByProviderID(provider, providerID string) (*models.User, error) {
	var user models.User
	result := database.DB.First(&user, "provider = ? AND provider_id = ?", provider, providerID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
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
			return nil, errors.New("failed to update user")
		}

		return &existingUser, nil
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
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(req models.CreateUserRequest) (models.UserResponse, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.UserResponse{}, errors.New("failed to hash password")
	}

	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	result := database.DB.Create(&user)
	if result.Error != nil {
		return models.UserResponse{}, errors.New("failed to create user")
	}

	return user.ToResponse(), nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(id string, req models.UpdateUserRequest) (models.UserResponse, error) {
	var user models.User
	result := database.DB.First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.UserResponse{}, errors.New("user not found")
		}
		return models.UserResponse{}, result.Error
	}

	// Update fields if provided
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.UserResponse{}, errors.New("failed to hash password")
		}
		user.Password = string(hashedPassword)
	}
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}

	result = database.DB.Save(&user)
	if result.Error != nil {
		return models.UserResponse{}, errors.New("failed to update user")
	}

	return user.ToResponse(), nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(id string) error {
	result := database.DB.Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		return errors.New("failed to delete user")
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
