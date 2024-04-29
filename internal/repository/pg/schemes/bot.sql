CREATE TABLE IF NOT EXISTS bots (
    id            BIGINT,
    token         TEXT,
    username      TEXT,
    first_name    TEXT,
    is_donor      INT,
    ch_id         BIGINT DEFAULT 0,
    ch_link       TEXT DEFAULT '',
    group_link_id INT DEFAULT 0,
    lichka        TEXT DEFAULT '',
    user_creator  INT  DEFAULT 0,
    is_disable    INT  DEFAULT 0,
    created_at    TIMESTAMP DEFAULT now(),
    ch_is_skam    INT DEFAULT 0,
    PRIMARY KEY (id, token)
);