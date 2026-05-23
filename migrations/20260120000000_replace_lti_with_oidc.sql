-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS lti_users;

CREATE TABLE oidc_providers (
    name VARCHAR(100) PRIMARY KEY,
    issuer TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_secret TEXT,
    auth_url TEXT NOT NULL,
    token_url TEXT NOT NULL,
    jwks_url TEXT NOT NULL,
    redirect_url TEXT NOT NULL,
    frontend_redirect_url TEXT NOT NULL DEFAULT '/',
    scopes TEXT[] NOT NULL DEFAULT ARRAY['openid', 'email', 'profile'],
    allowed_domains TEXT[],
    require_email_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE oidc_identities (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_name VARCHAR(100) NOT NULL,
    subject TEXT NOT NULL,
    email TEXT,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (provider_name, subject)
);

CREATE INDEX idx_oidc_identities_user_id ON oidc_identities(user_id);
CREATE INDEX idx_oidc_identities_provider_subject ON oidc_identities(provider_name, subject);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS oidc_identities;
DROP TABLE IF EXISTS oidc_providers;

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
-- +goose StatementEnd
