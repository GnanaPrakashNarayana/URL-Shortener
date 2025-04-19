package models

import (
	"time"
)

// User represents a registered user
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	OAuthAccounts []*OAuthAccount `json:"oauth_accounts,omitempty"`
}

// OAuthAccount represents an OAuth provider account linked to a user
type OAuthAccount struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	Provider       string    `json:"provider"`
	ProviderUserID string    `json:"provider_user_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// RoleAdmin is the admin role
const RoleAdmin = "admin"

// RoleUser is the regular user role
const RoleUser = "user"

// UserResponse represents a user response without sensitive data
type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// NewUser creates a new user
func NewUser(username, email, passwordHash string) *User {
	now := time.Now()
	return &User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         RoleUser, // Default role
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ToResponse converts a user to a user response
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

// HasRole checks if the user has the given role
func (u *User) HasRole(role string) bool {
	if role == RoleUser {
		// All users have the user role
		return true
	}
	return u.Role == role
}

// IsAdmin checks if the user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}