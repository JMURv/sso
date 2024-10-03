package db

import (
	"context"
	"database/sql"
	"fmt"
	repo "github.com/JMURv/sso/internal/repository"
	conf "github.com/JMURv/sso/pkg/config"
	"github.com/JMURv/sso/pkg/model"
	dbutils "github.com/JMURv/sso/pkg/utils/db"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	conn *sql.DB
}

func New(conf *conf.DBConfig) *Repository {
	conn, err := sql.Open("postgres", fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.Database,
	))
	if err != nil {
		zap.L().Fatal("Failed to connect to the database", zap.Error(err))
	}

	if err := conn.Ping(); err != nil {
		zap.L().Fatal("Failed to ping the database", zap.Error(err))
	}

	if err := dbutils.ApplyMigrations(conn, conf); err != nil {
		zap.L().Fatal("Failed to apply migrations", zap.Error(err))
	}

	model.MustPrecreateUsers(conn)
	return &Repository{conn: conn}
}

func (r *Repository) Close() error {
	return r.conn.Close()
}

func (r *Repository) UserSearch(ctx context.Context, query string, page, size int) (*utils.PaginatedData, error) {
	const op = "users.UserSearch.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE name ILIKE $1 OR email ILIKE $2`, "%"+query+"%", "%"+query+"%").Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.QueryContext(ctx,
		`SELECT id, name, password, email, avatar, address, phone, created_at, updated_at 
		 FROM users 
		 WHERE name ILIKE $1 OR email ILIKE $2 
		 ORDER BY created_at DESC 
		 LIMIT $3 OFFSET $4`, "%"+query+"%", "%"+query+"%", size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*model.User
	for rows.Next() {
		user := &model.User{}
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
		); err != nil {
			return nil, err
		}
		res = append(res, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((count + int64(size) - 1) / int64(size))
	hasNextPage := page < totalPages
	return &utils.PaginatedData{
		Data:        res,
		Count:       count,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNextPage: hasNextPage,
	}, nil
}

func (r *Repository) ListUsers(ctx context.Context, page, size int) (*utils.PaginatedData, error) {
	const op = "users.ListUsers.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return nil, err
	}

	rows, err := r.conn.QueryContext(ctx,
		`SELECT id, name, password, email, avatar, address, phone, created_at, updated_at 
		 FROM users 
		 ORDER BY created_at DESC 
		 LIMIT $1 OFFSET $2`, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*model.User
	for rows.Next() {
		user := &model.User{}
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
		); err != nil {
			return nil, err
		}
		res = append(res, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((count + int64(size) - 1) / int64(size))
	hasNextPage := page < totalPages
	return &utils.PaginatedData{
		Data:        res,
		Count:       count,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNextPage: hasNextPage,
	}, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	const op = "users.GetUserByID.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := &model.User{}
	err := r.conn.QueryRowContext(ctx, `
		SELECT id, name, password, email, avatar, address, phone, created_at, updated_at
		FROM users
		WHERE id = $1
		`, userID).
		Scan(&res.ID,
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

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	const op = "users.GetUserByEmail.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := &model.User{}
	err := r.conn.QueryRowContext(ctx, `
		SELECT id, name, password, email, avatar, address, phone, created_at, updated_at
		FROM users
		WHERE email = $1
		`, email).
		Scan(&res.ID,
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

func (r *Repository) CreateUser(ctx context.Context, req *model.User) (uuid.UUID, error) {
	const op = "users.CreateUser.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var idx uuid.UUID
	err := r.conn.QueryRow(`SELECT id FROM users WHERE id=$1 OR email=$2`, req.ID, req.Email).Scan(&idx)
	if err == nil {
		return uuid.Nil, repo.ErrAlreadyExists
	} else if err != nil && err != sql.ErrNoRows {
		return uuid.Nil, err
	}

	if req.Password == "" {
		return uuid.Nil, repo.ErrPasswordIsRequired
	}

	password, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, repo.ErrGeneratingPassword
	}
	req.Password = string(password)

	if req.Name == "" {
		return uuid.Nil, repo.ErrUsernameIsRequired
	}

	if req.Email == "" {
		return uuid.Nil, repo.ErrEmailIsRequired
	}

	var id uuid.UUID
	err = r.conn.QueryRowContext(
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
		return uuid.Nil, err
	}

	return id, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id uuid.UUID, req *model.User) error {
	const op = "users.UpdateUser.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx,
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

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return repo.ErrNotFound
	}

	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	const op = "users.DeleteUser.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
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
