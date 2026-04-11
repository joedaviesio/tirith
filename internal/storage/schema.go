package storage

const createTableSQL = `
CREATE TABLE IF NOT EXISTS api_calls (
    id                 TEXT PRIMARY KEY,
    timestamp          DATETIME NOT NULL,
    provider           TEXT NOT NULL,
    model              TEXT NOT NULL,

    input_tokens       INTEGER NOT NULL,
    output_tokens      INTEGER NOT NULL,
    cache_read_tokens  INTEGER DEFAULT 0,
    cache_write_tokens INTEGER DEFAULT 0,

    cost_cents         INTEGER NOT NULL,

    latency_ms         INTEGER NOT NULL,
    status_code        INTEGER NOT NULL,
    streaming          BOOLEAN DEFAULT FALSE,

    tag                TEXT,
    user_tag           TEXT,
    session_tag        TEXT,
    environment        TEXT DEFAULT 'default',

    endpoint           TEXT
);

CREATE INDEX IF NOT EXISTS idx_timestamp ON api_calls(timestamp);
CREATE INDEX IF NOT EXISTS idx_tag ON api_calls(tag);
CREATE INDEX IF NOT EXISTS idx_user_tag ON api_calls(user_tag);
CREATE INDEX IF NOT EXISTS idx_model ON api_calls(model);
CREATE INDEX IF NOT EXISTS idx_environment ON api_calls(environment);
CREATE INDEX IF NOT EXISTS idx_provider ON api_calls(provider);
`
