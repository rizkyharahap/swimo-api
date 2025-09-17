-- USERS: only for persist email
CREATE TABLE IF NOT EXISTS users (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email           citext UNIQUE NOT NULL,
    password_hash   boolean NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- SESSION: save session login & guests, + refresh token
CREATE TABLE IF NOT EXISTS sessions (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid REFERENCES users(id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('guests', 'user')),
    ip inet,
    user_agent text,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL DEFAULT now(),
    refresh_token_hash text,
    refresh_expires_at timestamptz,
    revoked_at timestamptz,
    CONSTRAINT guest_no_user CHECK(
        (kind = 'guests' AND user_id IS NULL) OR
        (kind = 'user' AND user_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_expires ON sessions(refresh_expires_at);

