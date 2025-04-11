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

func (r *Repository) ListRoles(ctx context.Context, page, size int, filters map[string]any) (*dto.PaginatedRoleResponse, error) {
	const op = "roles.ListRoles.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	idx := 1
	args := make([]any, 0, len(filters))
	clauses := make([]any, 0, len(filters))
	if name, ok := filters["name"]; ok && name != "" {
		clauses = append(clauses, "WHERE r.name ILIKE $1")
		args = append(args, fmt.Sprintf("%%%s%%", name))
		idx++
	}

	var count int64
	q := fmt.Sprintf(roleSelect, clauses)
	if err := r.conn.QueryRowContext(ctx, q, args...).Scan(&count); err != nil {
		zap.L().Error("failed to count roles", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	q = fmt.Sprintf(roleList, clauses, idx, idx+1)
	rows, err := r.conn.QueryContext(ctx, q, size, (page-1)*size)
	if err != nil {
		zap.L().Error(
			"failed to list roles",
			zap.String("op", op),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Error(err),
		)
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
		perms := make([]string, 0, 5)
		if err = rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			pq.Array(&perms),
		); err != nil {
			zap.L().Error("failed to scan role", zap.String("op", op), zap.Error(err))
			return nil, err
		}

		if p.Permissions, err = ScanPerms(perms); err != nil {
			zap.L().Error("failed to scan permissions", zap.String("op", op), zap.Error(err))
			return nil, err
		}
		res = append(res, &p)
	}

	if err = rows.Err(); err != nil {
		zap.L().Error("failed to iterate rows", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	totalPages := int((count + int64(size) - 1) / int64(size))
	return &dto.PaginatedRoleResponse{
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

	res := md.Role{}
	err := r.conn.GetContext(ctx, &res, roleGet, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			zap.L().Debug(
				"failed to find role",
				zap.String("op", op),
				zap.Uint64("id", id),
			)
			return nil, repo.ErrNotFound
		}
		zap.L().Error(
			"failed to get role",
			zap.String("op", op),
			zap.Uint64("id", id),
			zap.Error(err),
		)
		return nil, err
	}

	return &res, nil
}

func (r *Repository) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (uint64, error) {
	const op = "roles.CreateRole.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var id uint64
	err := r.conn.QueryRowContext(ctx, roleCreate, req.Name, req.Description).Scan(&id)
	if err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == "23505" {
			zap.L().Debug(
				"role already exists",
				zap.String("op", op),
			)
			return 0, repo.ErrAlreadyExists
		}
		zap.L().Error(
			"failed to create role",
			zap.String("op", op),
			zap.Error(err),
		)
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
		zap.L().Error(
			"failed to update role",
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
			"failed to find role",
			zap.String("op", op),
			zap.Uint64("id", id),
		)
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
		zap.L().Error(
			"failed to delete role",
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
			"failed to find role",
			zap.String("op", op),
			zap.Uint64("id", id),
		)
		return repo.ErrNotFound
	}
	return nil
}
