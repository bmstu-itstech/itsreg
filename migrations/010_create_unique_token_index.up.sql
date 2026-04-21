CREATE UNIQUE INDEX IF NOT EXISTS uniq_bots_token
    ON bots (token)
    WHERE deleted_at IS NULL;
