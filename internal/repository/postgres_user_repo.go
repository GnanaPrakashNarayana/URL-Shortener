package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
	"github.com/lib/pq"
)

// PostgresUserRepository is a PostgreSQL implementation of the UserRepository interface
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sql.DB) (*PostgresUserRepository, error) {
	return &PostgresUserRepository{
		db: db,
	}, nil
}

// Create creates a new user in the database
func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert the user
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO users (username, email, password_hash, role, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5, $6) 
         RETURNING id`,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		// Check for unique violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrUserConflict
		}
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, email, password_hash, role, created_at, updated_at 
         FROM users 
         WHERE id = $1`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Get OAuth accounts
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, user_id, provider, provider_user_id, created_at
         FROM oauth_accounts
         WHERE user_id = $1`,
		user.ID,
	)
	if err != nil {
		return &user, nil // Return user without OAuth accounts
	}
	defer rows.Close()

	user.OAuthAccounts = []*models.OAuthAccount{}
	for rows.Next() {
		var account models.OAuthAccount
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.ProviderUserID,
			&account.CreatedAt,
		)
		if err != nil {
			return &user, nil // Return user without all OAuth accounts
		}
		user.OAuthAccounts = append(user.OAuthAccounts, &account)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, email, password_hash, role, created_at, updated_at 
         FROM users 
         WHERE username = $1`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, email, password_hash, role, created_at, updated_at 
         FROM users 
         WHERE email = $1`,
		email,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// Update updates a user
func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update the timestamp
	user.UpdatedAt = time.Now()

	// Update the user
	result, err := tx.ExecContext(
		ctx,
		`UPDATE users 
         SET username = $1, email = $2, password_hash = $3, role = $4, updated_at = $5 
         WHERE id = $6`,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		// Check for unique violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrUserConflict
		}
		return err
	}

	// Check if the user was updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	// Commit the transaction
	return tx.Commit()
}

// Delete deletes a user
func (r *PostgresUserRepository) Delete(ctx context.Context, id int) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete the user
	result, err := tx.ExecContext(
		ctx,
		`DELETE FROM users WHERE id = $1`,
		id,
	)

	if err != nil {
		return err
	}

	// Check if the user was deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	// Commit the transaction
	return tx.Commit()
}

// List lists all users
func (r *PostgresUserRepository) List(ctx context.Context) ([]*models.User, error) {
	// Query all users
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, username, email, password_hash, role, created_at, updated_at 
         FROM users 
         ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse the rows
	users := []*models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// CreateOAuthAccount creates a new OAuth account
func (r *PostgresUserRepository) CreateOAuthAccount(ctx context.Context, account *models.OAuthAccount) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert the account
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO oauth_accounts (user_id, provider, provider_user_id, created_at) 
         VALUES ($1, $2, $3, $4) 
         RETURNING id`,
		account.UserID,
		account.Provider,
		account.ProviderUserID,
		time.Now(),
	).Scan(&account.ID)

	if err != nil {
		// Check for unique violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrUserConflict
		}
		// Check for foreign key violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return ErrUserNotFound
		}
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// GetUserByOAuthAccount retrieves a user by OAuth account
func (r *PostgresUserRepository) GetUserByOAuthAccount(ctx context.Context, provider, providerUserID string) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowContext(
		ctx,
		`SELECT u.id, u.username, u.email, u.password_hash, u.role, u.created_at, u.updated_at 
         FROM users u
         JOIN oauth_accounts oa ON u.id = oa.user_id
         WHERE oa.provider = $1 AND oa.provider_user_id = $2`,
		provider,
		providerUserID,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}