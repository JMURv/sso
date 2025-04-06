package db

import (
	"context"
	"database/sql"
	"errors"
	md "github.com/JMURv/sso/internal/models"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"time"
)

func (r *Repository) CreateToken(ctx context.Context, userID uuid.UUID, hashedT string, expiresAt time.Time, device *md.Device) error {
	const op = "auth.CreateToken.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			zap.L().Error(
				"error while transaction rollback",
				zap.String("op", op),
				zap.Error(err),
			)
		}
	}()

	_, err = tx.ExecContext(
		ctx, createUserDevice,
		device.ID, userID, device.Name, device.DeviceType, device.OS, device.Browser, device.UA, device.IP,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx, createRefreshToken,
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

	var stored string
	err := r.conn.QueryRowContext(ctx, isValidToken, userID, d.ID).Scan(&stored)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return token == stored, nil
}

func (r *Repository) RevokeAllTokens(ctx context.Context, userID uuid.UUID) error {
	const op = "auth.RevokeAllTokens.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	_, err := r.conn.ExecContext(ctx, revokeToken, userID)
	return err
}

func (r *Repository) GetByDevice(ctx context.Context, userID uuid.UUID, deviceID string) (*md.RefreshToken, error) {
	var token md.RefreshToken
	err := r.conn.QueryRowContext(ctx, getTokenByDevice, userID, deviceID).Scan(
		&token.ID,
		&token.ExpiresAt,
		&token.Revoked,
	)

	return &token, err
}

func (r *Repository) RevokeByDevice(ctx context.Context, userID uuid.UUID, deviceID string) error {
	_, err := r.conn.ExecContext(ctx, revokeTokenByDevice, userID, deviceID)
	return err
}
