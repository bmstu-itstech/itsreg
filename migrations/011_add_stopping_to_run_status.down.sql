UPDATE runs
SET status = 'active'
WHERE status = 'stopping';

ALTER TYPE run_status RENAME TO run_status_old;

CREATE TYPE run_status AS ENUM
(
    'starting',
    'active',
    'stopped',
    'failed'
);

ALTER TABLE runs
    ALTER COLUMN status TYPE run_status
        USING status::text::run_status;

DROP TYPE run_status_old;

