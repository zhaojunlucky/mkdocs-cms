package utils

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// GenerateGitHubAppJWT generates a JWT for GitHub App authentication
func GenerateGitHubAppJWT(appID int64, privateKey []byte) (string, error) {
	// Create the claims for the JWT
	now := time.Now()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    strconv.FormatInt(appID, 10),
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Parse the private key
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
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

// GetGitHubInstallationToken gets an installation token for a GitHub App
func GetGitHubInstallationToken(appID int64, privateKey []byte, installationID int64, opts *github.InstallationTokenOptions) (string, error) {
	// Generate JWT
	token, err := GenerateGitHubAppJWT(appID, privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %v", err)
	}

	// Create GitHub client with JWT
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
		},
	}
	client := github.NewClient(httpClient)

	// Get installation token
	installationToken, _, err := client.Apps.CreateInstallationToken(context.Background(), installationID, opts)
	if err != nil {
		return "", fmt.Errorf("failed to get installation token: %v", err)
	}

	return installationToken.GetToken(), nil
}
