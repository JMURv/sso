package db

import (
	"context"
	"database/sql"
	"errors"
	repo "github.com/JMURv/sso/internal/repository"
	"github.com/JMURv/sso/pkg/model"
	"github.com/opentracing/opentracing-go"
	"strings"
)

func (r *Repository) ListPermissions(ctx context.Context, page, size int) (*model.PaginatedPermission, error) {
	const op = "sso.ListPermissions.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.QueryRow(permSelect).Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.Query(permList, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*model.Permission, 0, size)
	for rows.Next() {
		var p model.Permission
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
	return &model.PaginatedPermission{
		Data:        res,
		Count:       count,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNextPage: page < totalPages,
	}, nil
}

func (r *Repository) GetPermission(ctx context.Context, id uint64) (*model.Permission, error) {
	const op = "sso.GetPermission.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := &model.Permission{}
	err := r.conn.QueryRow(permGet, id).Scan(&res.ID, &res.Name)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) CreatePerm(ctx context.Context, req *model.Permission) (uint64, error) {
	const op = "sso.CreatePerm.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	tx, err := r.conn.Begin()
	if err != nil {
		return 0, err
	}

	var id uint64
	err = tx.QueryRow(permCreate, req.Name).Scan(&id)
	if err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "unique constraint") {
			return 0, repo.ErrAlreadyExists
		}
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) UpdatePerm(ctx context.Context, id uint64, req *model.Permission) error {
	const op = "sso.UpdatePerm.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	res, err := tx.Exec(permUpdate, req.Name, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	if aff, _ := res.RowsAffected(); aff == 0 {
		return repo.ErrNotFound
	}

	return tx.Commit()
}

func (r *Repository) DeletePerm(ctx context.Context, id uint64) error {
	const op = "sso.DeletePerm.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.Exec(permDelete, id)
	if err != nil {
		return err
	}

	if aff, _ := res.RowsAffected(); aff == 0 {
		return repo.ErrNotFound
	}
	return nil
}
