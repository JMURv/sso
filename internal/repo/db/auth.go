package db

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func (r *Repository) CreateToken(ctx context.Context, userID uuid.UUID, hashedT string, expiresAt time.Time) error {
	const op = "auth.CreateToken.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	_, err := r.conn.ExecContext(ctx, `
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
        VALUES ($1, $2, $3)`,
		userID, hashedT, expiresAt,
	)

	return err
}

func (r *Repository) IsTokenValid(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
	const op = "auth.IsTokenValid.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var hash string
	var expiresAt time.Time
	var revoked bool

	err := r.conn.QueryRowContext(ctx, `
        SELECT token_hash, expires_at, revoked 
        FROM refresh_tokens 
        WHERE user_id = $1
        ORDER BY expires_at DESC
        LIMIT 1`,
		userID,
	).Scan(&hash, &expiresAt, &revoked)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if revoked || time.Now().After(expiresAt) {
		return false, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	return err == nil, nil
}

func (r *Repository) RevokeAllTokens(ctx context.Context, userID string) error {
	const op = "auth.RevokeAllTokens.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	_, err := r.conn.ExecContext(ctx, `
        UPDATE refresh_tokens 
        SET revoked = TRUE 
        WHERE user_id = $1`,
		userID,
	)
	return err
}
