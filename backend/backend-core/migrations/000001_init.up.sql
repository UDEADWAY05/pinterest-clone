CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE pins (
    id BIGSERIAL PRIMARY KEY,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL,
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    dominant_color TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE boards (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE board_pins (
    board_id BIGINT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    pin_id BIGINT NOT NULL REFERENCES pins(id) ON DELETE CASCADE,
    saved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (board_id, pin_id)
);

CREATE TABLE interactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    session_id TEXT NOT NULL,
    pin_id BIGINT NOT NULL REFERENCES pins(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    meta_json JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_pins_author_id ON pins(author_id);
CREATE INDEX idx_pins_created_at ON pins(created_at DESC);
CREATE INDEX idx_boards_owner_id ON boards(owner_id);
CREATE INDEX idx_interactions_pin_id ON interactions(pin_id);
CREATE INDEX idx_interactions_user_id ON interactions(user_id);
CREATE INDEX idx_interactions_session_id ON interactions(session_id);
