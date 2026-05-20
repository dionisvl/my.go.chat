-- +goose Up
-- +goose StatementBegin
CREATE TABLE messages (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username   TEXT NOT NULL,
    message    TEXT NOT NULL,
    color      TEXT NOT NULL DEFAULT '#FF00FF',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_created_at ON messages (created_at);

INSERT INTO messages (username, message) VALUES
    ('user1', 'Hello, world!'),
    ('user2', 'How are you?');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS messages;
-- +goose StatementEnd
