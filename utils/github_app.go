package utils

import (
	"fmt"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v45/github"
)

func CreateGitHubAppClient(ctx *core.APPContext) *github.Client {
	// Try to load private key if path is provided
	var privateKey []byte
	if ctx.GithubAppSettings.PrivateKeyPath != "" {
		var err error
		privateKey, err = os.ReadFile(ctx.GithubAppSettings.PrivateKeyPath)
		if err != nil {
			log.Fatalf("Warning: Failed to read GitHub App private key: %v\n", err)
		}
	}

	itr := &jwtTransport{
		settings:   ctx.GithubAppSettings,
		privateKey: privateKey,
	}

	// Create a GitHub App client with the JWT transport
	githubAppClient := github.NewClient(&http.Client{Transport: itr})

	return githubAppClient
}

// jwtTransport is an http.RoundTripper that adds a JWT to requests
type jwtTransport struct {
	settings   *models.GitHubAppSettings
	privateKey []byte
	base       http.RoundTripper
}

func (t *jwtTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Generate a new JWT for each request
	token, err := t.generateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %v", err)
	}

	// Clone the request to modify it
	req2 := req.Clone(req.Context())
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Accept", "application/vnd.github.v3+json")

	// Use the base transport or default if none set
	transport := t.base
	if transport == nil {
		transport = http.DefaultTransport
	}

	return transport.RoundTrip(req2)
}

func (t *jwtTransport) generateJWT() (string, error) {
	// Create the claims for the JWT
	now := time.Now()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    strconv.FormatInt(t.settings.AppID, 10),
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Parse the private key
	key, err := jwt.ParseRSAPrivateKeyFromPEM(t.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	// Sign the token
	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return signedToken, nil
}
