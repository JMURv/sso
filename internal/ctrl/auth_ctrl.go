package ctrl

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/cache"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/repo"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"math/rand"
	"strconv"
	"time"
)

type authRepo interface {
	CreateToken(ctx context.Context, userID uuid.UUID, hashedT string, expiresAt time.Time) error
	IsTokenValid(ctx context.Context, userID uuid.UUID, token string) (bool, error)
	RevokeAllTokens(ctx context.Context, userID uuid.UUID) error
}

const codeCacheKey = "code:%v"
const recoveryCacheKey = "recovery:%v"

func (c *Controller) Authenticate(ctx context.Context, req *dto.EmailAndPasswordRequest) (*dto.EmailAndPasswordResponse, error) {
	const op = "auth.Authenticate.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := c.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.String("op", op),
			zap.Any("req", req),
			zap.Error(err),
		)
		return nil, ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.String("op", op),
			zap.Any("req", req),
			zap.Error(err),
		)
		return nil, err
	}

	access, refresh, err := auth.Au.GenPair(ctx, res.ID, res.Permissions)
	if err != nil {
		return nil, err
	}

	if err = c.repo.CreateToken(ctx, res.ID, refresh, time.Now().Add(auth.RefreshTokenDuration)); err != nil {
		zap.L().Debug(
			"Failed to save token",
			zap.String("op", op),
			zap.String("refresh", refresh),
			zap.Error(err),
		)
		return nil, err
	}

	return &dto.EmailAndPasswordResponse{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (c *Controller) Refresh(ctx context.Context, req *dto.RefreshRequest) (*dto.RefreshResponse, error) {
	const op = "auth.Refresh.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	claims, err := auth.Au.ParseClaims(ctx, req.Refresh)
	if err != nil {
		return nil, err
	}

	isValid, err := c.repo.IsTokenValid(ctx, claims.UID, req.Refresh)
	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, auth.ErrTokenRevoked
	}

	access, refresh, err := auth.Au.GenPair(ctx, claims.UID, claims.Roles)
	if err != nil {
		return nil, err
	}

	if err = c.repo.RevokeAllTokens(ctx, claims.UID); err != nil {
		zap.L().Debug(
			"Failed to revoke tokens",
			zap.String("op", op),
			zap.Any("claims", claims),
			zap.Error(err),
		)
		return nil, err
	}

	if err = c.repo.CreateToken(ctx, claims.UID, refresh, time.Now().Add(auth.RefreshTokenDuration)); err != nil {
		zap.L().Debug(
			"Failed to save token",
			zap.String("op", op),
			zap.String("refresh", refresh),
			zap.Any("claims", claims),
			zap.Error(err),
		)
		return nil, err
	}

	return &dto.RefreshResponse{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (c *Controller) ParseClaims(ctx context.Context, token string) (*auth.Claims, error) {
	const op = "auth.ParseClaims.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	claims, err := auth.Au.ParseClaims(ctx, token)
	if err != nil {
		zap.L().Debug("invalid token", zap.Error(err))
		return nil, err
	}

	return claims, nil
}

func (c *Controller) Logout(ctx context.Context, uid uuid.UUID) error {
	const op = "auth.Logout.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	if err := c.repo.RevokeAllTokens(ctx, uid); err != nil {
		zap.L().Debug(
			"Failed to revoke tokens",
			zap.String("op", op),
			zap.Any("uid", uid),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (c *Controller) CheckForgotPasswordEmail(ctx context.Context, req *dto.CheckForgotPasswordEmailRequest) error {
	const op = "auth.CheckForgotPasswordEmail.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	storedCode, err := c.cache.GetInt(ctx, fmt.Sprintf(recoveryCacheKey, req.ID))
	if err != nil {
		zap.L().Debug(
			"Error getting from cache",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	if storedCode != req.Code {
		return ErrCodeIsNotValid
	}

	u, err := c.repo.GetUserByID(ctx, req.ID)
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

	newPass, err := auth.Au.Hash(req.Password)
	if err != nil {
		return err
	}

	u.Password = newPass
	if err = c.repo.UpdateUser(ctx, req.ID, u); err != nil {
		zap.L().Debug(
			"Error updating user",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	c.cache.Delete(ctx, fmt.Sprintf(userCacheKey, req.ID))
	return nil
}

func (c *Controller) SendForgotPasswordEmail(ctx context.Context, email string) error {
	const op = "auth.SendForgotPasswordEmail.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.String("op", op),
			zap.Error(err),
		)
		return auth.ErrInvalidCredentials
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	code := rand.Intn(9999-1000+1) + 1000

	if err = c.smtp.SendForgotPasswordEmail(ctx, strconv.Itoa(code), res.ID.String(), email); err != nil {
		zap.L().Debug(
			"failed to send email",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}

	c.cache.Set(ctx, time.Minute*15, fmt.Sprintf(recoveryCacheKey, res.ID.String()), code)
	return nil
}

func (c *Controller) SendLoginCode(ctx context.Context, email, password string) error {
	const op = "auth.SendLoginCode.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find user",
			zap.String("op", op),
			zap.Error(err),
		)
		return auth.ErrInvalidCredentials
	} else if err != nil {
		zap.L().Debug(
			"failed to get user",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}

	if err = auth.Au.ComparePasswords([]byte(res.Password), []byte(password)); err != nil {
		return err
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	code := rand.Intn(9999-1000+1) + 1000

	if err = c.smtp.SendLoginEmail(ctx, code, email); err != nil {
		zap.L().Debug(
			"failed to send an email",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}

	c.cache.Set(ctx, time.Minute*15, fmt.Sprintf(codeCacheKey, email), code)
	return nil
}

func (c *Controller) CheckLoginCode(ctx context.Context, req *dto.CheckLoginCodeRequest) (*dto.CheckLoginCodeResponse, error) {
	const op = "auth.CheckLoginCode.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	storedCode, err := c.cache.GetInt(ctx, fmt.Sprintf(codeCacheKey, req.Email))
	if err != nil && errors.Is(err, cache.ErrNotFoundInCache) {
		return nil, ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get from cache",
			zap.Error(err), zap.String("op", op),
		)
		return nil, err
	}

	if storedCode != req.Code {
		return nil, ErrCodeIsNotValid
	}

	res, err := c.repo.GetUserByEmail(ctx, req.Email)
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

	access, refresh, err := auth.Au.GenPair(ctx, res.ID, res.Permissions)
	if err != nil {
		return nil, ErrWhileGeneratingToken
	}

	return &dto.CheckLoginCodeResponse{
		Access:  access,
		Refresh: refresh,
	}, nil
}
