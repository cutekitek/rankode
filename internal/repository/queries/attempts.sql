-- name: CreateAttempt :one
INSERT INTO attempts(user_id, task_id, language, code, attempt_status, assignment_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetUserAttemptsByTask :many
SELECT * from attempts WHERE user_id = $1 and task_id = $2 AND (assignment_id = $3 OR $3 IS NULL) ORDER BY updated_at DESC; 

-- name: GetAttemptById :one
SELECT * FROM attempts WHERE id = $1;

-- name: UpdateAttemptStatus :exec
UPDATE attempts SET attempt_status = $1, running_time = $2, memory = $3, error = $4 WHERE id = $5;

-- name: CheckFirstSuccessfulAttempt :one
SELECT EXISTS(SELECT * from attempts where user_id = $1 and task_id = $2 and attempt_status = $3 AND (assignment_id = $4 OR $4 IS NULL));
