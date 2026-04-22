-- ====================
-- LIBRARY TABLES
-- ====================

CREATE TABLE movies (
    id              SERIAL PRIMARY KEY,
    tmdb_id         INT          UNIQUE NOT NULL,
    imdb_id         VARCHAR(20),
    title           VARCHAR(255) NOT NULL,
    original_title  VARCHAR(255),
    overview        TEXT,
    release_date    DATE,
    runtime         INT,
    rating          NUMERIC(3,1),
    vote_count      INT          DEFAULT 0,
    poster_path     TEXT,
    backdrop_path   TEXT,
    language        VARCHAR(10),
    cached_at       TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE genres (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE movie_genres (
    movie_id INT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    genre_id INT NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (movie_id, genre_id)
);
