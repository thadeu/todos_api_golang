CREATE TABLE IF NOT EXISTS todos (
  id integer primary key autoincrement,
  uuid text not null,
  title text not null,
  description text,
  completed boolean not null default false,
  user_id integer not null,
  created_at timestamp not null default current_timestamp,
  updated_at timestamp not null default current_timestamp,
  deleted_at timestamp,
  
  FOREIGN KEY (user_id) REFERENCES users (id)
);


CREATE UNIQUE INDEX IF NOT EXISTS idx_todos_uuid_unique ON todos (uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_todos_title_unique ON todos (title);
CREATE UNIQUE INDEX IF NOT EXISTS idx_todos_completed_unique ON todos (completed);
