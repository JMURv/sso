package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

func (r *Repository) ListDevices(ctx context.Context, uid uuid.UUID) ([]md.Device, error) {
	const op = "auth.ListDevices.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	devices := make([]md.Device, 0, config.DefaultSize)
	if err := r.conn.SelectContext(ctx, &devices, listDevices, uid); err != nil {
		zap.L().Error(
			"failed to list devices",
			zap.String("op", op),
			zap.String("uid", uid.String()),
			zap.Error(err),
		)
		return nil, err
	}

	return devices, nil
}

func (r *Repository) GetDevice(ctx context.Context, uid uuid.UUID, dID string) (*md.Device, error) {
	const op = "auth.GetDevice.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := md.Device{}
	err := r.conn.GetContext(ctx, &res, getDevice, dID, uid)
	if errors.Is(err, sql.ErrNoRows) {
		zap.L().Debug(
			"no device found",
			zap.String("op", op),
			zap.String("userID", uid.String()),
			zap.String("deviceID", dID),
		)
		return nil, repo.ErrNotFound
	} else if err != nil {
		zap.L().Error(
			"failed to get device",
			zap.String("op", op),
			zap.String("userID", uid.String()),
			zap.String("deviceID", dID),
			zap.Error(err),
		)
		return nil, err
	}

	return &res, nil
}

func (r *Repository) GetDeviceByID(ctx context.Context, dID string) (*md.Device, error) {
	const op = "auth.GetDevice.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := md.Device{}
	err := r.conn.GetContext(ctx, &res, getDeviceByID, dID)
	if errors.Is(err, sql.ErrNoRows) {
		zap.L().Debug(
			"no device found",
			zap.String("op", op),
			zap.String("deviceID", dID),
		)
		return nil, repo.ErrNotFound
	} else if err != nil {
		zap.L().Error(
			"failed to get device",
			zap.String("op", op),
			zap.String("deviceID", dID),
			zap.Error(err),
		)
		return nil, err
	}

	return &res, nil
}

func (r *Repository) UpdateDevice(ctx context.Context, uid uuid.UUID, dID string, req *dto.UpdateDeviceRequest) error {
	const op = "auth.UpdateDevice.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, updateDevice, req.Name, dID, uid)
	if err != nil {
		zap.L().Error(
			"failed to update device",
			zap.String("op", op),
			zap.String("userID", uid.String()),
			zap.String("deviceID", dID),
			zap.Error(err),
		)
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		zap.L().Error(
			"failed to get affected rows",
			zap.String("op", op),
			zap.String("userID", uid.String()),
			zap.String("deviceID", dID),
			zap.Error(err),
		)
		return err
	}

	if aff == 0 {
		zap.L().Debug(
			"failed to find device",
			zap.String("op", op),
			zap.String("userID", uid.String()),
			zap.String("deviceID", dID),
		)
		return repo.ErrNotFound
	}

	return nil
}

func (r *Repository) DeleteDevice(ctx context.Context, userID uuid.UUID, deviceID string) error {
	const op = "auth.DeleteDevice.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	if err := r.RevokeByDevice(ctx, userID, deviceID); err != nil {
		return err
	}

	res, err := r.conn.ExecContext(ctx, deleteDevice, deviceID, userID)
	if err != nil {
		zap.L().Error(
			"failed to delete device",
			zap.String("op", op),
			zap.String("userID", userID.String()),
			zap.String("deviceID", deviceID),
			zap.Error(err),
		)
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		zap.L().Error(
			"failed to get affected rows",
			zap.String("op", op),
			zap.String("userID", userID.String()),
			zap.String("deviceID", deviceID),
			zap.Error(err),
		)
		return err
	}

	if aff == 0 {
		zap.L().Debug(
			"failed to find device",
			zap.String("op", op),
			zap.String("userID", userID.String()),
			zap.String("deviceID", deviceID),
		)
		return repo.ErrNotFound
	}
	return err
}
