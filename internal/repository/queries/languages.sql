-- name: ListLanguages :many
SELECT * from languages;

-- name: LanguageExists :one
SELECT name from languages where 'name' = $1;