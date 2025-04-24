CREATE TABLE IF NOT EXISTS bio_pages (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    short_code VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    theme VARCHAR(50) DEFAULT 'default',
    profile_image_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    visits INT NOT NULL DEFAULT 0,
    last_visit_at TIMESTAMP,
    is_published BOOLEAN NOT NULL DEFAULT false,
    custom_css TEXT
);

CREATE TABLE IF NOT EXISTS bio_links (
    id SERIAL PRIMARY KEY,
    bio_page_id INT NOT NULL REFERENCES bio_pages(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    icon VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    visits INT NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT true
);

-- Create indexes
CREATE INDEX idx_bio_pages_user_id ON bio_pages(user_id);
CREATE INDEX idx_bio_pages_short_code ON bio_pages(short_code);
CREATE INDEX idx_bio_links_bio_page_id ON bio_links(bio_page_id);
CREATE INDEX idx_bio_links_display_order ON bio_links(display_order);