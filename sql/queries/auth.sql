-- name: StoreRefreshToken :one
INSERT INTO refresh_token (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    NOW() + interval '60 days',
    NULL
    )
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT user_id 
FROM refresh_token 
WHERE token = $1
AND expires_at > NOW()
AND revoked_at IS NULL;

-- name: RevokeRefershToken :exec
UPDATE refresh_token
SET updated_at = NOW(),
    revoked_at = NOW()
WHERE token = $1;
