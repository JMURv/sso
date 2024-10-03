CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name       VARCHAR(50)  NOT NULL,
    password   VARCHAR(255) NOT NULL,
    email      VARCHAR(50)  NOT NULL UNIQUE,

    avatar     VARCHAR(255),
    address    VARCHAR(255),
    phone      VARCHAR(20),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    value BOOLEAN DEFAULT FALSE
);
