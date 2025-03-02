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

type authRepo interface {
	CreateToken(
		ctx context.Context,
		userID uuid.UUID,
		hashedT string,
		expiresAt time.Time,
		device *md.Device,
	) error
	IsTokenValid(ctx context.Context, userID uuid.UUID, d *md.Device, token string) (bool, error)
	RevokeAllTokens(ctx context.Context, userID uuid.UUID) error

	GetUserByOAuth2(ctx context.Context, provider, providerID string) (*md.User, error)
	CreateOAuth2Connection(ctx context.Context, userID uuid.UUID, provider string, data *dto.ProviderResponse) error
}

const codeCacheKey = "code:%v"
const recoveryCacheKey = "recovery:%v"

func (c *Controller) Authenticate(ctx context.Context, d *dto.DeviceRequest, req *dto.EmailAndPasswordRequest) (*dto.EmailAndPasswordResponse, error) {
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

	if err = auth.Au.ComparePasswords([]byte(res.Password), []byte(req.Password)); err != nil {
		zap.L().Debug(
			"failed to compare password",
			zap.String("op", op),
			zap.Any("req", req),
			zap.Error(err),
		)
		return nil, auth.ErrInvalidCredentials
	}

	access, refresh, err := auth.Au.GenPair(ctx, res.ID, res.Permissions)
	if err != nil {
		return nil, err
	}

	hash, err := auth.Au.HashSHA256(refresh)
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

	return &dto.EmailAndPasswordResponse{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (c *Controller) Refresh(ctx context.Context, d *dto.DeviceRequest, req *dto.RefreshRequest) (*dto.RefreshResponse, error) {
	const op = "auth.Refresh.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	claims, err := auth.Au.ParseClaims(ctx, req.Refresh)
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

	access, refresh, err := auth.Au.GenPair(ctx, claims.UID, claims.Roles)
	if err != nil {
		return nil, err
	}

	hash, err := auth.Au.HashSHA256(refresh)
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

func (c *Controller) GetOAuth2AuthURL(ctx context.Context, provider string) (*dto.StartOAuth2Response, error) {
	const op = "auth.GetOAuth2AuthURL.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	pr, err := auth.Au.GetOAuth2Provider(ctx, provider)
	if err != nil {
		return nil, err
	}

	signedState, err := auth.Au.GenerateSignedState()
	if err != nil {
		return nil, err
	}

	return &dto.StartOAuth2Response{
		URL: pr.GetConfig().AuthCodeURL(signedState),
	}, nil
}

func (c *Controller) HandleOAuth2Callback(ctx context.Context, d *dto.DeviceRequest, provider, code, state string) (*dto.OAuth2CallbackResponse, error) {
	const op = "auth.HandleOAuth2Callback.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	isValid, err := auth.Au.ValidateSignedState(state, 5*time.Minute)
	if !isValid || err != nil {
		return nil, errors.New("invalid oauth state")
	}

	pr, err := auth.Au.GetOAuth2Provider(ctx, provider)
	if err != nil {
		return nil, err
	}

	oauthUser, err := pr.GetUser(ctx, code)
	if err != nil {
		return nil, err
	}

	user, err := c.repo.GetUserByOAuth2(ctx, provider, oauthUser.ProviderID)
	if errors.Is(err, repo.ErrNotFound) {
		user, err = c.repo.GetUserByEmail(ctx, oauthUser.Email)
		if errors.Is(err, repo.ErrNotFound) {
			hash, err := auth.Au.Hash(uuid.NewString())
			if err != nil {
				return nil, err
			}

			user = &md.User{
				Name:     oauthUser.Name,
				Email:    oauthUser.Email,
				Password: hash,
				Avatar:   oauthUser.Picture,
			}

			id, err := c.repo.CreateUser(ctx, user)
			if err != nil {
				zap.L().Debug(
					"failed to create user",
					zap.String("op", op),
					zap.Any("user", user),
					zap.Error(err),
				)
				return nil, err
			}
			user.ID = id

			if err = c.repo.CreateOAuth2Connection(ctx, user.ID, provider, oauthUser); err != nil {
				zap.L().Debug(
					"failed to create oauth2 connection",
					zap.String("op", op),
					zap.Any("oauthUser", oauthUser),
					zap.Error(err),
				)
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else if err == nil {
			// TODO: Привязать OAuth2 к существующему пользователю
		}
	}

	access, refresh, err := auth.Au.GenPair(ctx, user.ID, user.Permissions)
	hash, err := auth.Au.HashSHA256(refresh)
	if err != nil {
		return nil, err
	}

	device := auth.GenerateDevice(d)
	if err = c.repo.CreateToken(ctx, user.ID, hash, time.Now().Add(auth.RefreshTokenDuration), &device); err != nil {
		zap.L().Debug(
			"Failed to save token",
			zap.String("op", op),
			zap.String("refresh", refresh),
			zap.Error(err),
		)
		return nil, err
	}

	return &dto.OAuth2CallbackResponse{
		Access:  access,
		Refresh: refresh,
	}, err
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

func (c *Controller) CheckLoginCode(ctx context.Context, d *dto.DeviceRequest, req *dto.CheckLoginCodeRequest) (*dto.CheckLoginCodeResponse, error) {
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

	hash, err := auth.Au.Hash(refresh)
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

	return &dto.CheckLoginCodeResponse{
		Access:  access,
		Refresh: refresh,
	}, nil
}
