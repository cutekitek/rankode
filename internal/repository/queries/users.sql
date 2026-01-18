-- name: GetUsersByEmailOrUsername :one
SELECT * FROM users WHERE username = $1 OR email = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: NewUser :one
INSERT INTO users (username, email, password_hash, elo) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: BanUser :exec
UPDATE users SET roles = -1 WHERE id = $1;

-- name: IncreaseUserElo :exec
UPDATE users SET elo = elo + $1 WHERE id = $2;

-- name: GetUsersLeaderboard :many
SELECT username, elo from users ORDER BY elo desc limit $1;

-- name: UpdateUserRole :exec
UPDATE users SET roles = $1 WHERE id = $2;
