package google

import (
	"context"
	"fmt"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/coreos/go-oidc/v3/oidc"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type Provider struct {
	config     *oauth2.Config
	provider   *oidc.Provider
	oidcConfig *oidc.Config
}

func New(conf config.Config) *Provider {
	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		return nil
	}

	return &Provider{
		config: &oauth2.Config{
			ClientID:     conf.Auth.Providers.OIDC.Google.ClientID,
			ClientSecret: conf.Auth.Providers.OIDC.Google.ClientSecret,
			RedirectURL:  conf.Auth.Providers.OIDC.Google.RedirectURL,
			Scopes:       conf.Auth.Providers.OIDC.Google.Scopes,
			Endpoint:     provider.Endpoint(),
		},
		provider: provider,
		oidcConfig: &oidc.Config{
			ClientID: conf.Auth.Providers.OIDC.Google.ClientID,
		},
	}
}

func (p *Provider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *Provider) Exchange(ctx context.Context, code string) (*dto.ProviderResponse, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		zap.L().Error("failed to exchange token", zap.Error(err))
		return nil, err
	}

	return p.getUserInfo(ctx, token)
}

func (p *Provider) getUserInfo(ctx context.Context, token *oauth2.Token) (*dto.ProviderResponse, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		zap.L().Error("no id_token in token response")
		return nil, fmt.Errorf("no id_token in token response")
	}

	idToken, err := p.provider.Verifier(p.oidcConfig).Verify(ctx, rawIDToken)
	if err != nil {
		zap.L().Error("failed to verify ID token", zap.Error(err))
		return nil, err
	}

	var claims struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Subject string `json:"sub"`
	}
	if err = idToken.Claims(&claims); err != nil {
		zap.L().Error("failed to parse claims", zap.Error(err))
		return nil, err
	}

	return &dto.ProviderResponse{
		ProviderID:  claims.Subject,
		Email:       claims.Email,
		Name:        claims.Name,
		AccessToken: token.AccessToken,
		IDToken:     rawIDToken,
	}, nil
}
