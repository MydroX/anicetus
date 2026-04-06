-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_audiences (
    user_uuid VARCHAR(50) NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    audience_id INTEGER NOT NULL REFERENCES allowed_audiences(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_uuid, audience_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_audiences;
-- +goose StatementEnd
