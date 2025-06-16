-- name: CreateTask :one
INSERT INTO tasks (user_id, title, description, difficulty, passes, score, topics)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateTask :exec
UPDATE tasks
SET title = $2, description = $3, difficulty = $4, passes = $5, score = $6, topics = $7, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;

-- name: ListTasks :many
SELECT * FROM tasks;

-- name: GetTaskById :one
SELECT * FROM tasks where id = $1;

-- name: IncreaseTaskPasses :exec
UPDATE tasks SET passes = passes + 1 WHERE id = $1;