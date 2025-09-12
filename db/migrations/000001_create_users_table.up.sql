CREATE TABLE IF NOT EXISTS users (
    id integer primary key autoincrement,
    uuid text not null,
    name text null,
    email text null,
    encrypted_password text not null,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp,
    deleted_at timestamp null
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_uuid_unique ON users (uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users (email);