CREATE TABLE IF NOT EXISTS users
(
    id          BIGINT      PRIMARY KEY,
    username    VARCHAR     NOT NULL,
    updated_at  TIMESTAMPTZ DEFAULT now()
)
