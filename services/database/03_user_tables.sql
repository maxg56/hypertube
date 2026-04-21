-- ====================
-- USER TABLES
-- ====================

CREATE TABLE users (
    id          SERIAL PRIMARY KEY,
    username    VARCHAR(50)  UNIQUE NOT NULL,
    email       VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    first_name  VARCHAR(100),
    last_name   VARCHAR(100),
    avatar_url  TEXT,
    email_verified BOOLEAN   DEFAULT FALSE,
    created_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE email_verifications (
    id                SERIAL PRIMARY KEY,
    email             VARCHAR(255) NOT NULL,
    verification_code VARCHAR(6)   NOT NULL,
    expires_at        TIMESTAMP    NOT NULL,
    created_at        TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE password_resets (
    id         SERIAL PRIMARY KEY,
    user_id    INT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP   NOT NULL,
    used       BOOLEAN     DEFAULT FALSE,
    created_at TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
);
