-- ====================
-- HYPERTUBE DATABASE INITIALIZATION
-- ====================

\echo 'Starting Hypertube database initialization...'

\echo '1. Setting up schema...'
\i 01_schema_setup.sql

\echo '2. Creating enums...'
\i 02_enums.sql

\echo '3. Creating user tables...'
\i 03_user_tables.sql

\echo '4. Creating library tables...'
\i 04_library_tables.sql

\echo '5. Creating torrent & stream tables...'
\i 05_torrent_tables.sql

\echo '6. Creating comment tables...'
\i 06_comment_tables.sql

\echo '7. Creating worker tables...'
\i 07_worker_tables.sql

\echo '8. Creating indexes...'
\i 08_indexes.sql

\echo 'Hypertube database initialization completed successfully!'
