DROP INDEX IF EXISTS idx_pins_embedding;
ALTER TABLE pins DROP COLUMN IF EXISTS embedding;
