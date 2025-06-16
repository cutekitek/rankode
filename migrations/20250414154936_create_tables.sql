-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    roles INT NOT NULL DEFAULT 0, -- 0: user, 1: moderator, 2: admin
    elo int NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_elo ON users (elo);
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_username ON users (username);

CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    difficulty INT NOT NULL,
    passes INT NOT NULL DEFAULT 0,
    score float NOT NULL DEFAULT 0.0,
    topics int[],
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_difficulty ON tasks (difficulty);
CREATE INDEX idx_tasks_score ON tasks (score);
CREATE INDEX idx_tasks_passes ON tasks (passes);
CREATE INDEX on tasks using gin(topics);

CREATE TABLE topics (
    id serial primary key not null,
    name VARCHAR(255) UNIQUE NOT NULL,
    tasks_count int NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP 
);

CREATE INDEX ON topics(tasks_count);

CREATE TABLE languages (
    name text NOT NULL PRIMARY KEY
);

CREATE TABLE attempts (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    user_id INT REFERENCES users(id) NOT NULL,
    task_id INT REFERENCES tasks(id) NOT NULL,
    language TEXT REFERENCES languages(name) NOT NULL,
    code TEXT NOT NULL,
    attempt_status int NOT NULL,
    error TEXT,
    running_time INT,
    memory INT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_attempts_user_id_task_id ON attempts (user_id, task_id);
CREATE INDEX idx_attempts_running_time ON attempts(running_time);
CREATE INDEX idx_attempts_memory ON attempts(memory);

CREATE TABLE task_test_cases (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    task_id INT REFERENCES tasks(id) NOT NULL,
    case_order int NOT NULL,
    input_file VARCHAR(255) NOT NULL,
    output_file VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_test_cases_task_id ON task_test_cases(task_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS task_topics;
DROP TABLE IF EXISTS topics;
DROP TABLE IF EXISTS task_test_cases;
DROP TABLE IF EXISTS attempts;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS languages;
-- +goose StatementEnd