package db

import (
	"context"
	"database/sql"
	repo "github.com/JMURv/sso/internal/repository"
	md "github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/db"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

func (r *Repository) SearchUser(ctx context.Context, query string, page, size int) (*md.PaginatedUser, error) {
	const op = "sso.SearchUser.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.QueryRow(userSearchSelectQ, "%"+query+"%", "%"+query+"%").
		Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.Query(userSearchQ, "%"+query+"%", "%"+query+"%", size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*md.User, 0, size)
	for rows.Next() {
		user := &md.User{}
		perms := make([]string, 0, 5)
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Password,
			&user.Email,
			&user.Avatar,
			&user.Address,
			&user.Phone,
			&user.CreatedAt,
			&user.UpdatedAt,
			pq.Array(&perms),
		); err != nil {
			return nil, err
		}

		user.Permissions, err = utils.ScanPermissions(perms)
		if err != nil {
			return nil, err
		}

		res = append(res, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((count + int64(size) - 1) / int64(size))
	return &md.PaginatedUser{
		Data:        res,
		Count:       count,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNextPage: page < totalPages,
	}, nil
}

func (r *Repository) ListUsers(ctx context.Context, page, size int) (*md.PaginatedUser, error) {
	const op = "sso.ListUsers.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.QueryRow(userSelectQ).Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.Query(userListQ, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*md.User, 0, size)
	for rows.Next() {
		user := &md.User{}
		perms := make([]string, 0, 5)
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Password,
			&user.Email,
			&user.Avatar,
			&user.Address,
			&user.Phone,
			&user.CreatedAt,
			&user.UpdatedAt,
			pq.Array(&perms),
		); err != nil {
			return nil, err
		}

		user.Permissions, err = utils.ScanPermissions(perms)
		if err != nil {
			return nil, err
		}

		res = append(res, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((count + int64(size) - 1) / int64(size))
	return &md.PaginatedUser{
		Data:        res,
		Count:       count,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNextPage: page < totalPages,
	}, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (*md.User, error) {
	const op = "sso.GetUserByID.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := &md.User{}
	perms := make([]string, 0, 5)
	err := r.conn.QueryRow(userGetByIDQ, userID).Scan(
		&res.ID,
		&res.Name,
		&res.Password,
		&res.Email,
		&res.Avatar,
		&res.Address,
		&res.Phone,
		&res.CreatedAt,
		&res.UpdatedAt,
		pq.Array(&perms),
	)

	if err == sql.ErrNoRows {
		return nil, repo.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	res.Permissions, err = utils.ScanPermissions(perms)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*md.User, error) {
	const op = "sso.GetUserByEmail.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := &md.User{}
	err := r.conn.QueryRow(userGetByEmailQ, email).
		Scan(
			&res.ID,
			&res.Name,
			&res.Password,
			&res.Email,
			&res.Avatar,
			&res.Address,
			&res.Phone,
			&res.CreatedAt,
			&res.UpdatedAt,
		)

	if err == sql.ErrNoRows {
		return nil, repo.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) CreateUser(ctx context.Context, req *md.User) (uuid.UUID, error) {
	const op = "sso.CreateUser.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	password, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, repo.ErrGeneratingPassword
	}
	req.Password = string(password)

	tx, err := r.conn.Begin()
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID
	err = tx.QueryRow(
		userCreateQ,
		req.Name,
		req.Password,
		req.Email,
		req.Avatar,
		req.Address,
		req.Phone,
	).Scan(&id)

	if err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "unique constraint") {
			return uuid.Nil, repo.ErrAlreadyExists
		}
		return uuid.Nil, err
	}

	if len(req.Permissions) > 0 {
		for _, v := range req.Permissions {
			if _, err := tx.Exec(userCreatePermQ, id, v.ID, v.Value); err != nil {
				tx.Rollback()
				return uuid.Nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id uuid.UUID, req *md.User) error {
	const op = "sso.UpdateUser.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	res, err := tx.Exec(
		userUpdateQ,
		req.Name,
		req.Password,
		req.Email,
		req.Avatar,
		req.Address,
		req.Phone,
		id,
	)
	if err != nil {
		return err
	}

	if aff, _ := res.RowsAffected(); aff == 0 {
		return repo.ErrNotFound
	}

	if _, err = tx.Exec(userDeletePermQ, id); err != nil {
		tx.Rollback()
		return err
	}

	for _, v := range req.Permissions {
		if _, err := tx.Exec(
			userCreatePermQ,
			id, v.ID, v.Value,
		); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	const op = "sso.DeleteUser.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, userDeleteQ, id)
	if err != nil {
		return err
	}

	if aff, _ := res.RowsAffected(); aff == 0 {
		return repo.ErrNotFound
	}
	return nil
}
