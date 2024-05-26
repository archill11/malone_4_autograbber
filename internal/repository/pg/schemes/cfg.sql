CREATE TABLE IF NOT EXISTS cfg (
    id TEXT DEFAULT '',
    val TEXT DEFAULT '0',
    
    PRIMARY KEY (id)
);

-------------------------------------------

INSERT INTO cfg (id, val)
  VALUES ('auto-acc-media-gr', '0')
ON CONFLICT DO NOTHING;
