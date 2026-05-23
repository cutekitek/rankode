-- name: GetOidcIdentity :one
SELECT * FROM oidc_identities
WHERE provider_name = $1 AND subject = $2;

-- name: UpsertOidcIdentity :one
INSERT INTO oidc_identities (user_id, provider_name, subject, email, email_verified)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (provider_name, subject) DO UPDATE
SET email = EXCLUDED.email,
    email_verified = EXCLUDED.email_verified,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetOidcProvider :one
SELECT * FROM oidc_providers WHERE name = $1;

-- name: UpsertOidcProvider :one
INSERT INTO oidc_providers (
    name,
    issuer,
    client_id,
    client_secret,
    auth_url,
    token_url,
    jwks_url,
    redirect_url,
    frontend_redirect_url,
    scopes,
    allowed_domains,
    require_email_verified
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (name) DO UPDATE
SET issuer = EXCLUDED.issuer,
    client_id = EXCLUDED.client_id,
    client_secret = EXCLUDED.client_secret,
    auth_url = EXCLUDED.auth_url,
    token_url = EXCLUDED.token_url,
    jwks_url = EXCLUDED.jwks_url,
    redirect_url = EXCLUDED.redirect_url,
    frontend_redirect_url = EXCLUDED.frontend_redirect_url,
    scopes = EXCLUDED.scopes,
    allowed_domains = EXCLUDED.allowed_domains,
    require_email_verified = EXCLUDED.require_email_verified,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;
