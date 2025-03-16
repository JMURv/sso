package ctrl

import (
	"context"
	"errors"
	"fmt"
	wa "github.com/JMURv/sso/internal/auth/webauthn"
	"github.com/JMURv/sso/internal/cache"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"net/http"
)

const (
	webauthnSessionKey = "webauthn:%s:%s" // format: type:userID
)

func (c *Controller) StartRegistration(ctx context.Context, uid uuid.UUID) (*protocol.CredentialCreation, error) {
	const op = "webauthn.StartRegistration.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	user, err := c.GetUserForWA(ctx, uid)
	if err != nil {
		return nil, err
	}

	res, sessionData, err := c.au.Wa.BeginRegistration(user)
	if err != nil {
		return nil, err
	}

	if err = c.StoreWASession(ctx, wa.Register, uid, sessionData); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Controller) FinishRegistration(ctx context.Context, uid uuid.UUID, r *http.Request) error {
	const op = "webauthn.FinishRegistration.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	user, err := c.GetUserForWA(ctx, uid)
	if err != nil {
		zap.L().Debug("Failed to get user", zap.String("op", op), zap.Error(err))
		return err
	}

	sessionData, err := c.GetWASession(ctx, wa.Register, uid)
	if err != nil {
		zap.L().Debug("Failed to get WASession", zap.String("op", op), zap.Error(err))
		return err
	}

	credential, err := c.au.Wa.FinishRegistration(user, *sessionData, r)
	if err != nil {
		zap.L().Debug("Failed to finish registration", zap.String("op", op), zap.Error(err))
		return err
	}

	if err = c.StoreWACredential(ctx, uid, credential); err != nil {
		zap.L().Debug("Failed to store credential", zap.String("op", op), zap.Error(err))
		return err
	}

	return nil
}

func (c *Controller) BeginLogin(ctx context.Context, email string) (*protocol.CredentialAssertion, error) {
	const op = "webauthn.BeginLogin.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	user, err := c.GetUserByEmailForWA(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	options, sessionData, err := c.au.Wa.BeginLogin(user)
	if err != nil {
		zap.L().Error("Failed to begin login", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	err = c.StoreWASession(ctx, wa.Login, user.ID, sessionData)
	if err != nil {
		zap.L().Error("Failed to StoreWASession", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	return options, nil
}

func (c *Controller) FinishLogin(ctx context.Context, email string, d dto.DeviceRequest, r *http.Request) (dto.TokenPair, error) {
	const op = "webauthn.FinishLogin.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var res dto.TokenPair
	user, err := c.GetUserByEmailForWA(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return res, ErrNotFound
		}
		return res, ErrNotFound
	}

	sess, err := c.GetWASession(ctx, wa.Login, user.ID)
	if err != nil {
		return res, err
	}

	_, err = c.au.Wa.FinishLogin(user, *sess, r)
	if err != nil {
		return res, err
	}

	res, err = c.GenPair(ctx, &d, user.ID, []md.Permission{})
	if err != nil {
		return res, err
	}

	return res, nil
}

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

	d, err := json.Marshal(req)
	if err != nil {
		return err
	}

	c.cache.Set(ctx, config.MinCacheTime, fmt.Sprintf(webauthnSessionKey, sessionType, userID), d)
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
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
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
