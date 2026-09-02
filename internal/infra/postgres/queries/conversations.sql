-- name: GetUserConversations :many

SELECT
    c.id AS conversation_id,

    CASE
        WHEN dc.conversation_id IS NOT NULL THEN 'direct'
        ELSE 'group'
    END AS conversation_type,

    -- direct user
    u.id AS user_id,
    u.username,
    u.display_name,
    u.avatar_url,

    -- group
    g.name AS group_name,
    g.avatar_url AS group_avatar_url,

    -- last message
    c.last_message_id,
    c.last_message_content,
    c.last_message_created_at,

    c.updated_at

FROM conversations c

LEFT JOIN direct_conversations dc
    ON dc.conversation_id = c.id

LEFT JOIN users u
    ON u.id = CASE
        WHEN dc.user_one_id = sqlc.arg(user_id) THEN dc.user_two_id
        ELSE dc.user_one_id
    END

LEFT JOIN groups g
    ON g.conversation_id = c.id

LEFT JOIN group_members gm
    ON gm.group_id = c.id
    AND gm.user_id = sqlc.arg(user_id)

WHERE
    (
        dc.user_one_id = sqlc.arg(user_id)
        OR dc.user_two_id = sqlc.arg(user_id)
        OR gm.user_id IS NOT NULL
    )
    AND (
        sqlc.narg(cursor_updated_at)::timestamptz IS NULL
        OR c.updated_at < sqlc.narg(cursor_updated_at)
        OR (
            c.updated_at = sqlc.narg(cursor_updated_at)
            AND c.id < sqlc.narg(cursor_id)
        )
    )

ORDER BY
    c.updated_at DESC,
    c.id DESC

LIMIT sqlc.arg(query_limit);

-- name: GetConversationByID :one

SELECT
    c.id AS conversation_id,

    CASE
        WHEN dc.conversation_id IS NOT NULL THEN 'direct'
        ELSE 'group'
    END AS conversation_type,

    -- direct user
    u.id AS user_id,
    u.username,
    u.display_name,
    u.avatar_url,

    -- group
    g.name AS group_name,
    g.avatar_url AS group_avatar_url,

    -- last message
    c.last_message_id,
    c.last_message_content,
    c.last_message_created_at,

    c.updated_at

FROM conversations c

LEFT JOIN direct_conversations dc
    ON dc.conversation_id = c.id

LEFT JOIN users u
    ON u.id = CASE
        WHEN dc.user_one_id = sqlc.arg(user_id) THEN dc.user_two_id
        ELSE dc.user_one_id
    END

LEFT JOIN groups g
    ON g.conversation_id = c.id

LEFT JOIN group_members gm
    ON gm.group_id = c.id
    AND gm.user_id = sqlc.arg(user_id)

WHERE
    c.id = sqlc.arg(conversation_id)
    AND (
        dc.user_one_id = sqlc.arg(user_id)
        OR dc.user_two_id = sqlc.arg(user_id)
        OR gm.user_id IS NOT NULL
    );

-- name: GetDirectConversation :one
SELECT conversation_id
FROM direct_conversations
WHERE user_one_id = sqlc.arg(user_one_id)
AND user_two_id = sqlc.arg(user_two_id);

-- name: FindDirectConversation :one
SELECT conversation_id
FROM direct_conversations
WHERE user_one_id = $1
AND user_two_id = $2;

-- name: InsertConversation :one
INSERT INTO conversations(message_id_counter)
VALUES (0)
RETURNING id;

-- name: InsertDirectConversation :one
INSERT INTO direct_conversations(
    conversation_id,
    user_one_id,
    user_two_id
)
VALUES(
    $1,
    $2,
    $3
)
RETURNING conversation_id;

-- name: GetConversationMessages :many
SELECT
    message_id,
    conversation_id,
    sender_id,
    content,
    created_at,
    updated_at
FROM messages
WHERE conversation_id = sqlc.arg(conversation_id)
  AND (
      sqlc.arg(cursor_created_at)::timestamptz IS NULL
      OR (created_at, message_id) < (
          sqlc.narg(cursor_created_at)::timestamptz,
          sqlc.narg(cursor_id)::BIGINT
      )
  )
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit);

-- name: IsConversationMember :one
SELECT EXISTS (
    SELECT 1
    FROM direct_conversations dc
    WHERE dc.conversation_id = sqlc.arg(conversation_id) AND (dc.user_one_id = sqlc.arg(user_id) OR dc.user_two_id = sqlc.arg(user_id))

    UNION ALL

    SELECT 1
    FROM group_members gm
    WHERE gm.group_id = sqlc.arg(conversation_id) AND gm.user_id = sqlc.arg(user_id)
);

-- name: GetConvUsers :many
SELECT user_one_id AS user_id
FROM direct_conversations dc
WHERE dc.conversation_id = sqlc.arg(conversation_id)

UNION ALL

SELECT user_two_id AS user_id
FROM direct_conversations dc
WHERE dc.conversation_id = sqlc.arg(conversation_id)

UNION ALL

SELECT user_id
FROM group_members
WHERE group_id = sqlc.arg(conversation_id);