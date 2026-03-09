ALTER TABLE entries DROP CONSTRAINT entries_bot_id_fkey;
ALTER TABLE entries
    ADD CONSTRAINT entries_bot_id_fkey
        FOREIGN KEY (bot_id) REFERENCES bots (id);
