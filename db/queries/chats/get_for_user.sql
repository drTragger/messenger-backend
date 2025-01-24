SELECT c.id,
       c.user1_id,
       c.user2_id,
       c.last_message_id,
       c.created_at,
       c.updated_at,
       u1.id              AS user1_id,
       u1.username        AS user1_username,
       u1.first_name      AS user1_first_name,
       u1.last_name       AS user1_last_name,
       u1.phone           AS user1_phone,
       u1.last_seen       AS user1_last_seen,
       u1.profile_picture AS user1_profile_picture,
       u1.created_at      AS user1_created_at,
       u1.updated_at      AS user1_updated_at,
       u2.id              AS user2_id,
       u2.username        AS user2_username,
       u2.first_name      AS user2_first_name,
       u2.last_name       AS user2_last_name,
       u2.phone           AS user2_phone,
       u2.last_seen       AS user2_last_seen,
       u2.profile_picture AS user2_profile_picture,
       u2.created_at      AS user2_created_at,
       u2.updated_at      AS user2_updated_at,
       m.id               AS message_id,
       m.sender_id        AS last_message_sender_id,
       m.recipient_id     AS last_message_recipient_id,
       LEFT(
               CASE
                   WHEN LENGTH(m.content) > $4 THEN CONCAT(SUBSTRING(m.content, 1, $4), '...')
                   ELSE m.content
                   END,
               $5
       )                  AS message_content_trimmed,
       m.read_at          AS last_message_read_at,
       m.chat_id          AS last_message_chat_id,
       m.created_at       AS last_message_created_at,
       m.updated_at       AS last_message_updated_at,
       a.id               AS attachment_id,
       a.file_name        AS attachment_file_name,
       a.file_path        AS attachment_file_path,
       a.file_type        AS attachment_file_type,
       a.file_size        AS attachment_file_size,
       a.created_at       AS attachment_created_at,
       a.updated_at       AS attachment_updated_at
FROM chats c
         LEFT JOIN users u1 ON c.user1_id = u1.id
         LEFT JOIN users u2 ON c.user2_id = u2.id
         LEFT JOIN messages m ON c.last_message_id = m.id
         LEFT JOIN LATERAL (
    SELECT id, file_name, file_path, file_type, file_size, created_at, updated_at
    FROM attachments
    WHERE message_id = m.id
    ORDER BY created_at DESC
    LIMIT 1
    ) a ON true
WHERE c.user1_id = $1
   OR c.user2_id = $1
ORDER BY c.updated_at DESC
LIMIT $2 OFFSET $3