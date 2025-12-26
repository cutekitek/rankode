-- name: CreateOrUpdateGrade :one
INSERT INTO grades (assignment_id, task_id, user_id, grade, feedback, graded_by)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (assignment_id, task_id, user_id) DO UPDATE
SET grade = $4, feedback = $5, graded_by = $6, graded_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetGrade :one
SELECT * FROM grades WHERE assignment_id = $1 AND task_id = $2 AND user_id = $3;

-- name: GetGradesForAssignment :many
SELECT g.*, u.username as student_username, t.title as task_title
FROM grades g
JOIN users u ON g.user_id = u.id
JOIN tasks t ON g.task_id = t.id
WHERE g.assignment_id = $1
ORDER BY u.username, t.title;

-- name: GetStudentGradesForAssignment :many
SELECT g.*, t.title as task_title
FROM grades g
JOIN tasks t ON g.task_id = t.id
WHERE g.assignment_id = $1 AND g.user_id = $2
ORDER BY t.title;

-- name: GetStudentGradesForCourse :many
SELECT g.*, a.title as assignment_title, t.title as task_title
FROM grades g
JOIN assignments a ON g.assignment_id = a.id
JOIN tasks t ON g.task_id = t.id
WHERE a.course_id = $1 AND g.user_id = $2
ORDER BY a.due_date, t.title;

-- name: DeleteGrade :exec
DELETE FROM grades WHERE id = $1;

-- name: GetAverageGradeForAssignment :one
SELECT AVG(grade)::numeric(3,2) as average_grade
FROM grades
WHERE assignment_id = $1;