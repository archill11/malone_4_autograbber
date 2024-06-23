CREATE TABLE IF NOT EXISTS group_link (
    id SERIAL,
    title TEXT,
    link TEXT,
    user_creator INT  DEFAULT 0,
    PRIMARY KEY (title)
);

-------------------------------------------

ALTER TABLE group_link
  ADD COLUMN IF NOT EXISTS user_creator INT DEFAULT 0;