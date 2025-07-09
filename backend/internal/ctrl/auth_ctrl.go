package ctrl

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/auth/jwt"
	"github.com/JMURv/sso/internal/cache"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

const (
	codeCacheKey     = "code:%v"
	recoveryCacheKey = "recovery:%v"
)

type authRepo interface {
	CreateToken(ctx context.Context, userID uuid.UUID, hashedT string, expiresAt time.Time, device *md.Device) error
	IsTokenValid(ctx context.Context, userID uuid.UUID, d *md.Device, token string) (bool, error)
	RevokeAllTokens(ctx context.Context, userID uuid.UUID) error
}

func (c *Controller) GenPair(ctx context.Context, d *dto.DeviceRequest, uid uuid.UUID, p []md.Role) (dto.TokenPair, error) {
	const op = "auth.GenPair.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var res dto.TokenPair

	access, refresh, err := c.au.GenPair(ctx, uid, p)
	if err != nil {
		return res, err
	}

	device := auth.GenerateDevice(d)
	if err = c.repo.CreateToken(ctx, uid, refresh, c.au.GetRefreshTime(), &device); err != nil {
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
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	if err = c.au.ComparePasswords([]byte(res.Password), []byte(req.Password)); err != nil {
		return nil, auth.ErrInvalidCredentials
	}

	pair, err := c.GenPair(ctx, d, res.ID, res.Roles)
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
		zap.L().Info(
			"token is invalid",
			zap.String("op", op),
			zap.String("userID", claims.UID.String()),
		)

		return nil, auth.ErrTokenRevoked
	}

	access, refresh, err := c.au.GenPair(ctx, claims.UID, claims.Roles)
	if err != nil {
		return nil, err
	}

	if err = c.repo.RevokeAllTokens(ctx, claims.UID); err != nil {
		return nil, err
	}

	err = c.repo.CreateToken(ctx, claims.UID, refresh, c.au.GetRefreshTime(), &device)
	if err != nil {
		return nil, err
	}

	return &dto.TokenPair{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (c *Controller) ParseClaims(ctx context.Context, token string) (res jwt.Claims, err error) {
	const op = "auth.ParseClaims.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err = c.au.ParseClaims(ctx, token)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (c *Controller) Logout(ctx context.Context, uid uuid.UUID) error {
	const op = "auth.Logout.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	if err := c.repo.RevokeAllTokens(ctx, uid); err != nil {
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
		return err
	}

	if storedCode != req.Code {
		zap.L().Debug(
			"codes don't match",
			zap.String("op", op),
			zap.Int("stored code", storedCode),
			zap.Int("request code", req.Code),
		)

		return ErrCodeIsNotValid
	}

	u, err := c.repo.GetUserByID(ctx, req.ID)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	newPass, err := c.au.Hash(req.Password)
	if err != nil {
		return err
	}

	if err = c.repo.UpdateMe(
		ctx, req.ID, &dto.UpdateUserRequest{
			Name:     u.Name,
			Email:    u.Email,
			Password: newPass,
			Avatar:   u.Avatar,
		},
	); err != nil {
		return err
	}

	if err = c.repo.RevokeAllTokens(ctx, u.ID); err != nil {
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
		return auth.ErrInvalidCredentials
	} else if err != nil {
		return err
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	code := rand.Intn(9999-1000+1) + 1000

	go c.smtp.SendForgotPasswordEmail(ctx, strconv.Itoa(code), res.ID.String(), email)

	c.cache.Set(ctx, config.MinCacheTime, fmt.Sprintf(recoveryCacheKey, res.ID.String()), code)

	return nil
}

func (c *Controller) SendLoginCode(ctx context.Context, d *dto.DeviceRequest, email, password string) (dto.TokenPair, error) {
	const op = "auth.SendLoginCode.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var tokens dto.TokenPair
	res, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		return tokens, auth.ErrInvalidCredentials
	} else if err != nil {
		return tokens, err
	}

	if err = c.au.ComparePasswords([]byte(res.Password), []byte(password)); err != nil {
		return tokens, err
	}

	device := auth.GenerateDevice(d)
	devs, err := c.repo.ListDevices(ctx, res.ID)
	for i := 0; i < len(devs); i++ {
		if devs[i].ID == device.ID {
			access, refresh, err := c.au.GenPair(ctx, res.ID, res.Roles)
			if err != nil {
				return tokens, ErrWhileGeneratingToken
			}

			err = c.repo.CreateToken(ctx, res.ID, refresh, c.au.GetRefreshTime(), &device)
			if err != nil {
				return tokens, err
			}

			tokens.Access = access
			tokens.Refresh = refresh
			return tokens, nil
		}
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	code := rand.Intn(9999-1000+1) + 1000

	go c.smtp.SendLoginEmail(ctx, code, email)
	c.cache.Set(ctx, config.MinCacheTime, fmt.Sprintf(codeCacheKey, email), code)
	return tokens, nil
}

func (c *Controller) CheckLoginCode(ctx context.Context, d *dto.DeviceRequest, req *dto.CheckLoginCodeRequest) (*dto.TokenPair, error) {
	const op = "auth.CheckLoginCode.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	storedCode, err := c.cache.GetInt(ctx, fmt.Sprintf(codeCacheKey, req.Email))
	if err != nil && errors.Is(err, cache.ErrNotFoundInCache) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	if storedCode != req.Code {
		return nil, ErrCodeIsNotValid
	}

	res, err := c.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	access, refresh, err := c.au.GenPair(ctx, res.ID, res.Roles)
	if err != nil {
		return nil, ErrWhileGeneratingToken
	}

	device := auth.GenerateDevice(d)
	if err = c.repo.CreateToken(ctx, res.ID, refresh, c.au.GetRefreshTime(), &device); err != nil {
		return nil, err
	}

	return &dto.TokenPair{
		Access:  access,
		Refresh: refresh,
	}, nil
}
