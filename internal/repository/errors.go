package repository

import "errors"

// Common errors
var (
	// ErrSlugUnavailable is returned when a slug is already in use
	ErrSlugUnavailable = errors.New("slug is already in use")
)
