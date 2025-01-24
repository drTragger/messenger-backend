CREATE TABLE attachments
(
    id           SERIAL PRIMARY KEY,                     -- Unique identifier for the attachment
    message_id   INT          NOT NULL,                  -- Foreign key referencing messages
    file_name    VARCHAR(255) NOT NULL,                  -- Original file name
    file_path    varchar(50)  NOT NULL,                  -- Path or URL to the stored file
    file_type    VARCHAR(100) NOT NULL,                  -- MIME type (e.g., image/png, video/mp4)
    file_size    BIGINT       NOT NULL,                  -- File size in bytes
    thumbnail_id INT,                                    -- Foreign key referencing thumbnails (nullable)
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Timestamp when the attachment was created
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Timestamp when the attachment was updated
    CONSTRAINT fk_message FOREIGN KEY (message_id) REFERENCES messages (id) ON DELETE CASCADE,
    CONSTRAINT fk_thumbnail FOREIGN KEY (thumbnail_id) REFERENCES thumbnails (id) ON DELETE SET NULL
);

CREATE INDEX idx_attachments_message_id ON attachments (message_id);