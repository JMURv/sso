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

func (r *Repository) ListRoles(ctx context.Context, page, size int) (*md.PaginatedRole, error) {
	const op = "roles.ListRoles.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.QueryRowContext(ctx, roleSelect).Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.QueryContext(ctx, roleList, size, (page-1)*size)
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

	res := make([]*md.Role, 0, size)
	for rows.Next() {
		var p md.Role
		if err = rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
		); err != nil {
			return nil, err
		}
		res = append(res, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((count + int64(size) - 1) / int64(size))
	return &md.PaginatedRole{
		Data:        res,
		Count:       count,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNextPage: page < totalPages,
	}, nil
}

func (r *Repository) GetRole(ctx context.Context, id uint64) (*md.Role, error) {
	const op = "roles.GetRole.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := &md.Role{}
	err := r.conn.QueryRowContext(ctx, roleGet, id).Scan(&res.ID, &res.Name, &res.Description)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (uint64, error) {
	const op = "roles.CreateRole.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var id uint64
	err := r.conn.QueryRowContext(ctx, roleCreate, req.Name, req.Description).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return 0, repo.ErrAlreadyExists
		}
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateRole(ctx context.Context, id uint64, req *dto.UpdateRoleRequest) error {
	const op = "roles.UpdateRole.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, roleUpdate, req.Name, req.Description, id)
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

func (r *Repository) DeleteRole(ctx context.Context, id uint64) error {
	const op = "roles.DeleteRole.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, roleDelete, id)
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
