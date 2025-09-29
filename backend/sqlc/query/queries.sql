-- name: UpsertUser :exec
INSERT INTO users (uuid, username, email, hashed_password)
VALUES ($1, $2, $3, $4)
ON CONFLICT (uuid) DO UPDATE SET
    username = EXCLUDED.username,
    email = EXCLUDED.email,
    hashed_password = EXCLUDED.hashed_password;

-- name: GetUserByUUID :one
SELECT uuid, username, email FROM users WHERE uuid = $1;

-- name: GetUserByUsernameOrEmail :one
SELECT uuid, username, email, hashed_password
FROM users
WHERE username = $1 OR email = $1
LIMIT 1;
