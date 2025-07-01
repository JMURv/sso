package providers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JMURv/sso/internal/auth/providers/oauth2/google"
	g_oidc "github.com/JMURv/sso/internal/auth/providers/oidc/google"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Port interface {
	Get(provider Providers, flow Flow) (OAuthProvider, error)
	SuccessURL() string
	GenerateSignedState() string
	ValidateSignedState(signedState string, maxAge time.Duration) error
}

type (
	Flow      string
	Providers string
)

const (
	OIDC   Flow = "oidc"
	OAuth2 Flow = "oauth2"
)

const (
	Google Providers = "google"
)

type OAuthProvider interface {
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (*dto.ProviderResponse, error)
}

type Core struct {
	secret          []byte
	successURL      string
	OAuth2Providers OAuthProviders
	OIDCProviders   OIDCProviders
}

type OAuthProviders struct {
	Google OAuthProvider
}

type OIDCProviders struct {
	Google OAuthProvider
}

func New(conf config.Config) *Core {
	return &Core{
		secret:     []byte(conf.Auth.Providers.Secret),
		successURL: conf.Auth.Providers.SuccessURL,
		OAuth2Providers: OAuthProviders{
			Google: google.New(conf),
		},
		OIDCProviders: OIDCProviders{
			Google: g_oidc.New(conf),
		},
	}
}

func (c *Core) Get(provider Providers, flow Flow) (OAuthProvider, error) {
	switch provider {
	case Google:
		if flow == OIDC {
			return c.OIDCProviders.Google, nil
		}
		return c.OAuth2Providers.Google, nil
	default:
		zap.L().Error(
			"Unknown provider",
			zap.Any("provider", provider),
			zap.Any("flow", flow),
		)
		return nil, ErrUnknownProvider
	}
}

func (c *Core) SuccessURL() string {
	return c.successURL
}

func (c *Core) GenerateSignedState() string {
	rawState := uuid.New().String()
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	data := rawState + "|" + timestamp

	h := hmac.New(sha256.New, c.secret)
	h.Write([]byte(data))
	signature := h.Sum(nil)

	signedState := base64.URLEncoding.EncodeToString([]byte(data)) + "." + base64.URLEncoding.EncodeToString(signature)
	return signedState
}

func (c *Core) ValidateSignedState(signedState string, maxAge time.Duration) error {
	parts := strings.Split(signedState, ".")
	if len(parts) != 2 {
		zap.L().Error(
			"Invalid state format",
			zap.String("state", signedState),
			zap.Duration("maxAge", maxAge),
		)
		return ErrInvalidStateFormat
	}

	data, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		zap.L().Error(
			"Failed to decode string",
			zap.Any("parts", parts),
		)
		return err
	}

	expectedSignature, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		zap.L().Error(
			"Failed to decode string",
			zap.Any("parts", parts),
		)
		return err
	}

	h := hmac.New(sha256.New, c.secret)
	h.Write(data)
	actualSignature := h.Sum(nil)

	if !hmac.Equal(expectedSignature, actualSignature) {
		zap.L().Debug(
			"Invalid signature",
			zap.ByteString("exp", expectedSignature),
			zap.ByteString("actual", actualSignature),
		)
		return ErrInvalidSignature
	}

	dataParts := strings.Split(string(data), "|")
	if len(dataParts) != 2 {
		zap.L().Error(
			"Invalid data format",
			zap.Any("data", dataParts),
		)
		return ErrInvalidDataFormat
	}

	timestamp, err := strconv.ParseInt(dataParts[1], 10, 64)
	if err != nil {
		zap.L().Error(
			"failed to parse timestamp",
			zap.Any("data", dataParts[1]),
		)
		return err
	}

	if time.Since(time.Unix(timestamp, 0)) > maxAge {
		zap.L().Debug(
			"state has been expired",
			zap.Int64("tm", timestamp),
			zap.Duration("age", maxAge),
		)
		return ErrStateExpired
	}

	return nil
}
