CREATE TABLE IF NOT EXISTS chats (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  user_ids TEXT,
  is_group BOOLEAN NOT NULL DEFAULT 0
);