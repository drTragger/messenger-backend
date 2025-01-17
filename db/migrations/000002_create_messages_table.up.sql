CREATE TABLE messages
(
    id           SERIAL PRIMARY KEY,
    sender_id    INT                      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    recipient_id INT                      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content      TEXT                     NOT NULL,
    read_at      TIMESTAMP WITH TIME ZONE NULL,
    message_type VARCHAR(50)                       DEFAULT 'text', -- 'text', 'image', 'file', etc.
    chat_id      INT                      NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);