package ctrl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/cache"
	repo "github.com/JMURv/sso/internal/repository"
	"github.com/JMURv/sso/pkg/consts"
	"github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"strconv"
	"time"
)

const codeCacheKey = "code:%v"
const recoveryCacheKey = "recovery:%v"

const userCacheKey = "user:%v"
const usersSearchCacheKey = "users-search:%v:%v:%v"
const usersListKey = "users-list:%v:%v"
const invalidateUserRelatedCachePattern = "users-*"

func (c *Controller) invalidateUserRelatedCache() {
	ctx := context.Background()
	if err := c.cache.InvalidateKeysByPattern(ctx, invalidateUserRelatedCachePattern); err != nil {
		zap.L().Debug("failed to invalidate cache", zap.String("key", invalidateUserRelatedCachePattern), zap.Error(err))
	}
}

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

func (c *Controller) UserSearch(ctx context.Context, query string, page, size int) (*utils.PaginatedData, error) {
	const op = "users.UserSearch.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	cached := &utils.PaginatedData{}
	cacheKey := fmt.Sprintf(usersSearchCacheKey, query, page, size)
	if err := c.cache.GetToStruct(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.UserSearch(ctx, query, page, size)
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

func (c *Controller) ListUsers(ctx context.Context, page, size int) (*utils.PaginatedData, error) {
	const op = "users.GetUsersList.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	cached := &utils.PaginatedData{}
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
			zap.String("userID", userID.String()),
		)
		return nil, ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.Error(err), zap.String("op", op),
			zap.String("userID", userID.String()),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
		if err = c.cache.Set(ctx, consts.DefaultCacheTime, cacheKey, bytes); err != nil {
			zap.L().Debug(
				"failed to set to cache",
				zap.Error(err), zap.String("op", op),
				zap.String("userID", userID.String()),
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
		err := c.smtp.SendOptFile(ctx, u.Email, fileName, bytes)
		if err != nil {
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

	go c.invalidateUserRelatedCache()
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

	go c.invalidateUserRelatedCache()
	return nil
}

func (c *Controller) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const op = "users.DeleteUser.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	if err := c.repo.DeleteUser(ctx, userID); err != nil {
		zap.L().Debug(
			"failed to delete user",
			zap.Error(err), zap.String("op", op),
			zap.String("userID", userID.String()),
		)
		return err
	}

	if err := c.cache.Delete(ctx, fmt.Sprintf(userCacheKey, userID)); err != nil {
		zap.L().Debug(
			"failed to delete from cache",
			zap.Error(err), zap.String("op", op),
			zap.String("userID", userID.String()),
		)
	}

	go c.invalidateUserRelatedCache()
	return nil
}

func (c *Controller) SendSupportEmail(ctx context.Context, uid uuid.UUID, theme, text string) error {
	const op = "sso.SendSupportEmail.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	u, err := c.repo.GetUserByID(ctx, uid)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"Error find user",
			zap.Error(err), zap.String("op", op),
		)
		return ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"Error getting user",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	if err = c.smtp.SendSupportEmail(ctx, u, theme, text); err != nil {
		zap.L().Debug(
			"Error sending email",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	return nil
}

func (c *Controller) CheckForgotPasswordEmail(ctx context.Context, password string, uid uuid.UUID, code int) error {
	const op = "sso.CheckForgotPasswordEmail.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	storedCode, err := c.cache.GetCode(ctx, fmt.Sprintf(recoveryCacheKey, uid))
	if err != nil {
		zap.L().Debug(
			"Error getting from cache",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	if storedCode != code {
		return ErrCodeIsNotValid
	}

	u, err := c.repo.GetUserByID(ctx, uid)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"Error find user",
			zap.Error(err), zap.String("op", op),
		)
		return ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"Error getting user",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	newPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return repo.ErrGeneratingPassword
	}

	u.Password = string(newPassword)
	if err = c.repo.UpdateUser(ctx, uid, u); err != nil {
		zap.L().Debug(
			"Error updating user",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	if err = c.cache.Delete(ctx, fmt.Sprintf(userCacheKey, uid)); err != nil {
		zap.L().Debug(
			"Error deleting from cache",
			zap.Error(err), zap.String("op", op),
		)
	}

	return nil
}

func (c *Controller) SendForgotPasswordEmail(ctx context.Context, email string) error {
	const op = "sso.SendForgotPasswordEmail.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	u, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.Error(err), zap.String("op", op),
		)
		return ErrInvalidCredentials
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	code := rand.Intn(9999-1000+1) + 1000

	if err = c.cache.Set(ctx, time.Minute*15, fmt.Sprintf(recoveryCacheKey, u.ID.String()), code); err != nil {
		zap.L().Debug(
			"failed to set to cache",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	if err = c.smtp.SendForgotPasswordEmail(ctx, strconv.Itoa(code), u.ID.String(), email); err != nil {
		zap.L().Debug(
			"failed to send email",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	return nil
}

func (c *Controller) SendLoginCode(ctx context.Context, email, password string) error {
	const op = "sso.SendLoginCode.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	u, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.Error(err), zap.String("op", op),
		)
		return ErrInvalidCredentials
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	code := rand.Intn(9999-1000+1) + 1000

	if err = c.cache.Set(ctx, time.Minute*15, fmt.Sprintf(codeCacheKey, email), code); err != nil {
		zap.L().Debug(
			"failed to set to cache",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	if err = c.smtp.SendLoginEmail(ctx, code, email); err != nil {
		zap.L().Debug(
			"failed to send an email",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	return nil
}

func (c *Controller) CheckLoginCode(ctx context.Context, email string, code int) (string, string, error) {
	const op = "sso.CheckLoginCode.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	storedCode, err := c.cache.GetCode(ctx, fmt.Sprintf(codeCacheKey, email))
	if err != nil && errors.Is(err, cache.ErrNotFoundInCache) {
		return "", "", ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get from cache",
			zap.Error(err), zap.String("op", op),
		)
		return "", "", err
	}

	if storedCode != code {
		return "", "", ErrCodeIsNotValid
	}

	u, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.Error(err), zap.String("op", op),
		)
		return "", "", ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.Error(err), zap.String("op", op),
		)
		return "", "", err
	}

	accessToken, err := c.auth.NewToken(u, auth.AccessTokenDuration)
	if err != nil {
		zap.L().Debug(
			"failed to generate access token",
			zap.Error(err), zap.String("op", op),
		)
		return "", "", ErrWhileGeneratingToken
	}

	refreshToken, err := c.auth.NewToken(u, auth.RefreshTokenDuration)
	if err != nil {
		zap.L().Debug(
			"failed to generate refresh token",
			zap.Error(err), zap.String("op", op),
		)
		return "", "", ErrWhileGeneratingToken
	}
	return accessToken, refreshToken, nil
}
