package ctrl

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	repo "github.com/JMURv/sso/internal/repository"
	"github.com/JMURv/sso/pkg/consts"
	"github.com/JMURv/sso/pkg/model"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type userRepo interface {
	ListUsers(ctx context.Context, page, size int) (*model.PaginatedUser, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)

	CreateUser(ctx context.Context, req *model.User) (uuid.UUID, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *model.User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	SearchUser(ctx context.Context, query string, page int, size int) (*model.PaginatedUser, error)
}

const userCacheKey = "user:%v"
const usersSearchCacheKey = "users-search:%v:%v:%v"
const usersListKey = "users-list:%v:%v"
const userPattern = "users-*"

func (c *Controller) IsUserExist(ctx context.Context, email string) (isExist bool, err error) {
	const op = "sso.IsUserExist.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	_, err = c.repo.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		return false, nil
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.Error(err), zap.String("op", op),
		)
		return true, err
	}

	return true, nil
}

func (c *Controller) SearchUser(ctx context.Context, query string, page, size int) (*model.PaginatedUser, error) {
	const op = "users.UserSearch.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	cached := &model.PaginatedUser{}
	cacheKey := fmt.Sprintf(usersSearchCacheKey, query, page, size)
	if err := c.cache.GetToStruct(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.SearchUser(ctx, query, page, size)
	if err != nil {
		zap.L().Debug(
			"failed to search users",
			zap.Error(err), zap.String("op", op),
			zap.String("query", query),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
		if err = c.cache.Set(ctx, consts.DefaultCacheTime, cacheKey, bytes); err != nil {
			zap.L().Debug(
				"failed to set to cache",
				zap.Error(err), zap.String("op", op),
				zap.String("query", query),
			)
		}
	}

	return res, nil
}

func (c *Controller) ListUsers(ctx context.Context, page, size int) (*model.PaginatedUser, error) {
	const op = "users.GetUsersList.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	cached := &model.PaginatedUser{}
	cacheKey := fmt.Sprintf(usersListKey, page, size)
	if err := c.cache.GetToStruct(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.ListUsers(ctx, page, size)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to list users",
			zap.Error(err), zap.String("op", op),
			zap.Int("page", page), zap.Int("size", size),
		)
		return nil, ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to list users",
			zap.Error(err), zap.String("op", op),
			zap.Int("page", page), zap.Int("size", size),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
		if err = c.cache.Set(ctx, consts.DefaultCacheTime, cacheKey, bytes); err != nil {
			zap.L().Debug(
				"failed to set to cache",
				zap.Error(err), zap.String("op", op),
				zap.Int("page", page), zap.Int("size", size),
			)
		}
	}
	return res, nil
}

func (c *Controller) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	const op = "users.GetUserByID.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	cached := &model.User{}
	cacheKey := fmt.Sprintf(userCacheKey, userID)
	if err := c.cache.GetToStruct(ctx, cacheKey, cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.GetUserByID(ctx, userID)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.Error(err), zap.String("op", op),
			zap.String("id", userID.String()),
		)
		return nil, ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.Error(err), zap.String("op", op),
			zap.String("id", userID.String()),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
		if err = c.cache.Set(ctx, consts.DefaultCacheTime, cacheKey, bytes); err != nil {
			zap.L().Debug(
				"failed to set to cache",
				zap.Error(err), zap.String("op", op),
				zap.String("id", userID.String()),
			)
		}
	}

	return res, nil
}

func (c *Controller) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	const op = "users.GetUserByEmail.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	cached := &model.User{}
	cacheKey := fmt.Sprintf(userCacheKey, email)
	if err := c.cache.GetToStruct(ctx, cacheKey, cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.Error(err), zap.String("op", op),
			zap.String("email", email),
		)
		return nil, ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.Error(err), zap.String("op", op),
			zap.String("email", email),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
		if err = c.cache.Set(ctx, consts.DefaultCacheTime, cacheKey, bytes); err != nil {
			zap.L().Debug(
				"failed to set to cache",
				zap.Error(err), zap.String("op", op),
				zap.String("email", email),
			)
		}
	}

	return res, nil
}

func (c *Controller) CreateUser(ctx context.Context, u *model.User, fileName string, bytes []byte) (uuid.UUID, string, string, error) {
	const op = "users.CreateUser.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	id, err := c.repo.CreateUser(ctx, u)
	if err != nil && errors.Is(err, repo.ErrAlreadyExists) {
		zap.L().Debug(
			"user already exists",
			zap.Error(err), zap.String("op", op),
		)
		return uuid.Nil, "", "", ErrAlreadyExists
	} else if err != nil {
		zap.L().Debug(
			"failed to create user",
			zap.Error(err), zap.String("op", op),
		)
		return uuid.Nil, "", "", err
	}

	if bytes, err := json.Marshal(u); err == nil {
		if err = c.cache.Set(ctx, consts.DefaultCacheTime, fmt.Sprintf(userCacheKey, id), bytes); err != nil {
			zap.L().Debug(
				"failed to set to cache",
				zap.Error(err), zap.String("op", op),
			)
		}
	}

	if fileName != "" && bytes != nil {
		if err = c.smtp.SendOptFile(ctx, u.Email, fileName, bytes); err != nil {
			zap.L().Debug(
				"failed to send email",
				zap.Error(err),
				zap.String("op", op),
			)
		}
	}

	accessToken, err := c.auth.NewToken(u, auth.AccessTokenDuration)
	if err != nil {
		zap.L().Debug(
			"failed to create access token",
			zap.Error(err), zap.String("op", op),
		)
		return id, "", "", ErrWhileGeneratingToken
	}

	refreshToken, err := c.auth.NewToken(u, auth.RefreshTokenDuration)
	if err != nil {
		zap.L().Debug(
			"failed to create refresh token",
			zap.Error(err), zap.String("op", op),
		)
		return id, "", "", ErrWhileGeneratingToken
	}

	if err = c.smtp.SendUserCredentials(ctx, u.Email, u.Password); err != nil {
		return id, accessToken, refreshToken, err
	}

	go c.cache.InvalidateKeysByPattern(ctx, userPattern)
	return id, accessToken, refreshToken, nil
}

func (c *Controller) UpdateUser(ctx context.Context, id uuid.UUID, req *model.User) error {
	const op = "users.UpdateUser.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	err := c.repo.UpdateUser(ctx, id, req)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.Error(err), zap.String("op", op),
			zap.String("id", id.String()),
		)
		return ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to update user",
			zap.Error(err), zap.String("op", op),
			zap.String("id", id.String()),
		)
		return err
	}

	if err := c.cache.Delete(ctx, fmt.Sprintf(userCacheKey, id)); err != nil {
		zap.L().Debug(
			"failed to delete from cache",
			zap.Error(err), zap.String("op", op),
			zap.String("id", id.String()),
		)
	}

	go c.cache.InvalidateKeysByPattern(ctx, userPattern)
	return nil
}

func (c *Controller) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const op = "users.DeleteUser.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	if err := c.repo.DeleteUser(ctx, userID); err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.Error(err), zap.String("op", op),
			zap.String("id", userID.String()),
		)
		return ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to delete user",
			zap.Error(err), zap.String("op", op),
			zap.String("id", userID.String()),
		)
		return err
	}

	if err := c.cache.Delete(ctx, fmt.Sprintf(userCacheKey, userID)); err != nil {
		zap.L().Debug(
			"failed to delete from cache",
			zap.Error(err), zap.String("op", op),
			zap.String("id", userID.String()),
		)
	}

	go c.cache.InvalidateKeysByPattern(ctx, userPattern)
	return nil
}
