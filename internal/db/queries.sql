-- Queries for User Management

-- name: CreateUser :one
INSERT INTO user (email, password, metadata)
VALUES (?, ?, ?)
RETURNING id;

-- name: CreateSession :exec
INSERT INTO sessions (user_id, session_id, expires_at)
VALUES (?, ?, ?);

-- name: GetUserBySessionID :one
SELECT u.*
FROM user u
         JOIN sessions t ON u.id = t.user_id
WHERE t.session_id = ?
  AND t.expires_at > datetime('now');

-- name: CreateRememberToken :exec
INSERT INTO remember_tokens (user_id, selector, validator_hash, expires_at)
VALUES (?, ?, ?, ?);

-- name: RotateRememberToken :exec
UPDATE remember_tokens
SET validator_hash = ?,
    expires_at = ?
WHERE selector = ?;

-- name: GetUserAndValidatorBySelector :one
SELECT u.*, r.validator_hash
FROM user u
         JOIN remember_tokens r ON u.id = r.user_id
WHERE r.selector = ?
  AND r.expires_at > datetime('now');

-- name: GetUserByID :one
SELECT *
FROM user
WHERE id = ?;

-- name: GetUserByEmail :one
SELECT *
FROM user
WHERE email = ?;

-- name: UpdateUserMetadata :exec
UPDATE user SET metadata = json_patch(COALESCE(metadata, '{}'), ?) WHERE id = ?;

-- name: DeleteSessionByID :exec
DELETE FROM sessions
WHERE session_id = ?;

-- name: DeleteRememberTokenBySelector :exec
DELETE FROM remember_tokens
WHERE selector = ?;

-- name: DeleteAllUserSessions :exec
DELETE FROM sessions
WHERE user_id = ?;

-- name: DeleteAllUserRememberTokens :exec
DELETE FROM remember_tokens
WHERE user_id = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < datetime('now');

-- name: DeleteExpiredRememberTokens :exec
DELETE FROM remember_tokens
WHERE expires_at < datetime('now');
