CREATE TABLE IF NOT EXISTS users (
    id integer primary key autoincrement,
    uuid text,
    name text,
    email text,
    created_at timestamp,
    updated_at timestamp,
    deleted_at timestamp
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_uuid_unique ON users (uuid);