package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type AuthController struct {
	BaseController
	userService        *services.UserService
	githubOAuthConfig  *oauth2.Config
	googleOAuthConfig  *oauth2.Config
	jwtSecret          []byte
	jwtExpirationHours int
}

func (c *AuthController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	c.ctx = ctx
	c.userService = ctx.MustGetService("userService").(*services.UserService)
	c.initConfig()
	c.registerRoutes(router)
}

func (c *AuthController) initConfig() {
	// Initialize GitHub OAuth config
	githubConfig := &oauth2.Config{
		ClientID:     c.ctx.Config.GitHub.OAuth.ClientID,
		ClientSecret: c.ctx.Config.GitHub.OAuth.ClientSecret,
		RedirectURL:  c.ctx.Config.OAuth.RedirectURL,
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
	}
	c.githubOAuthConfig = githubConfig

	// Initialize Google OAuth config
	googleConfig := &oauth2.Config{
		ClientID:     c.ctx.Config.Google.OAuth.ClientID,
		ClientSecret: c.ctx.Config.Google.OAuth.ClientSecret,
		RedirectURL:  c.ctx.Config.OAuth.RedirectURL,
		Scopes:       c.ctx.Config.Google.OAuth.Scopes,
		Endpoint:     google.Endpoint,
	}
	c.googleOAuthConfig = googleConfig

	// Get JWT secret from configuration
	jwtSecret := []byte(c.ctx.Config.JWT.Secret)
	c.jwtSecret = jwtSecret
	c.jwtExpirationHours = c.ctx.Config.JWT.ExpirationHours
}

// RegisterRoutes registers the auth routes
func (c *AuthController) registerRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.GET("/github", c.GithubLogin)
		auth.GET("/github/callback", c.GithubCallback)
		auth.GET("/google", c.GoogleLogin)
		auth.GET("/google/callback", c.GoogleCallback)
		auth.GET("/user", c.AuthMiddleware(), c.GetCurrentUser)
	}
}

// GithubLogin initiates GitHub OAuth flow
func (c *AuthController) GithubLogin(ctx *gin.Context) {
	// Generate a random state
	state, err := generateRandomState()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	// Store state in cookie
	ctx.SetCookie("oauth_state", state, 3600, "/", "", false, true)

	// Redirect to GitHub OAuth page
	url := c.githubOAuthConfig.AuthCodeURL(state)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

// GithubCallback handles GitHub OAuth callback
func (c *AuthController) GithubCallback(ctx *gin.Context) {
	// Get state from cookie
	state, err := ctx.Cookie("oauth_state")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// Verify state
	if state != ctx.Query("state") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// Exchange code for token
	code := ctx.Query("code")
	token, err := c.githubOAuthConfig.Exchange(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	// Get user info from GitHub
	client := c.githubOAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	// Parse user info
	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &githubUser); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	// If email is not provided, get it from the emails API
	if githubUser.Email == "" {
		emails, err := c.getGithubEmails(client)
		if err == nil && len(emails) > 0 {
			for _, email := range emails {
				if email.Primary && email.Verified {
					githubUser.Email = email.Email
					break
				}
			}
			// If no primary email, use the first verified one
			if githubUser.Email == "" {
				for _, email := range emails {
					if email.Verified {
						githubUser.Email = email.Email
						break
					}
				}
			}
		}
	}

	// Create or update user in database
	user := &models.User{
		Username:   githubUser.Login,
		Name:       githubUser.Name,
		Email:      githubUser.Email,
		AvatarURL:  githubUser.AvatarURL,
		Provider:   "github",
		ProviderID: fmt.Sprintf("%d", githubUser.ID),
	}

	// Save user to database
	savedUser, err := c.userService.CreateOrUpdateUser(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
		return
	}

	// Generate JWT token
	jwtToken, err := c.generateJWT(savedUser)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Redirect to frontend with token
	frontendURL := fmt.Sprintf("%s/login", c.ctx.Config.FrontendURL)
	redirectURL := fmt.Sprintf("%s?token=%s", frontendURL, url.QueryEscape(jwtToken))
	ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GoogleLogin initiates Google OAuth flow
func (c *AuthController) GoogleLogin(ctx *gin.Context) {
	// Generate a random state
	state, err := generateRandomState()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	// Store state in cookie
	ctx.SetCookie("oauth_state", state, 3600, "/", "", false, true)

	// Redirect to Google OAuth page
	url := c.googleOAuthConfig.AuthCodeURL(state)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles Google OAuth callback
func (c *AuthController) GoogleCallback(ctx *gin.Context) {
	// Get state from cookie
	state, err := ctx.Cookie("oauth_state")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// Verify state
	if state != ctx.Query("state") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// Exchange code for token
	code := ctx.Query("code")
	token, err := c.googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	// Get user info from Google
	client := c.googleOAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	// Parse user info
	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	// Create or update user in database
	user := &models.User{
		Username:   googleUser.Email,
		Name:       googleUser.Name,
		Email:      googleUser.Email,
		AvatarURL:  googleUser.Picture,
		Provider:   "google",
		ProviderID: googleUser.ID,
	}

	// Save user to database
	savedUser, err := c.userService.CreateOrUpdateUser(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
		return
	}

	// Generate JWT token
	jwtToken, err := c.generateJWT(savedUser)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Redirect to frontend with token
	frontendURL := fmt.Sprintf("%s/login", c.ctx.Config.FrontendURL)
	redirectURL := fmt.Sprintf("%s?token=%s", frontendURL, url.QueryEscape(jwtToken))
	ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GetCurrentUser returns the current authenticated user
func (c *AuthController) GetCurrentUser(ctx *gin.Context) {
	// Get user from context (set by AuthMiddleware)
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// AuthMiddleware is a middleware to authenticate requests
func (c *AuthController) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get token from header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// Check if the header has the Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(c.jwtSecret), nil
		})

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check token expiration
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
					return
				}
			}

			// Get user ID from claims
			userIDStr, ok := claims["sub"].(string)
			if !ok {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				return
			}

			// Get user from database
			user, err := c.userService.GetUserByID(userIDStr)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				return
			}

			// Set user in context
			ctx.Set("user", user)
			ctx.Next()
		} else {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
	}
}

// Helper function to generate a random state
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Helper function to generate JWT token
func (c *AuthController) generateJWT(user *models.User) (string, error) {
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID
	claims["name"] = user.Name
	claims["email"] = user.Email
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(c.jwtExpirationHours)).Unix() // Token expires in jwtExpirationHours hours

	// Sign token
	tokenString, err := token.SignedString(c.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Helper function to get GitHub emails
func (c *AuthController) getGithubEmails(client *http.Client) ([]struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.Unmarshal(body, &emails); err != nil {
		return nil, err
	}

	return emails, nil
}
