-- ====================
-- SCHEMA SETUP
-- ====================
CREATE SCHEMA IF NOT EXISTS public;
SET search_path TO public;

-- ====================
-- DROP TABLES (reverse dependency order)
-- ====================
DROP TABLE IF EXISTS jobs CASCADE;
DROP TABLE IF EXISTS watch_history CASCADE;
DROP TABLE IF EXISTS comments CASCADE;
DROP TABLE IF EXISTS torrents CASCADE;
DROP TABLE IF EXISTS movie_genres CASCADE;
DROP TABLE IF EXISTS genres CASCADE;
DROP TABLE IF EXISTS movies CASCADE;
DROP TABLE IF EXISTS password_resets CASCADE;
DROP TABLE IF EXISTS email_verifications CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- ====================
-- DROP TYPES
-- ====================
DROP TYPE IF EXISTS torrent_status_enum CASCADE;
DROP TYPE IF EXISTS job_status_enum CASCADE;
DROP TYPE IF EXISTS job_type_enum CASCADE;
