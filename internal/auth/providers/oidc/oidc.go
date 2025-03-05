package providers

import (
	"context"
	"fmt"
	"github.com/JMURv/sso/internal/dto"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	config     *oauth2.Config
	provider   *oidc.Provider
	oidcConfig *oidc.Config
}

func NewOIDCProvider(issuer, clientID, clientSecret, redirectURL string, scopes []string) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %v", err)
	}

	return &OIDCProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     provider.Endpoint(),
		},
		provider: provider,
		oidcConfig: &oidc.Config{
			ClientID: clientID,
		},
	}, nil
}

func (p *OIDCProvider) GetConfig() *oauth2.Config {
	return p.config
}

func (p *OIDCProvider) GetUser(ctx context.Context, code string) (*dto.ProviderResponse, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %v", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in token response")
	}

	idToken, err := p.provider.Verifier(p.oidcConfig).Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %v", err)
	}

	var claims struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Subject string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %v", err)
	}

	return &dto.ProviderResponse{
		ProviderID:  claims.Subject,
		Email:       claims.Email,
		Name:        claims.Name,
		AccessToken: token.AccessToken,
		IDToken:     rawIDToken,
	}, nil
}
