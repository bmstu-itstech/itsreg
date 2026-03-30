ALTER TABLE threads
    DROP CONSTRAINT IF EXISTS threads_bot_id_user_id_fkey;

DROP TABLE IF EXISTS participants;
