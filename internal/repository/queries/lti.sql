-- name: CreateLtiLink :exec
INSERT INTO lti_users (user_id, lti_subject, lti_issuer, lti_deployment_id)
VALUES ($1, $2, $3, $4);

-- name: GetLtiUser :one
SELECT * FROM lti_users WHERE lti_subject = $1 AND lti_issuer = $2;
