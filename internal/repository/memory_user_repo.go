package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
)

// ErrUserNotFound is returned when a user is not found
var ErrUserNotFound = errors.New("user not found")

// ErrUserConflict is returned when a user with the same username or email already exists
var ErrUserConflict = errors.New("user with the same username or email already exists")

// MemoryUserRepository is an in-memory repository
type MemoryUserRepository struct {
	users         map[int]*models.User
	oauthAccounts map[string]map[string]*models.OAuthAccount // provider -> providerUserID -> account
	mutex         sync.RWMutex
	nextID        int
}

// NewMemoryUserRepository creates a new in-memory user repository
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users:         make(map[int]*models.User),
		oauthAccounts: make(map[string]map[string]*models.OAuthAccount),
		nextID:        1,
	}
}

// Create creates a new user
func (r *MemoryUserRepository) Create(ctx context.Context, user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if username or email already exists
	for _, existingUser := range r.users {
		if existingUser.Username == user.Username || existingUser.Email == user.Email {
			return ErrUserConflict
		}
	}

	// Assign an ID
	user.ID = r.nextID
	r.nextID++

	// Set creation and update times
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Store the user
	r.users[user.ID] = user

	return nil
}

// GetByID retrieves a user by ID
func (r *MemoryUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *MemoryUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, ErrUserNotFound
}

// GetByEmail retrieves a user by email
func (r *MemoryUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, ErrUserNotFound
}

// Update updates a user
func (r *MemoryUserRepository) Update(ctx context.Context, user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.users[user.ID]; !ok {
		return ErrUserNotFound
	}

	// Check for username/email conflicts
	for id, existingUser := range r.users {
		if id == user.ID {
			continue
		}
		if existingUser.Username == user.Username || existingUser.Email == user.Email {
			return ErrUserConflict
		}
	}

	// Update the user
	user.UpdatedAt = time.Now()
	r.users[user.ID] = user

	return nil
}

// Delete deletes a user
func (r *MemoryUserRepository) Delete(ctx context.Context, id int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.users[id]; !ok {
		return ErrUserNotFound
	}

	delete(r.users, id)
	return nil
}

// List lists all users
func (r *MemoryUserRepository) List(ctx context.Context) ([]*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}

// CreateOAuthAccount creates a new OAuth account
func (r *MemoryUserRepository) CreateOAuthAccount(ctx context.Context, account *models.OAuthAccount) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if user exists
	if _, ok := r.users[account.UserID]; !ok {
		return ErrUserNotFound
	}

	// Create provider map if it doesn't exist
	if _, ok := r.oauthAccounts[account.Provider]; !ok {
		r.oauthAccounts[account.Provider] = make(map[string]*models.OAuthAccount)
	}

	// Check if account already exists
	if _, ok := r.oauthAccounts[account.Provider][account.ProviderUserID]; ok {
		return ErrUserConflict
	}

	// Store the account
	r.oauthAccounts[account.Provider][account.ProviderUserID] = account

	return nil
}

// GetUserByOAuthAccount retrieves a user by OAuth account
func (r *MemoryUserRepository) GetUserByOAuthAccount(ctx context.Context, provider, providerUserID string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Check if provider exists
	providerAccounts, ok := r.oauthAccounts[provider]
	if !ok {
		return nil, ErrUserNotFound
	}

	// Check if account exists
	account, ok := providerAccounts[providerUserID]
	if !ok {
		return nil, ErrUserNotFound
	}

	// Get the user
	user, ok := r.users[account.UserID]
	if !ok {
		return nil, ErrUserNotFound
	}

	return user, nil
}