ALTER TABLE boards
ADD COLUMN is_default BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX idx_boards_default_per_user ON boards(owner_id) WHERE is_default = TRUE;
