package providers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	gh_oauth2 "github.com/JMURv/sso/internal/auth/providers/oauth2/github"
	g_oauth2 "github.com/JMURv/sso/internal/auth/providers/oauth2/google"
	g_oidc "github.com/JMURv/sso/internal/auth/providers/oidc/google"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"strconv"
	"strings"
	"time"
)

type Flow string

const (
	OIDC   Flow = "oidc"
	OAuth2 Flow = "oauth2"
)

type Providers string

const (
	Google Providers = "google"
	GitHub Providers = "github"
)

type OAuth2Provider interface {
	GetConfig() *oauth2.Config
	GetUser(ctx context.Context, code string) (*dto.ProviderResponse, error)
	GetSuccessURL() string
}

type Provider struct {
	secret          []byte
	OAuth2Providers struct {
		Google OAuth2Provider
		GitHub OAuth2Provider
	}
	OIDCProviders struct {
		Google OAuth2Provider
	}
}

func New(conf config.Config) *Provider {
	return &Provider{
		secret: []byte(conf.Auth.ProviderSignSecret),
		OAuth2Providers: struct {
			Google OAuth2Provider
			GitHub OAuth2Provider
		}{
			Google: g_oauth2.NewGoogleOAuth2(conf),
			GitHub: gh_oauth2.NewGitHubOAuth2(conf),
		},
		OIDCProviders: struct {
			Google OAuth2Provider
		}{
			Google: g_oidc.NewGoogleOIDC(conf),
		},
	}
}

func (p *Provider) Get(_ context.Context, provider Providers, flow Flow) (OAuth2Provider, error) {
	switch provider {
	case Google:
		if flow == OIDC {
			return p.OIDCProviders.Google, nil
		}
		return p.OAuth2Providers.Google, nil
	case GitHub:
		return p.OAuth2Providers.GitHub, nil
	default:
		zap.L().Error(
			"Unknown provider",
			zap.Any("provider", provider),
		)
		return nil, ErrUnknownProvider
	}
}

func (p *Provider) GenerateSignedState() string {
	rawState := uuid.New().String()
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	data := rawState + "|" + timestamp

	h := hmac.New(sha256.New, p.secret)
	h.Write([]byte(data))
	signature := h.Sum(nil)

	signedState := base64.URLEncoding.EncodeToString([]byte(data)) + "." + base64.URLEncoding.EncodeToString(signature)
	return signedState
}

func (p *Provider) ValidateSignedState(signedState string, maxAge time.Duration) error {
	parts := strings.Split(signedState, ".")
	if len(parts) != 2 {
		zap.L().Debug(
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

	h := hmac.New(sha256.New, p.secret)
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
		zap.L().Debug(
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
