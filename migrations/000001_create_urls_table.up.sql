CREATE TABLE IF NOT EXISTS urls (
    id VARCHAR(255) PRIMARY KEY,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    visits INT NOT NULL DEFAULT 0,
    last_visit_at TIMESTAMP NULL
);

CREATE INDEX idx_urls_created_at ON urls(created_at);