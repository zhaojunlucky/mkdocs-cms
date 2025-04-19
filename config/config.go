package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	JWT         JWTConfig      `yaml:"jwt"`
	OAuth       OAuthConfig    `yaml:"oauth"`
	GitHub      GitHubConfig   `yaml:"github"`
	Security    SecurityConfig `yaml:"security"`
	WorkingDir  string         `yaml:"working_dir"`
	LogLevel    string         `yaml:"log_level"`
	FrontendURL string         `yaml:"frontend_url"`
	APIURL      string         `yaml:"api_url"`
	MinIOConfig MinIOConfig    `yaml:"minio"`
	RateLimit   RateLimit      `yaml:"rate_limit"`
}

type MinIOConfig struct {
	APIURL    string `yaml:"api_url"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

type RateLimit struct {
	MaxRequests       int `yaml:"max_requests"`
	RequestsPerSecond int `yaml:"per_second"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret          string `yaml:"secret"`
	ExpirationHours int    `yaml:"expiration_hours"`
}

// OAuthConfig represents general OAuth configuration
type OAuthConfig struct {
	RedirectURL string `yaml:"redirect_url"`
}

// GitHubConfig represents GitHub configuration
type GitHubConfig struct {
	OAuth GitHubOAuthConfig `yaml:"oauth"`
	App   GitHubAppConfig   `yaml:"app"`
}

// GitHubOAuthConfig represents GitHub OAuth configuration
type GitHubOAuthConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// GitHubAppConfig represents GitHub App configuration
type GitHubAppConfig struct {
	AppID          int64  `yaml:"app_id"`
	PrivateKeyPath string `yaml:"private_key_path"`
	WebhookSecret  string `yaml:"webhook_secret"`
	Name           string `yaml:"app_name"`
	Description    string `yaml:"description"`
	HomepageURL    string `yaml:"homepage_url"`
	WebhookURL     string `yaml:"webhook_url"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	CORS CORSConfig `yaml:"cors"`
	CSRF CSRFConfig `yaml:"csrf"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

// CSRFConfig represents CSRF configuration
type CSRFConfig struct {
	Enabled            bool `yaml:"enabled"`
	TokenExpiryMinutes int  `yaml:"token_expiry_minutes"`
}

// LoadConfig loads the configuration from the specified file
func LoadConfig(configPath string) (*Config, error) {
	// If configPath is not absolute, make it relative to the current working directory
	if !filepath.IsAbs(configPath) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		configPath = filepath.Join(cwd, configPath)
	}

	// Read the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the YAML configuration
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if they exist
	overrideWithEnvVars(&config)

	return &config, nil
}

// overrideWithEnvVars overrides configuration values with environment variables if they exist
func overrideWithEnvVars(config *Config) {
	// JWT
	if val := os.Getenv("JWT_SECRET"); val != "" {
		config.JWT.Secret = val
	}

	// OAuth
	if val := os.Getenv("OAUTH_REDIRECT_URL"); val != "" {
		config.OAuth.RedirectURL = val
	}

	// GitHub OAuth
	if val := os.Getenv("GITHUB_OAUTH_CLIENT_ID"); val != "" {
		config.GitHub.OAuth.ClientID = val
	}
	if val := os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"); val != "" {
		config.GitHub.OAuth.ClientSecret = val
	}

	// GitHub App
	if val := os.Getenv("GITHUB_APP_ID"); val != "" {
		// Try to parse the value as an integer
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			fmt.Printf("Warning: Failed to parse GITHUB_APP_ID as an integer: %v\n", err)
			panic(err)
		}
		config.GitHub.App.AppID = intVal
	}
	if val := os.Getenv("GITHUB_APP_PRIVATE_KEY_PATH"); val != "" {
		config.GitHub.App.PrivateKeyPath = val
	}
	if val := os.Getenv("GITHUB_WEBHOOK_SECRET"); val != "" {
		config.GitHub.App.WebhookSecret = val
	}
}
