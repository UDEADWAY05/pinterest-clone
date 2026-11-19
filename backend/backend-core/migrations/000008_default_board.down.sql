ALTER TABLE boards
DROP COLUMN is_default;
DROP INDEX idx_boards_default_per_user;
