CREATE EXTENSION IF NOT EXISTS vector;

ALTER TABLE pins ADD COLUMN IF NOT EXISTS embedding vector(512);

CREATE INDEX IF NOT EXISTS idx_pins_embedding ON pins
USING hnsw (embedding vector_cosine_ops);
