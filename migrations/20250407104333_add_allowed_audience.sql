-- +goose Up
-- +goose StatementBegin
CREATE TABLE allowed_audiences (
    id SERIAL PRIMARY KEY,
    uuid VARCHAR(50) UNIQUE NOT NULL,
    audience VARCHAR(255) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    description TEXT,
    permissions JSONB,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(audience)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE allowed_audiences;
-- +goose StatementEnd
