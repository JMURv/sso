package ctrl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/cache"
	"github.com/JMURv/sso/internal/dto"
	repo "github.com/JMURv/sso/internal/repository"
	"github.com/JMURv/sso/pkg/consts"
	"github.com/JMURv/sso/pkg/model"
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

func (c *Controller) Authenticate(ctx context.Context, req *dto.EmailAndPasswordRequest) (*dto.EmailAndPasswordResponse, error) {
	const op = "sso.Authenticate.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	u, err := c.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.Error(err), zap.String("op", op),
		)
		return nil, ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.Error(err), zap.String("op", op),
		)
		return nil, err
	}

	accessToken, err := c.auth.NewToken(u, auth.AccessTokenDuration)
	if err != nil {
		zap.L().Debug(
			"failed to generate access token",
			zap.Error(err), zap.String("op", op),
		)
		return nil, ErrWhileGeneratingToken
	}

	return &dto.EmailAndPasswordResponse{
		Token: accessToken,
	}, nil
}

func (c *Controller) ParseClaims(ctx context.Context, token string) (map[string]any, error) {
	const op = "sso.ParseClaims.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	claims, err := c.auth.VerifyToken(token)
	if err != nil {
		zap.L().Debug("invalid token", zap.Error(err))
		return nil, err
	}

	if _, ok := claims["uid"].(string); !ok {
		zap.L().Debug("failed to parse uid", zap.String("op", op))
		return nil, ErrParseUUID
	}

	return claims, nil
}

func (c *Controller) GetUserByToken(ctx context.Context, token string) (*model.User, error) {
	const op = "sso.GetUserByToken.ctrl"
	span, _ := opentracing.StartSpanFromContext(ctx, op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	claims, err := c.ParseClaims(ctx, token)
	if err != nil {
		zap.L().Debug("invalid token", zap.Error(err))
		return nil, err
	}

	uid, err := uuid.Parse(claims["uid"].(string))
	if err != nil {
		zap.L().Debug("failed to parse uuid", zap.String("op", op))
		return nil, ErrParseUUID
	}

	cached := &model.User{}
	cacheKey := fmt.Sprintf(userCacheKey, uid)
	if err := c.cache.GetToStruct(ctx, cacheKey, cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.GetUserByID(ctx, uid)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.Error(err), zap.String("op", op),
			zap.String("id", uid.String()),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
		if err = c.cache.Set(ctx, consts.DefaultCacheTime, cacheKey, bytes); err != nil {
			zap.L().Debug(
				"failed to set to cache",
				zap.Error(err), zap.String("op", op),
				zap.String("id", uid.String()),
			)
		}
	}

	return res, nil
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
