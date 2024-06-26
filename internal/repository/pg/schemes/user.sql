CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL,
    username TEXT,
    firstname TEXT,
    is_admin INT DEFAULT 0,
    is_super_admin INT DEFAULT 0,
    is_user INT DEFAULT 0,
    PRIMARY KEY (id)
);

-------------------------------------------

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS is_user INT DEFAULT 0;