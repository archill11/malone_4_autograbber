CREATE TABLE IF NOT EXISTS posts (
    ch_id BIGINT,
    post_id BIGINT,
    donor_ch_post_id BIGINT,
    created_at TIMESTAMP DEFAULT now(),
    PRIMARY KEY (ch_id, post_id, donor_ch_post_id)
);