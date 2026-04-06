-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_audiences (
    user_uuid UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    audience_uuid UUID NOT NULL REFERENCES allowed_audiences(uuid) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_uuid, audience_uuid)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_audiences;
-- +goose StatementEnd
