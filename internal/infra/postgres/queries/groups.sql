-- name: InsertGroup :one
INSERT INTO groups (
    conversation_id,
    name,
    bio,
    invite_code,
    created_by
)
VALUES (
    sqlc.arg(conversation_id),
    sqlc.arg(name),
    sqlc.arg(bio),
    sqlc.arg(invite_code),
    sqlc.arg(created_by)
)
RETURNING
    conversation_id,
    name,
    avatar_url,
    invite_code,
    bio,
    created_by,
    created_at,
    updated_at;

-- name: InsertGroupMember :exec
INSERT INTO group_members (
    group_id,
    user_id,
    role
)
VALUES (
    sqlc.arg(group_id),
    sqlc.arg(user_id),
    sqlc.arg(role)
)
ON CONFLICT (group_id, user_id) DO NOTHING;

-- name: GetGroupByInviteCode :one
SELECT
    g.conversation_id,
    g.name,
    g.avatar_url,
    g.invite_code,
    g.bio,
    g.created_by,
    g.created_at,
    g.updated_at,
    (
        SELECT COUNT(*)
        FROM group_members gm
        WHERE gm.group_id = g.conversation_id
    ) AS member_count
FROM groups g
WHERE g.invite_code = sqlc.arg(invite_code);

-- name: GetGroupMemberRole :one
SELECT role
FROM group_members
WHERE group_id = sqlc.arg(group_id)
  AND user_id = sqlc.arg(user_id);

-- name: CountGroupMembers :one
SELECT COUNT(*) AS member_count
FROM group_members
WHERE group_id = sqlc.arg(group_id);

-- name: RemoveGroupMember :exec
DELETE FROM group_members
WHERE group_id = sqlc.arg(group_id)
  AND user_id = sqlc.arg(user_id);

-- name: DeleteGroup :exec
DELETE FROM conversations
WHERE id = sqlc.arg(group_id);
