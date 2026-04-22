-- ====================
-- WORKER TABLES
-- ====================

CREATE TABLE jobs (
    id           SERIAL PRIMARY KEY,
    type         job_type_enum   NOT NULL,
    status       job_status_enum DEFAULT 'queued',
    payload      JSONB,
    result       JSONB,
    error_msg    TEXT,
    attempts     INT             DEFAULT 0,
    max_attempts INT             DEFAULT 3,
    scheduled_at TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    started_at   TIMESTAMP,
    finished_at  TIMESTAMP,
    created_at   TIMESTAMP       DEFAULT CURRENT_TIMESTAMP
);
