-- ====================
-- ENUMS
-- ====================

CREATE TYPE user_role_enum AS ENUM (
    'user',
    'admin'
);

CREATE TYPE torrent_status_enum AS ENUM (
    'pending',
    'downloading',
    'ready',
    'error'
);

CREATE TYPE job_status_enum AS ENUM (
    'queued',
    'running',
    'done',
    'failed'
);

CREATE TYPE job_type_enum AS ENUM (
    'fetch_metadata',
    'download_torrent',
    'transcode',
    'fetch_subtitles',
    'cleanup'
);
