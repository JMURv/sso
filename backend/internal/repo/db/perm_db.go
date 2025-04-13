package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

func (r *Repository) ListPermissions(ctx context.Context, page, size int, filters map[string]any) (*dto.PaginatedPermissionResponse, error) {
	const op = "sso.ListPermissions.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	idx := 1
	var clause string
	args := make([]any, 0, len(filters))
	if name, ok := filters["search"]; ok && name != "" {
		clause = "WHERE p.name ILIKE $1"
		args = append(args, fmt.Sprintf("%%%s%%", name))
		idx++
	}

	var count int64
	q := fmt.Sprintf(permSelect, clause)
	if err := r.conn.QueryRowContext(ctx, q, args...).Scan(&count); err != nil {
		zap.L().Error("failed to count permissions", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	res := make([]*md.Permission, 0, size)
	q = fmt.Sprintf(permList, clause, idx, idx+1)
	args = append(args, size, (page-1)*size)
	err := r.conn.SelectContext(ctx, &res, q, args...)
	if err != nil {
		zap.L().Error(
			"failed to list permissions",
			zap.String("op", op),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Error(err),
		)
		return nil, err
	}

	totalPages := int((count + int64(size) - 1) / int64(size))
	return &dto.PaginatedPermissionResponse{
		Data:        res,
		Count:       count,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNextPage: page < totalPages,
	}, nil
}

func (r *Repository) GetPermission(ctx context.Context, id uint64) (*md.Permission, error) {
	const op = "sso.GetPermission.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := md.Permission{}
	err := r.conn.GetContext(ctx, &res, permGet, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		zap.L().Debug(
			"failed to find permission",
			zap.String("op", op),
			zap.Uint64("id", id),
		)
		return nil, repo.ErrNotFound
	} else if err != nil {
		zap.L().Error(
			"failed to find permission",
			zap.String("op", op),
			zap.Uint64("id", id),
			zap.Error(err),
		)
		return nil, err
	}

	return &res, nil
}

func (r *Repository) CreatePerm(ctx context.Context, req *dto.CreatePermissionRequest) (uint64, error) {
	const op = "sso.CreatePerm.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var id uint64
	err := r.conn.QueryRowContext(ctx, permCreate, req.Name, req.Description).Scan(&id)
	if err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == "23505" {
			zap.L().Debug(
				"permission already exists",
				zap.String("op", op),
			)
			return 0, repo.ErrAlreadyExists
		}
		zap.L().Error(
			"failed to create permission",
			zap.String("op", op),
			zap.Error(err),
		)
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdatePerm(ctx context.Context, id uint64, req *dto.UpdatePermissionRequest) error {
	const op = "sso.UpdatePerm.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, permUpdate, req.Name, req.Description, id)
	if err != nil {
		zap.L().Error(
			"failed to update permission",
			zap.String("op", op),
			zap.Uint64("id", id),
			zap.Error(err),
		)
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		zap.L().Error(
			"failed to get affected rows",
			zap.String("op", op),
			zap.Uint64("id", id),
			zap.Error(err),
		)
		return err
	}

	if aff == 0 {
		zap.L().Debug(
			"failed to find permission",
			zap.String("op", op),
			zap.Uint64("id", id),
		)
		return repo.ErrNotFound
	}

	return nil
}

func (r *Repository) DeletePerm(ctx context.Context, id uint64) error {
	const op = "sso.DeletePerm.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, permDelete, id)
	if err != nil {
		zap.L().Error(
			"failed to delete permission",
			zap.String("op", op),
			zap.Uint64("id", id),
			zap.Error(err),
		)
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		zap.L().Error(
			"failed to get affected rows",
			zap.String("op", op),
			zap.Uint64("id", id),
			zap.Error(err),
		)
		return err
	}

	if aff == 0 {
		zap.L().Debug(
			"failed to find permission",
			zap.String("op", op),
			zap.Uint64("id", id),
		)
		return repo.ErrNotFound
	}
	return nil
}
