-- +goose Up
-- +goose StatementBegin
CREATE TABLE sessions (
    id SERIAL PRIMARY KEY,
    uuid VARCHAR(50) UNIQUE NOT NULL,
    user_uuid VARCHAR(50) NOT NULL REFERENCES users(uuid),
    refresh_token_hash TEXT,
    last_used_at TIMESTAMP,
    os VARCHAR(50),
    os_version VARCHAR(50),
    browser VARCHAR(50),
    browser_version VARCHAR(50),
    ipv4_address VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sessions;
-- +goose StatementEnd
