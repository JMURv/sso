package ctrl

import (
	"context"
	"errors"
	"time"

	"github.com/JMURv/sso/internal/auth/providers"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type oauth2Repo interface {
	GetUserByOAuth2(ctx context.Context, provider, providerID string) (*md.User, error)
	CreateOAuth2Connection(ctx context.Context, userID uuid.UUID, provider string, data *dto.ProviderResponse) error
}

func (c *Controller) GetOAuth2AuthURL(ctx context.Context, provider string) (*dto.StartProviderResponse, error) {
	const op = "auth.GetOAuth2AuthURL.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)

	defer span.Finish()

	pr, err := c.au.Get(providers.Providers(provider), providers.OAuth2)
	if err != nil {
		return nil, err
	}

	return &dto.StartProviderResponse{
		URL: pr.AuthCodeURL(
			c.au.GenerateSignedState(),
		),
	}, nil
}

func (c *Controller) HandleOAuth2Callback(
	ctx context.Context,
	d *dto.DeviceRequest,
	provider, code, state string,
) (*dto.HandleCallbackResponse, error) {
	const op = "auth.HandleOAuth2Callback.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)

	defer span.Finish()

	err := c.au.ValidateSignedState(state, 5*time.Minute)
	if err != nil {
		return nil, err
	}

	pr, err := c.au.Get(providers.Providers(provider), providers.OAuth2)
	if err != nil {
		return nil, err
	}

	oauthUser, err := pr.Exchange(ctx, code)
	if err != nil {
		zap.L().Error(
			"Failed to get user",
			zap.String("op", op),
			zap.String("code", code),
			zap.Error(err),
		)

		return nil, err
	}

	user, err := c.repo.GetUserByOAuth2(ctx, provider, oauthUser.ProviderID)
	if errors.Is(err, repo.ErrNotFound) {
		user, err = c.repo.GetUserByEmail(ctx, oauthUser.Email)
		if errors.Is(err, repo.ErrNotFound) {
			user = &md.User{
				Name:   oauthUser.Name,
				Email:  oauthUser.Email,
				Avatar: oauthUser.Picture,
			}

			id, err := c.repo.CreateUser(
				ctx, &dto.CreateUserRequest{
					Name:   oauthUser.Name,
					Email:  oauthUser.Email,
					Avatar: oauthUser.Picture,
				},
			)
			if err != nil {
				return nil, err
			}

			user.ID = id
		} else if err != nil {
			return nil, err
		}
	}

	if err = c.repo.CreateOAuth2Connection(ctx, user.ID, provider, oauthUser); err != nil {
		return nil, err
	}

	pair, err := c.GenPair(ctx, d, user.ID, user.Roles)
	if err != nil {
		return nil, err
	}

	return &dto.HandleCallbackResponse{
		Access:     pair.Access,
		Refresh:    pair.Refresh,
		SuccessURL: c.au.SuccessURL(),
	}, nil
}
