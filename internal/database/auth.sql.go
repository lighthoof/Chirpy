// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: auth.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const getUserFromRefreshToken = `-- name: GetUserFromRefreshToken :one
SELECT user_id 
FROM refresh_token 
WHERE token = $1
AND expires_at > NOW()
AND revoked_at IS NULL
`

func (q *Queries) GetUserFromRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getUserFromRefreshToken, token)
	var user_id uuid.UUID
	err := row.Scan(&user_id)
	return user_id, err
}

const revokeRefershToken = `-- name: RevokeRefershToken :exec
UPDATE refresh_token
SET updated_at = NOW(),
    revoked_at = NOW()
WHERE token = $1
`

func (q *Queries) RevokeRefershToken(ctx context.Context, token string) error {
	_, err := q.db.ExecContext(ctx, revokeRefershToken, token)
	return err
}

const storeRefreshToken = `-- name: StoreRefreshToken :one
INSERT INTO refresh_token (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    NOW() + interval '60 days',
    NULL
    )
RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at
`

type StoreRefreshTokenParams struct {
	Token  string
	UserID uuid.UUID
}

func (q *Queries) StoreRefreshToken(ctx context.Context, arg StoreRefreshTokenParams) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, storeRefreshToken, arg.Token, arg.UserID)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.ExpiresAt,
		&i.RevokedAt,
	)
	return i, err
}
