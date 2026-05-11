-- ====================
-- TORRENT & STREAM TABLES
-- ====================

CREATE TABLE torrents (
    id           SERIAL PRIMARY KEY,
    movie_id     INT          NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    magnet_uri   TEXT         NOT NULL,
    info_hash    VARCHAR(64)  UNIQUE NOT NULL,
    status       torrent_status_enum DEFAULT 'pending',
    file_path    TEXT,
    file_size    BIGINT,
    downloaded   BIGINT       DEFAULT 0,
    progress     NUMERIC(5,2) DEFAULT 0,
    quality      VARCHAR(20),
    source       VARCHAR(50),
    error_msg    TEXT,
    created_at   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE watch_history (
    id           SERIAL PRIMARY KEY,
    user_id      INT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id     INT          NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    watched_at   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    progress_sec INT          DEFAULT 0,
    UNIQUE (user_id, movie_id)
);
