-- +goose Up
-- +goose StatementBegin
ALTER TABLE attempts ADD COLUMN assignment_id INT REFERENCES assignments(id) NULL;
CREATE INDEX idx_attempts_assignment_id ON attempts(assignment_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_attempts_assignment_id;
ALTER TABLE attempts DROP COLUMN IF EXISTS assignment_id;
-- +goose StatementEnd
