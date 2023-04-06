CREATE TABLE IF NOT EXISTS bots (
    id BIGINT,
    token VARCHAR(255),
    username VARCHAR(255),
    first_name VARCHAR(255),
    is_donor INT,
    ch_id BIGINT DEFAULT 0,
    ch_link VARCHAR(255) DEFAULT '',
    PRIMARY KEY (id, token)
);