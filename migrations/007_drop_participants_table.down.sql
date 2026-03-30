CREATE TABLE IF NOT EXISTS participants (
    bot_id        VARCHAR NOT NULL
        REFERENCES bots
            ON DELETE CASCADE,
    user_id       BIGINT  NOT NULL,
    active_thread VARCHAR,

    PRIMARY KEY (bot_id, user_id)
);

ALTER TABLE threads
    ADD CONSTRAINT threads_bot_id_user_id_fkey
        FOREIGN KEY (bot_id, user_id) REFERENCES participants (bot_id, user_id)
