CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL,
    username VARCHAR(255),
    firstname VARCHAR(255),
    is_admin INT DEFAULT 0,
    PRIMARY KEY (id)
);