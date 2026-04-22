-- ====================
-- INDEXES
-- ====================

-- users
CREATE INDEX idx_users_email    ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- email_verifications
CREATE INDEX idx_email_verifications_email ON email_verifications(email);

-- password_resets
CREATE INDEX idx_password_resets_token      ON password_resets(token);
CREATE INDEX idx_password_resets_user_id    ON password_resets(user_id);
CREATE INDEX idx_password_resets_expires_at ON password_resets(expires_at);

-- movies
CREATE INDEX idx_movies_tmdb_id      ON movies(tmdb_id);
CREATE INDEX idx_movies_imdb_id      ON movies(imdb_id);
CREATE INDEX idx_movies_release_date ON movies(release_date);
CREATE INDEX idx_movies_rating       ON movies(rating DESC);

-- torrents
CREATE INDEX idx_torrents_movie_id  ON torrents(movie_id);
CREATE INDEX idx_torrents_info_hash ON torrents(info_hash);
CREATE INDEX idx_torrents_status    ON torrents(status);

-- watch_history
CREATE INDEX idx_watch_history_user_id  ON watch_history(user_id);
CREATE INDEX idx_watch_history_movie_id ON watch_history(movie_id);

-- comments
CREATE INDEX idx_comments_movie_id ON comments(movie_id);
CREATE INDEX idx_comments_user_id  ON comments(user_id);

-- jobs
CREATE INDEX idx_jobs_status       ON jobs(status);
CREATE INDEX idx_jobs_type         ON jobs(type);
CREATE INDEX idx_jobs_scheduled_at ON jobs(scheduled_at);
