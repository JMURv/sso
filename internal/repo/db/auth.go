package db

import (
	"context"
	"database/sql"
	"github.com/JMURv/sso/internal/auth"
	md "github.com/JMURv/sso/internal/models"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// TODO: Make hash
func (r *Repository) CreateToken(ctx context.Context, userID uuid.UUID, hashedT string, expiresAt time.Time) error {
	const op = "auth.CreateToken.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	_, err := r.conn.ExecContext(
		ctx, `
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

	err := r.conn.QueryRowContext(
		ctx, `
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

	_, err := r.conn.ExecContext(
		ctx, `
        UPDATE refresh_tokens 
        SET revoked = TRUE 
        WHERE user_id = $1`,
		userID,
	)
	return err
}

func (r *Repository) Create(
	ctx context.Context,
	userID,
	token,
	deviceID string,
	deviceInfo auth.DeviceInfo,
	expiresAt time.Time,
) error {

	deviceInfoJSON, _ := json.Marshal(deviceInfo)

	_, err := r.conn.ExecContext(
		ctx, `
        INSERT INTO refresh_tokens 
        (user_id, token_hash, expires_at, device_id, device_info)
        VALUES ($1, $2, $3, $4, $5)`,
		userID,
		token,
		expiresAt,
		deviceID,
		deviceInfoJSON,
	)
	return err
}

func (r *Repository) GetByDevice(ctx context.Context, userID uuid.UUID, deviceID string) (*md.RefreshToken, error) {
	var token md.RefreshToken
	err := r.conn.QueryRowContext(
		ctx, `
        SELECT id, expires_at, revoked, device_info
        FROM refresh_tokens
        WHERE user_id = $1 AND device_id = $2
        ORDER BY created_at DESC
        LIMIT 1`,
		userID,
		deviceID,
	).Scan(&token.ID, &token.ExpiresAt, &token.Revoked, &token.DeviceInfo)

	return &token, err
}

func (r *Repository) RevokeByDevice(ctx context.Context, userID uuid.UUID, deviceID string) error {
	_, err := r.conn.ExecContext(
		ctx, `
        UPDATE refresh_tokens
        SET revoked = TRUE
        WHERE user_id = $1 AND device_id = $2`,
		userID,
		deviceID,
	)
	return err
}
