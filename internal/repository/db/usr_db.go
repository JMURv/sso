package db

import (
	"context"
	"database/sql"
	repo "github.com/JMURv/sso/internal/repository"
	md "github.com/JMURv/sso/pkg/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

func (r *Repository) SearchUser(ctx context.Context, query string, page, size int) (*md.PaginatedUser, error) {
	const op = "sso.SearchUser.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM users WHERE name ILIKE $1 OR email ILIKE $2`, "%"+query+"%", "%"+query+"%",
	).Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.QueryContext(
		ctx,
		`SELECT 
    	 u.id, 
    	 u.name, 
    	 u.password, 
    	 u.email, 
    	 u.avatar, 
    	 u.address, 
    	 u.phone, 
    	 u.created_at, 
    	 u.updated_at,
    	 ARRAY_AGG(
    	     ARRAY[p.id::TEXT, p.name, up.value::TEXT]
		 ) FILTER (WHERE p.id IS NOT NULL) AS permissions
		 FROM users u
		 LEFT JOIN user_permission up ON up.user_id = u.id
         LEFT JOIN permission p ON p.id = up.permission_id
		 WHERE u.name ILIKE $1 OR u.email ILIKE $2
		 GROUP BY u.id, u.name
		 ORDER BY u.name DESC 
		 LIMIT $3 OFFSET $4`, "%"+query+"%", "%"+query+"%", size, (page-1)*size,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*md.User, 0, size)
	for rows.Next() {
		user := &md.User{}
		perms := make([][]string, 0, 5)
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

		for i := range perms {
			id, err := strconv.ParseUint(perms[i][0], 10, 64)
			if err != nil {
				return nil, err
			}
			user.Permissions = append(
				user.Permissions, md.Permission{
					ID:    id,
					Name:  perms[i][1],
					Value: perms[i][2] == "true",
				},
			)
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
	if err := r.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.QueryContext(
		ctx,
		`SELECT 
    	 u.id, 
    	 u.name, 
    	 u.password, 
    	 u.email, 
    	 u.avatar, 
    	 u.address, 
    	 u.phone, 
    	 u.created_at, 
    	 u.updated_at,
    	 COALESCE(
			ARRAY_AGG(
				ARRAY[p.id::TEXT, p.name, up.value::TEXT]
				) FILTER (WHERE p.id IS NOT NULL), '{}'
		 ) AS permissions
		 FROM users u
		 LEFT JOIN user_permission up ON up.user_id = u.id
         LEFT JOIN permission p ON p.id = up.permission_id
		 GROUP BY u.id, u.created_at
		 ORDER BY created_at DESC 
		 LIMIT $1 OFFSET $2`, size, (page-1)*size,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*md.User, 0, size)
	for rows.Next() {
		user := &md.User{}
		perms := make([][]string, 0, 5)
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

		for i := range perms {
			id, err := strconv.ParseUint(perms[i][0], 10, 64)
			if err != nil {
				return nil, err
			}
			user.Permissions = append(
				user.Permissions, md.Permission{
					ID:    id,
					Name:  perms[i][1],
					Value: perms[i][2] == "true",
				},
			)
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
	perms := make([][]string, 0, 5)
	err := r.conn.QueryRow(
		`SELECT 
    	 u.id, 
    	 u.name, 
    	 u.password, 
    	 u.email, 
    	 u.avatar, 
    	 u.address, 
    	 u.phone, 
    	 u.created_at, 
    	 u.updated_at,
    	 COALESCE(
			ARRAY_AGG(
				ARRAY[p.id::TEXT, p.name, up.value::TEXT]
				) FILTER (WHERE p.id IS NOT NULL), '{}'
		 ) AS permissions
		 FROM users u
		 LEFT JOIN user_permission up ON up.user_id = u.id
         LEFT JOIN permission p ON p.id = up.permission_id
		 WHERE u.id = $1
		 GROUP BY u.id, u.created_at`,
		userID,
	).Scan(
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

	for i := range perms {
		id, err := strconv.ParseUint(perms[i][0], 10, 64)
		if err != nil {
			return nil, err
		}
		res.Permissions = append(
			res.Permissions, md.Permission{
				ID:    id,
				Name:  perms[i][1],
				Value: perms[i][2] == "true",
			},
		)
	}

	if err == sql.ErrNoRows {
		return nil, repo.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*md.User, error) {
	const op = "sso.GetUserByEmail.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := &md.User{}
	err := r.conn.QueryRowContext(
		ctx, `
		SELECT id, name, password, email, avatar, address, phone, created_at, updated_at
		FROM users
		WHERE email = $1
		`, email,
	).
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

	var idx uuid.UUID
	if err := r.conn.
		QueryRow(`SELECT id FROM users WHERE email=$1`, req.Email).
		Scan(&idx); err == nil {
		return uuid.Nil, repo.ErrAlreadyExists
	} else if err != nil && err != sql.ErrNoRows {
		return uuid.Nil, err
	}

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
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO users (name, password, email, avatar, address, phone) 
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		req.Name,
		req.Password,
		req.Email,
		req.Avatar,
		req.Address,
		req.Phone,
	).Scan(&id)

	if err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}

	if len(req.Permissions) > 0 {
		for _, v := range req.Permissions {
			if _, err := tx.Exec(
				`INSERT INTO user_permission (user_id, permission_id, value) VALUES ($1, $2, $3)`,
				id, v.ID, v.Value,
			); err != nil {
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
		`UPDATE users 
		 SET name = $1, password = $2, email = $3, avatar = $4, address = $5, phone = $6 
		 WHERE id = $7`,
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

	aff, _ := res.RowsAffected()
	if aff == 0 {
		return repo.ErrNotFound
	}

	_, err = tx.Exec(`DELETE FROM user_permission WHERE user_id = $1`, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, v := range req.Permissions {
		if _, err := tx.Exec(
			`INSERT INTO user_permission (user_id, permission_id, value) VALUES ($1, $2, $3)`,
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

	res, err := r.conn.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if aff, _ := res.RowsAffected(); aff == 0 {
		return repo.ErrNotFound
	}
	return nil
}
