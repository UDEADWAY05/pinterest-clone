DROP INDEX IF EXISTS idx_pins_active;

ALTER TABLE pins
DROP COLUMN IF EXISTS deleted_at;
