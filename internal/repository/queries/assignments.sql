-- name: CreateAssignment :one
INSERT INTO assignments (course_id, title, description, start_date, due_date, max_attempts_per_task, group_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAssignmentByID :one
SELECT * FROM assignments WHERE id = $1;

-- name: ListAssignmentsByCourse :many
SELECT * FROM assignments WHERE course_id = $1 ORDER BY created_at DESC;

-- name: UpdateAssignment :exec
UPDATE assignments 
SET title = $2, description = $3, start_date = $4, due_date = $5, 
    max_attempts_per_task = $6, group_id = $7, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;


-- name: DeleteAssignment :exec
DELETE FROM assignments WHERE id = $1;

-- name: GetAssignmentWithTasks :one
SELECT 
    a.*,
    json_agg(
        json_build_object(
            'task_id', at.task_id,
            'order_index', at.order_index,
            'weight', at.weight
        ) ORDER BY at.order_index
    ) as tasks
FROM assignments a
LEFT JOIN assignment_tasks at ON a.id = at.assignment_id
WHERE a.id = $1
GROUP BY a.id;

-- name: ListAssignmentsForStudent :many
SELECT a.* FROM assignments a
JOIN enrollments e ON a.course_id = e.course_id
LEFT JOIN group_students gs ON a.group_id = gs.group_id AND e.user_id = gs.user_id
WHERE e.user_id = $1 
  AND (a.start_date IS NULL OR a.start_date <= CURRENT_TIMESTAMP)
  AND (a.group_id IS NULL OR gs.user_id IS NOT NULL)
ORDER BY a.due_date ASC NULLS LAST, a.created_at DESC;

-- name: CountAssignmentSubmissions :one
SELECT COUNT(*) FROM attempts WHERE assignment_id = $1::int;

-- name: CountStudentsWhoStartedAssignment :one
SELECT COUNT(DISTINCT user_id) FROM attempts WHERE assignment_id = $1::int;

-- name: CountStudentsWhoFinishedAssignment :one
SELECT COUNT(*) 
FROM (
    SELECT a.user_id
    FROM attempts a
    WHERE a.assignment_id = $1::int AND a.attempt_status = 0
    GROUP BY a.user_id
    HAVING COUNT(DISTINCT a.task_id) >= (
        SELECT COUNT(*) FROM assignment_tasks WHERE assignment_id = $1::int
    )
) AS finished_students;
