package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
	"github.com/lib/pq"
)

// PostgresBioPageRepository is a PostgreSQL implementation of the BioPageRepository interface
type PostgresBioPageRepository struct {
	db *sql.DB
}

// NewPostgresBioPageRepository creates a new PostgreSQL bio page repository
func NewPostgresBioPageRepository(db *sql.DB) (*PostgresBioPageRepository, error) {
	return &PostgresBioPageRepository{
		db: db,
	}, nil
}

// CreateBioPage creates a new bio page
func (r *PostgresBioPageRepository) CreateBioPage(ctx context.Context, bioPage *models.BioPage) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert the bio page
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO bio_pages (user_id, short_code, title, description, theme, profile_image_url, 
                              created_at, updated_at, visits, last_visit_at, is_published, custom_css) 
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) 
         RETURNING id`,
		bioPage.UserID,
		bioPage.ShortCode,
		bioPage.Title,
		bioPage.Description,
		bioPage.Theme,
		bioPage.ProfileImageURL,
		bioPage.CreatedAt,
		bioPage.UpdatedAt,
		bioPage.Visits,
		nil, // last_visit_at starts as NULL
		bioPage.IsPublished,
		bioPage.CustomCSS,
	).Scan(&bioPage.ID)

	if err != nil {
		// Check for unique violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrSlugUnavailable
		}
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// GetBioPageByID retrieves a bio page by ID
func (r *PostgresBioPageRepository) GetBioPageByID(ctx context.Context, id int) (*models.BioPage, error) {
	var bioPage models.BioPage
	var lastVisitAt sql.NullTime
	var description sql.NullString
	var profileImageURL sql.NullString
	var customCSS sql.NullString

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, short_code, title, description, theme, profile_image_url, 
                created_at, updated_at, visits, last_visit_at, is_published, custom_css 
         FROM bio_pages 
         WHERE id = $1`,
		id,
	).Scan(
		&bioPage.ID,
		&bioPage.UserID,
		&bioPage.ShortCode,
		&bioPage.Title,
		&description,
		&bioPage.Theme,
		&profileImageURL,
		&bioPage.CreatedAt,
		&bioPage.UpdatedAt,
		&bioPage.Visits,
		&lastVisitAt,
		&bioPage.IsPublished,
		&customCSS,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Handle nullable fields
	if description.Valid {
		bioPage.Description = description.String
	}
	if profileImageURL.Valid {
		bioPage.ProfileImageURL = profileImageURL.String
	}
	if lastVisitAt.Valid {
		bioPage.LastVisitAt = lastVisitAt.Time
	}
	if customCSS.Valid {
		bioPage.CustomCSS = customCSS.String
	}

	// Get the links for this bio page
	links, err := r.ListBioLinksByBioPageID(ctx, bioPage.ID)
	if err != nil {
		return &bioPage, nil // Return page without links
	}

	bioPage.Links = links
	return &bioPage, nil
}

// GetBioPageByShortCode retrieves a bio page by short code
func (r *PostgresBioPageRepository) GetBioPageByShortCode(ctx context.Context, shortCode string) (*models.BioPage, error) {
	var bioPage models.BioPage
	var lastVisitAt sql.NullTime
	var description sql.NullString
	var profileImageURL sql.NullString
	var customCSS sql.NullString

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, short_code, title, description, theme, profile_image_url, 
                created_at, updated_at, visits, last_visit_at, is_published, custom_css 
         FROM bio_pages 
         WHERE short_code = $1`,
		shortCode,
	).Scan(
		&bioPage.ID,
		&bioPage.UserID,
		&bioPage.ShortCode,
		&bioPage.Title,
		&description,
		&bioPage.Theme,
		&profileImageURL,
		&bioPage.CreatedAt,
		&bioPage.UpdatedAt,
		&bioPage.Visits,
		&lastVisitAt,
		&bioPage.IsPublished,
		&customCSS,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Handle nullable fields
	if description.Valid {
		bioPage.Description = description.String
	}
	if profileImageURL.Valid {
		bioPage.ProfileImageURL = profileImageURL.String
	}
	if lastVisitAt.Valid {
		bioPage.LastVisitAt = lastVisitAt.Time
	}
	if customCSS.Valid {
		bioPage.CustomCSS = customCSS.String
	}

	// Get the links for this bio page
	links, err := r.ListBioLinksByBioPageID(ctx, bioPage.ID)
	if err != nil {
		return &bioPage, nil // Return page without links
	}

	bioPage.Links = links
	return &bioPage, nil
}

// ListBioPagesByUserID lists all bio pages for a user
func (r *PostgresBioPageRepository) ListBioPagesByUserID(ctx context.Context, userID int) ([]*models.BioPage, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, user_id, short_code, title, description, theme, profile_image_url, 
                created_at, updated_at, visits, last_visit_at, is_published, custom_css 
         FROM bio_pages 
         WHERE user_id = $1
         ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bioPages := []*models.BioPage{}
	for rows.Next() {
		var bioPage models.BioPage
		var lastVisitAt sql.NullTime
		var description sql.NullString
		var profileImageURL sql.NullString
		var customCSS sql.NullString

		err := rows.Scan(
			&bioPage.ID,
			&bioPage.UserID,
			&bioPage.ShortCode,
			&bioPage.Title,
			&description,
			&bioPage.Theme,
			&profileImageURL,
			&bioPage.CreatedAt,
			&bioPage.UpdatedAt,
			&bioPage.Visits,
			&lastVisitAt,
			&bioPage.IsPublished,
			&customCSS,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if description.Valid {
			bioPage.Description = description.String
		}
		if profileImageURL.Valid {
			bioPage.ProfileImageURL = profileImageURL.String
		}
		if lastVisitAt.Valid {
			bioPage.LastVisitAt = lastVisitAt.Time
		}
		if customCSS.Valid {
			bioPage.CustomCSS = customCSS.String
		}

		bioPages = append(bioPages, &bioPage)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get links for each bio page
	for _, bioPage := range bioPages {
		links, err := r.ListBioLinksByBioPageID(ctx, bioPage.ID)
		if err != nil {
			continue // Skip links for this page
		}
		bioPage.Links = links
	}

	return bioPages, nil
}

// UpdateBioPage updates a bio page
func (r *PostgresBioPageRepository) UpdateBioPage(ctx context.Context, bioPage *models.BioPage) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update the timestamp
	bioPage.UpdatedAt = time.Now()

	// If LastVisitAt is zero, set it to NULL
	var lastVisitAt interface{}
	if bioPage.LastVisitAt.IsZero() {
		lastVisitAt = nil
	} else {
		lastVisitAt = bioPage.LastVisitAt
	}

	// Update the bio page
	result, err := tx.ExecContext(
		ctx,
		`UPDATE bio_pages 
         SET title = $1, description = $2, theme = $3, profile_image_url = $4, 
             updated_at = $5, visits = $6, last_visit_at = $7, is_published = $8, custom_css = $9 
         WHERE id = $10`,
		bioPage.Title,
		bioPage.Description,
		bioPage.Theme,
		bioPage.ProfileImageURL,
		bioPage.UpdatedAt,
		bioPage.Visits,
		lastVisitAt,
		bioPage.IsPublished,
		bioPage.CustomCSS,
		bioPage.ID,
	)

	if err != nil {
		return err
	}

	// Check if the bio page was updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	// Commit the transaction
	return tx.Commit()
}

// DeleteBioPage deletes a bio page
func (r *PostgresBioPageRepository) DeleteBioPage(ctx context.Context, id int) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete the bio page
	result, err := tx.ExecContext(
		ctx,
		`DELETE FROM bio_pages WHERE id = $1`,
		id,
	)

	if err != nil {
		return err
	}

	// Check if the bio page was deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	// Commit the transaction
	return tx.Commit()
}

// CreateBioLink creates a new bio link
func (r *PostgresBioPageRepository) CreateBioLink(ctx context.Context, bioLink *models.BioLink) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert the bio link
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO bio_links (bio_page_id, title, url, display_order, icon, created_at, updated_at, visits, is_enabled) 
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
         RETURNING id`,
		bioLink.BioPageID,
		bioLink.Title,
		bioLink.URL,
		bioLink.DisplayOrder,
		bioLink.Icon,
		bioLink.CreatedAt,
		bioLink.UpdatedAt,
		bioLink.Visits,
		bioLink.IsEnabled,
	).Scan(&bioLink.ID)

	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// GetBioLinkByID retrieves a bio link by ID
func (r *PostgresBioPageRepository) GetBioLinkByID(ctx context.Context, id int) (*models.BioLink, error) {
	var bioLink models.BioLink
	var icon sql.NullString

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, bio_page_id, title, url, display_order, icon, created_at, updated_at, visits, is_enabled 
         FROM bio_links 
         WHERE id = $1`,
		id,
	).Scan(
		&bioLink.ID,
		&bioLink.BioPageID,
		&bioLink.Title,
		&bioLink.URL,
		&bioLink.DisplayOrder,
		&icon,
		&bioLink.CreatedAt,
		&bioLink.UpdatedAt,
		&bioLink.Visits,
		&bioLink.IsEnabled,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Handle nullable fields
	if icon.Valid {
		bioLink.Icon = icon.String
	}

	return &bioLink, nil
}

// ListBioLinksByBioPageID lists all bio links for a bio page
func (r *PostgresBioPageRepository) ListBioLinksByBioPageID(ctx context.Context, bioPageID int) ([]*models.BioLink, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, bio_page_id, title, url, display_order, icon, created_at, updated_at, visits, is_enabled 
         FROM bio_links 
         WHERE bio_page_id = $1
         ORDER BY display_order ASC`,
		bioPageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bioLinks := []*models.BioLink{}
	for rows.Next() {
		var bioLink models.BioLink
		var icon sql.NullString

		err := rows.Scan(
			&bioLink.ID,
			&bioLink.BioPageID,
			&bioLink.Title,
			&bioLink.URL,
			&bioLink.DisplayOrder,
			&icon,
			&bioLink.CreatedAt,
			&bioLink.UpdatedAt,
			&bioLink.Visits,
			&bioLink.IsEnabled,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if icon.Valid {
			bioLink.Icon = icon.String
		}

		bioLinks = append(bioLinks, &bioLink)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bioLinks, nil
}

// UpdateBioLink updates a bio link
func (r *PostgresBioPageRepository) UpdateBioLink(ctx context.Context, bioLink *models.BioLink) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update the timestamp
	bioLink.UpdatedAt = time.Now()

	// Update the bio link
	result, err := tx.ExecContext(
		ctx,
		`UPDATE bio_links 
         SET title = $1, url = $2, display_order = $3, icon = $4, 
             updated_at = $5, visits = $6, is_enabled = $7 
         WHERE id = $8`,
		bioLink.Title,
		bioLink.URL,
		bioLink.DisplayOrder,
		bioLink.Icon,
		bioLink.UpdatedAt,
		bioLink.Visits,
		bioLink.IsEnabled,
		bioLink.ID,
	)

	if err != nil {
		return err
	}

	// Check if the bio link was updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	// Commit the transaction
	return tx.Commit()
}

// DeleteBioLink deletes a bio link
func (r *PostgresBioPageRepository) DeleteBioLink(ctx context.Context, id int) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete the bio link
	result, err := tx.ExecContext(
		ctx,
		`DELETE FROM bio_links WHERE id = $1`,
		id,
	)

	if err != nil {
		return err
	}

	// Check if the bio link was deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	// Commit the transaction
	return tx.Commit()
}

// ReorderBioLinks updates the display order of bio links
func (r *PostgresBioPageRepository) ReorderBioLinks(ctx context.Context, bioPageID int, linkIDs []int) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update the display order for each link
	for i, linkID := range linkIDs {
		_, err := tx.ExecContext(
			ctx,
			`UPDATE bio_links 
             SET display_order = $1, updated_at = $2
             WHERE id = $3 AND bio_page_id = $4`,
			i,
			time.Now(),
			linkID,
			bioPageID,
		)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	return tx.Commit()
}