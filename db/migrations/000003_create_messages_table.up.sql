CREATE TABLE messages
(
    id           SERIAL PRIMARY KEY,
    sender_id    INT                      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    recipient_id INT                      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content      TEXT,
    read_at      TIMESTAMP WITH TIME ZONE NULL,
    chat_id      INT                      NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    parent_id    INT                      REFERENCES messages (id) ON DELETE SET NULL,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);