-- name: CreateMessage :one
INSERT INTO messages (
    conversation_id,
    sender_id,
    content
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING message_id, created_at, updated_at;

-- name: UpdateMessage :one
UPDATE messages
SET content = sqlc.arg(msgText),
    updated_at = NOW()
WHERE message_id = sqlc.arg(message_id) and conversation_id = sqlc.arg(conversation_id) and sender_id = sqlc.arg(user_id)
RETURNING updated_at;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE message_id = sqlc.arg(message_id) and conversation_id = sqlc.arg(conversation_id) and sender_id = sqlc.arg(user_id);