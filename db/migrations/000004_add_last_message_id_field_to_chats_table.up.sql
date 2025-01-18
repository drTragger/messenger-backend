ALTER TABLE chats
    ADD COLUMN last_message_id INT REFERENCES messages (id) ON DELETE SET NULL;