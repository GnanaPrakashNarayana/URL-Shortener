-- Create the urls table with all necessary columns including short_code and user_id
CREATE TABLE IF NOT EXISTS urls (
    id VARCHAR(255) NOT NULL, -- Changed from PRIMARY KEY here, short_code will be PK or have UNIQUE index
    original_url TEXT NOT NULL,
    short_code VARCHAR(50) NOT NULL UNIQUE, -- Added short_code, ensure it's UNIQUE
    user_id VARCHAR(255) NULL,              -- Added user_id (nullable)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Use TIMESTAMPTZ and DEFAULT
    visit_count INTEGER NOT NULL DEFAULT 0, -- Renamed from visits, added DEFAULT
    -- Removed last_visit_at as it's not in the Go model currently

    -- Define PRIMARY KEY (can be id or short_code, short_code makes sense for lookups)
    PRIMARY KEY (short_code) -- Making short_code the primary key automatically makes it unique and indexed

    -- Add FOREIGN KEY constraint if users table exists (assuming migration 000002 creates users(id))
    -- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Optional: If you keep 'id' as PK, you still need a UNIQUE index on short_code
-- CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_short_code_unique ON urls(short_code);

-- Add index on user_id for faster lookups of user's URLs
CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);

-- Index on created_at might be useful depending on queries
CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at);

-- Note: If migration 000002_create_users_table.up.sql exists and creates the users table,
-- uncomment the FOREIGN KEY line above AFTER the users table is created.
-- If you add the foreign key here, make sure migration 000002 runs *before* this one,
-- or add the foreign key in a later migration using ALTER TABLE.
-- For simplicity now, let's assume users table is created in migration 2 and we don't add the FK here yet.
-- It can be added later if needed.

