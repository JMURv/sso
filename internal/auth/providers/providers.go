package providers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	g_oauth2 "github.com/JMURv/sso/internal/auth/providers/oauth2/google"
	g_oidc "github.com/JMURv/sso/internal/auth/providers/oidc/google"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
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
}

type Provider struct {
	secret          []byte
	OAuth2Providers struct {
		Google OAuth2Provider
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
		}{
			Google: g_oauth2.NewGoogleOAuth2(conf),
		},
		OIDCProviders: struct {
			Google OAuth2Provider
		}{
			Google: g_oidc.NewGoogleOIDC(conf),
		},
	}
}

func (p *Provider) Get(ctx context.Context, provider Providers, flow Flow) (OAuth2Provider, error) {
	const op = "providers.Get.auth"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	switch provider {
	case Google:
		if flow == OIDC {
			return p.OIDCProviders.Google, nil
		}
		return p.OAuth2Providers.Google, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func (p *Provider) GenerateSignedState() (string, error) {
	rawState := uuid.New().String()
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	data := rawState + "|" + timestamp

	h := hmac.New(sha256.New, p.secret)
	h.Write([]byte(data))
	signature := h.Sum(nil)

	signedState := base64.URLEncoding.EncodeToString([]byte(data)) + "." + base64.URLEncoding.EncodeToString(signature)
	return signedState, nil
}

func (p *Provider) ValidateSignedState(signedState string, maxAge time.Duration) (bool, error) {
	parts := strings.Split(signedState, ".")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid state format")
	}

	data, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}

	expectedSignature, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	h := hmac.New(sha256.New, p.secret)
	h.Write(data)
	actualSignature := h.Sum(nil)

	if !hmac.Equal(expectedSignature, actualSignature) {
		return false, fmt.Errorf("invalid signature")
	}

	dataParts := strings.Split(string(data), "|")
	if len(dataParts) != 2 {
		return false, fmt.Errorf("invalid data format")
	}

	timestamp, err := strconv.ParseInt(dataParts[1], 10, 64)
	if err != nil {
		return false, err
	}

	if time.Since(time.Unix(timestamp, 0)) > maxAge {
		return false, fmt.Errorf("state expired")
	}

	return true, nil
}
