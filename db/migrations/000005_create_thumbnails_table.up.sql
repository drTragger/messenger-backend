CREATE TABLE thumbnails
(
    id         SERIAL PRIMARY KEY,                     -- Unique identifier for the thumbnail
    file_path  VARCHAR(50) NOT NULL,                   -- Path or URL to the stored thumbnail
    file_type  VARCHAR(50) NOT NULL,                   -- MIME type (e.g., image/png)
    file_size  BIGINT      NOT NULL,                   -- Thumbnail file size in bytes
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Timestamp when the thumbnail was created
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()  -- Timestamp when the thumbnail was created
);

CREATE INDEX idx_thumbnails_file_type ON thumbnails (file_type);