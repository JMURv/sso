-- USERS

CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(50)  NOT NULL,
    password   VARCHAR(255) NULL,
    email      VARCHAR(50)  NOT NULL UNIQUE,
    avatar     VARCHAR(255),

    created_at TIMESTAMPTZ      DEFAULT NOW(),
    updated_at TIMESTAMPTZ      DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permission (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS user_permission (
    user_id       UUID   NOT NULL,
    permission_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, permission_id),

    value         BOOLEAN DEFAULT FALSE,

    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_permission FOREIGN KEY (permission_id) REFERENCES permission (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS users_email_idx ON users (email);
CREATE INDEX IF NOT EXISTS permission_name_idx ON permission (name);

-- USER DEVICES

CREATE TABLE IF NOT EXISTS user_devices (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     UUID         NOT NULL REFERENCES users (id),
    name        VARCHAR(100) NOT NULL,
    device_type VARCHAR(50),
    os          VARCHAR(50),
    browser     VARCHAR(50),
    user_agent  TEXT,
    ip          VARCHAR(45),
    last_active TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_devices_user_id ON user_devices (user_id);
CREATE INDEX IF NOT EXISTS idx_user_devices_ip ON user_devices (ip);
CREATE INDEX IF NOT EXISTS idx_user_devices_last_active ON user_devices (last_active);

-- REFRESH TOKEN

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id           SERIAL PRIMARY KEY,
    user_id      UUID                     NOT NULL,
    token_hash   TEXT                     NOT NULL UNIQUE,
    expires_at   TIMESTAMPTZ              NOT NULL,
    revoked      BOOLEAN                  NOT NULL DEFAULT FALSE,
    device_id    VARCHAR(36)              NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ              NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,

    CONSTRAINT fk_device
        FOREIGN KEY (device_id) REFERENCES user_devices (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens (expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_device_id ON refresh_tokens (device_id);

-- OAUTH2
CREATE TABLE IF NOT EXISTS oauth2_connections (
    id            SERIAL PRIMARY KEY,
    user_id       UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider      VARCHAR(50)  NOT NULL, -- google, github etc.
    provider_id   VARCHAR(255) NOT NULL, -- User ID from provider
    access_token  TEXT,
    refresh_token TEXT,
    id_token      TEXT,
    expires_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
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
    user_id          UUID REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (id, user_id)
);

INSERT INTO permission (name)
VALUES ('admin'),
       ('staff')
ON CONFLICT (name) DO NOTHING;