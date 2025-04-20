package services

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/config"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token expired")
	ErrInvalidOAuthState  = errors.New("invalid OAuth state")
)

// Provider type
type Provider string

// OAuth providers
const (
	ProviderGoogle Provider = "google"
	ProviderGitHub Provider = "github"
)

// AuthService handles authentication
type AuthService struct {
	userRepo       repository.UserRepository
	config         *config.AuthConfig
	oauthProviders map[Provider]*oauth2.Config
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepository, config *config.AuthConfig) *AuthService {
	// Create OAuth providers
	oauthProviders := make(map[Provider]*oauth2.Config)

	// Google OAuth
	if config.OAuth.GoogleClientID != "" && config.OAuth.GoogleClientSecret != "" {
		oauthProviders[ProviderGoogle] = &oauth2.Config{
			ClientID:     config.OAuth.GoogleClientID,
			ClientSecret: config.OAuth.GoogleClientSecret,
			RedirectURL:  config.OAuth.GoogleRedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		}
	}

	// GitHub OAuth
	if config.OAuth.GitHubClientID != "" && config.OAuth.GitHubClientSecret != "" {
		oauthProviders[ProviderGitHub] = &oauth2.Config{
			ClientID:     config.OAuth.GitHubClientID,
			ClientSecret: config.OAuth.GitHubClientSecret,
			RedirectURL:  config.OAuth.GitHubRedirectURL,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	}

	return &AuthService{
		userRepo:       userRepo,
		config:         config,
		oauthProviders: oauthProviders,
	}
}

// RegisterUser registers a new user
func (s *AuthService) RegisterUser(ctx context.Context, username, email, password string) (*models.User, error) {
	// Check if username or email already exists
	_, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil {
		return nil, ErrUserExists
	} else if err != repository.ErrUserNotFound {
		return nil, err
	}

	_, err = s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return nil, ErrUserExists
	} else if err != repository.ErrUserNotFound {
		return nil, err
	}

	// Hash the password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create the user
	user := models.NewUser(username, email, string(passwordHash))
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// LoginUser logs in a user
func (s *AuthService) LoginUser(ctx context.Context, usernameOrEmail, password string) (*models.User, error) {
	// Try finding the user by username or email
	var user *models.User
	var err error

	user, err = s.userRepo.GetByUsername(ctx, usernameOrEmail)
	if err != nil {
		if err != repository.ErrUserNotFound {
			return nil, err
		}
		// Try email
		user, err = s.userRepo.GetByEmail(ctx, usernameOrEmail)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
	}

	// Verify the password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GenerateToken generates a JWT token for a user
func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	// Set expiration time
	expirationTime := time.Now().Add(time.Duration(s.config.JWTExpirationMinutes) * time.Minute)

	// Create the JWT claims
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      expirationTime.Unix(),
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*models.User, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Validate the token
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Get the claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Check if the token has expired
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > int64(exp) {
		return nil, ErrExpiredToken
	}

	// Get the user ID
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Create a user from the claims
	user := &models.User{
		ID:       int(userID),
		Username: claims["username"].(string),
		Email:    claims["email"].(string),
		Role:     claims["role"].(string),
	}

	return user, nil
}

// GetOAuthURL returns the URL for OAuth authentication
func (s *AuthService) GetOAuthURL(provider Provider, state string) (string, error) {
	// Get the provider
	oauthConfig, ok := s.oauthProviders[provider]
	if !ok {
		return "", fmt.Errorf("unknown provider: %s", provider)
	}

	// Get the URL
	url := oauthConfig.AuthCodeURL(state)
	return url, nil
}

// HandleOAuthCallback handles the OAuth callback
func (s *AuthService) HandleOAuthCallback(ctx context.Context, provider Provider, code, state, expectedState string) (*models.User, error) {
	// Validate state
	if state != expectedState {
		return nil, ErrInvalidOAuthState
	}

	// Get the provider
	oauthConfig, ok := s.oauthProviders[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	// Exchange the code for a token
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	// Get the user info
	var userInfo map[string]interface{}
	var providerUserID string
	var email string
	var name string

	// Provider-specific user info fetching
	client := oauthConfig.Client(ctx, token)

	switch provider {
	case ProviderGoogle:
		// Get user info from Google
		resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Parse the response
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			return nil, err
		}

		// Extract user info
		providerUserID = userInfo["sub"].(string)
		email = userInfo["email"].(string)
		name = userInfo["name"].(string)

	case ProviderGitHub:
		// Get user info from GitHub
		resp, err := client.Get("https://api.github.com/user")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Parse the response
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			return nil, err
		}

		// Extract user info
		providerUserID = fmt.Sprintf("%v", userInfo["id"])
		name = userInfo["login"].(string)

		// GitHub doesn't return email in the user endpoint, get it from the emails endpoint
		resp, err = client.Get("https://api.github.com/user/emails")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var emails []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
			return nil, err
		}

		// Find the primary email
		for _, emailObj := range emails {
			if emailObj["primary"].(bool) {
				email = emailObj["email"].(string)
				break
			}
		}

		// If no primary email found, use the first one
		if email == "" && len(emails) > 0 {
			email = emails[0]["email"].(string)
		}
	}

	// Check if the user already exists by OAuth account
	user, err := s.userRepo.GetUserByOAuthAccount(ctx, string(provider), providerUserID)
	if err != nil && err != repository.ErrUserNotFound {
		return nil, err
	}

	// If the user exists, return it
	if err == nil {
		return user, nil
	}

	// Check if the user exists by email
	user, err = s.userRepo.GetByEmail(ctx, email)
	if err != nil && err != repository.ErrUserNotFound {
		return nil, err
	}

	// If the user exists, link the OAuth account
	if err == nil {
		// Create the OAuth account
		account := &models.OAuthAccount{
			UserID:         user.ID,
			Provider:       string(provider),
			ProviderUserID: providerUserID,
		}
		if err := s.userRepo.CreateOAuthAccount(ctx, account); err != nil {
			return nil, err
		}
		return user, nil
	}

	// Create a new user
	// Generate a random password
	password := make([]byte, 32)
	if _, err := rand.Read(password); err != nil {
		return nil, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create the user
	user = models.NewUser(name, email, string(passwordHash))
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create the OAuth account
	account := &models.OAuthAccount{
		UserID:         user.ID,
		Provider:       string(provider),
		ProviderUserID: providerUserID,
	}
	if err := s.userRepo.CreateOAuthAccount(ctx, account); err != nil {
		return nil, err
	}

	return user, nil
}

// HasProvider checks if a specific OAuth provider is configured
func (s *AuthService) HasProvider(provider Provider) bool {
	_, ok := s.oauthProviders[provider]
	return ok
}