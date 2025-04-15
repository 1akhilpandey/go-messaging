CREATE TABLE user_tokens (
    token_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token_value TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);