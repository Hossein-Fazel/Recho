-- name: CreateMessage :one
INSERT INTO messages (
    conversation_id,
    sender_id,
    type
)
VALUES (
    sqlc.arg(conversation_id),
    sqlc.arg(sender_id),
    sqlc.arg(message_type)
)
RETURNING
    message_id,
    conversation_id,
    created_at,
    updated_at;

-- name: CreateTextMessage :exec
INSERT INTO text_messages (
    conversation_id,
    message_id,
    content
)
VALUES (
    sqlc.arg(conversation_id),
    sqlc.arg(message_id),
    sqlc.arg(content)
);

-- name: UpdateMessage :one
UPDATE messages
SET updated_at = NOW()
WHERE conversation_id = sqlc.arg(conversation_id) AND message_id = sqlc.arg(message_id) AND sender_id = sqlc.arg(user_id)
RETURNING updated_at;

-- name: UpdateTextMessage :exec
UPDATE text_messages AS tm
SET
    content = sqlc.arg(msg_text)
WHERE tm.conversation_id = sqlc.arg(conversation_id)
  AND tm.message_id = sqlc.arg(message_id)
  AND EXISTS (
      SELECT 1
      FROM messages AS m
      WHERE m.conversation_id = tm.conversation_id
        AND m.message_id = tm.message_id
        AND m.sender_id = sqlc.arg(user_id)
        AND m.type = 'text'
  );

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE message_id = sqlc.arg(message_id) and conversation_id = sqlc.arg(conversation_id) and sender_id = sqlc.arg(user_id);
