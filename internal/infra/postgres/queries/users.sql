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
    display_name,
    avatar_url,
    bio,
    created_at,
    updated_at;


-- name: GetUserByUsername :one
SELECT
    id,
    username,
    password_hash,
    display_name,
    avatar_url,
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
    avatar_url,
    bio,
    created_at,
    updated_at
FROM users
WHERE id = $1
LIMIT 1;


-- name: UsernameExists :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE username = $1
);

-- name: GetUserForLogin :one
SELECT
    id,
    username,
    password_hash
FROM users
WHERE username = $1
LIMIT 1;