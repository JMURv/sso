package ctrl

import (
	"context"
	"errors"
	"fmt"
	wa "github.com/JMURv/sso/internal/auth/webauthn"
	"github.com/JMURv/sso/internal/cache"
	"github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

const (
	webauthnSessionKey = "webauthn:%s:%s" // format: type:userID
)

func (c *Controller) GetUserForWA(ctx context.Context, uid uuid.UUID) (*md.WebauthnUser, error) {
	const op = "webauthn.GetUserForWA.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	user, err := c.repo.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			zap.L().Debug("User not found", zap.String("op", op), zap.Any("uid", uid))
			return nil, ErrNotFound
		}
		zap.L().Error("Failed to get user", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	credentials, err := c.repo.GetWACredentials(ctx, uid)
	if err != nil {
		zap.L().Error("Failed to get credentials", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	return &md.WebauthnUser{
		ID:          user.ID,
		Email:       user.Email,
		Credentials: credentials,
	}, nil
}

func (c *Controller) StoreWASession(ctx context.Context, sessionType wa.SessionType, userID uuid.UUID, req *webauthn.SessionData) error {
	const op = "webauthn.StoreWASession.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	c.cache.Set(ctx, config.MinCacheTime, fmt.Sprintf(webauthnSessionKey, sessionType, userID), req)
	return nil
}

func (c *Controller) GetWASession(ctx context.Context, sessionType wa.SessionType, userID uuid.UUID) (*webauthn.SessionData, error) {
	const op = "webauthn.GetWASession.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var session webauthn.SessionData
	err := c.cache.GetToStruct(ctx, fmt.Sprintf(webauthnSessionKey, sessionType, userID), &session)
	if err != nil {
		if errors.Is(err, cache.ErrNotFoundInCache) {
			return nil, ErrNotFound
		}
		zap.L().Error("Failed to get session", zap.String("op", op), zap.Error(err))
		return nil, err
	}
	return &session, nil
}

func (c *Controller) StoreWACredential(ctx context.Context, userID uuid.UUID, credential *webauthn.Credential) error {
	const op = "webauthn.StoreWACredential.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	if err := c.repo.CreateWACredential(
		ctx, userID, &md.WebauthnCredential{
			ID:              credential.ID,
			PublicKey:       credential.PublicKey,
			AttestationType: credential.AttestationType,
			Authenticator:   credential.Authenticator,
			UserID:          userID,
		},
	); err != nil {
		zap.L().Error("Failed to store credential", zap.String("op", op), zap.Error(err))
		return err
	}
	return nil
}

func (c *Controller) GetUserByEmailForWA(ctx context.Context, email string) (*md.WebauthnUser, error) {
	const op = "webauthn.GetUserByEmailForWA.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	user, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			zap.L().Debug(
				repo.ErrNotFound.Error(),
				zap.String("op", op),
				zap.String("email", email),
			)
			return nil, ErrNotFound
		}
		zap.L().Error(
			"Failed to get user",
			zap.String("op", op),
			zap.String("email", email),
			zap.Error(err),
		)
		return nil, err
	}

	credentials, err := c.repo.GetWACredentials(ctx, user.ID)
	if err != nil {
		zap.L().Error(
			"Failed to get credentials",
			zap.String("op", op),
			zap.Error(err),
		)
		return nil, err
	}

	return &md.WebauthnUser{
		ID:          user.ID,
		Email:       user.Email,
		Credentials: credentials,
	}, nil
}
