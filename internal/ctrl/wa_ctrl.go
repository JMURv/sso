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
	"strings"
)

type waRepo interface {
	GetWACredentials(ctx context.Context, userID uuid.UUID) ([]webauthn.Credential, error)
	CreateWACredential(ctx context.Context, userID uuid.UUID, cred *webauthn.Credential) error
	UpdateWACredential(ctx context.Context, cred *webauthn.Credential) error
}

const (
	webauthnSessionKey = "webauthn:%s:%s"
)

func (c *Controller) StartRegistration(ctx context.Context, uid uuid.UUID) (*protocol.CredentialCreation, error) {
	const op = "webauthn.StartRegistration.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	user, err := c.GetUserForWA(ctx, uid, "")
	if err != nil {
		return nil, err
	}

	opts, sess, err := c.au.Wa.BeginRegistration(
		user, func(credCreationOpts *protocol.PublicKeyCredentialCreationOptions) {
			credCreationOpts.CredentialExcludeList = user.ExcludeCredentialDescriptorList()
		},
	)
	if err != nil {
		zap.L().Error(
			"Failed to begin registration",
			zap.String("op", op),
			zap.String("uid", uid.String()),
			zap.Any("user", user),
			zap.Error(err),
		)
		return nil, err
	}

	if err = c.StoreWASession(ctx, wa.Register, uid, sess); err != nil {
		return nil, err
	}

	return opts, nil
}

func (c *Controller) FinishRegistration(ctx context.Context, uid uuid.UUID, r *http.Request) error {
	const op = "webauthn.FinishRegistration.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	sess, err := c.GetWASession(ctx, wa.Register, uid)
	if err != nil {
		return err
	}

	user, err := c.GetUserForWA(ctx, uid, "")
	if err != nil {
		return err
	}

	credential, err := c.au.Wa.FinishRegistration(user, *sess, r)
	if err != nil {
		zap.L().Error(
			"failed to finish registration",
			zap.String("op", op),
			zap.String("uid", uid.String()),
			zap.Any("user", user),
			zap.Any("sess", sess),
			zap.Error(err),
		)
		return err
	}

	if err = c.repo.CreateWACredential(ctx, uid, credential); err != nil {
		zap.L().Error(
			"failed to create webAuthn credential",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (c *Controller) BeginLogin(ctx context.Context, email string) (*protocol.CredentialAssertion, error) {
	const op = "webauthn.BeginLogin.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	user, err := c.GetUserForWA(ctx, uuid.Nil, email)
	if err != nil {
		return nil, err
	}

	opts, sess, err := c.au.Wa.BeginLogin(user)
	if err != nil {
		if strings.Contains(err.Error(), "Found no credentials for user") {
			return nil, ErrNotFound
		}
		zap.L().Error(
			"failed to begin login",
			zap.String("op", op),
			zap.String("email", email),
			zap.Any("user", user),
			zap.Error(err),
		)
		return nil, err
	}

	if err = c.StoreWASession(ctx, wa.Login, user.ID, sess); err != nil {
		return nil, err
	}
	return opts, nil
}

func (c *Controller) FinishLogin(ctx context.Context, email string, d dto.DeviceRequest, r *http.Request) (dto.TokenPair, error) {
	const op = "webauthn.FinishLogin.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var res dto.TokenPair
	user, err := c.GetUserForWA(ctx, uuid.Nil, email)
	if err != nil {
		return res, err
	}

	sess, err := c.GetWASession(ctx, wa.Login, user.ID)
	if err != nil {
		return res, err
	}

	cred, err := c.au.Wa.FinishLogin(user, *sess, r)
	if err != nil {
		zap.L().Error(
			"Failed to finish login",
			zap.String("op", op),
			zap.String("email", email),
			zap.Any("user", user),
			zap.Any("sess", sess),
			zap.Error(err),
		)
		return res, err
	}

	if cred.Authenticator.CloneWarning {
		zap.L().Error(
			"credential appears to be cloned",
			zap.String("op", op),
			zap.Any("credential", cred),
		)
		return res, errors.New("credential appears to be cloned") // http.StatusForbiddens
	}

	if err = c.repo.UpdateWACredential(ctx, cred); err != nil {
		zap.L().Error(
			"user failed to update credential during finish login",
			zap.Error(err),
			zap.Any("user", user),
		)
		return res, err
	}

	res, err = c.GenPair(ctx, &d, user.ID, user.Permissions)
	if err != nil {
		return res, err
	}
	return res, nil
}

func (c *Controller) StoreWASession(ctx context.Context, sessionType wa.SessionType, userID uuid.UUID, req *webauthn.SessionData) error {
	const op = "webauthn.StoreWASession.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	d, err := json.Marshal(req)
	if err != nil {
		zap.L().Error(
			"failed to marshall req",
			zap.Any("sessType", sessionType),
			zap.String("uid", userID.String()),
			zap.Any("req", req),
		)
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
		zap.L().Error(
			"failed to get session",
			zap.String("op", op),
			zap.Error(err),
		)
		return nil, err
	}
	return &session, nil
}

func (c *Controller) GetUserForWA(ctx context.Context, uid uuid.UUID, email string) (*md.WebauthnUser, error) {
	const op = "webauthn.GetUserForWA.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var err error
	var user *md.User
	if uid != uuid.Nil {
		user, err = c.repo.GetUserByID(ctx, uid)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				zap.L().Debug(
					"user not found",
					zap.String("op", op),
					zap.Any("uid", uid),
				)
				return nil, ErrNotFound
			}
			zap.L().Error(
				"failed to get user",
				zap.String("op", op),
				zap.Any("uid", uid),
				zap.Error(err),
			)
			return nil, err
		}
	} else {
		user, err = c.repo.GetUserByEmail(ctx, email)
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
				"failed to get user",
				zap.String("op", op),
				zap.String("email", email),
				zap.Error(err),
			)
			return nil, err
		}
	}

	credentials, err := c.repo.GetWACredentials(ctx, user.ID)
	if err != nil {
		zap.L().Error(
			"failed to get webAuthn credentials",
			zap.String("op", op),
			zap.Error(err),
		)
		return nil, err
	}
	return &md.WebauthnUser{
		ID:          user.ID,
		Email:       user.Email,
		Permissions: user.Permissions,
		Credentials: credentials,
	}, nil
}
