-- ====================
-- COMMENT TABLES
-- ====================

CREATE TABLE comments (
    id         SERIAL PRIMARY KEY,
    movie_id   INT          NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    user_id    INT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT         NOT NULL,
    created_at TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);
