package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zhaojunlucky/mkdocs-cms/config"
)

// AuthMiddleware is a middleware that checks if the user is authenticated
type AuthMiddleware struct {
	jwtSecret []byte
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware() gin.HandlerFunc {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/oauth_config.yaml"
	}

	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration for auth middleware: %v", err)
	}

	authMiddleware := &AuthMiddleware{
		jwtSecret: []byte(appConfig.JWT.Secret),
	}
	return authMiddleware.RequireAuth
}

// RequireAuth is a middleware that checks if the user is authenticated
func (m *AuthMiddleware) RequireAuth(c *gin.Context) {
	// Get the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
		c.Abort()
		return
	}

	// Check if the header format is valid
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		c.Abort()
		return
	}

	// Extract the token
	tokenString := parts[1]

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return m.jwtSecret, nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
		c.Abort()
		return
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Set user ID in the context
		userId, ok := claims["sub"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		c.Set("userId", userId)
		c.Next()
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}
}

// RequireAuth is a standalone middleware function for routes that need authentication
func RequireAuth() gin.HandlerFunc {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/oauth_config.yaml"
	}

	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration for auth middleware: %v", err)
	}

	jwtSecret := []byte(appConfig.JWT.Secret)

	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header format is valid
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := parts[1]

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Check if the token is valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Set user ID in the context
			userId, ok := claims["userId"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
				c.Abort()
				return
			}

			c.Set("userId", userId)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
	}
}
