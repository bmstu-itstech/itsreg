INSERT INTO scripts
    (id, owner_id, "desc", created_at, updated_at, deleted_at)
VALUES
    ('sc0001', 1, 'Test script sc0001', '2026-04-11 10:00:00', '2026-04-11 10:00:00', NULL),
    ('sc0002', 1, 'Test script sc0002', '2026-04-11 10:00:00', '2026-04-11 11:00:00', NULL),
    ('sc0003', 2, 'Test script sc0003', '2026-04-11 10:00:00', '2026-04-11 10:00:00', NULL),
    ('sc0004', 2, 'Test script sc0003', '2026-04-11 10:00:00', '2026-04-11 10:00:00', '2026-04-11 11:00:00');

INSERT INTO nodes
    (script_id, state, title)
VALUES
    ('sc0001', 1, 'sc0001:1'),
    ('sc0001', 2, 'sc0001:2'),
    ('sc0002', 1, 'sc0002:1'),
    ('sc0003', 1, 'sc0003:1'),
    ('sc0004', 1, 'sc0004:1');

INSERT INTO entries
    (script_id, key, start)
VALUES
    ('sc0001', 'start', 1),
    ('sc0002', 'start', 1),
    ('sc0003', 'start', 1),
    ('sc0004', 'start', 1);

INSERT INTO edges
    (script_id, state, "index", to_state, operation, pred_type, pred_data)
VALUES
    ('sc0001', 1, 1, 2, 'save', 'regex', '^Далее$'),
    ('sc0001', 1, 2, 2, 'append', 'always', ''),
    ('sc0001', 2, 1, 1, 'noop', 'exact', 'Назад');

INSERT INTO messages
    (script_id, state, "index", "text")
VALUES
    ('sc0001', 1, 1, 'Message 1 for sc0001:1'),
    ('sc0001', 1, 2, 'Message 2 for sc0001:1'),
    ('sc0001', 2, 1, 'Message for sc0001:2'),
    ('sc0002', 1, 1, 'Message for sc0002:1'),
    ('sc0003', 1, 1, 'Message for sc0003:1'),
    ('sc0004', 1, 1, 'Message for sc0004:1');

INSERT INTO options
    (script_id, state, "index", "text")
VALUES
    ('sc0001', 1, 1, 'Option 1 for sc0001:1'),
    ('sc0001', 1, 2, 'Option 2 for sc0001:1'),
    ('sc0001', 2, 1, 'Option for sc0001:2');

INSERT INTO bots
    (id, owner_id, script_id, token, "desc", created_at, updated_at, deleted_at)
VALUES
    ('b0001', 1, 'sc0001', 'token_b0001', 'Test bot b0001', '2026-04-11 10:00:00', '2026-04-11 10:00:00', NULL),
    ('b0002', 1, 'sc0002', 'token_b0002', 'Test bot b0002', '2026-04-11 10:00:00', '2026-04-11 11:00:00', NULL),
    ('b0003', 2, 'sc0003', 'token_b0003', 'Test bot b0003', '2026-04-11 10:00:00', '2026-04-11 10:00:00', NULL),
    ('b0004', 2, 'sc0004', 'token_b0004', 'Test bot b0004', '2026-04-11 10:00:00', '2026-04-11 10:00:00', '2026-04-11 11:00:00');

INSERT INTO runs
    (id, bot_id, token, status, error_msg, started_at, stopped_at)
VALUES
    ('r0001', 'b0001', 'token_b0001', 'starting', NULL, '2026-04-11 10:00:00', '2026-04-11 10:30:00'),
    ('r0002', 'b0001', 'token_b0001', 'failed', 'Some error occurred', '2026-04-11 10:30:00', '2026-04-11 10:45:00'),
    ('r0003', 'b0002', 'token_b0002', 'active', NULL, '2026-04-11 10:15:00', NULL),
    ('r0004', 'b0003', 'token_b0003', 'stopped', NULL, '2026-04-11 10:20:00', '2026-04-11 10:50:00');

INSERT INTO users
    (id, username, updated_at)
VALUES
    (1, 'user1', '2026-04-11 10:00:00'),
    (2, 'user2', '2026-04-11 10:00:00');

INSERT INTO threads
    (id, bot_id, user_id, key, state, started_at, updated_at)
VALUES
    ('t0001', 'b0001', 1, 'start', 1, '2026-04-11 10:00:00', '2026-04-11 10:30:00'),
    ('t0002', 'b0001', 2, 'start', 2, '2026-04-11 10:05:00', '2026-04-11 10:35:00'),
    ('t0003', 'b0002', 1, 'start', 1, '2026-04-11 10:10:00', '2026-04-11 10:40:00'),
    ('t0004', 'b0003', 2, 'start', 1, '2026-04-11 10:15:00', '2026-04-11 10:45:00');

INSERT INTO answers
    (thread_id, state, "text")
VALUES
    ('t0001', 1, 'Answer 1 for t0001:1'),
    ('t0001', 2, 'Answer 2 for t0001:2'),
    ('t0002', 1, 'Answer for t0002:1'),
    ('t0003', 1, 'Answer for t0003:1'),
    ('t0004', 1, 'Answer for t0004:1');
