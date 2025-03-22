package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"strings"
)

func (r *Repository) ListPermissions(ctx context.Context, page, size int) (*md.PaginatedPermission, error) {
	const op = "sso.ListPermissions.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.QueryRowContext(ctx, permSelect).Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.QueryContext(ctx, permList, size, (page-1)*size)
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

	res := make([]*md.Permission, 0, size)
	for rows.Next() {
		var p md.Permission
		if err = rows.Scan(
			&p.ID,
			&p.Name,
		); err != nil {
			return nil, err
		}
		res = append(res, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((count + int64(size) - 1) / int64(size))
	return &md.PaginatedPermission{
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

	res := &md.Permission{}
	err := r.conn.QueryRowContext(ctx, permGet, id).Scan(&res.ID, &res.Name)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) CreatePerm(ctx context.Context, req *dto.CreatePermissionRequest) (uint64, error) {
	const op = "sso.CreatePerm.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var id uint64
	err := r.conn.QueryRowContext(ctx, permCreate, req.Name).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return 0, repo.ErrAlreadyExists
		}
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdatePerm(ctx context.Context, id uint64, req *dto.UpdatePermissionRequest) error {
	const op = "sso.UpdatePerm.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, permUpdate, req.Name, id)
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

func (r *Repository) DeletePerm(ctx context.Context, id uint64) error {
	const op = "sso.DeletePerm.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, permDelete, id)
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
