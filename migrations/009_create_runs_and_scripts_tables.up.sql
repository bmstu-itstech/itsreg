ALTER SCHEMA public
    RENAME TO old;

CREATE SCHEMA IF NOT EXISTS public;

ALTER TABLE IF EXISTS old.schema_migrations
    SET SCHEMA public;


CREATE TYPE run_status AS ENUM
(
    'starting',
    'active',
    'stopped',
    'failed'
);

CREATE TYPE edge_operation AS ENUM
(
    'noop',
    'save',
    'append'
);

CREATE TYPE edge_pred_type AS ENUM
(
    'always',
    'exact',
    'regex'
);



CREATE TABLE IF NOT EXISTS scripts
(
    id          VARCHAR(6)              PRIMARY KEY,
    owner_id    BIGINT      NOT NULL,
    "desc"      VARCHAR     NOT NULL    DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL    DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL    DEFAULT now(),
    deleted_at  TIMESTAMPTZ             DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS nodes
(
    script_id   VARCHAR(6)  NOT NULL,
    state       INTEGER     NOT NULL,
    title       VARCHAR     NOT NULL,

    PRIMARY KEY (script_id, state)
);

CREATE TABLE IF NOT EXISTS entries
(
    script_id   VARCHAR(6)  NOT NULL,
    key         VARCHAR     NOT NULL,
    start       INTEGER     NOT NULL,

    PRIMARY KEY (script_id, key)
);

CREATE TABLE IF NOT EXISTS edges
(
    script_id   VARCHAR(6)      NOT NULL,
    state       INTEGER         NOT NULL,
    index       INTEGER         NOT NULL,
    to_state    INTEGER         NOT NULL,
    operation   EDGE_OPERATION  NOT NULL,
    pred_type   EDGE_PRED_TYPE  NOT NULL,
    pred_data   VARCHAR         NOT NULL,

    PRIMARY KEY (script_id, state, index)
);

CREATE TABLE IF NOT EXISTS messages
(
    script_id   VARCHAR(6)  NOT NULL,
    state       INTEGER     NOT NULL,
    index       INTEGER     NOT NULL,
    text        VARCHAR     NOT NULL,

    PRIMARY KEY (script_id, state, index)
);

CREATE TABLE IF NOT EXISTS options
(
    script_id   VARCHAR(6)  NOT NULL,
    state       INTEGER     NOT NULL,
    index       INTEGER     NOT NULL,
    text        VARCHAR     NOT NULL,

    PRIMARY KEY (script_id, state, index)
);

CREATE TABLE IF NOT EXISTS bots
(
    id          VARCHAR(6)              PRIMARY KEY,
    owner_id    BIGINT      NOT NULL,
    script_id   VARCHAR(8)  NOT NULL,
    token       VARCHAR     NOT NULL,
    "desc"      VARCHAR     NOT NULL    DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL    DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL    DEFAULT now(),
    deleted_at  TIMESTAMPTZ             DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS runs
(
    id          VARCHAR(6)              PRIMARY KEY,
    bot_id      VARCHAR(6)  NOT NULL,
    token       VARCHAR     NOT NULL,
    status      RUN_STATUS  NOT NULL    DEFAULT 'starting',
    error_msg   VARCHAR                 DEFAULT NULL,
    started_at  TIMESTAMPTZ             DEFAULT NULL,
    stopped_at  TIMESTAMPTZ             DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS users
(
    id          BIGINT                  PRIMARY KEY,
    username    VARCHAR     NOT NULL,
    updated_at  TIMESTAMPTZ             DEFAULT now()
);

CREATE TABLE IF NOT EXISTS threads
(
    id          VARCHAR(8)              PRIMARY KEY,
    bot_id      VARCHAR(6)  NOT NULL,
    user_id     BIGINT      NOT NULL,
    key         VARCHAR     NOT NULL,
    state       INTEGER     NOT NULL,
    started_at  TIMESTAMPTZ NOT NULL    DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL    DEFAULT now()
);

CREATE TABLE IF NOT EXISTS answers
(
    thread_id   VARCHAR(8)  NOT NULL,
    state       INTEGER     NOT NULL,
    text        VARCHAR     NOT NULL,

    PRIMARY KEY (thread_id, state)
);



ALTER TABLE nodes
    ADD CONSTRAINT fk_nodes_script_id_scripts
        FOREIGN KEY (script_id) REFERENCES scripts (id)
            ON DELETE CASCADE;

ALTER TABLE entries
    ADD CONSTRAINT fk_entries_script_id_state_nodes
        FOREIGN KEY (script_id, start) REFERENCES nodes (script_id, state)
            ON DELETE CASCADE;

ALTER TABLE edges
    ADD CONSTRAINT fk_edges_script_id_state_nodes
        FOREIGN KEY (script_id, state) REFERENCES nodes (script_id, state)
            ON DELETE CASCADE,
    ADD CONSTRAINT fk_edges_script_id_to_state_nodes
        FOREIGN KEY (script_id, to_state) REFERENCES nodes (script_id, state)
            ON DELETE CASCADE;

ALTER TABLE messages
    ADD CONSTRAINT fk_messages_bot_id_state_nodes
        FOREIGN KEY (script_id, state) REFERENCES nodes (script_id, state)
            ON DELETE CASCADE;

ALTER TABLE options
    ADD CONSTRAINT fk_options_script_id_state_nodes
        FOREIGN KEY (script_id, state) REFERENCES nodes (script_id, state)
            ON DELETE CASCADE;

ALTER TABLE bots
    ADD CONSTRAINT fk_bots_script_id_scripts
        FOREIGN KEY (script_id) REFERENCES scripts (id)
            ON DELETE CASCADE;

ALTER TABLE runs
    ADD CONSTRAINT fk_runs_bot_id_bots
        FOREIGN KEY (bot_id) REFERENCES bots (id)
            ON DELETE CASCADE;

ALTER TABLE threads
    ADD CONSTRAINT fk_threads_user_id_users
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE;

ALTER TABLE answers
    ADD CONSTRAINT fk_answers_thread_id_threads
        FOREIGN KEY (thread_id) REFERENCES threads (id)
            ON DELETE CASCADE;



CREATE UNIQUE INDEX uniq_active_run_per_bot
    ON runs (bot_id)
    WHERE status IN ('starting', 'active');
