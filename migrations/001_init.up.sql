CREATE TABLE IF NOT EXISTS bots (
    id          VARCHAR     PRIMARY KEY,
    token       VARCHAR     NOT NULL,
    author      BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS nodes (
    bot_id      VARCHAR     NOT NULL,
    state       INTEGER     NOT NULL,
    title       VARCHAR     NOT NULL,

    PRIMARY KEY (bot_id, state),

    FOREIGN KEY (bot_id)
        REFERENCES bots (id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS entries (
    bot_id      VARCHAR     NOT NULL,
    key         VARCHAR     NOT NULL,
    start       INTEGER     NOT NULL,

    PRIMARY KEY (bot_id, key),

    FOREIGN KEY (bot_id)
        REFERENCES  bots (id),

    FOREIGN KEY (bot_id, start)
        REFERENCES nodes (bot_id, state)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS bot_messages (
    id          SERIAL      PRIMARY KEY,
    bot_id      VARCHAR     NOT NULL,
    state       INTEGER     NOT NULL,
    text        TEXT        NOT NULL,

    FOREIGN KEY (bot_id, state)
        REFERENCES nodes (bot_id, state)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS options (
    id           SERIAL     PRIMARY KEY,
    bot_id       VARCHAR    NOT NULL,
    state        INTEGER    NOT NULL,
    text         VARCHAR    NOT NULL,

    FOREIGN KEY (bot_id, state)
        REFERENCES nodes (bot_id, state)
        ON DELETE CASCADE
);

DO $$ BEGIN
    CREATE TYPE PREDICATE_T
    AS ENUM (
        'always',
        'exact',
        'regexp'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE OPERATION
        AS ENUM (
            'noop',
            'save',
            'append'
        );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS edges (
    id          SERIAL      PRIMARY KEY,
    bot_id      VARCHAR     NOT NULL,
    state       INTEGER     NOT NULL,
    to_state    INTEGER     NOT NULL,
    operation   OPERATION   NOT NULL,
    pred_type   PREDICATE_T NOT NULL,
    pred_data   VARCHAR     DEFAULT NULL,

    FOREIGN KEY (bot_id, state)
        REFERENCES nodes (bot_id, state)
        ON DELETE CASCADE
);

CREATE TABLE participants (
    bot_id      VARCHAR     NOT NULL,
    user_id     BIGINT      NOT NULL,
    cthread     VARCHAR     DEFAULT NULL,

    PRIMARY KEY (bot_id, user_id),

    FOREIGN KEY (bot_id)
        REFERENCES bots (id)
        ON DELETE CASCADE
);

CREATE TABLE threads (
    id          VARCHAR     PRIMARY KEY,
    bot_id      VARCHAR     NOT NULL,
    user_id     BIGINT      NOT NULL,
    key         VARCHAR     NOT NULL,
    state       INTEGER     NOT NULL,
    started_at  TIMESTAMPTZ NOT NULL,

    FOREIGN KEY (bot_id, key)
        REFERENCES entries (bot_id, key),

    FOREIGN KEY (bot_id, user_id)
        REFERENCES participants (bot_id, user_id),

    FOREIGN KEY (bot_id, state)
        REFERENCES nodes (bot_id, state)
        ON DELETE CASCADE
);

CREATE TABLE answers (
    thread_id   VARCHAR     NOT NULL,
    state       INTEGER     NOT NULL,
    text        TEXT        DEFAULT NULL,

    PRIMARY KEY (thread_id, state),

    FOREIGN KEY (thread_id)
        REFERENCES  threads (id)
        ON DELETE CASCADE
);


CREATE TABLE usernames (
    user_id     BIGINT      PRIMARY KEY,
    username    VARCHAR(64) NOT NULL
);