package services

import (
	"errors"

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
func (s *UserService) GetUserByID(id uint) (models.UserResponse, error) {
	var user models.User
	result := database.DB.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.UserResponse{}, errors.New("user not found")
		}
		return models.UserResponse{}, result.Error
	}

	return user.ToResponse(), nil
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
func (s *UserService) UpdateUser(id uint, req models.UpdateUserRequest) (models.UserResponse, error) {
	var user models.User
	result := database.DB.First(&user, id)
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
func (s *UserService) DeleteUser(id uint) error {
	result := database.DB.Delete(&models.User{}, id)
	if result.Error != nil {
		return errors.New("failed to delete user")
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
