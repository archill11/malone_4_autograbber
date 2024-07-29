CREATE TABLE IF NOT EXISTS bots (
    id                BIGINT,
    token             TEXT,
    username          TEXT,
    first_name        TEXT,
    is_donor          INT,
    ch_id             BIGINT DEFAULT 0,
    ch_link           TEXT DEFAULT '',
    group_link_id     INT DEFAULT 0,
    lichka            TEXT DEFAULT '',
    user_creator      BIGINT  DEFAULT 0,
    is_disable        INT  DEFAULT 0,
    created_at        TIMESTAMP DEFAULT now(),
    ch_is_skam        INT DEFAULT 0,
    personal_link     TEXT DEFAULT '',
    PRIMARY KEY (id, token)
);

-------------------------------------------

ALTER TABLE bots
  ADD COLUMN IF NOT EXISTS personal_link TEXT DEFAULT '';