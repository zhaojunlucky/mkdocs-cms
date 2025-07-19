package controllers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/env"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type AuthController struct {
	BaseController
	userService        *services.UserService
	githubOAuthConfig  *oauth2.Config
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
		auth.GET("/user", c.GetCurrentUser)
		auth.DELETE("/logout", c.Logout)
	}
}

// GithubLogin initiates GitHub OAuth flow
func (c *AuthController) GithubLogin(ctx *gin.Context) {
	// Generate a random state
	state, err := generateRandomState()
	if err != nil {
		log.Errorf("Failed to generate state: %v", err)
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
	errUrl := fmt.Sprintf("%s/error", c.ctx.Config.FrontendURL)
	frontendURL := fmt.Sprintf("%s/login", c.ctx.Config.FrontendURL)

	reqParam := core.NewRequestParam()
	oauth_state := reqParam.AddCookieParam("oauth_state", false, nil)
	state := reqParam.AddQueryParam("state", false, nil)
	code := reqParam.AddQueryParam("code", false, nil)

	if err := reqParam.Handle(ctx); err != nil {
		log.Errorf("error: %s", err)
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?error=%s&redirect=%s", errUrl, err.Error(), frontendURL))
	}

	// Verify state
	if oauth_state.String() != state.String() {
		log.Errorf("State mismatch")
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?error=%s&redirect=%s", errUrl, "Invalid state", frontendURL))

		return
	}
	token, err := c.githubOAuthConfig.Exchange(ctx, code.String())
	if err != nil {
		log.Errorf("Failed to exchange code for token: %v", err)
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?error=%s&redirect=%s", errUrl, "Failed to exchange code for token", frontendURL))

		return
	}

	// Get user info from GitHub
	client := c.githubOAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		log.Errorf("Failed to get user info: %v", err)
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?error=%s&redirect=%s", errUrl, "Failed to get user info", frontendURL))
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response body: %v", err)
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?error=%s&redirect=%s", errUrl, "Failed to read response body", frontendURL))
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
		log.Errorf("Failed to parse user info: %v", err)
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?error=%s&redirect=%s", errUrl, "Failed to parse user info", frontendURL))
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
		log.Errorf("Failed to save user: %v", err)
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?error=%s&redirect=%s", errUrl, "Failed to save user.", frontendURL))
		return
	}

	if !savedUser.IsActive {
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?error=%s&redirect=%s", errUrl, "User is not active, please send an email to admin@gundamz.de!", frontendURL))
		return
	}

	// Generate JWT token
	jwtToken, err := c.generateJWT(savedUser)
	if err != nil {
		log.Errorf("Failed to generate token: %v", err)
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?error=%s&redirect=%s", errUrl, "Failed to generate token", frontendURL))
		return
	}
	hash := sha256.Sum256([]byte(savedUser.ID))
	hashHex := hex.EncodeToString(hash[:])
	// Redirect to frontend with token
	redirectURL := fmt.Sprintf("%s?token=%s", frontendURL, url.QueryEscape(hashHex))

	ctx.SetCookie("session_id", jwtToken, 3600*c.jwtExpirationHours, "/", c.ctx.CookieDomain, env.IsProduction, true)
	ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GetCurrentUser returns the current authenticated user
func (c *AuthController) GetCurrentUser(ctx *gin.Context) {
	reqParam := core.NewRequestParam()
	userId := reqParam.AddContextParam("userId", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	userExpiresAt := reqParam.AddContextParam("userExpiresAt", false, nil).
		SetError(http.StatusUnauthorized, "Unauthorized")
	if err := reqParam.Handle(ctx); err != nil {
		core.HandleError(ctx, err)
		return
	}

	user, err := c.userService.GetUserByID(userId.String())
	if err != nil {
		log.Errorf("Failed to retrieve user: %v", err)
		core.ResponseErr(ctx, http.StatusInternalServerError, err)
		return
	}

	expireInt, err := userExpiresAt.Int64()
	if err != nil {
		log.Errorf("Failed to convert expiresAt: %v", err)
		core.ResponseErr(ctx, http.StatusUnprocessableEntity, err)
		return
	}

	ctx.JSON(http.StatusOK, user.ToResponseWithExpires(expireInt))
}

// Helper function to generate a random state
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Errorf("Failed to generate random state: %v", err)
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
		log.Errorf("Failed to sign token: %v", err)
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
		log.Errorf("Failed to read GitHub emails: %v", err)
		return nil, err
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.Unmarshal(body, &emails); err != nil {
		log.Errorf("Failed to unmarshal GitHub emails: %v", err)
		return nil, err
	}

	return emails, nil
}

func (c *AuthController) Logout(context *gin.Context) {
	context.SetCookie("session_id", "", -1, "/", c.ctx.CookieDomain, env.IsProduction, true)
	context.Status(http.StatusNoContent)
}
