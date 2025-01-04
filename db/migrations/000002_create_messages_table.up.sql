CREATE TABLE messages
(
    id           SERIAL PRIMARY KEY,
    sender_id    INT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    recipient_id INT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content      TEXT      NOT NULL,
    read_at      TIMESTAMP NULL,
    message_type VARCHAR(50) DEFAULT 'text', -- 'text', 'image', 'file', etc.
    created_at   TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
);