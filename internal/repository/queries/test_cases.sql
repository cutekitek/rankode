-- name: CreateTaskTestCase :one
INSERT INTO task_test_cases(task_id, case_order, input_file, output_file) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetTaskTestCases :many
SELECT * from task_test_cases WHERE  task_id = $1 ORDER BY case_order;

-- name: GetTestCaseByID :one
SELECT * from task_test_cases WHERE id = $1;

-- name: DeleteTaskTestCase :exec
UPDATE attempts SET attempt_status = $1, running_time = $2, memory = $3, error = $4 WHERE id = $5;