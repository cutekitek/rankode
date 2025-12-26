-- name: CreateGroup :one
INSERT INTO groups (course_id, name)
VALUES ($1, $2)
RETURNING *;

-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = $1;

-- name: ListGroupsByCourse :many
SELECT * FROM groups WHERE course_id = $1 ORDER BY name ASC;

-- name: DeleteGroup :exec
DELETE FROM groups WHERE id = $1;

-- name: AddStudentToGroup :exec
INSERT INTO group_students (group_id, user_id)
VALUES ($1, $2)
ON CONFLICT (group_id, user_id) DO NOTHING;

-- name: RemoveStudentFromGroup :exec
DELETE FROM group_students WHERE group_id = $1 AND user_id = $2;

-- name: ListStudentsInGroup :many
SELECT u.* FROM users u
JOIN group_students gs ON u.id = gs.user_id
WHERE gs.group_id = $1
ORDER BY u.username;

-- name: GetGroupsForStudent :many
SELECT g.* FROM groups g
JOIN group_students gs ON g.id = gs.group_id
WHERE gs.user_id = $1 AND g.course_id = $2;
