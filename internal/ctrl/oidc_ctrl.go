package ctrl

import (
	"context"
	"errors"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/auth/providers"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"time"
)

func (c *Controller) GetOIDCAuthURL(ctx context.Context, provider string) (*dto.StartProviderResponse, error) {
	const op = "auth.GetOIDCAuthURL.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	pr, err := auth.Au.Provider.Get(ctx, providers.Providers(provider), providers.OIDC)
	if err != nil {
		return nil, err
	}

	signedState, err := auth.Au.Provider.GenerateSignedState()
	if err != nil {
		return nil, err
	}

	return &dto.StartProviderResponse{
		URL: pr.GetConfig().AuthCodeURL(signedState),
	}, nil
}

func (c *Controller) HandleOIDCCallback(ctx context.Context, d *dto.DeviceRequest, provider, code, state string) (*dto.ProviderCallbackResponse, error) {
	const op = "auth.HandleOIDCCallback.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	isValid, err := auth.Au.Provider.ValidateSignedState(state, 5*time.Minute)
	if !isValid || err != nil {
		return nil, errors.New("invalid oauth state")
	}

	pr, err := auth.Au.Provider.Get(ctx, providers.Providers(provider), providers.OIDC)
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

	return &dto.ProviderCallbackResponse{
		Access:  access,
		Refresh: refresh,
	}, err
}
