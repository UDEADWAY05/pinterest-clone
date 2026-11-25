DROP INDEX IF EXISTS idx_boards_default_per_user;

ALTER TABLE boards
DROP COLUMN IF EXISTS is_default;
