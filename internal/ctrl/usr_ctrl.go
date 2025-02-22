package ctrl

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type userRepo interface {
	SearchUser(ctx context.Context, query string, page int, size int) (*dto.PaginatedUserResponse, error)
	ListUsers(ctx context.Context, page, size int) (*dto.PaginatedUserResponse, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*md.User, error)
	GetUserByEmail(ctx context.Context, email string) (*md.User, error)
	CreateUser(ctx context.Context, req *md.User) (uuid.UUID, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *md.User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

const userCacheKey = "user:%v"
const usersSearchCacheKey = "users-search:%v:%v:%v"
const usersListKey = "users-list:%v:%v"
const userPattern = "users-*"

func (c *Controller) IsUserExist(ctx context.Context, email string) (*dto.ExistsUserResponse, error) {
	const op = "users.IsUserExist.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := &dto.ExistsUserResponse{Exists: false}
	_, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		return res, nil
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.String("op", op),
			zap.Error(err),
		)
		return nil, err
	}

	res.Exists = true
	return res, nil
}

func (c *Controller) SearchUser(ctx context.Context, query string, page, size int) (*dto.PaginatedUserResponse, error) {
	const op = "users.SearchUser.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	cached := &dto.PaginatedUserResponse{}
	cacheKey := fmt.Sprintf(usersSearchCacheKey, query, page, size)
	if err := c.cache.GetToStruct(ctx, cacheKey, cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.SearchUser(ctx, query, page, size)
	if err != nil {
		zap.L().Debug(
			"failed to search users",
			zap.String("op", op),
			zap.String("query", query),
			zap.Error(err),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
		c.cache.Set(ctx, config.DefaultCacheTime, cacheKey, bytes)
	}

	return res, nil
}

func (c *Controller) ListUsers(ctx context.Context, page, size int) (*dto.PaginatedUserResponse, error) {
	const op = "users.ListUsers.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	cached := &dto.PaginatedUserResponse{}
	cacheKey := fmt.Sprintf(usersListKey, page, size)
	if err := c.cache.GetToStruct(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.ListUsers(ctx, page, size)
	if err != nil {
		zap.L().Debug(
			"failed to list users",
			zap.String("op", op),
			zap.Int("page", page), zap.Int("size", size),
			zap.Error(err),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
		c.cache.Set(ctx, config.DefaultCacheTime, cacheKey, bytes)
	}
	return res, nil
}

func (c *Controller) GetUserByID(ctx context.Context, userID uuid.UUID) (*md.User, error) {
	const op = "users.GetUserByID.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	cached := &md.User{}
	cacheKey := fmt.Sprintf(userCacheKey, userID)
	if err := c.cache.GetToStruct(ctx, cacheKey, cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.GetUserByID(ctx, userID)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.String("op", op),
			zap.String("id", userID.String()),
			zap.Error(err),
		)
		return nil, ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.String("op", op),
			zap.String("id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
		c.cache.Set(ctx, config.DefaultCacheTime, cacheKey, bytes)
	}
	return res, nil
}

func (c *Controller) GetUserByEmail(ctx context.Context, email string) (*md.User, error) {
	const op = "users.GetUserByEmail.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	cached := &md.User{}
	cacheKey := fmt.Sprintf(userCacheKey, email)
	if err := c.cache.GetToStruct(ctx, cacheKey, cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.String("op", op),
			zap.String("email", email),
			zap.Error(err),
		)
		return nil, ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.String("op", op),
			zap.String("email", email),
			zap.Error(err),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
		c.cache.Set(ctx, config.DefaultCacheTime, cacheKey, bytes)
	}
	return res, nil
}

func (c *Controller) CreateUser(ctx context.Context, u *md.User) (*dto.CreateUserResponse, error) {
	const op = "users.CreateUser.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	hash, err := auth.Au.Hash(u.Password)
	if err != nil {
		return nil, err
	}
	u.Password = hash

	id, err := c.repo.CreateUser(ctx, u)
	if err != nil && errors.Is(err, repo.ErrAlreadyExists) {
		zap.L().Debug(
			"user already exists",
			zap.Error(err), zap.String("op", op),
		)
		return nil, ErrAlreadyExists
	} else if err != nil {
		zap.L().Debug(
			"failed to create user",
			zap.Error(err), zap.String("op", op),
		)
		return nil, err
	}

	u.ID = id
	if bytes, err := json.Marshal(u); err == nil {
		c.cache.Set(ctx, config.DefaultCacheTime, fmt.Sprintf(userCacheKey, id), bytes)
	}
	go c.cache.InvalidateKeysByPattern(ctx, userPattern)
	return &dto.CreateUserResponse{
		ID: id,
	}, nil
}

func (c *Controller) UpdateUser(ctx context.Context, id uuid.UUID, req *md.User) error {
	const op = "users.UpdateUser.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	err := c.repo.UpdateUser(ctx, id, req)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.String("op", op),
			zap.String("id", id.String()),
			zap.Error(err),
		)
		return ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to update user",
			zap.String("op", op),
			zap.String("id", id.String()),
			zap.Error(err),
		)
		return err
	}

	c.cache.Delete(ctx, fmt.Sprintf(userCacheKey, id))
	go c.cache.InvalidateKeysByPattern(ctx, userPattern)
	return nil
}

func (c *Controller) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const op = "users.DeleteUser.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	err := c.repo.DeleteUser(ctx, userID)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.String("op", op),
			zap.String("id", userID.String()),
			zap.Error(err),
		)
		return ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to delete user",
			zap.String("op", op),
			zap.String("id", userID.String()),
			zap.Error(err),
		)
		return err
	}

	c.cache.Delete(ctx, fmt.Sprintf(userCacheKey, userID))
	go c.cache.InvalidateKeysByPattern(ctx, userPattern)
	return nil
}
