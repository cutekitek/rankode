-- +goose Up
-- +goose StatementBegin
CREATE TABLE lti_users (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lti_subject VARCHAR(255) NOT NULL,
    lti_issuer VARCHAR(255) NOT NULL,
    lti_deployment_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (lti_subject, lti_issuer)
);

CREATE INDEX idx_lti_users_user_id ON lti_users(user_id);
CREATE INDEX idx_lti_users_identity ON lti_users(lti_subject, lti_issuer);

-- Update users table to ensure we can store longer emails if needed (LTI emails can sometimes be long)
-- But standard VARCHAR(255) is usually enough. Just a note.

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lti_users;
-- +goose StatementEnd
