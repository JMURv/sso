package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func (r *Repository) CreateToken(
	ctx context.Context,
	userID uuid.UUID,
	hashedT string,
	expiresAt time.Time,
	device *md.Device,
) error {
	const op = "auth.CreateToken.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			zap.L().Debug(
				"error while transaction rollback",
				zap.String("op", op),
				zap.Error(err),
			)
		}
	}()

	_, err = tx.ExecContext(
		ctx, `
        INSERT INTO user_devices (id, user_id, name, device_type, os, browser, user_agent, ip)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (id) DO UPDATE 
        SET last_active = NOW()`,
		device.ID, userID, device.Name, device.DeviceType, device.OS, device.Browser, device.UA, device.IP,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx, `
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at, device_id)
        VALUES ($1, $2, $3, $4)`,
		userID, hashedT, expiresAt, device.ID,
	)
	if err != nil {
		return err
	}

	if tx.Commit() != nil {
		return err
	}
	return nil
}

func (r *Repository) IsTokenValid(ctx context.Context, userID uuid.UUID, d *md.Device, token string) (bool, error) {
	const op = "auth.IsTokenValid.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var hash string
	err := r.conn.QueryRowContext(
		ctx, `
        SELECT token_hash
        FROM refresh_tokens 
        WHERE user_id = $1 AND device_id = $2 AND expires_at > NOW() AND revoked = FALSE
        ORDER BY expires_at DESC
        LIMIT 1
        `, userID, d.ID,
	).Scan(&hash)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	shaHash := sha256.Sum256([]byte(token))
	hexHash := hex.EncodeToString(shaHash[:])
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(hexHash)) == nil, nil
}

func (r *Repository) RevokeAllTokens(ctx context.Context, userID uuid.UUID) error {
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

func (r *Repository) GetByDevice(ctx context.Context, userID uuid.UUID, deviceID string) (*md.RefreshToken, error) {
	var token md.RefreshToken
	err := r.conn.QueryRowContext(
		ctx, `
        SELECT id, expires_at, revoked
        FROM refresh_tokens
        WHERE user_id = $1 AND device_id = $2
        ORDER BY created_at DESC
        LIMIT 1`,
		userID,
		deviceID,
	).Scan(&token.ID, &token.ExpiresAt, &token.Revoked)

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

func (r *Repository) GetUserDevices(ctx context.Context, userID uuid.UUID) ([]md.Device, error) {
	const op = "auth.GetUserDevices.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	rows, err := r.conn.QueryContext(
		ctx, `
        SELECT id, name, device_type, os, user_agent, browser, ip, last_active 
        FROM user_devices 
        WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			zap.L().Debug(
				"failed to close rows",
				zap.String("op", op),
				zap.Error(err),
			)
		}
	}(rows)

	devices := make([]md.Device, 0, config.DefaultSize)
	for rows.Next() {
		var device md.Device
		if err := rows.Scan(
			&device.ID,
			&device.Name,
			&device.DeviceType,
			&device.OS,
			&device.UA,
			&device.Browser,
			&device.IP,
			&device.LastActive,
		); err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (r *Repository) DeleteDevice(ctx context.Context, userID uuid.UUID, deviceID string) error {
	const op = "auth.DeleteDevice.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(
		ctx, `DELETE FROM user_devices WHERE id = $1 AND user_id = $2`,
		deviceID, userID,
	)

	if err != nil {
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if aff == 0 {
		return repo.ErrNotFound
	}
	return err
}
