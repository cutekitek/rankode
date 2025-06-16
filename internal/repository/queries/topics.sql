-- name: CreateTopic :one
INSERT INTO topics (name)
VALUES ($1) RETURNING *;

-- name: DeleteTopic :exec
DELETE FROM topics
WHERE name = $1;

-- name: ListTopics :many
SELECT * from topics ORDER BY tasks_count DESC;

-- name: ListTopicsByIDs :many
SELECT * from topics WHERE id = ANY(sqlc.arg(topic_ids)::int[]);

-- name: TopicsByName :many
SELECT * FROM topics WHERE name ILIKE '%' + sqlc.arg(name)::varchar + '%' ORDER BY tasks_count DESC;

-- name: UpdateTopicsCounters :exec
UPDATE topics SET tasks_count = tasks_count + sqlc.arg(diff) WHERE id = ANY(sqlc.arg(topic_ids)::int[]);