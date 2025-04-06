package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

func (r *Repository) SearchUser(ctx context.Context, query string, page, size int) (*dto.PaginatedUserResponse, error) {
	const op = "users.SearchUser.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.QueryRowContext(ctx, userSearchSelectQ, "%"+query+"%", "%"+query+"%").
		Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.QueryContext(ctx, userSearchQ, "%"+query+"%", "%"+query+"%", size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		if err = rows.Close(); err != nil {
			zap.L().Debug(
				"failed to close rows",
				zap.String("op", op),
				zap.Error(err),
			)
		}
	}(rows)

	res := make([]*md.User, 0, size)
	for rows.Next() {
		user := &md.User{}
		roles := make([]string, 0, 5)
		if err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Avatar,
			&user.CreatedAt,
			&user.UpdatedAt,
			pq.Array(&roles),
		); err != nil {
			return nil, err
		}

		user.Roles, err = ScanRoles(roles)
		if err != nil {
			return nil, err
		}
		res = append(res, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((count + int64(size) - 1) / int64(size))
	return &dto.PaginatedUserResponse{
		Data:        res,
		Count:       count,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNextPage: page < totalPages,
	}, nil
}

func (r *Repository) ListUsers(ctx context.Context, page, size int) (*dto.PaginatedUserResponse, error) {
	const op = "users.ListUsers.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.QueryRowContext(ctx, userSelectQ).Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.QueryContext(ctx, userListQ, size, (page-1)*size)
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

	res := make([]*md.User, 0, size)
	for rows.Next() {
		user := &md.User{}
		roles := make([]string, 0, 5)
		if err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Avatar,
			&user.CreatedAt,
			&user.UpdatedAt,
			pq.Array(&roles),
		); err != nil {
			return nil, err
		}

		user.Roles, err = ScanRoles(roles)
		if err != nil {
			return nil, err
		}

		res = append(res, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((count + int64(size) - 1) / int64(size))
	return &dto.PaginatedUserResponse{
		Data:        res,
		Count:       count,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNextPage: page < totalPages,
	}, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (*md.User, error) {
	const op = "users.GetUserByID.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var err error
	res := &md.User{}
	roles := make([]string, 0, 5)
	oauth2 := make([]string, 0, 5)
	err = r.conn.QueryRowContext(ctx, userGetByIDQ, userID).Scan(
		&res.ID,
		&res.Name,
		&res.Email,
		&res.Avatar,
		&res.IsWA,
		&res.IsActive,
		&res.IsEmailVerified,
		&res.CreatedAt,
		&res.UpdatedAt,
		pq.Array(&roles),
		pq.Array(&oauth2),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	res.Roles, err = ScanRoles(roles)
	if err != nil {
		return nil, err
	}

	res.Oauth2Connections, err = ScanOauth2Connections(oauth2)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*md.User, error) {
	const op = "users.GetUserByEmail.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := &md.User{}
	roles := make([]string, 0, 5)
	err := r.conn.QueryRowContext(ctx, userGetByEmailQ, email).
		Scan(
			&res.ID,
			&res.Name,
			&res.Email,
			&res.Password,
			&res.Avatar,
			&res.IsWA,
			&res.IsActive,
			&res.IsEmailVerified,
			&res.CreatedAt,
			&res.UpdatedAt,
			pq.Array(&roles),
		)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	res.Roles, err = ScanRoles(roles)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *Repository) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (uuid.UUID, error) {
	const op = "users.CreateUser.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	tx, err := r.conn.Begin()
	if err != nil {
		return uuid.Nil, err
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

	var id uuid.UUID
	err = tx.QueryRowContext(
		ctx,
		userCreateQ,
		req.Name,
		req.Password,
		req.Email,
		req.Avatar,
		req.IsActive,
		req.IsEmail,
	).Scan(&id)

	if err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == "23505" {
			return uuid.Nil, repo.ErrAlreadyExists
		}
		return uuid.Nil, err
	}

	if len(req.Roles) > 0 {
		for i := 0; i < len(req.Roles); i++ {
			if _, err = tx.ExecContext(ctx, userAddRoleQ, id, req.Roles[i]); err != nil {
				return uuid.Nil, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest) error {
	const op = "users.UpdateUser.repo"
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

	res, err := tx.ExecContext(
		ctx,
		userUpdateQ,
		req.Name,
		req.Password,
		req.Email,
		req.Avatar,
		req.IsActive,
		req.IsEmail,
		id,
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

	if _, err = tx.ExecContext(ctx, userRemoveRoleQ, id); err != nil {
		return err
	}

	for i := 0; i < len(req.Roles); i++ {
		if _, err = tx.ExecContext(ctx, userAddRoleQ, id, req.Roles[i]); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	const op = "users.DeleteUser.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, userDeleteQ, id)
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
