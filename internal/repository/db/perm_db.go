package db

import (
	"context"
	"database/sql"
	"errors"
	repo "github.com/JMURv/sso/internal/repository"
	"github.com/JMURv/sso/pkg/model"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
)

func (r *Repository) ListPermissions(ctx context.Context, page, size int) (*model.PaginatedPermission, error) {
	const op = "sso.ListPermissions.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM permission`,
	).Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.QueryContext(
		ctx,
		`SELECT id, name
		FROM permission
		ORDER BY name
		LIMIT $1 OFFSET $2`, size, (page-1)*size,
	)
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
	err := r.conn.QueryRowContext(
		ctx, `
		SELECT id, name
		FROM permission
		WHERE id = $1
		`, id,
	).Scan(
		&res.ID,
		&res.Name,
	)

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

	var idx uuid.UUID
	if err := r.conn.
		QueryRow(`SELECT id FROM permission WHERE name=$1`, req.Name).
		Scan(&idx); err == nil {
		return 0, repo.ErrAlreadyExists
	} else if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	tx, err := r.conn.Begin()
	if err != nil {
		return 0, err
	}

	var id uint64
	err = tx.QueryRow(`INSERT INTO permission (name) VALUES ($1) RETURNING id`, req.Name).Scan(&id)
	if err != nil {
		tx.Rollback()
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

	res, err := r.conn.Exec(`UPDATE permission SET name = $1 WHERE id = $2`, req.Name, id)
	if err != nil {
		return err
	}

	if aff, _ := res.RowsAffected(); aff == 0 {
		return repo.ErrNotFound
	}

	return nil
}

func (r *Repository) DeletePerm(ctx context.Context, id uint64) error {
	const op = "sso.DeletePerm.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.Exec(`DELETE FROM permission WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if aff, _ := res.RowsAffected(); aff == 0 {
		return repo.ErrNotFound
	}
	return nil
}
