package ctrl

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/cache"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"math/rand"
	"strconv"
	"time"
)

const codeCacheKey = "code:%v"
const recoveryCacheKey = "recovery:%v"

func (c *Controller) GenPair(ctx context.Context, d *dto.DeviceRequest, uid uuid.UUID, p []md.Permission) (dto.TokenPair, error) {
	const op = "auth.GenPair.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var res dto.TokenPair
	access, refresh, err := c.au.GenPair(ctx, uid, p)
	if err != nil {
		return res, err
	}

	hash, err := c.au.HashSHA256(refresh)
	if err != nil {
		return res, err
	}

	device := auth.GenerateDevice(d)
	if err = c.repo.CreateToken(ctx, uid, hash, time.Now().Add(auth.RefreshTokenDuration), &device); err != nil {
		zap.L().Debug(
			"Failed to save token",
			zap.String("op", op),
			zap.String("refresh", refresh),
			zap.Error(err),
		)
		return res, err
	}

	res.Access = access
	res.Refresh = refresh
	return res, nil
}

func (c *Controller) Authenticate(ctx context.Context, d *dto.DeviceRequest, req *dto.EmailAndPasswordRequest) (*dto.TokenPair, error) {
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

	if err = c.au.ComparePasswords([]byte(res.Password), []byte(req.Password)); err != nil {
		zap.L().Debug(
			"failed to compare password",
			zap.String("op", op),
			zap.Any("req", req),
			zap.Error(err),
		)
		return nil, auth.ErrInvalidCredentials
	}

	pair, err := c.GenPair(ctx, d, res.ID, res.Permissions)
	if err != nil {
		return nil, err
	}

	return &dto.TokenPair{
		Access:  pair.Access,
		Refresh: pair.Refresh,
	}, nil
}

func (c *Controller) Refresh(ctx context.Context, d *dto.DeviceRequest, req *dto.RefreshRequest) (*dto.TokenPair, error) {
	const op = "auth.Refresh.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	claims, err := c.au.ParseClaims(ctx, req.Refresh)
	if err != nil {
		return nil, err
	}

	device := auth.GenerateDevice(d)
	isValid, err := c.repo.IsTokenValid(ctx, claims.UID, &device, req.Refresh)
	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, auth.ErrTokenRevoked
	}

	access, refresh, err := c.au.GenPair(ctx, claims.UID, claims.Roles)
	if err != nil {
		return nil, err
	}

	hash, err := c.au.HashSHA256(refresh)
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

	err = c.repo.CreateToken(ctx, claims.UID, hash, time.Now().Add(auth.RefreshTokenDuration), &device)
	if err != nil {
		zap.L().Debug(
			"Failed to save token",
			zap.String("op", op),
			zap.String("refresh", refresh),
			zap.Any("claims", claims),
			zap.Error(err),
		)
		return nil, err
	}

	return &dto.TokenPair{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (c *Controller) ParseClaims(ctx context.Context, token string) (res auth.Claims, err error) {
	const op = "auth.ParseClaims.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err = c.au.ParseClaims(ctx, token)
	if err != nil {
		zap.L().Debug("invalid token", zap.Error(err))
		return res, err
	}

	return res, nil
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

	newPass, err := c.au.Hash(req.Password)
	if err != nil {
		return err
	}

	if err = c.repo.UpdateUser(
		ctx, req.ID, &dto.UpdateUserRequest{
			Name:     u.Name,
			Email:    u.Email,
			Password: newPass,
			Avatar:   u.Avatar,
		},
	); err != nil {
		zap.L().Debug(
			"Error updating user",
			zap.Error(err), zap.String("op", op),
		)
		return err
	}

	if err = c.repo.RevokeAllTokens(ctx, u.ID); err != nil {
		zap.L().Debug(
			"Failed to revoke tokens",
			zap.String("op", op),
			zap.Any("uid", u.ID),
			zap.Error(err),
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

	if err = c.au.ComparePasswords([]byte(res.Password), []byte(password)); err != nil {
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

func (c *Controller) CheckLoginCode(ctx context.Context, d *dto.DeviceRequest, req *dto.CheckLoginCodeRequest) (*dto.TokenPair, error) {
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

	access, refresh, err := c.au.GenPair(ctx, res.ID, res.Permissions)
	if err != nil {
		return nil, ErrWhileGeneratingToken
	}

	hash, err := c.au.Hash(refresh)
	if err != nil {
		return nil, err
	}

	device := auth.GenerateDevice(d)
	if err = c.repo.CreateToken(ctx, res.ID, hash, time.Now().Add(auth.RefreshTokenDuration), &device); err != nil {
		zap.L().Debug(
			"Failed to save token",
			zap.String("op", op),
			zap.String("refresh", refresh),
			zap.Error(err),
		)
		return nil, err
	}

	return &dto.TokenPair{
		Access:  access,
		Refresh: refresh,
	}, nil
}
