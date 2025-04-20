ALTER TABLE urls ADD COLUMN expires_at TIMESTAMP NULL;

-- Add an index to improve query performance when checking expired URLs
CREATE INDEX idx_urls_expires_at ON urls(expires_at);