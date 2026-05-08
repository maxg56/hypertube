-- ====================
-- FAVORITES TABLE
-- ====================

CREATE TABLE favorites (
    id        SERIAL PRIMARY KEY,
    user_id   INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tmdb_id   INT NOT NULL,
    added_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, tmdb_id)
);

CREATE INDEX ON favorites (user_id);
