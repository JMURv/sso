package db

import (
	"context"
	"errors"
	repo "github.com/JMURv/sso/internal/repository"
	"github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

func (r *Repository) UserSearch(ctx context.Context, query string, page, size int) (*utils.PaginatedData, error) {
	const op = "users.UserSearch.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var count int64
	if err := r.conn.
		Model(&model.User{}).
		Where("name ILIKE ? OR email ILIKE ?", "%"+query+"%", "%"+query+"%").
		Count(&count).
		Error; err != nil {
		return nil, err
	}
	totalPages := int((count + int64(size) - 1) / int64(size))
	hasNextPage := page < totalPages

	var users []*model.User
	if err := r.conn.
		Offset((page-1)*size).
		Limit(size).
		Where("name ILIKE ? OR email ILIKE ?", "%"+query+"%", "%"+query+"%").
		Find(&users).Error; err != nil {
		return nil, err
	}

	return &utils.PaginatedData{
		Data:        users,
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
	if err := r.conn.Model(&model.User{}).Count(&count).Error; err != nil {
		return nil, err
	}
	totalPages := int((count + int64(size) - 1) / int64(size))
	hasNextPage := page < totalPages

	var res []*model.User
	if err := r.conn.
		Offset((page - 1) * size).
		Limit(size).
		Order("created_at desc").
		Find(&res).
		Error; err != nil {
		return nil, err
	}

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

	var u model.User
	if err := r.conn.Where("id=?", userID).First(&u).Error; err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repo.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	const op = "users.GetUserByEmail.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var u model.User
	if err := r.conn.Where("email=?", email).First(&u).Error; err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repo.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateUser(ctx context.Context, u *model.User) error {
	const op = "users.CreateUser.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	if u.Password == "" {
		return repo.ErrPasswordIsRequired
	}

	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return repo.ErrGeneratingPassword
	}
	u.Password = string(password)

	if u.Name == "" {
		return repo.ErrUsernameIsRequired
	}

	if u.Email == "" {
		return repo.ErrEmailIsRequired
	}

	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	if err = r.conn.Create(&u).Error; err != nil && errors.Is(err, gorm.ErrDuplicatedKey) {
		return repo.ErrAlreadyExists
	} else if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateUser(ctx context.Context, userID uuid.UUID, newData *model.User) (*model.User, error) {
	const op = "users.UpdateUser.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	u, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if newData.Email != "" {
		u.Email = newData.Email
	}

	if newData.Password != "" {
		u.Password = newData.Password
	}

	if newData.Name != "" {
		u.Name = newData.Name
	}

	if newData.Avatar != "" {
		u.Avatar = newData.Avatar
	}

	if newData.Address != "" {
		u.Address = newData.Address
	}

	if newData.Phone != "" {
		u.Phone = newData.Phone

	}

	u.IsAdmin = newData.IsAdmin
	u.IsOpt = newData.IsOpt
	u.UpdatedAt = time.Now()
	r.conn.Save(&u)
	return u, nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const op = "users.DeleteUser.repo"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var u model.User
	if err := r.conn.Delete(&u, userID).Error; err != nil {
		return err
	}

	return nil
}
