-- name: AddTaskToAssignment :exec
INSERT INTO assignment_tasks (assignment_id, task_id, order_index, weight)
VALUES ($1, $2, $3, $4)
ON CONFLICT (assignment_id, task_id) DO UPDATE
SET order_index = $3, weight = $4;

-- name: RemoveTaskFromAssignment :exec
DELETE FROM assignment_tasks WHERE assignment_id = $1 AND task_id = $2;

-- name: GetTasksForAssignment :many
SELECT t.*, at.order_index, at.weight
FROM tasks t
JOIN assignment_tasks at ON t.id = at.task_id
WHERE at.assignment_id = $1
ORDER BY at.order_index;

-- name: GetAssignmentForTask :one
SELECT a.* FROM assignments a
JOIN assignment_tasks at ON a.id = at.assignment_id
WHERE at.task_id = $1
LIMIT 1;

-- name: UpdateTaskOrderAndWeight :exec
UPDATE assignment_tasks
SET order_index = $3, weight = $4
WHERE assignment_id = $1 AND task_id = $2;

-- name: CountTasksInAssignment :one
SELECT COUNT(*) FROM assignment_tasks WHERE assignment_id = $1;