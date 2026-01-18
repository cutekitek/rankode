-- name: CreateTask :one
INSERT INTO tasks (user_id, title, description, difficulty, passes, score, topics, course_id, is_public, verification_file)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateTask :exec
UPDATE tasks
SET title = $2, description = $3, difficulty = $4, passes = $5, score = $6, topics = $7, course_id = $8, is_public = $9, verification_file = $10, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateTaskVerificationFile :exec
UPDATE tasks SET verification_file = $2 WHERE id = $1;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;

-- name: ListTasks :many
SELECT * FROM tasks;

-- name: GetTaskById :one
SELECT * FROM tasks where id = $1;

-- name: IncreaseTaskPasses :exec
UPDATE tasks SET passes = passes + 1 WHERE id = $1;

-- name: ListTasksByCourse :many
SELECT * FROM tasks WHERE course_id = $1 ORDER BY created_at DESC;

-- name: ListPublicTasks :many
SELECT * FROM tasks WHERE is_public = true ORDER BY created_at DESC;

-- name: UpdateTaskCourseAndVisibility :exec
UPDATE tasks SET course_id = $2, is_public = $3 WHERE id = $1;