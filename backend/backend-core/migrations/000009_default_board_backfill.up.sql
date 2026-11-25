-- Пользователям с досками, но без дефолтной —
-- назначить одну существующую доску.
WITH candidates AS (
    SELECT DISTINCT ON (owner_id) id
    FROM boards
    WHERE owner_id NOT IN (
        SELECT owner_id
        FROM boards
        WHERE is_default = TRUE
    )
    ORDER BY owner_id, created_at ASC, id ASC
)
UPDATE boards b
SET is_default = TRUE
FROM candidates c
WHERE b.id = c.id;

-- Пользователям вообще без досок создать "Избранное".
INSERT INTO boards (owner_id, title, is_private, is_default)
SELECT u.id, 'Избранное', FALSE, TRUE
FROM users u
WHERE NOT EXISTS (
    SELECT 1
    FROM boards b
    WHERE b.owner_id = u.id
);