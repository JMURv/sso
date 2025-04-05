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

	rows, err := r.conn.QueryContext(ctx, listDevices, uid)
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			zap.L().Error(
				"failed to close rows",
				zap.String("op", op),
				zap.Error(err),
			)
		}
	}(rows)

	devices := make([]md.Device, 0, config.DefaultSize)
	for rows.Next() {
		var device md.Device
		if err = rows.Scan(
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

func (r *Repository) GetDevice(ctx context.Context, uid uuid.UUID, dID string) (*md.Device, error) {
	const op = "auth.GetDevice.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var err error
	res := &md.Device{}
	err = r.conn.QueryRowContext(ctx, getDevice, dID, uid).Scan(
		&res.ID,
		&res.Name,
		&res.DeviceType,
		&res.OS,
		&res.Browser,
		&res.UA,
		&res.IP,
		&res.LastActive,
		&res.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) UpdateDevice(ctx context.Context, uid uuid.UUID, dID string, req *dto.UpdateDeviceRequest) error {
	const op = "auth.UpdateDevice.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, updateDevice, req.Name, dID, uid)
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
