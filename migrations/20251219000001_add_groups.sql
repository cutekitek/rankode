-- +goose Up
-- +goose StatementBegin
CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    course_id INT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE group_students (
    group_id INT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, user_id)
);

ALTER TABLE assignments ADD COLUMN group_id INT REFERENCES groups(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE assignments DROP COLUMN IF EXISTS group_id;
DROP TABLE IF EXISTS group_students;
DROP TABLE IF EXISTS groups;
-- +goose StatementEnd
