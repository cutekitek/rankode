-- +goose Up
-- +goose StatementBegin

-- Add teacher role to existing roles system (0: user, 1: moderator, 2: admin, 3: teacher)
-- Note: roles is an integer column, not an enum, so no schema change needed
-- We'll handle the teacher role (value 3) in application logic

-- Create courses table (single teacher per course)
CREATE TABLE courses (
    id SERIAL PRIMARY KEY,
    teacher_id INT NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    join_code VARCHAR(10) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_courses_teacher_id ON courses(teacher_id);
CREATE INDEX idx_courses_join_code ON courses(join_code);

-- Create assignments table (contains due dates, start dates, max attempts)
CREATE TABLE assignments (
    id SERIAL PRIMARY KEY,
    course_id INT NOT NULL REFERENCES courses(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_date TIMESTAMP,
    due_date TIMESTAMP,
    max_attempts_per_task INT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_assignments_course_id ON assignments(course_id);
CREATE INDEX idx_assignments_due_date ON assignments(due_date);

-- Create assignment_tasks mapping (many-to-many with ordering)
CREATE TABLE assignment_tasks (
    assignment_id INT NOT NULL REFERENCES assignments(id),
    task_id INT NOT NULL REFERENCES tasks(id),
    order_index INT NOT NULL,
    weight DECIMAL(3,2) DEFAULT 1.0,
    PRIMARY KEY (assignment_id, task_id)
);

CREATE INDEX idx_assignment_tasks_assignment_id ON assignment_tasks(assignment_id);
CREATE INDEX idx_assignment_tasks_task_id ON assignment_tasks(task_id);

-- Create enrollments table
CREATE TABLE enrollments (
    course_id INT NOT NULL REFERENCES courses(id),
    user_id INT NOT NULL REFERENCES users(id),
    enrolled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    role VARCHAR(10) DEFAULT 'student',
    PRIMARY KEY (course_id, user_id)
);

CREATE INDEX idx_enrollments_course_id ON enrollments(course_id);
CREATE INDEX idx_enrollments_user_id ON enrollments(user_id);

-- Create grades table (1-5 per task)
CREATE TABLE grades (
    id SERIAL PRIMARY KEY,
    assignment_id INT NOT NULL REFERENCES assignments(id),
    task_id INT NOT NULL REFERENCES tasks(id),
    user_id INT NOT NULL REFERENCES users(id),
    grade SMALLINT NOT NULL CHECK (grade BETWEEN 1 AND 5),
    feedback TEXT,
    graded_by INT REFERENCES users(id),
    graded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (assignment_id, task_id, user_id)
);

CREATE INDEX idx_grades_assignment_id ON grades(assignment_id);
CREATE INDEX idx_grades_task_id ON grades(task_id);
CREATE INDEX idx_grades_user_id ON grades(user_id);
CREATE INDEX idx_grades_assignment_user ON grades(assignment_id, user_id);

-- Modify tasks table to support course association and privacy
ALTER TABLE tasks 
ADD COLUMN course_id INT REFERENCES courses(id),
ADD COLUMN is_public BOOLEAN NOT NULL DEFAULT true;

CREATE INDEX idx_tasks_course_id ON tasks(course_id);
CREATE INDEX idx_tasks_is_public ON tasks(is_public);

-- Optional: Link attempts to assignments for better tracking
-- ALTER TABLE attempts ADD COLUMN assignment_id INT REFERENCES assignments(id) NULL;
-- CREATE INDEX idx_attempts_assignment_id ON attempts(assignment_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop new indexes
DROP INDEX IF EXISTS idx_tasks_is_public;
DROP INDEX IF EXISTS idx_tasks_course_id;

-- Drop new columns from tasks table
ALTER TABLE tasks 
DROP COLUMN IF EXISTS course_id,
DROP COLUMN IF EXISTS is_public;

-- Drop new tables (in reverse order of creation)
DROP TABLE IF EXISTS grades;
DROP TABLE IF EXISTS enrollments;
DROP TABLE IF EXISTS assignment_tasks;
DROP TABLE IF EXISTS assignments;
DROP TABLE IF EXISTS courses;

-- +goose StatementEnd