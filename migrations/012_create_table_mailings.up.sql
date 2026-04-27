CREATE TYPE mailing_status AS ENUM
(
    'scheduled',
    'started',
    'completed',
    'failed'
);

CREATE TABLE IF NOT EXISTS mailings
(
    id              VARCHAR(6)      PRIMARY KEY,
    name            VARCHAR         NOT NULL,
    bot_id          VARCHAR         NOT NULL,
    entry_key       VARCHAR         NOT NULL,
    status          MAILING_STATUS  NOT NULL,
    created_at      TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    started_at      TIMESTAMPTZ                 DEFAULT NULL,
    completed_at    TIMESTAMPTZ                 DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS mailing_recipients
(
    mailing_id VARCHAR(6) NOT NULL,
    user_id    BIGINT     NOT NULL,

    PRIMARY KEY (mailing_id, user_id)
);

CREATE TABLE IF NOT EXISTS mailing_results
(
    mailing_id VARCHAR(6) NOT NULL,
    user_id    BIGINT     NOT NULL,
    success    BOOLEAN    NOT NULL,
    error_msg  VARCHAR              DEFAULT NULL,

    PRIMARY KEY (mailing_id, user_id)
);


ALTER TABLE mailing_recipients
    ADD CONSTRAINT fk_mailing_recipients_mailing_id_mailings
        FOREIGN KEY (mailing_id) REFERENCES mailings (id)
            ON DELETE CASCADE,
    ADD CONSTRAINT fk_mailing_recipients_user_id_users
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE;

ALTER TABLE mailing_results
    ADD CONSTRAINT fk_mailing_results_mailing_id_mailings
        FOREIGN KEY (mailing_id) REFERENCES mailings (id)
            ON DELETE CASCADE,
    ADD CONSTRAINT fk_mailing_results_user_id_users
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE,
    ADD CONSTRAINT fk_mailing_results_mailing_id_user_id_mailing_recipients
        FOREIGN KEY (mailing_id, user_id) REFERENCES mailing_recipients (mailing_id, user_id)
            ON DELETE CASCADE;

ALTER TABLE mailings
    ADD CONSTRAINT fk_mailings_bot_id_bots
        FOREIGN KEY (bot_id) REFERENCES bots (id)
            ON DELETE CASCADE;
