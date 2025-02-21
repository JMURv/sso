CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(50)  NOT NULL,
    password   VARCHAR(255) NOT NULL,
    email      VARCHAR(50)  NOT NULL UNIQUE,

    avatar     VARCHAR(255),
    address    VARCHAR(255),
    phone      VARCHAR(20),

    created_at TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permission (
    id    SERIAL PRIMARY KEY,
    name  VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS user_permission (
    user_id       UUID NOT NULL,
    permission_id BIGINT NOT NULL ,
    PRIMARY KEY (user_id, permission_id),

    value BOOLEAN DEFAULT FALSE,

    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_permission FOREIGN KEY (permission_id) REFERENCES permission (id) ON DELETE CASCADE
);

INSERT INTO permission (name) VALUES('admin'), ('staff') ON CONFLICT (name) DO NOTHING;

CREATE INDEX IF NOT EXISTS users_email_idx ON users (email);
CREATE INDEX IF NOT EXISTS permission_name_idx ON permission (name);