CREATE TABLE chats
(
    id              SERIAL PRIMARY KEY,
    user1_id        INT                      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    user2_id        INT                      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user1_id, user2_id)
);