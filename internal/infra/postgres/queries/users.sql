-- name: CreateUser :one
INSERT INTO users (
    username,
    password_hash
)
VALUES (
    $1,
    $2
)
RETURNING
    id,
    username,
    password_hash,
    display_name,
    avatar_key,
    bio,
    created_at,
    updated_at;


-- name: GetUserByUsername :one
SELECT
    id,
    username,
    password_hash,
    display_name,
    avatar_key,
    bio,
    created_at,
    updated_at
FROM users
WHERE username = $1
LIMIT 1;


-- name: GetUserByID :one
SELECT
    id,
    username,
    password_hash,
    display_name,
    avatar_key,
    bio,
    created_at,
    updated_at
FROM users
WHERE id = $1
LIMIT 1;


-- name: UpdateUser :one
UPDATE users
SET
    username     = $2,
    display_name = $3,
    avatar_key   = $4,
    bio          = $5,
    updated_at   = NOW()
WHERE id = $1
RETURNING
    id,
    username,
    password_hash,
    display_name,
    avatar_key,
    bio,
    created_at,
    updated_at;


-- name: UsernameExists :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE username = $1
);

-- name: SearchUsers :many
SELECT
    id,
    username,
    display_name,
    avatar_key
FROM users
WHERE username LIKE sqlc.arg(query) || '%'
ORDER BY username
LIMIT 20;