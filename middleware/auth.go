package middleware

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware is a middleware that checks if the user is authenticated
type AuthMiddleware struct {
	jwtSecret []byte
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(ctx *core.APPContext) gin.HandlerFunc {

	authMiddleware := &AuthMiddleware{
		jwtSecret: []byte(ctx.Config.JWT.Secret),
	}
	return authMiddleware.RequireAuth
}

// RequireAuth is a middleware that checks if the user is authenticated
func (m *AuthMiddleware) RequireAuth(c *gin.Context) {

	if shouldSkipAuth(c.Request.URL.Path) {
		log.Infof("Skipping auth for path: %s", c.Request.URL.Path)
		c.Next()
		return
	}

	// Get the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Warnf("Authorization header is required")
		core.ResponseErrStr(c, http.StatusUnauthorized, "Authorization header is required")
		c.Abort()
		return
	}

	// Check if the header format is valid
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		log.Warnf("Invalid authorization header format")
		core.ResponseErrStr(c, http.StatusUnauthorized, "Invalid authorization header format")
		c.Abort()
		return
	}

	// Extract the token
	tokenString := parts[1]

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Errorf("unable to find signing method: %v", token.Header["alg"])
			return nil, jwt.ErrSignatureInvalid
		}

		return m.jwtSecret, nil
	})

	if err != nil {
		log.Errorf("invalid token: %v", err)
		core.ResponseErrStr(c, http.StatusUnauthorized, "invalid token: "+err.Error())
		c.Abort()
		return
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Set user ID in the context
		userId, ok := claims["sub"].(string)
		if !ok {
			log.Errorf("invalid token claims")
			core.ResponseErrStr(c, http.StatusUnauthorized, "invalid token claims")
			c.Abort()
			return
		}

		c.Set("userId", userId)
		c.Next()
	} else {
		core.ResponseErrStr(c, http.StatusUnauthorized, "invalid token")
		c.Abort()
		return
	}
}

func shouldSkipAuth(path string) bool {
	// List of paths that should bypass auth
	skipPaths := []string{
		"/api/auth/github",
		"/api/auth/github/callback",
		"/api/auth/google",
		"/api/auth/google/callback",
		"/api/github/webhook",
	}

	for _, skipPath := range skipPaths {
		if strings.HasSuffix(path, skipPath) {
			return true
		}
	}
	return false
}
