CREATE TABLE IF NOT EXISTS short_urls (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NULL,
    click_count BIGINT NOT NULL DEFAULT 0,
    last_accessed_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_short_urls_expires_at ON short_urls (expires_at);
