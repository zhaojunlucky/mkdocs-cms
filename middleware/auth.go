package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/database"
	"github.com/zhaojunlucky/mkdocs-cms/models"

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
	authHeader, err := c.Cookie("session_id")
	if err != nil {
		log.Warnf("cookie session_id is required")
		core.ResponseErrStr(c, http.StatusUnauthorized, "cookie session_id is required")
		c.Abort()
		return
	}
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		log.Warnf("cookie session_id is required")
		core.ResponseErrStr(c, http.StatusUnauthorized, "cookie session_id is required")
		c.Abort()
		return
	}

	// Parse the token
	token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
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
		var user models.User
		err := database.DB.Preload("Roles").Where("id = ?", userId).First(&user).Error
		if err != nil {
			log.Errorf("user not found: %v", err)
			core.ResponseErrStr(c, http.StatusUnauthorized, "user not found: "+err.Error())
			c.Abort()
			return
		}
		if !user.IsActive {
			log.Errorf("user is not active")
			core.ResponseErrStr(c, http.StatusUnauthorized, "user is not active")
			c.Abort()
			return
		}
		expires := claims["exp"].(float64)
		c.Set("userId", userId)
		c.Set("userExpiresAt", fmt.Sprintf("%d", int64(expires)))
		c.Next()
	} else {
		core.ResponseErrStr(c, http.StatusUnauthorized, "invalid token")
		c.Abort()
		return
	}
}

func shouldSkipAuth(path string) bool {
	// List of paths that should bypass auth
	skipPaths := []*regexp.Regexp{
		regexp.MustCompile("^/api/auth/github"),
		regexp.MustCompile("^/api/auth/github/callback"),
		regexp.MustCompile("^/api/auth/logout"),
		regexp.MustCompile("^/api/github/webhook"),
		regexp.MustCompile("^/api/site/version"),
		regexp.MustCompile("^/api/v1/storage/.+"),
		regexp.MustCompile("^/metrics"),
	}

	for _, skipPath := range skipPaths {
		if skipPath.MatchString(path) {
			return true
		}
	}
	return false
}
