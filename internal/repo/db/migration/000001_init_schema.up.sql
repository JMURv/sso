-- USERS
CREATE TABLE IF NOT EXISTS users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(50)  NOT NULL,
    password          VARCHAR(255) NULL,
    email             VARCHAR(50)  NOT NULL UNIQUE,
    avatar            VARCHAR(255),
    is_wa             BOOLEAN          DEFAULT FALSE,
    is_active         BOOLEAN          DEFAULT FALSE,
    is_email_verified BOOLEAN          DEFAULT FALSE,
    created_at        TIMESTAMPTZ      DEFAULT NOW(),
    updated_at        TIMESTAMPTZ      DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS users_email_idx ON users (email);

-- ROLES
CREATE TABLE IF NOT EXISTS permission (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS roles (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    CONSTRAINT fk_permission FOREIGN KEY (permission_id) REFERENCES permission (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID   NOT NULL,
    role_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_permissions_name ON permission (name);
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles (name);
CREATE INDEX IF NOT EXISTS idx_user_roles_user ON user_roles(user_id);

-- USER DEVICES
CREATE TABLE IF NOT EXISTS user_devices (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     UUID         NOT NULL,
    name        VARCHAR(100) NOT NULL,
    device_type VARCHAR(50),
    os          VARCHAR(50),
    browser     VARCHAR(50),
    user_agent  TEXT,
    ip          VARCHAR(45),
    last_active TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_devices_user_id ON user_devices (user_id);
CREATE INDEX IF NOT EXISTS idx_user_devices_ip ON user_devices (ip);
CREATE INDEX IF NOT EXISTS idx_user_devices_last_active ON user_devices (last_active);

-- REFRESH TOKEN
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id           SERIAL PRIMARY KEY,
    user_id      UUID        NOT NULL,
    token_hash   TEXT        NOT NULL UNIQUE,
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked      BOOLEAN     NOT NULL DEFAULT FALSE,
    device_id    VARCHAR(36) NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_device FOREIGN KEY (device_id) REFERENCES user_devices (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens (expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_device_id ON refresh_tokens (device_id);

-- OAUTH2
CREATE TABLE IF NOT EXISTS oauth2_connections (
    id            SERIAL PRIMARY KEY,
    user_id       UUID         NOT NULL,
    provider      VARCHAR(50)  NOT NULL, -- google, github etc.
    provider_id   VARCHAR(255) NOT NULL, -- User ID from provider
    access_token  TEXT,
    refresh_token TEXT,
    id_token      TEXT,
    expires_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    UNIQUE (provider, provider_id)
);

CREATE INDEX IF NOT EXISTS idx_oauth2_user ON oauth2_connections (user_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_provider ON oauth2_connections (provider);
CREATE INDEX IF NOT EXISTS idx_oauth2_provider_id ON oauth2_connections (provider_id);

-- WebAuthn
CREATE TABLE IF NOT EXISTS wa_credentials (
    id               BYTEA NOT NULL,
    public_key       BYTEA NOT NULL,
    attestation_type TEXT  NOT NULL,
    authenticator    JSONB NOT NULL,
    user_id          UUID  NOT NULL,
    PRIMARY KEY (id, user_id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Initial inserts
INSERT INTO permission (name, description)
VALUES ('manage_users', 'Create/update/delete users'),
       ('manage_roles', 'Manage roles and permissions'),
       ('manage_projects', 'Create/edit projects'),
       ('manage_tasks', 'Create/assign development tasks'),
       ('edit_code', 'Commit code to repositories'),
       ('review_code', 'Review pull requests'),
       ('deploy', 'Deploy applications to environments'),
       ('access_dev_tools', 'Access development tools'),
       ('view_analytics', 'View project analytics'),
       ('manage_docs', 'Manage technical documentation')
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles (name, description)
VALUES ('admin', 'Full system access'),
       ('project_manager', 'Manages projects and tasks'),
       ('developer', 'Technical team member'),
       ('tester', 'Quality assurance specialist'),
       ('devops', 'Deployment and infrastructure')
ON CONFLICT (name) DO NOTHING;


INSERT INTO role_permissions (role_id, permission_id)
VALUES
-- Admin
((SELECT id FROM roles WHERE name = 'admin'), (SELECT id FROM permission WHERE name = 'manage_users')),
((SELECT id FROM roles WHERE name = 'admin'), (SELECT id FROM permission WHERE name = 'manage_roles')),
((SELECT id FROM roles WHERE name = 'admin'), (SELECT id FROM permission WHERE name = 'manage_projects')),
((SELECT id FROM roles WHERE name = 'admin'), (SELECT id FROM permission WHERE name = 'manage_tasks')),
((SELECT id FROM roles WHERE name = 'admin'), (SELECT id FROM permission WHERE name = 'edit_code')),
((SELECT id FROM roles WHERE name = 'admin'), (SELECT id FROM permission WHERE name = 'review_code')),
((SELECT id FROM roles WHERE name = 'admin'), (SELECT id FROM permission WHERE name = 'deploy')),
((SELECT id FROM roles WHERE name = 'admin'), (SELECT id FROM permission WHERE name = 'access_dev_tools')),
((SELECT id FROM roles WHERE name = 'admin'), (SELECT id FROM permission WHERE name = 'view_analytics')),
((SELECT id FROM roles WHERE name = 'admin'), (SELECT id FROM permission WHERE name = 'manage_docs')),

-- Project Manager
((SELECT id FROM roles WHERE name = 'project_manager'), (SELECT id FROM permission WHERE name = 'manage_projects')),
((SELECT id FROM roles WHERE name = 'project_manager'), (SELECT id FROM permission WHERE name = 'manage_tasks')),

-- Developer
((SELECT id FROM roles WHERE name = 'developer'), (SELECT id FROM permission WHERE name = 'edit_code')),
((SELECT id FROM roles WHERE name = 'developer'), (SELECT id FROM permission WHERE name = 'manage_docs')),

-- DevOps
((SELECT id FROM roles WHERE name = 'devops'), (SELECT id FROM permission WHERE name = 'deploy')),

-- Tester
((SELECT id FROM roles WHERE name = 'tester'), (SELECT id FROM permission WHERE name = 'review_code'))
ON CONFLICT (role_id, permission_id) DO NOTHING;
