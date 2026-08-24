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