-- name: EnrollStudent :exec
INSERT INTO enrollments (course_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (course_id, user_id) DO UPDATE
SET role = $3;

-- name: UnenrollStudent :exec
DELETE FROM enrollments WHERE course_id = $1 AND user_id = $2;

-- name: GetEnrollment :one
SELECT * FROM enrollments WHERE course_id = $1 AND user_id = $2;

-- name: ListStudentsInCourse :many
SELECT u.*, e.enrolled_at, e.role as enrollment_role
FROM users u
JOIN enrollments e ON u.id = e.user_id
WHERE e.course_id = $1
ORDER BY e.enrolled_at DESC;

-- name: CheckIfEnrolled :one
SELECT EXISTS(SELECT 1 FROM enrollments WHERE course_id = $1 AND user_id = $2);

-- name: CountStudentsInCourse :one
SELECT COUNT(*) FROM enrollments WHERE course_id = $1;