-- name: CreateCourse :one
INSERT INTO courses (teacher_id, name, description, join_code)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCourseByID :one
SELECT * FROM courses WHERE id = $1;

-- name: ListCoursesByTeacher :many
SELECT * FROM courses WHERE teacher_id = $1 ORDER BY created_at DESC;

-- name: GetCourseByJoinCode :one
SELECT * FROM courses WHERE join_code = $1;

-- name: UpdateCourse :exec
UPDATE courses 
SET name = $2, description = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND teacher_id = $4;

-- name: DeleteCourse :exec
DELETE FROM courses WHERE id = $1 AND teacher_id = $2;

-- name: ListCoursesForStudent :many
SELECT c.* FROM courses c
JOIN enrollments e ON c.id = e.course_id
WHERE e.user_id = $1
ORDER BY c.created_at DESC;

-- name: GetCourseWithStats :one
SELECT 
    c.*,
    COUNT(DISTINCT e.user_id) as student_count,
    COUNT(DISTINCT a.id) as assignment_count
FROM courses c
LEFT JOIN enrollments e ON c.id = e.course_id
LEFT JOIN assignments a ON c.id = a.course_id
WHERE c.id = $1
GROUP BY c.id;