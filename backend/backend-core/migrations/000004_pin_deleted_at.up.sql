ALTER TABLE pins
ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;

CREATE INDEX idx_pins_active
ON pins (created_at DESC)
WHERE deleted_at IS NULL;
