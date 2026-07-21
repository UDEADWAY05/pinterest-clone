CREATE TABLE IF NOT EXISTS pin_tags (
    pin_id BIGINT NOT NULL REFERENCES pins(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (pin_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_pin_tags_tag ON pin_tags(tag);
