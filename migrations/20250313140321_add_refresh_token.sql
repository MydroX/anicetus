-- +goose Up
-- +goose StatementBegin
CREATE TABLE sessions (
    uuid UUID PRIMARY KEY DEFAULT uuidv7(),
    user_uuid UUID NOT NULL REFERENCES users(uuid),
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

CREATE INDEX idx_sessions_user_uuid ON sessions(user_uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sessions;
-- +goose StatementEnd
