CREATE TABLE IF NOT EXISTS posts (
    ch_id BIGINT,
    post_id BIGINT,
    donor_ch_post_id BIGINT,
    PRIMARY KEY (ch_id, post_id)
);