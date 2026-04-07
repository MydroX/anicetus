-- +goose Up
-- +goose StatementBegin
CREATE TABLE roles (
    uuid UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE permissions (
    uuid UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE role_permissions (
    role_uuid UUID NOT NULL REFERENCES roles(uuid) ON DELETE CASCADE,
    permission_uuid UUID NOT NULL REFERENCES permissions(uuid) ON DELETE CASCADE,
    PRIMARY KEY (role_uuid, permission_uuid)
);

CREATE TABLE user_roles (
    user_uuid UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    role_uuid UUID NOT NULL REFERENCES roles(uuid) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_uuid, role_uuid)
);

-- Seed default roles
INSERT INTO roles (uuid, name, description) VALUES (uuidv7(), 'admin', 'Full system administrator');
INSERT INTO roles (uuid, name, description) VALUES (uuidv7(), 'user', 'Standard user');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
-- +goose StatementEnd
