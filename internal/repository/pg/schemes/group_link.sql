CREATE TABLE IF NOT EXISTS group_link (
    id SERIAL,
    title VARCHAR(255),
    link VARCHAR(255),
    PRIMARY KEY (title)
);